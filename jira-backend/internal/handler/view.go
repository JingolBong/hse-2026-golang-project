package handler

import (
	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/service"
)

type ProjectView struct {
	Existence bool   `json:"Existence"`
	Id        int64  `json:"Id"`
	Key       string `json:"Key"`
	Name      string `json:"Name"`
	Url       string `json:"Url"`
}

func projectFromModel(p models.Project) ProjectView {
	return ProjectView{
		Existence: true,
		Id:        p.JiraID,
		Key:       p.Key,
		Name:      p.Name,
		Url:       p.URL,
	}
}

func projectsFromModels(ps []models.Project) []ProjectView {
	views := make([]ProjectView, 0, len(ps))
	for _, p := range ps {
		views = append(views, projectFromModel(p))
	}
	return views
}

func projectsFromCatalog(ps []service.CatalogProject) []ProjectView {
	views := make([]ProjectView, 0, len(ps))
	for _, p := range ps {
		views = append(views, ProjectView{
			Existence: p.Existence,
			Id:        p.JiraID,
			Key:       p.Key,
			Name:      p.Name,
			Url:       p.URL,
		})
	}
	return views
}
