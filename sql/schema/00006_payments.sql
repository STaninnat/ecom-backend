-- +goose Up
CREATE TABLE
    payments (
        id TEXT PRIMARY KEY,
        order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
        user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        amount DECIMAL(10, 2) NOT NULL,
        currency TEXT NOT NULL,
        status TEXT NOT NULL CHECK (
            status IN ('pending', 'succeeded', 'failed', 'cancelled', 'refunded')),
        provider TEXT NOT NULL,
        provider_payment_id TEXT,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE UNIQUE INDEX idx_payments_provider_payment_id ON payments(provider_payment_id);
CREATE INDEX idx_payments_updated_at ON payments(updated_at);

-- +goose Down
DROP INDEX IF EXISTS idx_payments_order_id;
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_provider_payment_id;
DROP INDEX IF EXISTS idx_payments_updated_at;
DROP TABLE IF EXISTS payments;