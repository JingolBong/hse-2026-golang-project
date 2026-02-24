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
	err := s.readWithFallback(ctx, func(db *sql.DB) error {
		return db.QueryRowContext(ctx, query, jiraID).
			Scan(&projectFound.JiraID, &projectFound.Key, &projectFound.Name, &projectFound.URL)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get project by jira_id %d: %w", jiraID, err)
	}
	return &projectFound, nil
}

func (s *Storage) GetAllProjects(ctx context.Context) ([]models.Project, error) {
	const query = `
	SELECT jira_id, key, name, url
	FROM project;
	`
	var projects []models.Project
	err := s.readWithFallback(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var project models.Project
			if err := rows.Scan(&project.JiraID, &project.Key, &project.Name, &project.URL); err != nil {
				return err
			}
			projects = append(projects, project)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get all projects: %w", err)
	}

	return projects, nil
}

func (s *Storage) GetAuthorByJiraID(ctx context.Context, jiraID int64) (*models.Author, error) {
	const query = `
        SELECT jira_id, username, email
        FROM author
        WHERE jira_id = $1;`

	var author models.Author
	var email sql.NullString
	err := s.readWithFallback(ctx, func(db *sql.DB) error {
		return db.QueryRowContext(ctx, query, jiraID).
			Scan(&author.JiraID, &author.Username, &email)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get author by jira_id %d: %w", jiraID, err)
	}

	if email.Valid {
		author.Email = email.String
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
	err := s.readWithFallback(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, query, projectJiraID)
		if err != nil {
			return err
		}
		defer rows.Close()

		issues = nil
		for rows.Next() {
			var (
				i          models.Issue
				updatedAt  sql.NullTime
				closedAt   sql.NullTime
				timeSpent  sql.NullInt32
				creatorID  sql.NullInt64
				assigneeID sql.NullInt64
			)
			if err := rows.Scan(
				&i.JiraID, &i.ProjectID, &i.Key,
				&i.Summary, &i.Status, &i.Priority,
				&i.CreatedAt, &updatedAt, &closedAt,
				&timeSpent, &creatorID, &assigneeID,
			); err != nil {
				return err
			}
			if updatedAt.Valid {
				i.UpdatedAt = &updatedAt.Time
			}
			if closedAt.Valid {
				i.ClosedAt = &closedAt.Time
			}
			if timeSpent.Valid {
				i.TimeSpent = &timeSpent.Int32
			}
			if creatorID.Valid {
				i.CreatorID = &creatorID.Int64
			}
			if assigneeID.Valid {
				i.AssigneeID = &assigneeID.Int64
			}
			issues = append(issues, i)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get issues by project_id %d: %w", projectJiraID, err)
	}
	return issues, nil
}

func (s *Storage) GetStatusChangesByIssue(ctx context.Context, issueJiraID int64) ([]models.StatusChange, error) {
	const query = `
		SELECT id, issue_id, old_status, new_status, change_time
		FROM status_change
		WHERE issue_id = $1
		ORDER BY change_time ASC`

	var changes []models.StatusChange
	err := s.readWithFallback(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, query, issueJiraID)
		if err != nil {
			return err
		}
		defer rows.Close()

		changes = nil
		for rows.Next() {
			var (
				sc        models.StatusChange
				oldStatus sql.NullString
				newStatus sql.NullString
			)
			if err := rows.Scan(
				&sc.ID, &sc.IssueID,
				&oldStatus, &newStatus,
				&sc.ChangeTime,
			); err != nil {
				return err
			}
			if oldStatus.Valid {
				sc.OldStatus = &oldStatus.String
			}
			if newStatus.Valid {
				sc.NewStatus = &newStatus.String
			}
			changes = append(changes, sc)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("get status changes by issue_id %d: %w", issueJiraID, err)
	}
	return changes, nil
}
