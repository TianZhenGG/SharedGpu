package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func main() {
	// 建立于服务端tcp连接
	conn, err := net.Dial("tcp", "47.96.225.81:8080")
	// conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	clientID := "tian\n"
	// 将客户端 ID 发送给服务端
	_, err = conn.Write([]byte(clientID))
	if err != nil {
		log.Fatalf("Failed to send client ID to server: %v", err)
	}
	// 从服务端读取端口号和客户端 ID
	reader := bufio.NewReader(conn)
	portAndClientID, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read port and client ID: %v", err)
	}
	// 将端口号和客户端 ID 分开
	portAndClientID = strings.TrimSpace(portAndClientID)
	port, err := strconv.Atoi(portAndClientID[:5])

	//打印客户端ID，端口号
	fmt.Printf("Client ID: %s\n%d", portAndClientID, port)
}
