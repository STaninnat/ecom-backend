// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: users.sql

package database

import (
	"context"
	"database/sql"
	"time"
)

const checkExistsAndGetIDByEmail = `-- name: CheckExistsAndGetIDByEmail :one
SELECT 
    (id IS NOT NULL)::boolean AS exists, 
    COALESCE(id, '') AS id
FROM users
WHERE email = $1
LIMIT 1
`

type CheckExistsAndGetIDByEmailRow struct {
	Exists bool
	ID     string
}

func (q *Queries) CheckExistsAndGetIDByEmail(ctx context.Context, email string) (CheckExistsAndGetIDByEmailRow, error) {
	row := q.db.QueryRowContext(ctx, checkExistsAndGetIDByEmail, email)
	var i CheckExistsAndGetIDByEmailRow
	err := row.Scan(&i.Exists, &i.ID)
	return i, err
}

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
INSERT INTO users (id, name, email, password, provider, provider_id, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type CreateUserParams struct {
	ID         string
	Name       string
	Email      string
	Password   sql.NullString
	Provider   string
	ProviderID sql.NullString
	Role       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.ExecContext(ctx, createUser,
		arg.ID,
		arg.Name,
		arg.Email,
		arg.Password,
		arg.Provider,
		arg.ProviderID,
		arg.Role,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, name, email, password, provider, provider_id, phone, address, role, created_at, updated_at FROM users
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
		&i.Phone,
		&i.Address,
		&i.Role,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, name, email, password, provider, provider_id, phone, address, role, created_at, updated_at FROM users
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetUserByID(ctx context.Context, id string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Password,
		&i.Provider,
		&i.ProviderID,
		&i.Phone,
		&i.Address,
		&i.Role,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserInfo = `-- name: UpdateUserInfo :exec
UPDATE users
SET  name = $2, email = $3, phone = $4, address = $5, updated_at = $6
WHERE id = $1
`

type UpdateUserInfoParams struct {
	ID        string
	Name      string
	Email     string
	Phone     sql.NullString
	Address   sql.NullString
	UpdatedAt time.Time
}

func (q *Queries) UpdateUserInfo(ctx context.Context, arg UpdateUserInfoParams) error {
	_, err := q.db.ExecContext(ctx, updateUserInfo,
		arg.ID,
		arg.Name,
		arg.Email,
		arg.Phone,
		arg.Address,
		arg.UpdatedAt,
	)
	return err
}

const updateUserRole = `-- name: UpdateUserRole :exec
UPDATE users 
SET role = $2 WHERE id = $1
`

type UpdateUserRoleParams struct {
	ID   string
	Role string
}

func (q *Queries) UpdateUserRole(ctx context.Context, arg UpdateUserRoleParams) error {
	_, err := q.db.ExecContext(ctx, updateUserRole, arg.ID, arg.Role)
	return err
}

const updateUserSigninStatusByEmail = `-- name: UpdateUserSigninStatusByEmail :exec
UPDATE users
SET provider = $2, provider_id = $3, updated_at = $4
WHERE email = $1
`

type UpdateUserSigninStatusByEmailParams struct {
	Email      string
	Provider   string
	ProviderID sql.NullString
	UpdatedAt  time.Time
}

func (q *Queries) UpdateUserSigninStatusByEmail(ctx context.Context, arg UpdateUserSigninStatusByEmailParams) error {
	_, err := q.db.ExecContext(ctx, updateUserSigninStatusByEmail,
		arg.Email,
		arg.Provider,
		arg.ProviderID,
		arg.UpdatedAt,
	)
	return err
}

const updateUserStatusByID = `-- name: UpdateUserStatusByID :exec
UPDATE users
SET provider = $2, updated_at = $3
WHERE id = $1
`

type UpdateUserStatusByIDParams struct {
	ID        string
	Provider  string
	UpdatedAt time.Time
}

func (q *Queries) UpdateUserStatusByID(ctx context.Context, arg UpdateUserStatusByIDParams) error {
	_, err := q.db.ExecContext(ctx, updateUserStatusByID, arg.ID, arg.Provider, arg.UpdatedAt)
	return err
}
