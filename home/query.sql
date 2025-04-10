-- name: CreateNewUser :one
INSERT INTO users (email, username, password_hash, role_id)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = $1 LIMIT 1;

-- name: CreateRecruiterProfile :exec
INSERT INTO recruiter_profile(user_id) VALUES ($1);

-- name: CreateApplicantProfile :exec
INSERT INTO applicant_profile(user_id) VALUES ($1);

-- name: GetRecruiterProfile :one
SELECT * FROM recruiter_profile WHERE user_id = $1 LIMIT 1;

-- name: GetApplicantProfile :one
SELECT * FROM applicant_profile WHERE user_id = $1 LIMIT 1; 

-- name: UpdateUser :exec
UPDATE users SET email = $1, username = $2 WHERE user_id = $3;

-- name: UpdateRecruiterProfile :exec
UPDATE recruiter_profile SET company_name = $1, company_description = $2 WHERE user_id = $3;
