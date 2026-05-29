package service

import (
	"context"
	"errors"

	"hse-2026-golang-project/internal/models"
)

var ErrProjectNotFound = errors.New("project not found")

type issueRepo interface {
	GetByKey(ctx context.Context, key string) (*models.Project, error)
	GetIssuesByProject(ctx context.Context, id int64) ([]models.Issue, error)
}

type IssueService struct {
	repo issueRepo
}

func NewIssueService(repo issueRepo) *IssueService {
	return &IssueService{repo: repo}
}

func (s *IssueService) GetByProjectKey(ctx context.Context, key string) ([]models.Issue, error) {
	project, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	return s.repo.GetIssuesByProject(ctx, project.JiraID)
}
