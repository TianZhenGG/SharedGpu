package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

func main() {
	// 明文地址
	plaintext := ""

	// 密钥
	key := "the-key-has-to-be-32-bytes-long!"

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCFBEncrypter(block, []byte(key)[:block.BlockSize()])
	stream.XORKeyStream(ciphertext, []byte(plaintext))

	// 输出加密后的地址
	fmt.Println(base64.StdEncoding.EncodeToString(ciphertext))
}
