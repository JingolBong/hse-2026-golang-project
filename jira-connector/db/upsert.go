package db

import (
	"context"
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

func (s *Storage) UpsertAuthor(ctx context.Context, a models.Author) (int64, error) {
	const query = `
	INSERT INTO author (jira_id, username, email)
	VALUES ($1, $2, $3)
	ON CONFLICT (jira_id)
	DO UPDATE SET
		username = EXCLUDED.username,
		email = EXCLUDED.email
	RETURNING jira_id;
	`

	var id int64
	err := s.db.QueryRowContext(ctx, query,
		a.JiraID, a.Username, a.Email,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("upsert author %d: %w", a.JiraID, err)
	}

	return id, nil
}
