package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	bdfs "github.com/qjfoidnh/BaiduPCS-Go"
	"github.com/urfave/cli"
)

var (
	app *cli.App
)

func Upload(localFile, remoteDir string) error {

	err := app.Run([]string{"BaiduPCS-Go", "upload", "--policy", "overwrite", "-p", "10", "-l", "10", localFile, remoteDir})
	if err != nil {
		return err
	}
	return nil
}

func CreateDir(remoteDir string) error {
	err := app.Run([]string{"BaiduPCS-Go", "mkdir", remoteDir})
	if err != nil {
		return err
	}
	return nil
}

func DeleteDir(remoteDir string) error {
	err := app.Run([]string{"BaiduPCS-Go", "rm", remoteDir})
	if err != nil {
		return err
	}
	return nil
}

func Download(remoteDir, remoteFile string, savepath string) error {
	err := app.Run([]string{"BaiduPCS-Go", "download", "--ow", "--status", "-p", "10", "-l", "10", "--saveto", fmt.Sprintf("%s", savepath), remoteDir + "/" + remoteFile})
	if err != nil {
		return err

	}
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

func InitBdfs(bduss string) error {
	app = bdfs.GetApp()
	err := app.Run([]string{"BaiduPCS-Go", "login", "-bduss=" + bduss})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	bduss := "N3NUNuR0J3dU0zdnFVTHV0dmxMSTlWZDQ4Q2dqYzRJNFpGY0VvbnhTVTZXcFZsSUFBQUFBJCQAAAAAAAAAAAEAAACm0-4~s~TLrrm10fjT4wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADrNbWU6zW1lMj"
	err := InitBdfs(bduss)
	if err != nil {
		fmt.Println(err)
	}
	err = Download("e1aa6a13-df2a-5bc6-a3e4-c30ed0fd468e", "test.zip", "./")
	if err != nil {
		fmt.Println(err)
	}
	// err = Upload("test.go", "/")
	// if err != nil {
	// 	fmt.Println(err)
	// }

}
