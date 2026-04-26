package initialize

import (
	"github.com/redis/go-redis/v9"
)

func GetRedisClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}
