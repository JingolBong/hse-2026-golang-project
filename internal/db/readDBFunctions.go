package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hse-2026-golang-project/internal/models"
)

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

func (s *Storage) GetStatusChangesByIssue(ctx context.Context, issueJiraID int64) ([]models.StatusChange, error) {
	const query = `
	SELECT sc.id, sc.issue_id, sc.old_status, sc.new_status, sc.change_time
	FROM status_change sc
	WHERE sc.issue_id = $1
	ORDER BY change_time ASC;
	`
	var changes []models.StatusChange
	rows, err := s.db.QueryContext(ctx, query, issueJiraID)
	if err != nil {
		return nil, fmt.Errorf("query status changes by issue id %d: %w", issueJiraID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var statusChange models.StatusChange
		if err := rows.Scan(&statusChange.ID, &statusChange.IssueID, &statusChange.OldStatus, &statusChange.NewStatus, &statusChange.ChangeTime); err != nil {
			return nil, fmt.Errorf("scan status change: %w", err)
		}
		changes = append(changes, statusChange)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return changes, nil
}
