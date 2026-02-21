package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JingolBong/jira-connector/pkg/models"
)

func (s *Storage) UpsertProject(ctx context.Context, p models.Project) (int64, error) {
	const query = `
	INSERT INTO project (jira_id, key, name, url)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (jira_id)
	DO UPDATE SET
		key = EXCLUDED.key,
		name = EXCLUDED.name,
		url = EXCLUDED.url
	RETURNING jira_id;
	`

	var id int64
	err := s.db.QueryRowContext(ctx, query,
		p.JiraID, p.Key, p.Name, p.URL,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("upsert project %d: %w", p.JiraID, err)
	}

	return id, nil
}
func (s *Storage) GetProjectByJiraID(ctx context.Context, jiraID int64) (*models.Project, error) {
	const query = `
	SELECT jira_id, key, name, url
	FROM project
	WHERE jira_id = $1;
	`
	var projectFound models.Project
	err := s.db.QueryRowContext(ctx, query, jiraID).Scan(&projectFound.JiraID, &projectFound.Key, &projectFound.Name, &projectFound.URL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get project by jira_id %d: %w", jiraID, err)
	}

	return &projectFound, nil
}
