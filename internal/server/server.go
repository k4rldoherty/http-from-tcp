// Package server
package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/k4rldoherty/tcp-from-http/internal/response"
)

type Server struct {
	Port     int
	Listener net.Listener
	IsOpen   *atomic.Bool
}

func Serve(port int) (*Server, error) {
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
	defer conn.Close()
	if err := response.WriteStatusLine(conn, response.OK); err != nil {
		log.Printf("Handle: %v", err)
		return
	}
	if err := response.WriteHeaders(conn, response.GetDefaultHeaders(0)); err != nil {
		log.Printf("Handle: %v", err)
		return
	}
	conn.Write([]byte("\r\n"))
}
