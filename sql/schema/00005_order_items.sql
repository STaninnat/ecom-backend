-- +goose Up
CREATE TABLE
    order_items (
        id TEXT PRIMARY KEY,
        order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
        product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
        quantity INT NOT NULL,
        price DECIMAL(10, 2) NOT NULL,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );

CREATE INDEX idx_order_items_order_id ON  order_items(order_id);
CREATE INDEX idx_order_items_product_id ON  order_items(product_id);
CREATE INDEX idx_order_items_quantity ON  order_items(quantity);
CREATE INDEX idx_order_items_price ON  order_items(price);

-- +goose Down
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_quantity;
DROP INDEX IF EXISTS idx_order_items_price;
DROP TABLE IF EXISTS order_items;