-- +goose Up
CREATE TABLE IF NOT EXISTS job_applications (
    id SERIAL PRIMARY KEY,
    job_title TEXT NOT NULL,
    company TEXT,
    job_link TEXT,
    applied_date DATE,
    result_date DATE,
    status TEXT,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS coding_problems (
    id SERIAL PRIMARY KEY,
    leetcode_number INT,
    title TEXT,
    pattern TEXT,
    problem_link TEXT,
    difficulty TEXT,
    already_solved BOOLEAN DEFAULT false,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    repo_url TEXT,
    active BOOLEAN DEFAULT false,
    tech_stack TEXT[],
    summary TEXT
);

CREATE TABLE IF NOT EXISTS networking_contacts (
    id SERIAL PRIMARY KEY,
    person_name TEXT NOT NULL,
    how_met TEXT,
    linkedin_connected BOOLEAN DEFAULT false,
    company TEXT,
    position TEXT,
    notes TEXT
);

CREATE TABLE IF NOT EXISTS daily_goals (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    target_date DATE NOT NULL,
    completed BOOLEAN DEFAULT false,
    job_application_id INT REFERENCES job_applications(id) ON DELETE SET NULL,
    coding_problem_id INT REFERENCES coding_problems(id) ON DELETE SET NULL,
    project_id INT REFERENCES projects(id) ON DELETE SET NULL,
    contact_id INT REFERENCES networking_contacts(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS weekly_goals (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    week_of DATE NOT NULL,
    completed BOOLEAN DEFAULT false,
    job_application_id INT REFERENCES job_applications(id) ON DELETE SET NULL,
    coding_problem_id INT REFERENCES coding_problems(id) ON DELETE SET NULL,
    project_id INT REFERENCES projects(id) ON DELETE SET NULL,
    contact_id INT REFERENCES networking_contacts(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS monthly_goals (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    month_of DATE NOT NULL,
    completed BOOLEAN DEFAULT false,
    job_application_id INT REFERENCES job_applications(id) ON DELETE SET NULL,
    coding_problem_id INT REFERENCES coding_problems(id) ON DELETE SET NULL,
    project_id INT REFERENCES projects(id) ON DELETE SET NULL,
    contact_id INT REFERENCES networking_contacts(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS meetings (
    id SERIAL PRIMARY KEY,
    session_name TEXT NOT NULL,
    session_type TEXT CHECK (session_type IN ('virtual','in-person')) NOT NULL,
    session_time TIMESTAMPTZ,
    location TEXT,
    organizer TEXT,
    company TEXT,
    notes TEXT
);

-- +goose Down
DROP TABLE IF EXISTS meetings;
DROP TABLE IF EXISTS monthly_goals;
DROP TABLE IF EXISTS weekly_goals;
DROP TABLE IF EXISTS daily_goals;
DROP TABLE IF EXISTS networking_contacts;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS coding_problems;
DROP TABLE IF EXISTS job_applications;
