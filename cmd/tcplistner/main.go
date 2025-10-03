package main

import (
	"fmt"
	"net"

	"github.com/k4rldoherty/tcp-from-http/internal/request"
)

const ADDRESS = "127.0.0.1:42069"

func main() {
	l, err := net.Listen("tcp", ADDRESS)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	fmt.Printf("Listening on address: %v\n", ADDRESS)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Connection accepted from %v\n", conn.RemoteAddr())
		r, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", r.RequestLine.Method)
		fmt.Printf("- Target: %v\n", r.RequestLine.Target)
		fmt.Printf("- Version: %v\n", r.RequestLine.HTTPVersion)
		fmt.Println("Headers:")
		for k, v := range r.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Printf("Connection closed from %v\n", conn.RemoteAddr())
	}
}
