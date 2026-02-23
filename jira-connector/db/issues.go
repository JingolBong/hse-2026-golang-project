package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/JingolBong/jira-connector/pkg/models"
)

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
	err := s.db.QueryRowContext(ctx, query,
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

	tx, err := s.db.BeginTx(ctx, nil)
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

func (s *Storage) GetIssuesByProject(ctx context.Context, projectJiraID int64) ([]models.Issue, error) {
	const query = `
	SELECT i.jira_id, i.project_id, i.key, i.summary, i.status, i.priority, i.created_time, i.updated_time, i.closed_time, i.time_spent, i.creator_id, i.assignee_id
	FROM issue i
	WHERE i.project_id = $1
	ORDER BY i.created_time ASC;
	`
	var issues []models.Issue
	rows, err := s.db.QueryContext(ctx, query, projectJiraID)
	if err != nil {
		return nil, fmt.Errorf("get issues by project jira_id %d: %w", projectJiraID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var issue models.Issue
		err := rows.Scan(&issue.JiraID, &issue.ProjectID, &issue.Key, &issue.Summary, &issue.Status, &issue.Priority,
			&issue.CreatedAt, &issue.UpdatedAt, &issue.ClosedAt, &issue.TimeSpent, &issue.CreatorID, &issue.AssigneeID)
		if err != nil {
			return nil, fmt.Errorf("scan issue for project jira_id %d: %w", projectJiraID, err)
		}
		issues = append(issues, issue)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}
