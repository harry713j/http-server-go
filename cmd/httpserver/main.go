package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/harry713j/http-server/internal/header"
	"github.com/harry713j/http-server/internal/request"
	"github.com/harry713j/http-server/internal/response"
	"github.com/harry713j/http-server/internal/server"
)

const port = 42069

func main() {
	handler := func(w io.Writer, r *request.Request) *server.HandlerError {
		respWriter := response.NewWriter(w)

		if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/") {
			proxyPath := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin")
			proxyUrl := "https://httpbin.org" + proxyPath

			resp, err := http.Get(proxyUrl)

			if err != nil {
				return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
			}

			defer resp.Body.Close()

			headers := response.GetDefaultHeaders(0)
			headers.Add("Content-Type", resp.Header.Get("Content-Type"))
			headers.Remove("content-length")
			headers.Add("Transfer-Encoding", "chunked")

			headers.Add("Trailer", "X-Content-SHA256, X-Content-Length")

			if err := respWriter.WriteStatusLine(response.StatusOk); err != nil {
				return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
			}

			if err := respWriter.WriteHeaders(headers); err != nil {
				return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
			}

			buff := make([]byte, 1024)
			var fullBody bytes.Buffer // to calculate hash and length later

			for {
				n, err := resp.Body.Read(buff)

				if err != nil {
					if err == io.EOF {
						break
					}

					return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
				}

				if n > 0 {
					fullBody.Write(buff[:n])

					if _, err := respWriter.WriteChunkedBody(buff[:n]); err != nil {
						return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
					}
				}

			}

			if _, err := respWriter.WriteChunkedBodyDone(); err != nil {
				return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
			}

			// Compute SHA256 + length
			hash := sha256.Sum256(fullBody.Bytes())
			hashHex := fmt.Sprintf("%x", hash[:])
			length := strconv.Itoa(fullBody.Len())

			// Write trailers
			trailers := header.NewHeaders()
			trailers.Add("X-Content-SHA256", hashHex)
			trailers.Add("X-Content-Length", length)

			if err := respWriter.WriteTrailers(trailers); err != nil {
				return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: err.Error()}
			}

			return nil
		}

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
