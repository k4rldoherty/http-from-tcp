package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const ADDRESS = "localhost:42069"

func main() {
	addr, err := net.ResolveUDPAddr("udp", ADDRESS)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("connection opened")
	defer conn.Close()
	defer fmt.Println("connection closed")
	rd := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%v", ">")
		in, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		_, err = conn.Write([]byte(in))
		if err != nil {
			fmt.Println(err)
		}
	}
}
