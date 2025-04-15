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

-- name: UpdateApplicantProfile :exec
UPDATE applicant_profile SET first_name = $1, last_name = $2, school = $3, college = $4, age = $5 WHERE user_id = $6;


-- name: SearchSkillsFunc :many
SELECT name, similarity(name, $1) 
AS score
FROM skills
WHERE name % $1
ORDER BY score DESC;

-- name: GetApplicantSkills :many
SELECT applicant_skills.skill_id, applicant_skills.user_id, skills.name
FROM applicant_skills
JOIN skills
ON applicant_skills.skill_id = skills.skill_id
WHERE applicant_skills.user_id = $1;

-- name: GetSkill :one
SELECT * FROM skills WHERE name = $1;

-- name: CreateSkill :one
INSERT INTO skills(name) VALUES($1) RETURNING *;

-- name: GetApplicantSkill :one
SELECT * FROM applicant_skills WHERE skill_id = $1;

-- name: NewApplicantSkill :exec
INSERT INTO applicant_skills(user_id, skill_id) VALUES($1, $2);

-- name: DeleteApplicantSkill :exec
DELETE FROM applicant_skills WHERE user_id = $1 AND skill_id = $2;

-- name: AddApplicantProfile :exec
UPDATE applicant_profile SET resume_fileName = $1 WHERE user_id = $2;




