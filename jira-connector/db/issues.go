package db

import (
	"context"
	"fmt"

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

}

func GetIssuesByProject(ctx context.Context, projectJiraID int64) ([]models.Issue, error) {
	const query = `
	SELECT i.jira_id, i.project_id, i.key, i.summary, i.status, i.priority, i.created_time, i.updated_time, i.closed_time, i.time_spent, i.creator_id, i.assignee_id
	FROM issue i
	WHERE i.project_id = $1;
	`
}
