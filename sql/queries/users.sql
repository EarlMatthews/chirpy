-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email,hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;
-- name: UpgradeChripyRed :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;