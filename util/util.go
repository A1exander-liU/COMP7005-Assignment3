package util

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const CHUNK_SIZE int = 1024

func FormatResponse(wordCount int, charCount int, freqs map[rune]int) string {
	message := fmt.Sprintf("Word Count: %d | Char Count: %d\n", wordCount, charCount)

	message += "Character Frequencies\n"

	for key, value := range freqs {
		message += fmt.Sprintf("%c: %d/%d\n", key, value, charCount)
	}

	return message
}

func CharacterFrequencies(text string) map[rune]int {
	regex := regexp.MustCompile(`\s+`)
	cleaned := regex.ReplaceAllString(text, "")
	cleaned = strings.ToLower(cleaned)

	freqs := make(map[rune]int)
	for _, value := range cleaned {
		freqs[value]++
	}
	return freqs
}

func CharacterCount(text string) int {
	regex := regexp.MustCompile(`\s+`)
	cleaned := regex.ReplaceAllString(text, "")
	return len(strings.TrimSpace(cleaned))
}

func WordCount(text string) int {
	regex := regexp.MustCompile(`\s+`)
	cleaned := regex.ReplaceAllString(text, " ")
	cleaned = strings.TrimSpace(cleaned)

	if len(cleaned) == 0 {
		return 0
	}

	return len(strings.Split(cleaned, " "))
}

func EmbedSequenceNumber(data string, sequenceNumber int) string {
	numStr := fmt.Sprintf("%08d", sequenceNumber)
	return numStr + data
}

func ExtractSequenceNumber(data []byte) int {
	sequenceNumber, _ := strconv.Atoi(string(data[:8]))
	return sequenceNumber
}

func CombineData(data []string) string {
	final := ""
	for _, dataChunk := range data {
		final += dataChunk[8:]
	}
	return final
}

func SortData(data []string) []string {
	sort.Slice(data, func(i, j int) bool {
		num1, _ := strconv.Atoi(data[i][:8])
		num2, _ := strconv.Atoi(data[j][:8])
		return num1 < num2
	})
	return data
}

func CreateChunks(originalData string, chunkSize int) []string {
	var chunks = []string{}

	for i := 0; i < len(originalData); i += chunkSize {
		chunkEnd := chunkSize + i
		if chunkEnd > len(originalData) {
			chunkEnd = len(originalData)
		}

		chunk := originalData[i:chunkEnd]

		chunks = append(chunks, chunk)
	}

	return chunks
}

func PrintPacketData(data []string) {
	for _, dataPiece := range data {
		fmt.Println("Sequence Number: " + dataPiece[:8])
	}
}
