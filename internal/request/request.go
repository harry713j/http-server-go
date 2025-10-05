package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/harry713j/http-server/internal/header"
)

const (
	requestStateParsingRequestLine = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     header.Headers
	Body        []byte
	state       int
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
	buff := make([]byte, 8)
	readToIndex := 0
	req := Request{
		state:   requestStateParsingRequestLine,
		Headers: header.NewHeaders(),
	}

	for req.state != requestStateDone {
		// if the buffere is full
		if readToIndex == len(buff) {
			newBuff := make([]byte, 2*len(buff))
			copy(newBuff, buff)
			buff = newBuff
		}
		// read more data into buffer
		numOfBytesRead, err := reader.Read(buff[readToIndex:])

		if err != nil {
			// there is nothing left to read
			if err == io.EOF {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}

			return nil, err
		}

		readToIndex += numOfBytesRead

		numOfBytesParsed, err := req.parse(buff[:readToIndex])

		if err != nil {
			return nil, err
		}

		if numOfBytesParsed > 0 {
			copy(buff, buff[numOfBytesParsed:readToIndex])
			readToIndex -= numOfBytesParsed
		}

	}

	return &req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	dataStr := string(data)
	index := strings.Index(dataStr, "\r\n")
	if index == -1 {
		return nil, 0, nil
	}

	line := dataStr[:index]
	parts := strings.SplitN(line, " ", 3)

	if len(parts) != 3 {
		return nil, 0, ErrInvalidRequestline
	}

	method := parts[0]
	target := parts[1]
	version := parts[2]

	for _, r := range method {
		if !unicode.IsUpper(r) {
			return nil, 0, ErrInvalidHttpMethod
		}
	}

	if !strings.HasPrefix(version, "HTTP/") {
		return nil, 0, fmt.Errorf("invalid version prefix")
	}

	versionNumber := strings.TrimPrefix(version, "HTTP/")
	if versionNumber != "1.1" && versionNumber != "1.0" {
		return nil, 0, fmt.Errorf("unsupported HTTP version: %s", versionNumber)
	}

	if !strings.HasPrefix(target, "/") {
		return nil, 0, ErrInvalidTarget
	}

	return &RequestLine{Method: method, RequestTarget: target, HttpVersion: versionNumber}, index + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != requestStateDone {
		numOfBytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		totalBytesParsed += numOfBytesParsed

		if err != nil {
			return totalBytesParsed, err
		}

		if numOfBytesParsed == 0 {
			break // need more data
		}
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateParsingRequestLine:
		reqLine, n, err := parseRequestLine(data)

		if err != nil {
			return 0, err
		}

		if reqLine == nil {
			return 0, nil
		}

		r.RequestLine = *reqLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)

		if err != nil {
			return 0, err
		}

		if done {
			r.state = requestStateParsingBody
		}

		return n, nil

	case requestStateParsingBody:
		// if Content-Type header present then parse the body
		contentLengthStr := r.Headers.Get("Content-Length")

		if contentLengthStr == "" {
			r.state = requestStateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)

		if err != nil {
			return 0, errors.New("invalid content length " + err.Error())
		}

		remaining := contentLength - len(r.Body)
		if remaining <= 0 {
			r.state = requestStateDone
			return 0, nil
		}

		// Only take up to 'remaining' bytes from data
		take := len(data)
		if take > remaining {
			return 0, fmt.Errorf("body longer than Content-Length")
		}

		r.Body = append(r.Body, data[:take]...)

		if len(r.Body) == contentLength {
			r.state = requestStateDone
		}

		return take, nil
	default:
		return 0, fmt.Errorf("invalid state %d", r.state)
	}

}
