package config

import (
	"fmt"
	"os"
)

type Config struct {
	ContainerName      string
	ContainerAddress   string
	MongoContainerName string
	RedisContainerName string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Init() error {
	containerName := os.Getenv("CONTAINER_NAME")
	containerAddress := os.Getenv("CONTAINER_ADDRESS")
	mongoContainerName := os.Getenv("MONGO_CONTAINER_NAME")
	redisContainerName := os.Getenv("REDIS_CONTAINER_NAME")

	if containerName == "" || containerAddress == "" || mongoContainerName == "" || redisContainerName == "" {
		return fmt.Errorf("missing required environment variables")
	}

	c.ContainerName = containerName
	c.ContainerAddress = containerAddress
	c.MongoContainerName = mongoContainerName
	c.RedisContainerName = redisContainerName

	return nil
}
