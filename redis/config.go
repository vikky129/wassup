package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(containerName string) redis.UniversalClient {
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{fmt.Sprintf("%s:6379", containerName)},
		PoolSize: 10,
	})
	return redisClient
}
