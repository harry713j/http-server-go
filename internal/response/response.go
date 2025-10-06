package response

import (
	"fmt"
	"io"

	"github.com/harry713j/http-server/internal/header"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case StatusOk:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalServerError:
		reason = "Internal Server Error"
	default:
		reason = ""
	}

	if reason == "" {
		_, err := fmt.Fprintf(w, "HTTP/1.1 %d \r\n", statusCode)

		return err
	}

	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", statusCode, reason)
	return err
}

func GetDefaultHeaders(contentLen int) header.Headers {
	h := header.NewHeaders()

	h["Content-Length"] = fmt.Sprintf("%d", contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"

	return h
}

func WriteHeaders(w io.Writer, headers header.Headers) error {
	for key, value := range headers {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", key, value); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\r\n")
	return err
}
