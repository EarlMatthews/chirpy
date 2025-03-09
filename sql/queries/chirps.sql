-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: ShowChirpsAll :many
SELECT * FROM chirps;

-- name: ShowOneChirp :one
SELECT * from chirps where id = $1;

-- name: ShowOneChirpByauthor :many
SELECT * from chirps where user_id = $1;