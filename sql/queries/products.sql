-- name: CreateProduct :exec
INSERT INTO products (id, category_id, name, description, price, stock, image_url, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);


-- name: GetProductByID :one
SELECT * FROM products 
WHERE id = $1;

-- name: GetActiveProductByID :one
SELECT *
FROM products
WHERE id = $1 AND is_active = TRUE;

-- name: GetAllProducts :many
SELECT * FROM products 
ORDER BY updated_at DESC;

-- name: GetAllActiveProducts :many
SELECT *
FROM products
WHERE is_active = TRUE
ORDER BY updated_at DESC;

-- name: UpdateProduct :exec
UPDATE products
SET category_id = $2, name = $3, description = $4, price = $5, stock = $6, image_url = $7, is_active = $8, updated_at = $9
WHERE id = $1;

-- name: UpdateProductImageURL :exec
UPDATE products
SET image_url = $2, updated_at = $3
WHERE id = $1;

-- name: DeleteProductByID :exec
DELETE FROM products 
WHERE id = $1;

-- name: FilterProducts :many
SELECT *
FROM products
WHERE
    (category_id = sqlc.narg('category_id') OR sqlc.narg('category_id') IS NULL) AND
    (is_active = sqlc.narg('is_active') OR sqlc.narg('is_active') IS NULL) AND
    (price >= sqlc.narg('min_price') OR sqlc.narg('min_price') IS NULL) AND
    (price <= sqlc.narg('max_price') OR sqlc.narg('max_price') IS NULL)
ORDER BY created_at DESC;

