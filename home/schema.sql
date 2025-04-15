CREATE TABLE IF NOT EXISTS role_master(
    id SERIAL PRIMARY KEY,
    name VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS users(
    user_id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    password_hash TEXT,
    role_id INT,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY(role_id) REFERENCES role_master(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS recruiter_profile(
    user_id INT PRIMARY KEY,
    company_name VARCHAR(50) UNIQUE,
    company_description TEXT,
    FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS skills(
    skill_id SERIAL PRIMARY KEY,
    name TEXT
);

CREATE TABLE IF NOT EXISTS applicant_profile(
    user_id INT PRIMARY KEY,    
    first_name VARCHAR(40),
    last_name VARCHAR(40),
    resume_fileName TEXT UNIQUE,
    school VARCHAR(60),
    college TEXT,
    age INT,
    FOREIGN KEY(skill_id) REFERENCES skills(skill_id) ON DELETE SET NULL,
    FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS job_posting(
    posting_id SERIAL PRIMARY KEY,
    company_name VARCHAR(50) NOT NULL,
    FOREIGN KEY(company_name) REFERENCES recruiter_profile(company_name) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS skills_req(
    skill_id INT,
    posting_id INT,
    PRIMARY KEY(skill_id, posting_id),
    FOREIGN KEY(skill_id) REFERENCES skills(skill_id) ON DELETE CASCADE,
    FOREIGN KEY(posting_id) REFERENCES job_posting(posting_id) ON DELETE CASCADE
);

CREATE TABLE applicant_skills(
    skill_id INT,
    user_id INT,
    PRIMARY KEY(skill_id, user_id),
    FOREIGN KEY(skill_id) REFERENCES skills(skill_id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE
);


