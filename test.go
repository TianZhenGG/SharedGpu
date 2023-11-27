package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":3333")
	if err != nil {
		log.Fatalf("Failed to listen on port 3333: %v", err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept connection: %v", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// 这里处理连接，例如读取和写入数据
	fmt.Println("New connection!")
	defer conn.Close()
}
