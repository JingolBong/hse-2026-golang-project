package db

import (
	"context"
	"fmt"
	"hse-2026-golang-project/internal/models"
	"strings"
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
	err := s.writeDB.QueryRowContext(ctx, query,
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
	err := s.writeDB.QueryRowContext(ctx, query,
		a.JiraID, a.Username, a.Email,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("upsert author %d: %w", a.JiraID, err)
	}

	return id, nil
}

func (s *Storage) UpsertIssue(ctx context.Context, issue models.Issue) (int64, error) {
	const query = `
	INSERT INTO issue (jira_id, project_id, key, summary, status, priority, created_time, updated_time, closed_time, time_spent, creator_id, assignee_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT (jira_id)
	DO UPDATE SET
		project_id = EXCLUDED.project_id,
		key = EXCLUDED.key,
		summary = EXCLUDED.summary,
		status = EXCLUDED.status,
		priority = EXCLUDED.priority,
		created_time = EXCLUDED.created_time,
		updated_time = EXCLUDED.updated_time,
		closed_time = EXCLUDED.closed_time,
		time_spent = EXCLUDED.time_spent,
		creator_id = EXCLUDED.creator_id,
		assignee_id = EXCLUDED.assignee_id
	RETURNING jira_id;
	`
	var id int64
	err := s.writeDB.QueryRowContext(ctx, query,
		issue.JiraID, issue.ProjectID, issue.Key, issue.Summary, issue.Status, issue.Priority,
		issue.CreatedAt, issue.UpdatedAt, issue.ClosedAt, issue.TimeSpent, issue.CreatorID, issue.AssigneeID,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("upsert issue %d: %w", issue.JiraID, err)
	}

	return id, nil
}

func (s *Storage) UpsertIssuesBatch(ctx context.Context, issues []models.Issue) error {
	if len(issues) == 0 {
		return nil
	}

	tx, err := s.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	const fieldsPerIssue = 12
	valueStrings := make([]string, 0, len(issues))
	valueArgs := make([]interface{}, 0, len(issues)*fieldsPerIssue)

	for _, issue := range issues {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4,
			len(valueArgs)+5, len(valueArgs)+6, len(valueArgs)+7, len(valueArgs)+8,
			len(valueArgs)+9, len(valueArgs)+10, len(valueArgs)+11, len(valueArgs)+12,
		))
		valueArgs = append(valueArgs,
			issue.JiraID, issue.ProjectID, issue.Key, issue.Summary, issue.Status, issue.Priority,
			issue.CreatedAt, issue.UpdatedAt, issue.ClosedAt, issue.TimeSpent, issue.CreatorID, issue.AssigneeID,
		)
	}

	query := fmt.Sprintf(`
	INSERT INTO ISSUE (jira_id, project_id, key, summary, status, priority, created_time, updated_time, closed_time, time_spent, creator_id, assignee_id)
	VALUES %s
	ON CONFLICT (jira_id)
	DO UPDATE SET
		project_id = EXCLUDED.project_id,
		key = EXCLUDED.key,
		summary = EXCLUDED.summary,
		status = EXCLUDED.status,
		priority = EXCLUDED.priority,
		created_time = EXCLUDED.created_time,
		updated_time = EXCLUDED.updated_time,
		closed_time = EXCLUDED.closed_time,
		time_spent = EXCLUDED.time_spent,
		creator_id = EXCLUDED.creator_id,
		assignee_id = EXCLUDED.assignee_id;
	`, strings.Join(valueStrings, ", "))

	if _, err := tx.ExecContext(ctx, query, valueArgs...); err != nil {
		return fmt.Errorf("upsert issues batch (%d): %w", len(issues), err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) InsertStatusChanges(ctx context.Context, changes []models.StatusChange) error {
	if len(changes) == 0 {
		return nil
	}

	tx, err := s.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	const fieldsPerChange = 4
	valueStrings := make([]string, 0, len(changes))
	valueArgs := make([]interface{}, 0, len(changes)*fieldsPerChange)

	for _, change := range changes {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", len(valueArgs)+1, len(valueArgs)+2, len(valueArgs)+3, len(valueArgs)+4))
		valueArgs = append(valueArgs, change.IssueID, change.OldStatus, change.NewStatus, change.ChangeTime)
	}

	query := fmt.Sprintf(`
	INSERT INTO status_change (issue_id, old_status, new_status, change_time)
	VALUES %s;`, strings.Join(valueStrings, ", "))

	if _, err := tx.ExecContext(ctx, query, valueArgs...); err != nil {
		return fmt.Errorf("insert status changes batch (%d): %w", len(changes), err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
