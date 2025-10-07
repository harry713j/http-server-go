package response

import (
	"errors"
	"fmt"
	"io"

	"github.com/harry713j/http-server/internal/header"
)

type writerState int

const (
	StateInit writerState = iota
	StateWrittenStatus
	StateWrittenHeaders
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

	w.state = StateWrittenHeaders
	return nil
}
func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != StateWrittenHeaders {
		return 0, errors.New("body must be written after headers")
	}

	n, err := w.w.Write(p)
	w.state = StateDone
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	// chunk format
	// 	<size in hex>\r\n
	// <data>\r\n
	if w.state != StateWrittenHeaders && w.state != StateWrittenStatus {
		return 0, errors.New("chunked body must be written after headers")
	}

	chunkedHeader := fmt.Sprintf("%x\r\n", len(p))

	if _, err := w.w.Write([]byte(chunkedHeader)); err != nil {
		return 0, err
	}

	n, err := w.w.Write(p)

	if err != nil {
		return n, err
	}

	if _, err := w.w.Write([]byte("\r\n")); err != nil {
		return 0, err
	}

	w.state = StateWrittenHeaders
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != StateWrittenHeaders {
		return 0, fmt.Errorf("cannot finish chunked body before writing chunks")
	}
	n, err := w.w.Write([]byte("0\r\n\r\n"))
	if err == nil {
		w.state = StateDone
	}
	return n, err
}

func (w *Writer) WriteTrailers(h header.Headers) error {
	for name, value := range h {
		if _, err := fmt.Fprintf(w.w, "%s: %s\r\n", name, value); err != nil {
			return err
		}
	}
	// End of trailers block
	_, err := fmt.Fprint(w.w, "\r\n")
	return err
}
