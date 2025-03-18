-- +goose Up
CREATE TABLE
    products (
        id TEXT PRIMARY KEY,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL,
        name TEXT NOT NULL,
        description TEXT,
        price DECIMAL(10,2) NOT NULL,
        stock INT NOT NULL,
        image_url TEXT
    );

-- +goose Down
DROP TABLE IF EXISTS products;