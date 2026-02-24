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
	JiraID     int64      `db:"jira_id" json:"jiraId"`
	ProjectID  int64      `db:"project_id" json:"projectId"`
	Key        string     `db:"key" json:"key"`
	Summary    string     `db:"summary" json:"summary"`
	Status     string     `db:"status" json:"status"`
	Priority   string     `db:"priority" json:"priority"`
	CreatedAt  time.Time  `db:"created_time" json:"createdTime"`
	UpdatedAt  *time.Time `db:"updated_time" json:"updatedTime,omitempty"`
	ClosedAt   *time.Time `db:"closed_time" json:"closedTime,omitempty"`
	TimeSpent  *int32     `db:"time_spent" json:"timeSpent,omitempty"`
	CreatorID  *int64     `db:"creator_id" json:"creatorId,omitempty"`
	AssigneeID *int64     `db:"assignee_id" json:"assigneeId,omitempty"`
}

type StatusChange struct {
	ID         int64     `db:"id"`
	IssueID    int64     `db:"issue_id"`
	OldStatus  *string   `db:"old_status"`
	NewStatus  *string   `db:"new_status"`
	ChangeTime time.Time `db:"change_time"`
}
