package main

import (
	"7005-A3/util"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func sendSizeUdp(socket *net.UDPConn, to *net.UDPAddr, size int) error {
	sizeBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBuffer, uint64(size))

	_, err := socket.WriteToUDP(sizeBuffer, to)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

func receiveSizeUdp(conn *net.UDPConn) (int, *net.UDPAddr) {
	buf := make([]byte, 8) // 8 bytes for an int64
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Println(err)
		return 0, nil
	}

	return int(binary.BigEndian.Uint64(buf)), addr
}

func sendDataUdp(socket *net.UDPConn, to *net.UDPAddr, data string, chunkSize int) error {
	for i := 0; i < len(data); i += chunkSize {
		chunkEnd := chunkSize + i
		if chunkEnd > len(data) {
			chunkEnd = len(data)
		}

		chunk := data[i:chunkEnd]

		_, err := socket.WriteToUDP([]byte(chunk), to)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func receiveDataUdp(conn *net.UDPConn, size int) (string, error) {
	read := 0
	buffer := make([]byte, util.CHUNK_SIZE+8)
	dataSlices := []string{}

	for {
		nBytes, _, err := conn.ReadFromUDP(buffer)
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

	sorted := util.SortData(dataSlices)
	util.PrintPacketData(sorted)

	return util.CombineData(sorted), nil
}

func receive(conn *net.UDPConn) (string, *net.UDPAddr) {
	size, addr := receiveSizeUdp(conn)
	if addr == nil {
		return "", nil
	}

	content, err := receiveDataUdp(conn, size)
	if err != nil {
		return "", nil
	}

	return content, addr
}

func send(conn *net.UDPConn, addr *net.UDPAddr, content string) {
	chunks := util.CreateChunks(content, util.CHUNK_SIZE)
	for i := range chunks {
		chunks[i] = util.EmbedSequenceNumber(chunks[i], i)
	}

	data := strings.Join(chunks, "")

	err := sendSizeUdp(conn, addr, len(data))
	if err != nil {
		return
	}

	err = sendDataUdp(conn, addr, data, util.CHUNK_SIZE+8)
	if err != nil {
		return
	}
}

func listen(address string) *net.UDPConn {
	udpAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println("Resolve UDP Error:\n", err)
		os.Exit(1)
	}

	server, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		fmt.Println("Failed to listen:\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server listening on %s\n", address)
	return server
}

func handle(connection *net.UDPConn) {
	content, addr := receive(connection)

	wCount := util.WordCount(content)
	cCount := util.CharacterCount(content)
	freqs := util.CharacterFrequencies(content)

	response := util.FormatResponse(wCount, cCount, freqs)

	fmt.Println(response)

	send(connection, addr, response)
}

func cleanup(conn *net.UDPConn) {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			fmt.Println("Close Error:\n", err)
			os.Exit(1)
		}

		println("Server closed successfully")
	}
	println("Exiting")
	os.Exit(0)
}

func handleSigInt(channel chan os.Signal, exit func(*net.UDPConn), conn *net.UDPConn) {
	for {
		sig := <-channel

		switch sig {
		case os.Interrupt, syscall.SIGINT:
			exit(conn)
		}
	}
}

func main() {
	server := listen("[::]:8081")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)

	go handleSigInt(sigChan, cleanup, server)

	for {
		handle(server)
	}
}
