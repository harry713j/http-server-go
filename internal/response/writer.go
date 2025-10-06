package response

import (
	"errors"
	"io"

	"github.com/harry713j/http-server/internal/header"
)

type writerState int

const (
	StateInit writerState = iota
	StateWrittenStatus
	StateWritterHeaders
	StateDone
)

type Writer struct {
	w     io.Writer
	state writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w, state: StateInit}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StateInit {
		return errors.New("status line must be written first")
	}
	if err := WriteStatusLine(w.w, statusCode); err != nil {
		return err
	}
	w.state = StateWrittenStatus
	return nil
}
func (w *Writer) WriteHeaders(headers header.Headers) error {
	if w.state != StateWrittenStatus {
		return errors.New("header must be written after status line")
	}

	if err := WriteHeaders(w.w, headers); err != nil {
		return err
	}

	w.state = StateWritterHeaders
	return nil
}
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != StateWritterHeaders {
		return 0, errors.New("body must be written after headers")
	}

	n, err := w.w.Write(p)
	w.state = StateDone
	return n, err
}
