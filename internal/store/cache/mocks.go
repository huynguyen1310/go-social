package cache

import (
	"context"

	"github.com/huynguyen1310/social/internal/store"
)

func NewMockStore() Store {
	return Store{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct{}

func (m *MockUserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	return nil, nil
}

func (m *MockUserStore) Set(ctx context.Context, user *store.User) error {
	return nil
}
