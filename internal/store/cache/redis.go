package cache

import "github.com/redis/go-redis/v9"

func NewRedisClient(addr, pw string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pw,
		DB:       db,
	})
	return client
}
