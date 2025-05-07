-- name: CreatePayment :one
INSERT INTO payments (
  id, order_id, user_id, amount, currency, status, provider, provider_payment_id, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, updated_at;

-- name: GetPaymentByOrderID :one
SELECT * FROM payments
WHERE order_id = $1
ORDER BY updated_at DESC
LIMIT 1;

-- name: GetPaymentsByUserID :many
SELECT * FROM payments
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: GetAllPayments :many
SELECT *
FROM payments
ORDER BY updated_at DESC;

-- name: GetPaymentsByStatus :many
SELECT *
FROM payments
WHERE status = $1
ORDER BY updated_at DESC;

-- name: UpdatePaymentStatus :exec
UPDATE payments
SET status = $2, updated_at = $3
WHERE id = $1;

-- name: UpdatePaymentStatusByProviderPaymentID :exec
UPDATE payments
SET status = $2, updated_at = $3
WHERE provider_payment_id = $1;

-- name: UpdatePaymentStatusByID :exec
UPDATE payments
SET status = $2
WHERE id = $1;