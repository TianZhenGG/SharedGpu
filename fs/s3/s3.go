package main

import (
	"fmt"
	"sharedgpu/utils"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var endpoint = "irW2zAG1rHdetZe/b4NdGMuEr9pvkzWFAWiZ/UNSyguVdcgHCuig/ZYc"
var accessKeyId = "rpWD9Uf7uioDitP/SKkKHnPVd8jN/qvy"
var secretKeyId = "0YaKyivs0yp9rr6ldaA9My7dy/79q75YyLoZsDmm"
var bucket *oss.Bucket

func init() {
	endpoint, err := utils.DecryptAES("the-key-has-to-be-32-bytes-long!", endpoint)
	if err != nil {
		fmt.Println("Error:", err)
	}

	accessKeyId, err := utils.DecryptAES("the-key-has-to-be-32-bytes-long!", accessKeyId)
	if err != nil {
		fmt.Println("Error:", err)
	}

	secretKeyId, err := utils.DecryptAES("the-key-has-to-be-32-bytes-long!", secretKeyId)
	if err != nil {
		fmt.Println("Error:", err)
	}
	client, err := oss.New(endpoint, accessKeyId, secretKeyId)
	if err != nil {
		fmt.Println("Error:", err)

	}
	bucket, err = client.Bucket("sharedgpu")
	if err != nil {
		fmt.Println("Error:", err)
	}

}

func Upload(localFile, remoteDir string) error {
	// 上传文件
	remoteFile := remoteDir + "/" + localFile
	err := bucket.PutObjectFromFile(remoteFile, localFile)
	return err
}

func CreateDir(remoteDir string) error {
	// 创建目录
	err := bucket.PutObject(remoteDir+"/", nil)
	return err
}

func DeleteDir(remoteDir string) error {
	// 删除目录
	err := bucket.DeleteObject(remoteDir + "/")
	return err
}

func Download(remoteFile, localDir string) error {
	// 下载文件
	localFile := localDir + "/" + remoteFile
	err := bucket.GetObjectToFile(remoteFile, localFile)
	return err
}

func main() {
	// 上传文件
	err := Upload("fs/s3/s3.go", "e1aa6a13-df2a-5bc6-a3e4-c30ed0fd468e")
	if err != nil {
		fmt.Println("Error:", err)
	}

}
