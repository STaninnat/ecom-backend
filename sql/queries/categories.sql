-- name: CreateCategory :exec
INSERT INTO categories (id, name, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetAllCategories :many
SELECT * FROM categories ORDER BY name;

-- name: UpdateCategories :exec
UPDATE categories
SET name = $2, description = $3, updated_at = $4
WHERE id = $1;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = $1;