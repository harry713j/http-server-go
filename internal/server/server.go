package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/harry713j/http-server/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool // to prevent race condition
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)

	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", port, err)
	}

	srv := &Server{listener: listener}

	srv.listen()

	return srv, nil
}

func (s *Server) Close() error {
	if s.closed.Swap(true) {
		return nil
	}

	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()

		if err != nil {
			if s.closed.Load() {
				return // if the server closed
			}

			log.Printf("Connection error: %v\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	status := response.StatusOk
	responseBody := []byte("Hello World!")

	if err := response.WriteStatusLine(conn, status); err != nil {
		log.Printf("Error writing response line: %v\n", err)
		return
	}

	headers := response.GetDefaultHeaders(len(responseBody))
	if err := response.WriteHeaders(conn, headers); err != nil {
		log.Printf("Error writing headers: %v\n", err)
		return
	}

	if len(responseBody) > 0 {
		if _, err := conn.Write(responseBody); err != nil {
			log.Printf("Error writing response body: %v\n", err)
			return
		}
	}
}
