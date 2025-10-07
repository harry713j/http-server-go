package header

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlf := "\r\n"
	dataStr := string(data)
	crlfIndex := strings.Index(dataStr, crlf)

	if crlfIndex == -1 {
		return 0, false, nil
	}

	if crlfIndex == 0 {
		return 2, true, nil
	}
	line := dataStr[:crlfIndex]
	line = strings.TrimSpace(line)

	colonIndex := strings.Index(line, ":")

	if colonIndex == -1 || (colonIndex > 0 && line[colonIndex-1] == ' ') {
		return 0, false, fmt.Errorf("invalid header format")
	}

	key := strings.TrimSpace(line[:colonIndex])
	value := strings.TrimSpace(line[colonIndex+1:])

	if key == "" || value == "" {
		return 0, false, errors.New("invalid header key or value")
	}

	for _, r := range key {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r)) &&
			!strings.ContainsRune("!#$%&'*+-.^_`|~", r) {
			return 0, false, errors.New("invalid character in header key")
		}
	}

	key = strings.ToLower(key)

	if existingVal, ok := h[key]; !ok {
		h[key] = value
	} else {
		h[key] = fmt.Sprintf("%s, %s", existingVal, value)
	}

	return crlfIndex + 2, false, nil
}

func (h Headers) Get(key string) string {
	key = strings.ToLower(key)
	for k, v := range h {
		if strings.ToLower(k) == key {
			return v
		}
	}

	return ""
}

func (h Headers) Add(key, value string) {
	h[key] = value
}

func (h Headers) Remove(key string) {
	delete(h, key)
}
