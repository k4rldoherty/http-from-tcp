// Package server
package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/k4rldoherty/http-from-tcp/internal/headers"
	"github.com/k4rldoherty/http-from-tcp/internal/request"
	"github.com/k4rldoherty/http-from-tcp/internal/response"
)

type Server struct {
	Port     int
	Listener net.Listener
	IsOpen   *atomic.Bool
	Handler  Handler
}

type HandlerError struct {
	Code    int
	Message string
}

type Handler func(rw *response.Writer, r *request.Request)

func WriteError(w *response.Writer, err *HandlerError, body string) {
	if e := w.WriteStatusLine(response.StatusCode(err.Code)); e != nil {
		log.Printf("error writing status line: %v\n", e)
		return
	}
	h := response.GetDefaultHeaders(len([]byte(body)))
	h.Set("Content-Type", "text/html")
	if e := w.WriteHeaders(h); e != nil {
		log.Printf("error writing headers: %v\n", e)
		return
	}
	if _, e := w.WriteBody([]byte(body)); e != nil {
		log.Printf("error writing body: %v\n", e)
		return
	}
}

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Printf("Serve: %v", err)
		return nil, err
	}
	state := atomic.Bool{}
	state.Store(true)
	s := &Server{
		Port:     port,
		Listener: l,
		IsOpen:   &state,
		Handler:  handler,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.IsOpen.Store(false)
	err := s.Listener.Close()
	if err != nil {
		log.Printf("Close: %v", err)
		return err
	}
	return nil
}

func (s *Server) listen() {
	for s.IsOpen.Load() {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.IsOpen.Load() {
				log.Printf("Listen: %v", err)
			}
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing connection: %v\n", err)
			return
		}
	}()

	reqWriter := response.NewWriter(conn, &headers.Headers{})

	r, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("handle: %v", err)
		WriteError(reqWriter, &HandlerError{
			Code:    400,
			Message: "Bad Request",
		}, "Could not form a request from data recieved")
		return
	}
	s.Handler(reqWriter, r)
}
