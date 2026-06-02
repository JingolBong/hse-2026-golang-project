package service

import (
	"context"
	"fmt"

	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/internal/models"

	pb "hse-2026-golang-project/internal/proto/connector"
	"hse-2026-golang-project/jira-backend/internal/reqid"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func withRequestID(ctx context.Context) context.Context {
	if id := reqid.FromContext(ctx); id != "" {
		return metadata.AppendToOutgoingContext(ctx, reqid.MetadataKey, id)
	}
	return ctx
}

func (s *ProjectService) reqLog(ctx context.Context) *logrus.Entry {
	return s.log.WithField("request_id", reqid.FromContext(ctx))
}

type projectRepo interface {
	GetAll(ctx context.Context) ([]models.Project, error)
	GetByID(ctx context.Context, id int64) (*models.Project, error)
	GetIssuesByProject(ctx context.Context, id int64) ([]models.Issue, error)
}

type ProjectService struct {
	repo       projectRepo
	grpcClient pb.ConnectorServiceClient
	log        *logrus.Logger
}

func NewProjectService(repo projectRepo, client pb.ConnectorServiceClient, log *logrus.Logger) *ProjectService {
	return &ProjectService{
		repo:       repo,
		grpcClient: client,
		log:        log,
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
	s.reqLog(ctx).WithFields(logrus.Fields{
		"page":   page,
		"limit":  limit,
		"search": search,
	}).Debug("fetching project catalog via connector")

	resp, err := s.grpcClient.GetProjects(withRequestID(ctx), &pb.GetProjectsRequest{
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
	s.reqLog(ctx).WithField("project_id", id).Debug("deleting project via connector")

	req := &pb.DeleteProjectRequest{ProjectId: id}

	_, err := s.grpcClient.DeleteProject(withRequestID(ctx), req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return db.ErrNotFound
		}
		return fmt.Errorf("error deleting a project via the connector: %w", err)
	}

	return nil
}

func (s *ProjectService) Update(ctx context.Context, key string) error {
	s.reqLog(ctx).WithField("project", key).Debug("updating project via connector")

	req := &pb.UpdateProjectRequest{ProjectKey: key}

	_, err := s.grpcClient.UpdateProject(withRequestID(ctx), req)
	if err != nil {
		return fmt.Errorf("update project via connector: %w", err)
	}

	return nil
}
