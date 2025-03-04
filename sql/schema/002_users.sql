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
-- +goose Down
DROP TABLE chirps;
DROP TABLE users;
