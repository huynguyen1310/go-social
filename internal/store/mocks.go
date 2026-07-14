package store

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func NewMockStorage(t *testing.T) Storage {
	return Storage{
		Users:     &MockUsersStore{},
		Posts:     &MockPostsStore{},
		Comments:  &MockCommentsStore{},
		Followers: &MockFollowersStore{},
		Roles:     &MockRolesStore{},
	}
}

type MockUsersStore struct{}

func (m *MockUsersStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	return nil
}

func (m *MockUsersStore) Get(ctx context.Context, id int64) (*User, error) {
	return &User{
		ID:       id,
		Username: "testuser",
		Email:    "test@example.com",
		Role: &Role{
			ID:    1,
			Name:  "user",
			Level: 1,
		},
	}, nil
}

func (m *MockUsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return &User{
		ID:       1,
		Username: "testuser",
		Email:    email,
	}, nil
}

func (m *MockUsersStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return nil
}

func (m *MockUsersStore) Activate(ctx context.Context, token string) error {
	return nil
}

type MockPostsStore struct{}

func (m *MockPostsStore) Create(ctx context.Context, post *Post) error {
	return nil
}

func (m *MockPostsStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	return &Post{ID: id, Title: "test post", Content: "test content"}, nil
}

func (m *MockPostsStore) GetUserFeed(ctx context.Context, userId int64, fq PaginationFeedQuery) ([]*PostWithMetadata, error) {
	return nil, nil
}

func (m *MockPostsStore) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *MockPostsStore) Update(ctx context.Context, post *Post) error {
	return nil
}

type MockCommentsStore struct{}

func (m *MockCommentsStore) Create(ctx context.Context, comment *Comment) error {
	return nil
}

func (m *MockCommentsStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	return nil, nil
}

type MockFollowersStore struct{}

func (m *MockFollowersStore) Follow(ctx context.Context, followerID int64, userID int64) error {
	return nil
}

func (m *MockFollowersStore) Unfollow(ctx context.Context, followerID int64, userID int64) error {
	return nil
}

type MockRolesStore struct{}

func (m *MockRolesStore) GetByName(ctx context.Context, name string) (*Role, error) {
	return &Role{ID: 1, Name: "user", Level: 1}, nil
}
