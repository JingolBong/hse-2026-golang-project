package connector

type JiraProject struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name:`
	Self string `json:"self"` //url
}

type JiraIssues struct {
	StartAt    int         `json:"startAt"`
	MaxResults int         `json:"maxResults"`
	Total      int         `json:"total"`
	Issues     []JiraIssue `json:"issues"`
}

type JiraIssue struct {
	ID        string      `json:"id"`
	Key       string      `json:"key"`
	Fields    IssueFields `json:"fields"`
	ChangeLog ChangeLog   `json:"changelog"`
}

type IssueFields struct {
	Summary        string   `json:"summary"`
	Status         Status   `json:"status"`
	Priority       Priority `json:"priority"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
	ResolutionDate string   `json:"resolutiondate"`
	TimeSpent      int      `json:"timespent"`
	Creator        User     `json:"creator"`
	Assignee       User     `json:"assignee"`
}

type Status struct {
	Name string `json:"name"`
}

type Priority struct {
	Name string `json:"name"`
}

type User struct {
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

type ChangeLog struct {
	Histories []History `json:"histories"`
}

type History struct {
	Сreated string `json:"created"`
	Items   []Item `json:"items"`
}

type HistoryItem struct {
	Field      string `json:"field"`
	FromString string `json:"fromString"`
	ToString   string `json:"toString"`
}
