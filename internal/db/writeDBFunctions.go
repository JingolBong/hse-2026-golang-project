package db

import (
	"context"
	"database/sql"
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
	err := s.writeTx(ctx, nil, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query,
			p.JiraID, p.Key, p.Name, p.URL,
		).Scan(&id)
	})
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
	err := s.writeTx(ctx, nil, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query,
			a.JiraID, a.Username, a.Email,
		).Scan(&id)
	})

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
		updated_time = EXCLUDED.updated_time,
		closed_time = EXCLUDED.closed_time,
		time_spent = EXCLUDED.time_spent,
		creator_id = EXCLUDED.creator_id,
		assignee_id = EXCLUDED.assignee_id
	RETURNING jira_id;
	`
	var id int64
	err := s.writeTx(ctx, nil, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, query,
			issue.JiraID, issue.ProjectID, issue.Key, issue.Summary, issue.Status, issue.Priority,
			issue.CreatedAt, issue.UpdatedAt, issue.ClosedAt, issue.TimeSpent, issue.CreatorID, issue.AssigneeID,
		).Scan(&id)
	})
	if err != nil {
		return 0, fmt.Errorf("upsert issue %d: %w", issue.JiraID, err)
	}

	return id, nil
}

func (s *Storage) UpsertIssuesBatch(
	ctx context.Context,
	issues []models.Issue,
) error {

	if len(issues) == 0 {
		return nil
	}

	var (
		args   []interface{}
		values []string
	)
	cols := 12

	return s.writeTx(ctx, nil, func(tx *sql.Tx) error {

		for i, issue := range issues {
			offset := i*cols + 1

			values = append(values,
				fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
					offset, offset+1, offset+2, offset+3,
					offset+4, offset+5, offset+6, offset+7,
					offset+8, offset+9, offset+10, offset+11,
				),
			)

			args = append(args,
				issue.JiraID,
				issue.ProjectID,
				issue.Key,
				issue.Summary,
				issue.Status,
				issue.Priority,
				issue.CreatedAt,
				issue.UpdatedAt,
				issue.ClosedAt,
				issue.TimeSpent,
				issue.CreatorID,
				issue.AssigneeID,
			)
		}

		query := fmt.Sprintf(`
			INSERT INTO issue (
				jira_id,
				project_id,
				key,
				summary,
				status,
				priority,
				created_time,
				updated_time,
				closed_time,
				time_spent,
				creator_id,
				assignee_id
			)
			VALUES %s
			ON CONFLICT (jira_id)
			DO UPDATE SET
				project_id = EXCLUDED.project_id,
				key = EXCLUDED.key,
				summary = EXCLUDED.summary,
				status = EXCLUDED.status,
				priority = EXCLUDED.priority,
				updated_time = EXCLUDED.updated_time,
				closed_time = EXCLUDED.closed_time,
				time_spent = EXCLUDED.time_spent,
				creator_id = EXCLUDED.creator_id,
				assignee_id = EXCLUDED.assignee_id;
		`, strings.Join(values, ","))

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("batch upsert issues: %w", err)
		}

		return nil
	})
}

func (s *Storage) InsertStatusChangesBatch(
	ctx context.Context,
	changes []models.StatusChange,
) error {

	if len(changes) == 0 {
		return nil
	}

	var (
		args   []interface{}
		values []string
	)
	cols := 4

	return s.writeTx(ctx, nil, func(tx *sql.Tx) error {

		for i, sc := range changes {
			offset := i*cols + 1

			values = append(values,
				fmt.Sprintf("($%d,$%d,$%d,$%d)",
					offset, offset+1, offset+2, offset+3,
				),
			)

			args = append(args,
				sc.IssueID,
				sc.OldStatus,
				sc.NewStatus,
				sc.ChangeTime,
			)
		}

		query := fmt.Sprintf(`
			INSERT INTO status_change (
				issue_id,
				old_status,
				new_status,
				change_time
			)
			VALUES %s
			ON CONFLICT (issue_id, change_time, new_status)
			DO NOTHING;
		`, strings.Join(values, ","))

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("batch insert status_changes: %w", err)
		}

		return nil
	})
}
