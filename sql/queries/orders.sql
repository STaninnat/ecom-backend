-- name: CreateOrder :one
INSERT INTO orders (
    id, user_id, total_amount, status, payment_method,
    external_payment_id, tracking_number, shipping_address,
    contact_phone, created_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetOrderByUserID :many
SELECT * FROM orders 
WHERE user_id = $1;

-- name: GetOrderByID :one
SELECT * FROM orders 
WHERE id = $1;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = $3
WHERE id = $1;

-- name: ListAllOrders :many
SELECT * FROM orders 
ORDER BY created_at DESC;

-- name: DeleteOrderByID :exec
DELETE FROM orders 
WHERE id = $1;
