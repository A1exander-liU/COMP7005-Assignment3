package main

import (
	"7005-A3/util"
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

type Context struct {
	Address string
	Port    string
	Ip      string
	Socket  *net.UDPConn

	File    string
	Content string

	// Store send and receive size of data
	Size int

	Response string

	ExitCode int
}

func sendData(socket *net.UDPConn, data string, size int, chunkSize int) error {
	for i := 0; i < size; i += chunkSize {
		chunkEnd := chunkSize + i
		if chunkEnd > size {
			chunkEnd = size
		}

		chunk := data[i:chunkEnd]
		_, err := socket.Write([]byte(chunk))
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func sendSize(socket *net.UDPConn, size int) error {
	sizeBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBuffer, uint64(size))

	_, err := socket.Write(sizeBuffer)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func receiveData(conn *net.UDPConn, size int) (string, error) {
	read := 0
	buffer := make([]byte, util.CHUNK_SIZE+8)
	dataSlices := []string{}

	for {
		nBytes, err := conn.Read(buffer)
		read += nBytes

		if err != nil {
			fmt.Println("Error:", err)
			return "", err
		}

		dataSlices = append(dataSlices, string(buffer[:nBytes]))

		if read >= size {
			break
		}
	}

	return util.CombineData(util.SortData(dataSlices)), nil
}

func receiveSize(conn *net.UDPConn) (int, error) {
	buf := make([]byte, 8) // 8 bytes for an int64
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return int(binary.BigEndian.Uint64(buf)), nil
}

func receive(context *Context) {
	size, err := receiveSize(context.Socket)
	if err != nil {
		HandleUserInput(context)
	}
	context.Size = size

	response, err1 := receiveData(context.Socket, context.Size)
	if err1 != nil {
		HandleUserInput(context)
	}
	context.Response = response
}

func send(context *Context) {
	chunks := util.CreateChunks(context.Content, util.CHUNK_SIZE)
	for i := range chunks {
		chunks[i] = util.EmbedSequenceNumber(chunks[i], i)
	}

	context.Content = strings.Join(chunks, "")
	context.Size = len(strings.Join(chunks, ""))

	err := sendSize(context.Socket, context.Size)
	if err != nil {
		HandleUserInput(context)
	}

	err = sendData(context.Socket, context.Content, context.Size, util.CHUNK_SIZE+8)
	if err != nil {
		HandleUserInput(context)
	}
}

func exit(context *Context) {
	os.Exit(context.ExitCode)
}

func createSocket(context *Context) {
	addr, err := net.ResolveUDPAddr("udp", context.Address)
	if err != nil {
		fmt.Println("Resolve Udp Addr error:\n", err)
		context.ExitCode = 1
		exit(context)
	}

	client, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Socket Connect Error:\n", err)
		context.ExitCode = 1
		exit(context)
	}
	context.Socket = client
}

func readFile(context *Context) {
	content, err := os.ReadFile(context.File)
	if err != nil {
		fmt.Println("Read File Error:\n", err)
		HandleUserInput(context)
	}

	if len(content) == 0 {
		fmt.Println("File is empty")
		HandleUserInput(context)
	}

	context.Content = string(content)
	context.Size = len(context.Content)
}

func handleSendFile(context *Context) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter path to file: ")
	filePath, _ := reader.ReadString('\n')
	context.File = strings.TrimSpace(filePath)

	readFile(context)

	createSocket(context)
	defer context.Socket.Close()
	defer func(context *Context) { context.Socket = nil }(context)

	send(context)

	receive(context)

	fmt.Println(context.Response)
}

func HandleUserInput(context *Context) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("1. Send File\n2. Quit")
		choice, _ := reader.ReadString('\n')

		switch choice = strings.TrimSpace(choice); choice {
		case "1":
			handleSendFile(context)
		case "2":
			context.ExitCode = 0
			Exit(context)
		default:
			fmt.Println("Please enter 1 or 2")
		}
		fmt.Println()
	}
}

func transformAddress(ip string, port string) string {
	return ip + ":" + port
}

func transformIp(ip string) string {
	index := strings.Index(ip, ":")
	if index >= 0 {
		return fmt.Sprintf("[%s]", ip)
	}
	return ip
}

func Exit(context *Context) {
	if context.Socket != nil {
		err := context.Socket.Close()
		if err != nil {
			println("Close Socket Error:\n", err)
			context.ExitCode = 1
			os.Exit(context.ExitCode)
		}
	}
	os.Exit(context.ExitCode)
}

func parseArgs(args []string, context *Context) {
	if len(args) < 3 {
		fmt.Println("Need to enter address and a port")
		context.ExitCode = 1
		Exit(context)
	} else {
		context.Ip = transformIp(args[1])
		context.Port = args[2]
		context.Address = transformAddress(context.Ip, context.Port)

		HandleUserInput(context)
	}
}

func main() {
	context := Context{}
	parseArgs(os.Args, &context)
}
