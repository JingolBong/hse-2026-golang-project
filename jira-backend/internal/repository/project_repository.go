package repository

import (
	"context"

	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/internal/models"
)

type ProjectRepository struct {
	storage *db.Storage
}

func NewProjectRepository(storage *db.Storage) *ProjectRepository {
	return &ProjectRepository{storage: storage}
}

func (r *ProjectRepository) GetAll(ctx context.Context) ([]models.Project, error) {
	return r.storage.GetAllProjects(ctx)
}

func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*models.Project, error) {
	return r.storage.GetProjectByJiraID(ctx, id)
}

func (r *ProjectRepository) GetByKey(ctx context.Context, key string) (*models.Project, error) {
	return r.storage.GetProjectByKey(ctx, key)
}

func (r *ProjectRepository) GetByName(ctx context.Context, name string) (*models.Project, error) {
	return r.storage.GetProjectByName(ctx, name)
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	return r.storage.DeleteProject(ctx, id)
}

func (r *ProjectRepository) GetIssuesByProject(ctx context.Context, id int64) ([]models.Issue, error) {
	return r.storage.GetIssuesByProject(ctx, id)
}

func (r *ProjectRepository) GetStatusChangesByProject(ctx context.Context, id int64) ([]models.StatusChange, error) {
	return r.storage.GetStatusChangesByProject(ctx, id)
}
