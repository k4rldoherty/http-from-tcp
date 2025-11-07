package response

import (
	"fmt"
	"io"

	"github.com/k4rldoherty/http-from-tcp/internal/headers"
)

type Writer struct {
	Destination io.Writer
	StatusCode  StatusCode
	Headers     *headers.Headers
	State       WriterState
}

type WriterState int

const (
	WritingStatusLine WriterState = iota
	WritingHeaders
	WritingBody
	Done
)

func NewWriter(dest io.Writer, h *headers.Headers) *Writer {
	return &Writer{
		Destination: dest,
		Headers:     h,
		State:       WritingStatusLine,
	}
}

func (w *Writer) Write(b []byte) (int, error) {
	n, err := w.Destination.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (w *Writer) WriteStatusLine(s StatusCode) error {
	if w.State != WritingStatusLine {
		return fmt.Errorf("invalid state")
	}
	defer func() {
		w.State = WritingHeaders
	}()
	switch s {
	case OK:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		return err
	case BadRequest:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		return err
	case ServerError:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		return err
	default:
		r := fmt.Sprintf("HTTP/1.1 %v ", s)
		_, err := w.Write([]byte(r))
		return err
	}
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.State != WritingHeaders {
		return fmt.Errorf("invalid state")
	}
	defer func() {
		w.State = WritingBody
	}()
	for k, v := range h {
		_, err := w.Write(fmt.Appendf(nil, "%v: %v\r\n", k, v))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteBody(b []byte) (int, error) {
	if w.State != WritingBody {
		return 0, fmt.Errorf("invalid state")
	}
	defer func() {
		w.State = Done
	}()
	n, err := w.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBody(b []byte) error {
	if w.State != WritingBody {
		return fmt.Errorf("invalid state")
	}
	// Write the length of the chunk
	if _, err := fmt.Fprintf(w, "%X\r\n", len(b)); err != nil {
		return err
	}
	// Write the chunk, in raw bytes
	if _, err := w.Write(b); err != nil {
		return err
	}
	// Write the trailing newline
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteChunkedBodyDone() error {
	w.State = Done
	_, err := w.Write([]byte("0\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteTrailer(t headers.Headers) error {
	if w.State != Done {
		return fmt.Errorf("invalid state for writing trailers")
	}
	for k, v := range t {
		_, err := fmt.Fprintf(w, "%v: %v\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}
