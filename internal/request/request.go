package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const (
	stateInitialized = iota
	stateDone
)

type Request struct {
	RequestLine RequestLine
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
		state: stateInitialized,
	}

	for req.state != stateDone {
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
				if req.state != stateDone {
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
	parts := strings.Split(line, " ")

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

	if version != "HTTP/1.1" {
		return nil, 0, ErrInvalidHttpVersion
	}

	if !strings.HasPrefix(target, "/") {
		return nil, 0, ErrInvalidTarget
	}

	versionNumber := version[5:]

	return &RequestLine{Method: method, RequestTarget: target, HttpVersion: versionNumber}, index + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == stateDone {
		// nothing left to parse
		return 0, nil
	}
	r.state = stateInitialized
	reqLine, consumed, err := parseRequestLine(data)

	if err != nil {
		return 0, err
	}

	if consumed > 0 {
		r.RequestLine = *reqLine
		r.state = stateDone
		return consumed, nil
	}

	return consumed, nil
}
