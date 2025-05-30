package config

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file: " + err.Error())
	}
	RedisClient = redis.NewClient(&redis.Options{
		Addr:      os.Getenv("REDIS_URL"),
		Username:  os.Getenv("REDIS_USERNAME"),
		Password:  os.Getenv("REDIS_PASSWORD"),
		TLSConfig: nil,
	})

	_, err = RedisClient.Ping(Ctx).Result()
	if err != nil {
		panic("Redis connection failed: " + err.Error())
	}
}
