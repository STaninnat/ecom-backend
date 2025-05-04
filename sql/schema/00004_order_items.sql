-- +goose Up
CREATE TABLE
    orders_items (
        id TEXT PRIMARY KEY,
        order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
        product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
        quantity INT NOT NULL,
        price DECIMAL(10, 2) NOT NULL,
        total_price DECIMAL(10, 2) NOT NULL
    );

-- +goose Down
DROP TABLE IF EXISTS orders_items;