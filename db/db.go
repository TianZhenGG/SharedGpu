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
func GetAllValues() (map[string]interface{}, error) {
	// 连接到 Redis
	ctx := context.Background()

	// 查询所有的键
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	// 获取所有键对应的值
	values := make(map[string]interface{})
	for _, key := range keys {
		t, err := rdb.Type(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get type for key %s: %w", key, err)
		}

		var value interface{}
		switch t {
		case "string":
			value, err = rdb.Get(ctx, key).Result()
		case "list":
			value, err = rdb.LRange(ctx, key, 0, -1).Result()
		case "set":
			value, err = rdb.SMembers(ctx, key).Result()
		case "hash":
			value, err = rdb.HGetAll(ctx, key).Result()
		default:
			err = fmt.Errorf("unsupported type %s for key %s", t, key)
		}

		if err != nil {
			return nil, err
		}
		values[key] = value
	}

	return values, nil
}
