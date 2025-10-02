package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {

	listener, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Printf("Connection error: %v\n", err)
		}

		fmt.Printf("Connection accepted from %s\n", conn.RemoteAddr().String())

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}

		fmt.Printf("Connection has been closed with %v\n", conn.RemoteAddr().String())
		conn.Close()
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
