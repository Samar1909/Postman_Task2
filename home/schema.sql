CREATE TABLE IF NOT EXISTS role_master(
    id SERIAL PRIMARY KEY,
    name VARCHAR(20)
);

CREATE TABLE IF NOT EXISTS users(
    user_id SERIAL PRIMARY KEY,
    username VARCHAR(40) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role_id INT,
    FOREIGN KEY(role_id) REFERENCES role_master(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS recruiter_profile(
    user_id INT PRIMARY KEY,
    company_name VARCHAR(50) NOT NULL, 
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
    skill_id INT,
    resume_name TEXT,
    resume_data BYTEA,
    FOREIGN KEY(skill_id) REFERENCES skills(skill_id) ON DELETE SET NULL,
    FOREIGN KEY(user_id) REFERENCES users(user_id) ON DELETE CASCADE
);


