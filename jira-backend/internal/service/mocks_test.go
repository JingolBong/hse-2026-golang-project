package service

import (
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	"hse-2026-golang-project/internal/models"
	pb "hse-2026-golang-project/internal/proto/connector"
)

// mockRepo is a testify-based stub that implements all three consumer-side
// repository interfaces declared in this package (projectRepo, issueRepo,
// graphRepo). Tests configure expectations with .On(...).Return(...).
type mockRepo struct {
	mock.Mock
}

// Compile-time assertions: mockRepo must satisfy every repo interface the
// services depend on. If a service grows a new repo method, this breaks here.
var (
	_ projectRepo = (*mockRepo)(nil)
	_ issueRepo   = (*mockRepo)(nil)
	_ graphRepo   = (*mockRepo)(nil)
)

func (m *mockRepo) GetAll(ctx context.Context) ([]models.Project, error) {
	args := m.Called(ctx)
	var projects []models.Project
	if v := args.Get(0); v != nil {
		projects = v.([]models.Project)
	}
	return projects, args.Error(1)
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*models.Project, error) {
	args := m.Called(ctx, id)
	var p *models.Project
	if v := args.Get(0); v != nil {
		p = v.(*models.Project)
	}
	return p, args.Error(1)
}

func (m *mockRepo) GetByKey(ctx context.Context, key string) (*models.Project, error) {
	args := m.Called(ctx, key)
	var p *models.Project
	if v := args.Get(0); v != nil {
		p = v.(*models.Project)
	}
	return p, args.Error(1)
}

func (m *mockRepo) GetByName(ctx context.Context, name string) (*models.Project, error) {
	args := m.Called(ctx, name)
	var p *models.Project
	if v := args.Get(0); v != nil {
		p = v.(*models.Project)
	}
	return p, args.Error(1)
}

func (m *mockRepo) GetIssuesByProject(ctx context.Context, id int64) ([]models.Issue, error) {
	args := m.Called(ctx, id)
	var issues []models.Issue
	if v := args.Get(0); v != nil {
		issues = v.([]models.Issue)
	}
	return issues, args.Error(1)
}

func (m *mockRepo) GetStatusChangesByProject(ctx context.Context, id int64) ([]models.StatusChange, error) {
	args := m.Called(ctx, id)
	var changes []models.StatusChange
	if v := args.Get(0); v != nil {
		changes = v.([]models.StatusChange)
	}
	return changes, args.Error(1)
}

// mockConnectorClient stubs the gRPC ConnectorServiceClient. It embeds the
// interface so that any method we don't explicitly implement (e.g. Health,
// which ProjectService never calls) still satisfies the type; calling such a
// method would panic, which is the desired "unexpected call" signal in tests.
type mockConnectorClient struct {
	pb.ConnectorServiceClient
	mock.Mock
}

var _ pb.ConnectorServiceClient = (*mockConnectorClient)(nil)

func (m *mockConnectorClient) GetProjects(ctx context.Context, in *pb.GetProjectsRequest, opts ...grpc.CallOption) (*pb.GetProjectsResponse, error) {
	args := m.Called(ctx, in)
	var resp *pb.GetProjectsResponse
	if v := args.Get(0); v != nil {
		resp = v.(*pb.GetProjectsResponse)
	}
	return resp, args.Error(1)
}

func (m *mockConnectorClient) UpdateProject(ctx context.Context, in *pb.UpdateProjectRequest, opts ...grpc.CallOption) (*pb.UpdateProjectResponse, error) {
	args := m.Called(ctx, in)
	var resp *pb.UpdateProjectResponse
	if v := args.Get(0); v != nil {
		resp = v.(*pb.UpdateProjectResponse)
	}
	return resp, args.Error(1)
}

func (m *mockConnectorClient) DeleteProject(ctx context.Context, in *pb.DeleteProjectRequest, opts ...grpc.CallOption) (*pb.DeleteProjectResponse, error) {
	args := m.Called(ctx, in)
	var resp *pb.DeleteProjectResponse
	if v := args.Get(0); v != nil {
		resp = v.(*pb.DeleteProjectResponse)
	}
	return resp, args.Error(1)
}
