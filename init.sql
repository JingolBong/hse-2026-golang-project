DROP USER IF EXISTS pguser;
DROP USER IF EXISTS replicator;

CREATE USER pguser WITH PASSWORD 'pgpwd';
CREATE USER replicator WITH PASSWORD 'postgres' REPLICATION;

CREATE TABLE IF NOT EXISTS project (
    id SERIAL PRIMARY KEY,
    key VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500)
);

CREATE TABLE IF NOT EXISTS author (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS issues (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    key VARCHAR(50) NOT NULL,
    summary TEXT,
    status VARCHAR(100),
    priority VARCHAR(50),
    created_time TIMESTAMP,
    closed_time TIMESTAMP,
    updated_time TIMESTAMP,
    time_spent BIGINT,
    creator_id INTEGER,
    assignee_id INTEGER
);

CREATE TABLE IF NOT EXISTS statuschange (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL,
    old_status VARCHAR(100),
    new_status VARCHAR(100),
    change_time TIMESTAMP
);

ALTER TABLE issues ADD CONSTRAINT fk_issues_project 
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE;

ALTER TABLE issues ADD CONSTRAINT fk_issues_creator 
    FOREIGN KEY (creator_id) REFERENCES author(id);

ALTER TABLE issues ADD CONSTRAINT fk_issues_assignee 
    FOREIGN KEY (assignee_id) REFERENCES author(id);

ALTER TABLE statuschange ADD CONSTRAINT fk_statuschange_issue 
    FOREIGN KEY (issue_id) REFERENCES issues(id) ON DELETE CASCADE;

ALTER TABLE issues ADD CONSTRAINT unique_issue_key 
    UNIQUE (project_id, key);

CREATE INDEX IF NOT EXISTS idx_project_key ON project(key);
CREATE INDEX IF NOT EXISTS idx_author_username ON author(username);
CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_created_time ON issues(created_time);

GRANT ALL PRIVILEGES ON DATABASE testdb TO pguser;
GRANT ALL PRIVILEGES ON SCHEMA public TO pguser;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO pguser;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO pguser;

INSERT INTO project (key, name, url) VALUES 
('TEST', 'Test Project', 'https://jira.example.com/project/TEST')
ON CONFLICT (key) DO NOTHING;