package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("./message.txt")

	if err != nil {
		fmt.Println("Error reading file: ", err)
	}

	lines := getLinesChannel(file)

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {

	result := make(chan string)

	go func(result chan<- string) {
		defer f.Close()
		size := 8
		data := make([]byte, size)
		currentLine := ""
		for {
			n, err := f.Read(data)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				fmt.Println("Error reading chunks: ", err)
			}

			currentLine += string(data[:n])
			linePart := strings.Split(currentLine, "\n")

			if len(linePart) > 1 {
				for i := 0; i < len(linePart)-1; i++ {
					result <- linePart[i]
				}
			}

			currentLine = linePart[len(linePart)-1]
		}

		if currentLine != "" {
			result <- currentLine
		}

		close(result)
	}(result)

	return result
}
