package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/JingolBong/jira-connector/pkg/models"
)

func (s *Storage) InsertStatusChanges(ctx context.Context, changes []models.StatusChange) error {
	if len(changes) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
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
