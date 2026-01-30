package db

import (
	"time"
)

type Project struct {
	Id   int    `db:"id"`
	Key  string `db:"key"`
	Name string `db:"name"`
	URL  string `db:"url"`
}

type Author struct {
	Id       int    `db:"id"`
	Username string `db:"username"`
	Email    string `db:"email"`
}

type Issues struct {
	Id           int       `db:"id"`
	Project_id   int       `db:"project_id"`
	Key          string    `db:"key"`
	Summary      string    `db:"summary"`
	Status       string    `db:"status"`
	Priority     string    `db:"priority"`
	Created_time time.Time `db:"created_time"`
	Updated_time time.Time `db:"updated_time"`
	Closed_time  time.Time `db:"closed_time"`
	Time_spent   int       `db:"time_spent"`
	Creator_id   int       `db:"creator_id"`
	Assignee_id  int       `db:"assignee_id"`
}

type Status_change struct {
	Id          int       `db:"id"`
	Issue_id    int       `db:"issue_id"`
	Old_status  string    `db:"old_status"`
	New_status  string    `db:"new_status"`
	Change_time time.Time `db:"change_time"`
}
