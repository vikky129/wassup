package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	PubSubWassupChannel = "wassup_channel"
)

type PubSubHandler struct {
	redisClient redis.UniversalClient
}

func NewPubSubHandler() *PubSubHandler {
	return &PubSubHandler{}
}

func (psh *PubSubHandler) Init(ctx context.Context) error {
	redisClient := NewRedisClient("")
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return err
	}

	redisClient.Subscribe(ctx, PubSubWassupChannel)

	psh.redisClient = redisClient
	return nil
}

func (psh *PubSubHandler) PublishMessage(ctx context.Context, message string) error {
	return psh.redisClient.Publish(ctx, PubSubWassupChannel, message).Err()
}
