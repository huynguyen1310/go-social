package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huynguyen1310/social/internal/store"
	"github.com/redis/go-redis/v9"
)

type UserStore struct {
	rdb *redis.Client
}

const cacheExpiration = time.Minute * 15

func (s *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", id)
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user store.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, cacheKey, data, cacheExpiration).Err()
}
