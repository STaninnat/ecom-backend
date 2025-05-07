-- +goose Up
CREATE TABLE
    orders (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        total_amount DECIMAL(10, 2) NOT NULL,
        status TEXT NOT NULL CHECK (
            status IN ('pending', 'paid', 'shipped', 'delivered', 
                'cancelled', 'refunded')),
        payment_method TEXT,
        external_payment_id TEXT,
        tracking_number TEXT,
        shipping_address TEXT,
        contact_phone TEXT,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_total_amount ON orders(total_amount);
CREATE INDEX idx_orders_status ON orders(status);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_total_amount;
DROP INDEX IF EXISTS idx_orders_status;
DROP TABLE IF EXISTS orders;