package redis

import (
	"context"
	"fmt"
	"log"
	"personal/wassup/config"
	"strings"

	"github.com/redis/go-redis/v9"
)

type CacheHandler struct {
	redisClient redis.UniversalClient
	config      *config.Config
}

func NewCacheHandler(c *config.Config) *CacheHandler {
	return &CacheHandler{
		config: c,
	}
}

func (ch *CacheHandler) Init(ctx context.Context) error {
	redisClient := NewRedisClient(ch.config.RedisContainerName)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}

	ch.redisClient = redisClient
	return nil
}

func (ch *CacheHandler) StoreContainerDetails(ctx context.Context, userID, deviceID string) error {
	return ch.redisClient.HSet(ctx, fmt.Sprintf("%s:%s", userID, deviceID), map[string]any{
		fmt.Sprintf("%s:%s:name", userID, deviceID): ch.config.ContainerName,
		ch.config.ContainerName:                     ch.config.ContainerAddress,
	}).Err()
}

func (ch *CacheHandler) CheckKeyPrefixExists(key string) (bool, error) {
	ctx := context.Background()
	iter := ch.redisClient.Scan(ctx, 0, fmt.Sprintf("%s:*", key), 0).Iterator()
	for iter.Next(ctx) {
		return true, nil
	}

	if err := iter.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (ch *CacheHandler) DeleteKey(key string) error {
	ctx := context.Background()
	return ch.redisClient.Del(ctx, key).Err()
}

func (ch *CacheHandler) GetContainerDetails(ctx context.Context, userID string) ([]map[string]string, error) {
	var results []map[string]string

	iter := ch.redisClient.Scan(ctx, 0, fmt.Sprintf("%s:*", userID), 0).Iterator()
	for iter.Next(ctx) {
		result, err := ch.redisClient.HGetAll(ctx, iter.Val()).Result()
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (ch *CacheHandler) GetLiveUserAddressForUser(ctx context.Context, userID string, participants []string) (map[string][]string, error) {
	var userAddress = make(map[string][]string)
	for _, participant := range participants {
		if participant == userID {
			continue
		}

		recipientLive, _ := ch.CheckKeyPrefixExists(participant)
		log.Printf("User %s live status: %v\n", participant, recipientLive)
		if recipientLive {
			containerAddress, err := ch.GetContainerDetails(ctx, participant)
			if err != nil {
				fmt.Println("Error getting container details:", err)
				continue
			}

			for _, addressMap := range containerAddress {
				for key, address := range addressMap {
					if strings.HasSuffix(key, "name") {
						continue
					}
					userAddress[participant] = append(userAddress[participant], address)
				}
			}
		}
	}

	return userAddress, nil
}

func (ch *CacheHandler) Set(key string, otp any) error {
	ctx := context.Background()
	return ch.redisClient.Set(ctx, key, otp, 0).Err()
}

func (ch *CacheHandler) Get(key string) (string, error) {
	ctx := context.Background()
	return ch.redisClient.Get(ctx, key).Result()
}
