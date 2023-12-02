package db

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// 创建一个全局的 Redis 客户端
var ctx = context.Background()
var rdb *redis.Client

func InitRedis() (*redis.Client, error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "47.96.225.81:6379",
		Password: "", // 如果没有密码，留空
		DB:       0,  // 使用默认 DB
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

// hget all key field gpu字段 且status为0
func HgetallByValue(ctx context.Context, rdb *redis.Client, field string, selectedValue string) ([]string, error) {
	var keysWithSelectedValue []string

	// 获取所有的键
	keys, _, err := rdb.Scan(ctx, 0, "*", 0).Result()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		// 获取指定字段的值
		value, err := rdb.HGet(ctx, key, field).Result()
		if err != nil {
			return nil, err
		}

		// 如果值是 selectedValue将键添加到结果列表
		if value == selectedValue {
			keysWithSelectedValue = append(keysWithSelectedValue, key)
		}
	}

	return keysWithSelectedValue, nil
}
