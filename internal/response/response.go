// Package response
package response

import (
	"fmt"
	"io"

	"github.com/k4rldoherty/tcp-from-http/internal/headers"
)

type StatusCode int

const (
	OK          StatusCode = 200
	BadRequest  StatusCode = 400
	ServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, s StatusCode) error {
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

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%v\r\n", contentLen))
	h.Set("Connection", "close\r\n")
	h.Set("Content-Type", "text/plain\r\n")
	return h
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	for k, v := range h {
		_, err := w.Write(fmt.Appendf(nil, "%v: %v", k, v))
		if err != nil {
			return err
		}
	}
	return nil
}
