package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserId    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	Version   int       `json:"version"`
}

type PostStore struct {
	db *sql.DB
}

type PostWithMetadata struct {
	Post
	CommentCount int    `json:"comment_count"`
	Username     string `json:"username"`
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts (title, content, user_id, tags)
    VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	var post Post
	query := `SELECT id, title, content, user_id, tags, version, created_at, updated_at FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserId,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(
		ctx,
		query,
		id,
	)

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

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
        UPDATE posts
        SET title = $1, content = $2, tags = $3, version = version + 1
        WHERE id = $4
        AND version = $5 RETURNING version
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		post.ID,
		post.Version,
	).Scan(
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userId int64, fq PaginationFeedQuery) ([]*PostWithMetadata, error) {
	sortDir := "DESC"
	if fq.Sort == "asc" || fq.Sort == "ASC" {
		sortDir = "ASC"
	}

	args := []any{userId}
	paramIdx := 2

	query := `
	SELECT
		p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
		u.username,
		COUNT(c.id) AS comments_count
	FROM posts p
	LEFT JOIN comments c ON c.post_id = p.id
	LEFT JOIN users u ON p.user_id = u.id
	JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
	WHERE (f.user_id = $1 OR p.user_id = $1)`

	// Tags filter — comma-separated list, match any
	if fq.Tags != "" {
		tags := strings.Split(fq.Tags, ",")
		tagParams := make([]string, 0, len(tags))
		for _, tag := range tags {
			tagParams = append(tagParams, fmt.Sprintf("$%d", paramIdx))
			args = append(args, strings.TrimSpace(tag))
			paramIdx++
		}
		query += fmt.Sprintf(" AND p.tags @> ARRAY[%s]::text[]", strings.Join(tagParams, ", "))
	}

	// Search filter — matches title OR content
	if fq.Search != "" {
		query += fmt.Sprintf(" AND (p.title ILIKE '%%%%' || $%d || '%%%%' OR p.content ILIKE '%%%%' || $%d || '%%%%')", paramIdx, paramIdx)
		args = append(args, fq.Search)
		paramIdx++
	}

	// Since filter — posts created after this date
	if fq.Since != "" {
		query += fmt.Sprintf(" AND p.created_at >= $%d", paramIdx)
		args = append(args, fq.Since)
		paramIdx++
	}

	// Until filter — posts created before this date
	if fq.Until != "" {
		query += fmt.Sprintf(" AND p.created_at <= $%d", paramIdx)
		args = append(args, fq.Until)
		paramIdx++
	}

	query += fmt.Sprintf(`
	GROUP BY p.id, u.username
	ORDER BY p.created_at %s
	LIMIT $%d OFFSET $%d`, sortDir, paramIdx, paramIdx+1)

	args = append(args, fq.Limit, fq.Offset)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feeds []*PostWithMetadata
	for rows.Next() {
		var post PostWithMetadata
		if err := rows.Scan(
			&post.ID,
			&post.UserId,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.Username,
			&post.CommentCount,
		); err != nil {
			return nil, err
		}
		feeds = append(feeds, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feeds, nil
}
