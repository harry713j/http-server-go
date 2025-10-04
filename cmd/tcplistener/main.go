package main

import (
	"fmt"
	"log"
	"net"

	"github.com/harry713j/http-server/internal/request"
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

		request, err := request.RequestFromReader(conn)

		if err != nil {
			fmt.Println("Error parsing request: ", err)
			conn.Close()
			continue
		}

		fmt.Printf("Request Line:\n- Method: %v\n- Target: %v\n- Version: %v\n",
			request.RequestLine.Method, request.RequestLine.RequestTarget, request.RequestLine.HttpVersion)

		fmt.Printf("Connection has been closed with %v\n", conn.RemoteAddr().String())
		response := "HTTP/1.1 200 OK\r\nContent-Length: 13\r\nContent-Type: text/plain\r\n\r\nOK"
		conn.Write([]byte(response))
		conn.Close()
	}

}
