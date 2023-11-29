package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	// 下载文件
	err := download("sharedgpu/", "miniconda.zip")
	if err != nil {
		log.Fatal(err)
	}

	// 解压文件
	err = untar("path/to/local/file.tar", "path/to/dest/dir")
	if err != nil {
		log.Fatal(err)
	}

	// 压缩文件
	err = tarit("path/to/local/dir", "path/to/local/file.tar")
	if err != nil {
		log.Fatal(err)
	}

	// 上传文件
	err = upload("path/to/local/file.tar", "/path/to/remote/dir")
	if err != nil {
		log.Fatal(err)
	}
}

func download(remoteDir, remoteFile string) error {
	//if windows
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "bdfs/baidupcs-go.exe", "d  --ow --status --save -p 20 -l 20", remoteDir+remoteFile)
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	} else {
		fmt.Println("linux")
		cmd := exec.Command("bdfs/baidupcs-go", "download", "--ow --status --save -p 20 -l 20", remoteDir+remoteFile)
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
}

func upload(localFile, remoteDir string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "bdfs/baidupcs-go.exe", "upload", localFile, remoteDir)
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	} else {
		cmd := exec.Command("bdfs/baidupcs-go", "upload", localFile, remoteDir)
		err := cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}
}

func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func tarit(source, target string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(source, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(filepath.Base(source), strings.TrimPrefix(file, source))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}
