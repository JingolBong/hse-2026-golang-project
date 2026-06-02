package models

import "time"

type Project struct {
	JiraID int64  `db:"jira_id"`
	Key    string `db:"key"`
	Name   string `db:"name"`
	URL    string `db:"url"`
}

type Author struct {
	JiraID   int64   `db:"jira_id"`
	Username string  `db:"username"`
	Email    *string `db:"email"`
}

type Issue struct {
	JiraID     int64      `db:"jira_id"`
	ProjectID  int64      `db:"project_id"`
	Key        string     `db:"key"`
	Summary    string     `db:"summary"`
	Status     string     `db:"status"`
	Priority   string     `db:"priority"`
	CreatedAt  time.Time  `db:"created_time"`
	UpdatedAt  *time.Time `db:"updated_time"`
	ClosedAt   *time.Time `db:"closed_time"`
	TimeSpent  *int32     `db:"time_spent"`
	CreatorID  *int64     `db:"creator_id"`
	AssigneeID *int64     `db:"assignee_id"`

	// Resolved from the author table on read (GetIssuesByProject); not stored
	// on the issue row itself. Empty when the issue has no creator/assignee.
	CreatorName  string `db:"-"`
	AssigneeName string `db:"-"`
}

type StatusChange struct {
	ID         int64     `db:"id"`
	IssueID    int64     `db:"issue_id"`
	OldStatus  *string   `db:"old_status"`
	NewStatus  *string   `db:"new_status"`
	ChangeTime time.Time `db:"change_time"`
}
