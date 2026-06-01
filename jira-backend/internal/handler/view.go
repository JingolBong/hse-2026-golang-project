package handler

import (
	"net/http"
	"net/url"
	"strconv"

	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/service"
)

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
