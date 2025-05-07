-- name: CreateOrderItem :exec
INSERT INTO order_items (
    id, order_id, product_id, quantity, price,
    created_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7
);

-- name: GetOrderItemsByOrderID :many
SELECT * FROM order_items 
WHERE order_id = $1;