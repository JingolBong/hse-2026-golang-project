package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/service"
)

const issueTimeLayout = "2006-01-02 15:04"

// IssueView is the JSON shape the frontend issues table expects. The DB model
// (models.Issue) has no json tags and exposes raw ids/Go field names, so we
// map it here: CreatedAt -> CreatedTime, ClosedAt -> ClosedTime, and the
// author ids are resolved to usernames (Creator/Assignee) on read.
// Type has no column in the schema, so it is always empty for now.
type IssueView struct {
	Key         string `json:"Key"`
	Summary     string `json:"Summary"`
	Status      string `json:"Status"`
	Priority    string `json:"Priority"`
	Type        string `json:"Type"`
	CreatedTime string `json:"CreatedTime"`
	ClosedTime  string `json:"ClosedTime"`
	UpdatedTime string `json:"UpdatedTime"`
	Creator     string `json:"Creator"`
	Assignee    string `json:"Assignee"`
	TimeSpent   *int32 `json:"TimeSpent,omitempty"`
}

func formatIssueTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(issueTimeLayout)
}

func formatIssueTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatIssueTime(*t)
}

func issueFromModel(i models.Issue) IssueView {
	return IssueView{
		Key:         i.Key,
		Summary:     i.Summary,
		Status:      i.Status,
		Priority:    i.Priority,
		CreatedTime: formatIssueTime(i.CreatedAt),
		ClosedTime:  formatIssueTimePtr(i.ClosedAt),
		UpdatedTime: formatIssueTimePtr(i.UpdatedAt),
		Creator:     i.CreatorName,
		Assignee:    i.AssigneeName,
		TimeSpent:   i.TimeSpent,
	}
}

func issuesFromModels(items []models.Issue) []IssueView {
	views := make([]IssueView, 0, len(items))
	for _, i := range items {
		views = append(views, issueFromModel(i))
	}
	return views
}

type ProjectView struct {
	Existence bool            `json:"Existence"`
	Id        int64           `json:"Id,string"`
	Key       string          `json:"Key"`
	Name      string          `json:"Name"`
	Url       string          `json:"Url"`
	Links     map[string]Link `json:"_links"`
}

func projectLinks(r *http.Request, p models.Project, links *LinkBuilder) map[string]Link {
	id := strconv.FormatInt(p.JiraID, 10)
	projectQuery := url.Values{"project": []string{p.Key}}.Encode()

	return map[string]Link{
		"self":   links.RelForRequest(r, "/api/v1/projects/"+id),
		"delete": links.RelForRequest(r, "/api/v1/projects/"+id),
		"update": links.RelForRequest(r, "/api/v1/connector/updateProject?"+projectQuery),
		"issues": links.RelForRequest(r, "/api/v1/issues?"+projectQuery),
	}
}

func projectFromModel(r *http.Request, p models.Project, links *LinkBuilder) ProjectView {
	return ProjectView{
		Existence: true,
		Id:        p.JiraID,
		Key:       p.Key,
		Name:      p.Name,
		Url:       p.URL,
		Links:     projectLinks(r, p, links),
	}
}

func projectsFromModels(r *http.Request, ps []models.Project, links *LinkBuilder) []ProjectView {
	views := make([]ProjectView, 0, len(ps))
	for _, p := range ps {
		views = append(views, projectFromModel(r, p, links))
	}
	return views
}

func projectsFromCatalog(r *http.Request, ps []service.CatalogProject, links *LinkBuilder) []ProjectView {
	views := make([]ProjectView, 0, len(ps))
	for _, p := range ps {
		model := models.Project{
			JiraID: p.JiraID,
			Key:    p.Key,
		}
		views = append(views, ProjectView{
			Existence: p.Existence,
			Id:        p.JiraID,
			Key:       p.Key,
			Name:      p.Name,
			Url:       p.URL,
			Links:     projectLinks(r, model, links),
		})
	}
	return views
}
