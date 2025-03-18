-- name: CreateUser :exec
INSERT INTO users (id, name, email, password, provider, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: CheckUserExistsByName :one
SELECT EXISTS (SELECT name FROM users WHERE name = $1);

-- name: CheckUserExistsByEmail :one
SELECT EXISTS (SELECT email FROM users WHERE email = $1);