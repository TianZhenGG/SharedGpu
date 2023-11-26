package db

import (
	"context"
	"fmt"

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

func GetAllValues() ([]string, error) {
	// 连接到 Redis
	ctx := context.Background()

	// 查询所有的键
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	// 获取所有键对应的值
	var values []string
	for _, key := range keys {
		value, err := rdb.Get(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get value for key %s: %w", key, err)
		}
		values = append(values, value)
	}

	return values, nil
}
