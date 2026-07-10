package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Follower struct {
	FollowerID int64 `db:"follower_id"`
	UserID     int64 `db:"user_id"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followerID int64, userID int64) error {
	query := `INSERT INTO followers (follower_id, user_id) VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, followerID, userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}

	return nil
}

func (s *FollowerStore) Unfollow(ctx context.Context, followerID int64, userID int64) error {
	query := `DELETE FROM followers WHERE follower_id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, followerID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
