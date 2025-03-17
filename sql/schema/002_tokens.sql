-- +goose Up
CREATE TABLE
    refresh_tokens (
        id TEXT NOT NULL,
        user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        token TEXT NOT NULL,
        expires_at TIMESTAMP  NOT NULL,
        created_at TIMESTAMP  NOT NULL,
        updated_at TIMESTAMP  NOT NULL
    );

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP TABLE IF EXISTS refresh_tokens;