package models

import "time"

type ProjectJSON struct {
	JiraID int64  `json:"jira_id"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	URL    string `json:"url"`
}

type AuthorJSON struct {
	JiraID   int64   `json:"jira_id"`
	Username string  `json:"username"`
	Email    *string `json:"email,omitempty"`
}

type IssueJSON struct {
	JiraID     int64      `json:"jira_id"`
	ProjectID  int64      `json:"project_id"`
	Key        string     `json:"key"`
	Summary    string     `json:"summary"`
	Status     string     `json:"status"`
	Priority   string     `json:"priority"`
	CreatedAt  time.Time  `json:"created_time"`
	UpdatedAt  *time.Time `json:"updated_time,omitempty"`
	ClosedAt   *time.Time `json:"closed_time,omitempty"`
	TimeSpent  *int32     `json:"time_spent,omitempty"`
	CreatorID  *int64     `json:"creator_id,omitempty"`
	AssigneeID *int64     `json:"assignee_id,omitempty"`
}

type StatusChangeJSON struct {
	ID         int64     `json:"id"`
	IssueID    int64     `json:"issue_id"`
	OldStatus  *string   `json:"old_status,omitempty"`
	NewStatus  *string   `json:"new_status,omitempty"`
	ChangeTime time.Time `json:"change_time"`
}
