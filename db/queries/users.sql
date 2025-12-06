-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserFromValidRefreshToken :one
SELECT *
FROM users
WHERE id IN (
  SELECT user_id
  FROM refresh_tokens
  WHERE token = $1
  AND revoked_at IS NULL
  AND expires_at > NOW()
);

-- name: UpdateUser :one
UPDATE users
SET email = $2,
    hashed_password = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetChirpyRedActive :one
UPDATE users
SET is_chirpy_red = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
