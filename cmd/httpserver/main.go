package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/harry713j/http-server/internal/request"
	"github.com/harry713j/http-server/internal/response"
	"github.com/harry713j/http-server/internal/server"
)

const port = 42069

func main() {
	handler := func(w io.Writer, r *request.Request) *server.HandlerError {
		switch r.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your problem is not my problem\n",
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    "Woopsie, my bad\n",
			}
		default:
			io.WriteString(w, "All good, frfr\n")
			return nil
		}
	}

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
