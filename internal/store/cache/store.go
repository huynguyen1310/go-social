package cache

import (
	"context"

	"github.com/huynguyen1310/social/internal/store"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
	}
}

func NewRedisStore(rdb *redis.Client) *Store {
	return &Store{
		Users: &UserStore{rdb: rdb},
	}
}
