DROP ROLE IF EXISTS pguser;
DROP ROLE IF EXISTS replicator;

CREATE ROLE pguser WITH LOGIN PASSWORD 'pgpwd';
CREATE ROLE replicator WITH LOGIN PASSWORD 'postgres' REPLICATION;

CREATE TABLE project (
    jira_id BIGINT PRIMARY KEY,
    key VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    url TEXT
);

CREATE TABLE author (
    jira_id BIGINT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255)
);

CREATE TABLE issues (
    jira_id BIGINT PRIMARY KEY,
    project_id BIGINT NOT NULL,
    key VARCHAR(50) NOT NULL,
    summary TEXT,
    status VARCHAR(100),
    priority VARCHAR(50),
    created_time TIMESTAMP NOT NULL,
    updated_time TIMESTAMP,
    closed_time TIMESTAMP NULL,
    time_spent INTEGER,
    creator_id BIGINT,
    assignee_id BIGINT,

    FOREIGN KEY (project_id) REFERENCES project(jira_id) ON DELETE CASCADE,
    FOREIGN KEY (creator_id) REFERENCES author(jira_id),
    FOREIGN KEY (assignee_id) REFERENCES author(jira_id),

    CONSTRAINT unique_issue_key UNIQUE(project_id, key)
);

CREATE TABLE status_change (
    id BIGSERIAL PRIMARY KEY,
    issue_id BIGINT NOT NULL,
    old_status VARCHAR(100),
    new_status VARCHAR(100),
    change_time TIMESTAMP NOT NULL,

    FOREIGN KEY (issue_id) REFERENCES issues(jira_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_project_key ON project(key);
CREATE INDEX IF NOT EXISTS idx_author_username ON author(username);
CREATE INDEX IF NOT EXISTS idx_issues_project_id ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
CREATE INDEX IF NOT EXISTS idx_issues_priority ON issues(priority);
CREATE INDEX IF NOT EXISTS idx_issues_created_time ON issues(created_time);
CREATE INDEX IF NOT EXISTS idx_issues_closed_time ON issues(closed_time);
CREATE INDEX IF NOT EXISTS idx_statuschange_issue_id ON status_change(issue_id);
CREATE INDEX IF NOT EXISTS idx_statuschange_change_time ON status_change(change_time);

GRANT USAGE ON SCHEMA public TO pguser;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO pguser;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO pguser;

GRANT CONNECT ON DATABASE testdb TO replicator;
GRANT USAGE ON SCHEMA public TO replicator;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO replicator;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO pguser;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT ON SEQUENCES TO pguser;

--Logical replecation test version for CQRS and Logical replication in future
--ALTER SYSTEM SET wal_level = logical;
--ALTER SYSTEM SET max_replication_slots = 10;
--ALTER SYSTEM SET max_logical_replication_workers = 10;
--ALTER SYSTEM SET max_worker_processes = 10;
--TODO
--ALTER USER replicator REPLICATION LOGIN;