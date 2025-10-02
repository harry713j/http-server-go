package request

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var (
	ErrInvalidRequestline = errors.New("invalid request line of the request")
	ErrInvalidHttpMethod  = errors.New("invalid http method")
	ErrInvalidHttpVersion = errors.New("invalid http version")
	ErrInvalidTarget      = errors.New("invalid request target")
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	requestStr := string(data)
	requestParts := strings.SplitN(requestStr, "\r\n", 2)

	reqLine, err := parseRequestLine(requestParts[0])

	if err != nil {
		return nil, err
	}

	request := &Request{
		RequestLine: *reqLine,
	}

	return request, nil
}

func parseRequestLine(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return nil, ErrInvalidRequestline
	}

	method := parts[0]
	target := parts[1]
	version := parts[2]

	for _, r := range method {
		if !unicode.IsUpper(r) {
			return nil, ErrInvalidHttpMethod
		}
	}

	if version != "HTTP/1.1" {
		return nil, ErrInvalidHttpVersion
	}

	if !strings.HasPrefix(target, "/") {
		return nil, ErrInvalidTarget
	}

	versionNumber := version[5:]

	return &RequestLine{Method: method, RequestTarget: target, HttpVersion: versionNumber}, nil
}
