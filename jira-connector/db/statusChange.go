package db

import (
	"context"

	"github.com/JingolBong/jira-connector/pkg/models"
)

func (s *Storage) InsertStatusChanges(ctx context.Context, changes []models.StatusChange) error {}

func GetStatusChangesByIssue(ctx context.Context, issueJiraID int64) ([]models.StatusChange, error) {
	const query = `
	SELECT sc.id, sc.issue_id, sc.old_status, sc.new_status, sc.change_time
	FROM status_change sc
	WHERE sc.issue_id = $1;
	`
}
