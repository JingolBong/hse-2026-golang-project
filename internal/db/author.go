package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/JingolBong/jira-connector/internal/models"
)

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

func (s *Storage) GetAuthorByJiraID(ctx context.Context, jiraID int64) (*models.Author, error) {
	const query = `
        SELECT jira_id, username, email
        FROM author
        WHERE jira_id = $1;`

	var author models.Author
	err := s.db.QueryRowContext(ctx, query, jiraID).
		Scan(&author.JiraID, &author.Username, &author.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get author by jira_id %d: %w", jiraID, err)
	}

	return &author, nil
}
