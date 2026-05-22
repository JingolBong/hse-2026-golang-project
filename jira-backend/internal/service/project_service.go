package service

import (
	"context"
	"fmt"

	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/repository"

	pb "hse-2026-golang-project/internal/proto/connector"

	"github.com/sirupsen/logrus"
)

type ProjectService struct {
	repo *repository.ProjectRepository
	grpcClient pb.ConnectorServiceClient
	log *logrus.Logger
}

func NewProjectService(repo *repository.ProjectRepository, client pb.ConnectorServiceClient, log *logrus.Logger) *ProjectService {
	return &ProjectService{
		repo: repo,
		grpcClient: client,
		log: log,
	}
}

func (s *ProjectService) GetAll(ctx context.Context) ([]models.Project, error) {
	return s.repo.GetAll(ctx)
}

type CatalogProject struct {
	Existence bool
	JiraID    int64
	Key       string
	Name      string
	URL       string
}

type CatalogResult struct {
	Projects    []CatalogProject
	CurrentPage int
	PageCount   int
	TotalCount  int
}

func (s *ProjectService) GetCatalog(ctx context.Context, page, limit int, search string) (*CatalogResult, error) {
	resp, err := s.grpcClient.GetProjects(ctx, &pb.GetProjectsRequest{
		Limit:  int32(limit),
		Page:   int32(page),
		Search: search,
	})
	if err != nil {
		return nil, fmt.Errorf("get projects via connector: %w", err)
	}

	saved, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("load saved projects: %w", err)
	}
	idByKey := make(map[string]int64, len(saved))
	for _, p := range saved {
		idByKey[p.Key] = p.JiraID
	}

	projects := make([]CatalogProject, 0, len(resp.GetProjects()))
	for _, dto := range resp.GetProjects() {
		id, exists := idByKey[dto.GetKey()]
		projects = append(projects, CatalogProject{
			Existence: exists,
			JiraID:    id,
			Key:       dto.GetKey(),
			Name:      dto.GetName(),
			URL:       dto.GetUrl(),
		})
	}

	result := &CatalogResult{Projects: projects}
	if info := resp.GetPageInfo(); info != nil {
		result.CurrentPage = int(info.GetCurrentPage())
		result.PageCount = int(info.GetTotalPages())
		result.TotalCount = int(info.GetProjectsCount())
	}

	return result, nil
}



func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	req := &pb.DeleteProjectRequest{ProjectId: id}
	
	_, err := s.grpcClient.DeleteProject(ctx, req)
	if err != nil {
		return fmt.Errorf("Error deleting a project via the connector: %w", err)
	}

	return nil
}

func (s *ProjectService) Update(ctx context.Context, key string) error {
	req := &pb.UpdateProjectRequest{ProjectKey: key}

	_, err := s.grpcClient.UpdateProject(ctx, req)
	if err != nil {
		return fmt.Errorf("update project via connector: %w", err)
	}

	return nil
}