// server.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

var clientConns = make(map[string]net.Conn)
var mutex = &sync.Mutex{}
var portCounter = 30001

func main() {
	serverAddr := "0.0.0.0:8080"

	// 开始监听
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", serverAddr, err)
	}

	for {
		// 接受新的客户端连接
		clientConn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept client connection: %v", err)
		}

		// 从客户端读取客户端 ID
		reader := bufio.NewReader(clientConn)
		clientID, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read client ID: %v", err)
		}

		// 动态分配端口
		for {
			listener, err := net.Listen("tcp", fmt.Sprintf(":%d", portCounter))
			if err != nil {
				// 如果端口已经被占用，那么更换端口
				portCounter++
				if portCounter > 39999 {
					log.Fatalf("No more ports available for allocation")
				}
			} else {
				// 将端口号和客户端 ID 两个信息发送给客户端
				_, err = clientConn.Write([]byte(fmt.Sprintf("%d %s", portCounter, clientID)))
				if err != nil {
					log.Fatalf("Failed to send port to client: %v", err)
				}
				// 创建一个新的goroutine来处理端口映射
				go func() {
					for {
						conn, err := listener.Accept()
						if err != nil {
							log.Fatalf("Failed to accept connection on port %d: %v", portCounter, err)
						}
						go handleMapping(conn, "localhost:3333")
					}
				}()
				break
			}
		}
		fmt.Printf("Allocated port %d for client %s\n", portCounter, clientID)
		portCounter++
	}
}

func handleMapping(conn net.Conn, target string) {
	defer conn.Close()
	targetConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatalf("Failed to connect to target: %v", err)
	}
	defer targetConn.Close()

	go func() {
		_, err := io.Copy(targetConn, conn)
		if err != nil {
			log.Fatalf("Failed to copy from source to target: %v", err)
		}
	}()

	_, err = io.Copy(conn, targetConn)
	if err != nil {
		log.Fatalf("Failed to copy from target to source: %v", err)
	}
}

func copyIO(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()
	io.Copy(dst, src)
}
