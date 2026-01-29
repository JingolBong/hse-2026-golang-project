DROP USER IF EXISTS pguser;
DROP USER IF EXISTS replicator;
--оставляем для коректного использования в начале работы, для k8s будет не нужно

CREATE USER pguser WITH PASSWORD 'pgpwd';
CREATE USER replicator WITH PASSWORD 'postgres' REPLICATION LOGIN;
--два пользователя для работы бд и для потоковой репликации, user postgres автоматически создается докером

CREATE TABLE project (
    id SERIAL PRIMARY KEY,
    key VARCHAR(50) UNIQUE NOT NULL, --идентификатор из JIRA
    name VARCHAR(255) NOT NULL,
    url TEXT
);

CREATE TABLE author (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255)
);

CREATE TABLE issues (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    key VARCHAR(50) NOT NULL,
    summary TEXT,
    status VARCHAR(100),
    priority VARCHAR(50),
    created_time TIMESTAMP,
    updated_time TIMESTAMP,
    closed_time TIMESTAMP,
    time_spent INTEGER,
    creator_id INTEGER,
    assignee_id INTEGER,
    FOREIGN KEY (project_id) REFERENCES project(id) ON DELETE CASCADE, --если удалили проект удаляем все задачи
    FOREIGN KEY (creator_id) REFERENCES author(id),
    FOREIGN KEY (assignee_id) REFERENCES author(id),
    CONSTRAINT unique_issue_key UNIQUE(project_id, key) 
);

CREATE TABLE status_change (
    id SERIAL PRIMARY KEY,
    issue_id INTEGER NOT NULL,
    old_status VARCHAR(100),
    new_status VARCHAR(100),
    change_time TIMESTAMP,
    FOREIGN KEY (issue_id) REFERENCES issues(id) ON DELETE CASCADE
);

CREATE INDEX idx_project_key ON project(key);

CREATE INDEX idx_author_username ON author(username);

CREATE INDEX idx_issues_project_id ON issues(project_id);
CREATE INDEX idx_issues_status ON issues(status);
CREATE INDEX idx_issues_creation_time ON issues(created_time);
CREATE INDEX idx_issues_deletion_time ON issues(closed_time);

CREATE INDEX idx_statuschange_issue_id ON status_change(issue_id);
CREATE INDEX idx_statuschange_change_time ON status_change(change_time);

GRANT USAGE ON SCHEMA public TO pguser;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO pguser;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO pguser;

GRANT CONNECT ON DATABASE testdb TO replicator;
GRANT USAGE ON SCHEMA public TO replicator;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO replicator;

--Logical replecation test version for CQRS and Logical replication in future
ALTER SYSTEM SET wal_level = logical;
ALTER SYSTEM SET max_replication_slots = 10;
ALTER SYSTEM SET max_logical_replication_workers = 10;
ALTER SYSTEM SET max_worker_processes = 10;

ALTER USER replicator REPLICATION LOGIN;