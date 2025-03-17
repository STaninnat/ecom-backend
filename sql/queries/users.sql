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

-- name: CreateUserSession :exec
INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetUserBySessionID :one
SELECT user_id, token, expires_at 
FROM refresh_tokens WHERE token = $1 
LIMIT 1;

-- name: DeleteUserSession :exec
DELETE FROM refresh_tokens WHERE token = $1;