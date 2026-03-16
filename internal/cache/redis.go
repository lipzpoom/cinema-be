package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis(uri string) *redis.Client {
	// Basic setup from URI or hardcode if empty
	if uri == "" {
		uri = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: uri, // Use URI from config
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully.")

	return rdb
}
