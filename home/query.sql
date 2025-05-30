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
SELECT * FROM applicant_skills WHERE skill_id = $1 AND user_id = $2;

-- name: NewApplicantSkill :exec
INSERT INTO applicant_skills(user_id, skill_id) VALUES($1, $2);

-- name: DeleteApplicantSkill :exec
DELETE FROM applicant_skills WHERE user_id = $1 AND skill_id = $2;

-- name: AddApplicantProfile :exec
UPDATE applicant_profile SET resume_fileName = $1 WHERE user_id = $2;


-- name: GetApplicantResume :one
SELECT resume_fileName FROM applicant_profile WHERE user_id = $1;

-- name: CreateJobPosting :one
INSERT INTO job_posting(user_id) VALUES($1) RETURNING *;

-- name: GetJobPosting :one
SELECT * FROM job_posting WHERE posting_id = $1;

-- name: GetRequiredSkills :many
SELECT skills_req.skill_id, skills_req.posting_id, skills.name
FROM skills_req
JOIN skills
ON skills_req.skill_id = skills.skill_id
WHERE skills_req.posting_id = $1;

-- name: DeleteRequiredSkill :exec
DELETE FROM skills_req WHERE posting_id = $1 AND skill_id = $2;


-- name: CreateSkill_Req :exec
INSERT INTO skills_req(skill_id, posting_id) VALUES($1,$2);

-- name: GetSkill_req :one
SELECT * FROM skills_req WHERE posting_id = $1 AND skill_id = $2;

-- name: UpdateJobPosting :one
UPDATE job_posting SET job_title = $1, job_description = $2 WHERE posting_id = $3 RETURNING *;

-- name: DeleteJobPosting :exec
DELETE FROM job_posting
WHERE job_title IS NULL AND job_description IS NULL;

-- name: GetJobPostings :many
SELECT DISTINCT job_posting.job_title, job_posting.user_id, job_posting.posting_date,job_posting.posting_id, recruiter_profile.company_name
FROM job_posting
JOIN skills_req
ON job_posting.posting_id = skills_req.posting_id
JOIN applicant_skills
ON skills_req.skill_id = applicant_skills.skill_id
JOIN recruiter_profile
ON job_posting.user_id = recruiter_profile.user_id
WHERE job_posting.job_title IS NOT NULL AND job_posting.job_description IS NOT NULL;

-- name: CreateJobPosting_applicants :exec
INSERT INTO jobposting_applicants(posting_id, user_id) VALUES($1, $2);

-- name: GetJobPosting_applicants :many
SELECT DISTINCT job_posting.job_title, job_posting.posting_id, recruiter_profile.company_name
FROM job_posting
JOIN jobposting_applicants
ON jobposting_applicants.posting_id = job_posting.posting_id
JOIN recruiter_profile
ON recruiter_profile.user_id = job_posting.user_id
WHERE jobposting_applicants.user_id = $1;

-- name: GetJobPosting_recruiters :many
SELECT DISTINCT job_posting.job_title, job_posting.posting_id, job_posting.posting_date, recruiter_profile.company_name
FROM job_posting
JOIN recruiter_profile
ON job_posting.user_id = recruiter_profile.user_id
WHERE recruiter_profile.user_id = $1 AND job_posting.job_title IS NOT NULL AND job_posting.job_description IS NOT NULL;

-- name: GetJobApplicants :many
SELECT DISTINCT jobposting_applicants.user_id, jobposting_applicants.posting_id, users.username, users.email, users.user_id, job_posting.job_title
FROM jobposting_applicants
JOIN users
ON jobposting_applicants.user_id = users.user_id
JOIN job_posting
ON job_posting.posting_id = jobposting_applicants.posting_id
WHERE jobposting_applicants.posting_id = $1;

-- name: Interview_exists :one
SELECT EXISTS (
    SELECT 1 FROM interview
    WHERE posting_id = $1 AND user_id = $2
) AS exists;

-- name: CreateInterview :one
INSERT INTO interview(posting_id, user_id, interview_dateTime) VALUES($1, $2, $3) RETURNING *;

-- name: GetInterviews :many
SELECT DISTINCT job_posting.posting_id, job_posting.job_title, recruiter_profile.company_name, interview.accepted, interview.user_id
FROM job_posting
JOIN interview
ON interview.posting_id = job_posting.posting_id
JOIN recruiter_profile
ON recruiter_profile.user_id = job_posting.user_id
WHERE interview.user_id = $1;

-- name: GetInterviewDetails :one
SELECT DISTINCT interview.user_id, interview.posting_id, interview.interview_dateTime, interview.accepted, interview.anotherdate_req,interview.declined_complete,job_posting.job_title, job_posting.job_description, recruiter_profile.company_name, recruiter_profile.company_description, users.username, users.email 
FROM interview
JOIN job_posting
ON job_posting.posting_id = interview.posting_id
JOIN recruiter_profile
ON job_posting.user_id = recruiter_profile.user_id
JOIN users
ON users.user_id = recruiter_profile.user_id
WHERE interview.user_id = $1 and interview.posting_id = $2;

-- name: CheckJobPosting_applicant :one
SELECT EXISTS (
    SELECT 1 FROM jobposting_applicants
    WHERE posting_id = $1 AND user_id = $2
) AS exists;

-- name: GetInterview :one
SELECT * 
FROM interview
WHERE user_id = $1 AND posting_id = $2;

-- name: UpdateInterviewAnotherDateReq :exec
UPDATE
interview
SET anotherdate_req = TRUE, another_dateTime = $1
WHERE interview.user_id = $2 AND posting_id = $3;

-- name: UpdateInterviewDeclineComplete :exec
UPDATE
interview
SET declined_complete = $1, declined_complete = TRUE
WHERE interview.user_id = $2 AND posting_id = $3;

-- name: GetUsersByRoleID :many
SELECT DISTINCT role_master.id, users.user_id, users.username, users.email
FROM role_master
JOIN users
ON role_master.id = users.role_id
WHERE role_master.id = $1; 

-- name: GetRestrictedUsers :many
SELECT DISTINCT users.user_id, users.username, users.email, recruiter_profile.company_name
FROM users
JOIN recruiter_profile
ON users.user_id = recruiter_profile.user_id
WHERE users.role_id = 2 AND recruiter_profile.approved = FALSE;

-- name: ApproveRecruiter :exec
UPDATE recruiter_profile
SET approved = TRUE
WHERE user_id = $1;

-- name: DeclineRecruiter :exec
UPDATE recruiter_profile
SET declined_completely = TRUE
WHERE user_id = $1;

-- name: CheckUserByEmail :one
SELECT EXISTS (
    SELECT 1 FROM users
    WHERE email = $1
) AS exists;

-- name: UpdateUserRoleID :exec
UPDATE users
SET role_id = $1
WHERE user_id = $2;

