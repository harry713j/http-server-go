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
		respWriter := response.NewWriter(w)

		var (
			status response.StatusCode
			body   string
		)

		switch r.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = `
				<html>
					<head>
						<title>400 Bad Request</title>
					</head>
					<body>
						<h1>Bad Request</h1>
						<p>Your request honestly kinda sucked.</p>
					</body>
				</html>
			`
		case "/myproblem":
			status = response.StatusInternalServerError
			body = `
				<html>
					<head>
						<title>500 Internal Server Error</title>
					</head>
					<body>
						<h1>Internal Server Error</h1>
						<p>Okay, you know what? This one is on me.</p>
					</body>
				</html>
			`
		default:
			status = response.StatusOk
			body = `
				<html>
					<head>
						<title>200 OK</title>
					</head>
					<body>
						<h1>Success!</h1>
						<p>Your request was an absolute banger.</p>
					</body>
				</html>
			`
		}
		// Write response
		if err := respWriter.WriteStatusLine(status); err != nil {
			return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
		}

		h := response.GetDefaultHeaders(len(body))
		h.Add("Content-Type", "text/html")

		if err := respWriter.WriteHeaders(h); err != nil {
			return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
		}

		if _, err := respWriter.WriteBody([]byte(body)); err != nil {
			return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
		}

		return nil
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
