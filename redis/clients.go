package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
)

func InitRedis() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		DB:       0,
		Password: "",
	})

	_, err := redisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %v", err)
	}
	Client = redisClient
	return nil
}

func CloseRedis() error {
	return Client.Close()
}
