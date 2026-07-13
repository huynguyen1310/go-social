package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int64
	Name        string
	Level       int64
	Description string
}

type RoleStore struct {
	db *sql.DB
}

func (rs *RoleStore) GetByName(ctx context.Context, name string) (*Role, error) {
	query := `SELECT id, name, level, description FROM roles WHERE name = $1`
	var role Role
	err := rs.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Level, &role.Description)
	if err != nil {
		return nil, err
	}

	return &role, nil
}
