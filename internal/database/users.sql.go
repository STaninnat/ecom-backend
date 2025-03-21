// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package database

import (
	"context"
	"database/sql"
	"time"
)

const checkUserExistsByEmail = `-- name: CheckUserExistsByEmail :one
SELECT EXISTS (SELECT email FROM users WHERE email = $1)
`

func (q *Queries) CheckUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	row := q.db.QueryRowContext(ctx, checkUserExistsByEmail, email)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const checkUserExistsByName = `-- name: CheckUserExistsByName :one
SELECT EXISTS (SELECT name FROM users WHERE name = $1)
`

func (q *Queries) CheckUserExistsByName(ctx context.Context, name string) (bool, error) {
	row := q.db.QueryRowContext(ctx, checkUserExistsByName, name)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const createUser = `-- name: CreateUser :exec
INSERT INTO users (id, name, email, password, provider, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

type CreateUserParams struct {
	ID        string
	Name      string
	Email     string
	Password  sql.NullString
	Provider  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.ExecContext(ctx, createUser,
		arg.ID,
		arg.Name,
		arg.Email,
		arg.Password,
		arg.Provider,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, name, email, password, provider, provider_id, created_at, updated_at FROM users
WHERE email = $1
LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Password,
		&i.Provider,
		&i.ProviderID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
