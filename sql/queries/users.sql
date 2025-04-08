-- name: CreateUser :exec
INSERT INTO users (id, name, email, password, provider, provider_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: CheckExistsAndGetIDByEmail :one
SELECT 
    (id IS NOT NULL)::boolean AS exists, 
    COALESCE(id, '') AS id
FROM users
WHERE email = $1
LIMIT 1;

-- name: CheckUserExistsByName :one
SELECT EXISTS (SELECT name FROM users WHERE name = $1);

-- name: CheckUserExistsByEmail :one
SELECT EXISTS (SELECT email FROM users WHERE email = $1);

-- name: UpdateUserStatusByID :exec
UPDATE users
SET updated_at = $1, provider = $2
WHERE id = $3;

-- name: UpdateUserSigninStatusByEmail :exec
UPDATE users
SET updated_at = $1, provider = $2, Provider_id = $3
WHERE email = $4;