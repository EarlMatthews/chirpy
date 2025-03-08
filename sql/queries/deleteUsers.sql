-- name: DeleteUser :exec
DELETE FROM users;
-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1;