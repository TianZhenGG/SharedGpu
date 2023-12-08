package db

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

var encryptedAddr = "1vbshUShsWoE6NyjJ9JDTPs=" // 将这里替换为你的加密地址
var secretAddr = "uKmzjka2sm4A"                // 将这里替换为你的明文地址

func decryptAES(key, ciphertext string) (string, error) {
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

func InitRedis(ctx context.Context) (*redis.Client, error) {
	addr, err := decryptAES("the-key-has-to-be-32-bytes-long!", encryptedAddr)
	if err != nil {
		return nil, err
	}

	sercret, err := decryptAES("the-key-has-to-be-32-bytes-long!", secretAddr)
	if err != nil {
		return nil, err
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: sercret, // 没有密码
		DB:       0,
	})

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

// hget all key field gpu字段 且status为0
func HgetallByValue(ctx context.Context, rdb *redis.Client, field string, selectedValue string) ([]string, error) {
	var keys []string

	iter := rdb.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// 如果 key 的长度小于 36（UUID 的长度），则跳过
		if len(key) < 36 {
			continue
		}

		value, err := rdb.HGet(ctx, key, field).Result()
		if err != nil {
			fmt.Printf("Error getting field %s for key %s: %v\n", field, key, err)
			continue
		}
		// fmt.Printf("Key: %s, Value: %s\n", key, value)

		if strings.Contains(value, selectedValue) {

			keys = append(keys, key)
		}
	}
	if err := iter.Err(); err != nil {
		fmt.Printf("Error scanning keys: %v\n", err)
		return nil, err
	}
	// fmt.Printf("Keys: %v\n", keys)

	return keys, nil
}
