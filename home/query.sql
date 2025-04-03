-- name: CreateNewUser :exec
INSERT INTO users (email, username, password_hash, role_id)
VALUES ($1, $2, $3, $4);

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = $1 LIMIT 1;


