-- +goose Up
CREATE TABLE 
    categories (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL UNIQUE,
        description TEXT,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_categories_name ON categories(name);

-- +goose Down
DROP INDEX IF EXISTS idx_categories_name;
DROP TABLE IF EXISTS categories;
