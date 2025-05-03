-- +goose Up
CREATE TABLE
    payments (
        id TEXT PRIMARY KEY,
        order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
        payment_method TEXT NOT NULL,
        amount DECIMAL(10, 2) NOT NULL,
        payment_status TEXT NOT NULL,
        payment_date TIMESTAMP NOT NULL
    );

-- +goose Down
DROP TABLE IF EXISTS payments;