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
		rl, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", rl.RequestLine.Method)
		fmt.Printf("- Target: %v\n", rl.RequestLine.Target)
		fmt.Printf("- Version: %v\n", rl.RequestLine.HTTPVersion)
		fmt.Printf("Connection closed from %v\n", conn.RemoteAddr())
	}
}
