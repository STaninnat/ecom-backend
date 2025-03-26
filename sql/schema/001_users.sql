-- +goose Up
CREATE TABLE
    users (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT,
        provider TEXT NOT NULL CHECK (provider IN ('local', 'google')),
        provider_id TEXT,
        phone TEXT, 
        address TEXT, 
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );

CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP INDEX IF EXISTS idx_users_name;
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;