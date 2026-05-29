package handler

import (
	"context"

	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/service"
)

// These consumer-side interfaces describe exactly what each handler needs from
// the service layer. The concrete *service.XxxService types satisfy them, so
// wiring in main.go is unchanged; tests can pass lightweight fakes instead.

type projectService interface {
	GetAll(ctx context.Context) ([]models.Project, error)
	GetCatalog(ctx context.Context, page, limit int, search string) (*service.CatalogResult, error)
	GetStat(ctx context.Context, id int64) (*service.ProjectStat, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, key string) error
}

type issueService interface {
	GetByProjectKey(ctx context.Context, key string) ([]models.Issue, error)
}

type graphService interface {
	Make(ctx context.Context, project string, task int) error
	Get(ctx context.Context, project string, task int) (interface{}, error)
	Compare(ctx context.Context, keys []string, task int) (interface{}, error)
	IsAnalyzed(project string) bool
	IsEmpty(ctx context.Context, project string) (bool, error)
	DropAnalyzed(project string)
}
