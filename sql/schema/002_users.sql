-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    email TEXT NOT NULL UNIQUE
);

ALTER TABLE users
ADD hashed_password TEXT NOT NULL DEFAULT 'unset';

CREATE TABLE chirps( 
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    body TEXT,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE refresh_tokens(
    token       TEXT PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ,
    revoked_at  TIMESTAMPTZ
);

-- +goose Down
DROP TABLE chirps;
DROP TABLE users;
DROP TABLE refresh_tokens;
