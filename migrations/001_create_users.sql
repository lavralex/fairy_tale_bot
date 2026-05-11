-- +goose Up
CREATE TYPE user_role AS ENUM ('user', 'admin', 'superadmin');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    role user_role NOT NULL DEFAULT 'user',
    is_banned BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE users;
DROP TYPE user_role;