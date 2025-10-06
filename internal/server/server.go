package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/harry713j/http-server/internal/request"
	"github.com/harry713j/http-server/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool // to prevent race condition
	handler  Handler
}

type Handler func(w io.Writer, r *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %v", port, err)
	}

	srv := &Server{listener: listener, handler: handler}

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

	// parse request
	req, err := request.RequestFromReader(conn)

	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buff := bytes.NewBuffer([]byte{})

	if hErr := s.handler(buff, req); hErr != nil {
		hErr.Write(conn)
		return
	}

	status := response.StatusOk
	responseBody := buff.Bytes()

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

func (h HandlerError) Write(w io.Writer) {
	errRespBody := []byte(h.Message)
	headers := response.GetDefaultHeaders(len(errRespBody))

	if err := response.WriteStatusLine(w, h.StatusCode); err != nil {
		log.Printf("Error writing response line: %v\n", err)
		return
	}

	if err := response.WriteHeaders(w, headers); err != nil {
		log.Printf("Error writing headers: %v\n", err)
		return
	}

	if _, err := w.Write(errRespBody); err != nil {
		log.Printf("Error writing error body: %v\n", err)
		return
	}

}
