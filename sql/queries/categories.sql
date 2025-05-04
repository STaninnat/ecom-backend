-- name: CreateCategory :exec
INSERT INTO categories (id, name, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetAllCategories :many
SELECT * FROM categories ORDER BY name;
