-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    email TEXT NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL DEFAULT 'unset',
    is_chirpy_red BOOLEAN DEFAULT FALSE
);

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
DROP TABLE refresh_tokens;
DROP TABLE users;

