package bdfs

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/windows"
)

func LoginBd(bduss string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "bdfs/baidupcs-go.exe", "login", "-bduss=", bduss)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	} else {
		cmd := exec.Command("bdfs/baidupcs-go", "login", "-bduss="+bduss)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
}

func CreateDir(remoteDir string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "bdfs/baidupcs-go.exe", "mkdir", remoteDir)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	} else {
		cmd := exec.Command("bdfs/baidupcs-go", "mkdir", remoteDir)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
}

func DeleteDir(remoteDir string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "bdfs/baidupcs-go.exe", "rm", remoteDir)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	} else {
		cmd := exec.Command("bdfs/baidupcs-go", "rm", remoteDir)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
}

func Download(remoteDir, remoteFile string, savepath string) error {
	// 创建一个新的 Context 实例
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 如果是 windows
	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "powershell", "bdfs/baidupcs-go.exe", "d  --ow --status --saveto", savepath, "-p 20 -l 20", remoteDir+"/"+remoteFile)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
	} else {
		fmt.Println("linux")
		cmd := exec.CommandContext(ctx, "bdfs/baidupcs-go", "download", "--ow --status --save -p 20 -l 20", remoteDir+remoteFile)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	// 下载完成后取消进程
	cancel()

	return nil
}

func Upload(localFile, remoteDir string) error {
	// 创建一个新的 Context 实例
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "powershell", "bdfs/baidupcs-go.exe", "upload", "--policy overwrite -p 10 -l 10", localFile, remoteDir)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
	} else {
		cmd := exec.CommandContext(ctx, "bdfs/baidupcs-go", "upload", "-p 10 -l 10", localFile, remoteDir)
		cmd.SysProcAttr = &windows.SysProcAttr{HideWindow: true}
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	// 上传完成后取消进程
	cancel()

	return nil
}

func Zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果当前路径是 "miniconda"，则跳过这个路径
		if info.IsDir() && info.Name() == "miniconda" {
			return filepath.SkipDir
		}

		// 如果文件或文件夹的名称以 "." 开头，跳过它
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// 创建一个新的文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 设置文件头的名称为相对路径
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})
	return err
}

func Unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	// 确保在函数返回时关闭 reader
	defer reader.Close()

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// folderPath := "E:/SharedGpu"
	// uuid := "dadadasdadafasfa"
	// err := CreateDir(uuid)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = Upload(folderPath+".zip", uuid)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	uuidStr := "e1aa6a13-df2a-5bc6-a3e4-c30ed0fd468e"
	fmt.Println(uuidStr)
	// err := Download("miniconda", "miniconda.zip")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	err := Download(uuidStr, "test.py", "./")
	if err != nil {
		fmt.Println(err)
	}

	// // 解压文件
	// err := Unzip("miniconda.zip", "./")
	// if err != nil {
	// 	fmt.Println("failed to unzip file:", err)
	// }
}
