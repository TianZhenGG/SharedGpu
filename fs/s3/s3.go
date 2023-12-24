package s3

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var endpoint = "irW2zAG1rHdetZe/b4NdGMuEr9pvkzWFAWiZ/UNSyguVdcgHCuig/ZYc"
var accessKeyId = "rpWD9Uf7uioDitP/SKkKHnPVd8jN/qvy"
var secretKeyId = "0YaKyivs0yp9rr6ldaA9My7dy/79q75YyLoZsDmm"
var bucket *oss.Bucket

func DecryptAES(key, ciphertext string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	decodedCiphertext, _ := base64.StdEncoding.DecodeString(ciphertext)

	plaintext := make([]byte, len(decodedCiphertext))
	stream := cipher.NewCFBDecrypter(block, []byte(key)[:block.BlockSize()])
	stream.XORKeyStream(plaintext, decodedCiphertext)

	return string(plaintext), nil
}

func init() {
	endpoint, err := DecryptAES("the-key-has-to-be-32-bytes-long!", endpoint)
	if err != nil {
		fmt.Println("Error:", err)
	}

	accessKeyId, err := DecryptAES("the-key-has-to-be-32-bytes-long!", accessKeyId)
	if err != nil {
		fmt.Println("Error:", err)
	}

	secretKeyId, err := DecryptAES("the-key-has-to-be-32-bytes-long!", secretKeyId)
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

func ClearFiles(currentDir string) error {
	tempFilePattern := `^\.temp\d+$`
	tempFileRegexp := regexp.MustCompile(tempFilePattern)

	err := filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".zip") || tempFileRegexp.MatchString(info.Name())) {
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func Upload(localFile, remoteDir string) error {
	// 上传文件
	remoteFile := path.Join(remoteDir, filepath.Base(localFile))
	err := bucket.PutObjectFromFile(remoteFile, localFile)
	return err
}

func Uploadzip(localFile, remoteDir string) error {

	remoteFile := path.Join(remoteDir, filepath.Base(localFile))
	err := bucket.PutObjectFromFile(remoteFile, localFile)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	tempFile, err := ioutil.TempFile(".", ".temp")
	if err != nil {
		fmt.Println("failed to create temp file:", err)
		return err
	}
	defer os.Remove(tempFile.Name())

	return err
}

func CreateDir(remoteDir string) error {
	// 创建目录
	err := bucket.PutObject(remoteDir, nil)
	return err
}

func DeleteDir(remoteDir string) error {
	// 删除目录
	err := bucket.DeleteObject(remoteDir)
	return err
}

func Download(remoteDir, remoteFile string, savepath string) error {
	localFile := path.Join(savepath, remoteFile)
	remoteFile = path.Join(remoteDir, remoteFile)
	fmt.Println("remoteFile:", remoteFile)
	fmt.Println("localFile:", localFile)
	err := bucket.GetObjectToFile(remoteFile, localFile)
	return err
}

func main() {
	// 上传文件
	// err := Upload("miniconda.zip", "e1aa6a13-df2a-5bc6-a3e4-c30ed0fd468e")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	err := Download("e1aa6a13-df2a-5bc6-a3e4-c30ed0fd468e", "", ".")
	if err != nil {
		fmt.Println("failed to download file:", err)
	}

}
