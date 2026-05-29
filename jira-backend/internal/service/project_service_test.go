package service

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/models"
	pb "hse-2026-golang-project/internal/proto/connector"
)

func testLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}

func TestProjectService_GetAll(t *testing.T) {
	repo := &mockRepo{}
	projects := []models.Project{{JiraID: 1, Key: "A"}, {JiraID: 2, Key: "B"}}
	repo.On("GetAll", mock.Anything).Return(projects, nil)

	svc := NewProjectService(repo, &mockConnectorClient{}, testLogger())
	got, err := svc.GetAll(context.Background())

	require.NoError(t, err)
	assert.Equal(t, projects, got)
	repo.AssertExpectations(t)
}

func TestProjectService_GetCatalog_MarksExistence(t *testing.T) {
	repo := &mockRepo{}
	grpcClient := &mockConnectorClient{}

	resp := &pb.GetProjectsResponse{
		Projects: []*pb.ProjectDTO{
			{Key: "KEY1", Name: "First", Url: "u1"},
			{Key: "KEY2", Name: "Second", Url: "u2"},
		},
		PageInfo: &pb.PageInfo{CurrentPage: 2, TotalPages: 5, ProjectsCount: 47},
	}
	grpcClient.On("GetProjects", mock.Anything, mock.Anything).Return(resp, nil)

	// KEY1 already saved locally (so Existence=true, JiraID populated); KEY2 not.
	repo.On("GetAll", mock.Anything).Return([]models.Project{{JiraID: 100, Key: "KEY1"}}, nil)

	svc := NewProjectService(repo, grpcClient, testLogger())
	got, err := svc.GetCatalog(context.Background(), 2, 10, "")

	require.NoError(t, err)
	require.Len(t, got.Projects, 2)

	assert.True(t, got.Projects[0].Existence)
	assert.Equal(t, int64(100), got.Projects[0].JiraID)
	assert.Equal(t, "KEY1", got.Projects[0].Key)

	assert.False(t, got.Projects[1].Existence)
	assert.Equal(t, int64(0), got.Projects[1].JiraID)

	assert.Equal(t, 2, got.CurrentPage)
	assert.Equal(t, 5, got.PageCount)
	assert.Equal(t, 47, got.TotalCount)

	repo.AssertExpectations(t)
	grpcClient.AssertExpectations(t)
}

func TestProjectService_GetCatalog_ConnectorError(t *testing.T) {
	grpcClient := &mockConnectorClient{}
	grpcClient.On("GetProjects", mock.Anything, mock.Anything).Return(nil, errors.New("jira unavailable"))

	svc := NewProjectService(&mockRepo{}, grpcClient, testLogger())
	_, err := svc.GetCatalog(context.Background(), 1, 10, "")

	assert.Error(t, err)
	grpcClient.AssertExpectations(t)
}

func TestProjectService_GetCatalog_RepoError(t *testing.T) {
	repo := &mockRepo{}
	grpcClient := &mockConnectorClient{}
	grpcClient.On("GetProjects", mock.Anything, mock.Anything).
		Return(&pb.GetProjectsResponse{}, nil)
	repo.On("GetAll", mock.Anything).Return(nil, errors.New("db down"))

	svc := NewProjectService(repo, grpcClient, testLogger())
	_, err := svc.GetCatalog(context.Background(), 1, 10, "")

	assert.Error(t, err)
}

func TestProjectService_Delete(t *testing.T) {
	grpcClient := &mockConnectorClient{}
	grpcClient.On("DeleteProject", mock.Anything, mock.Anything).
		Return(&pb.DeleteProjectResponse{Status: "ok"}, nil)

	svc := NewProjectService(&mockRepo{}, grpcClient, testLogger())
	err := svc.Delete(context.Background(), 7)

	require.NoError(t, err)
	grpcClient.AssertExpectations(t)
}

func TestProjectService_Delete_Error(t *testing.T) {
	grpcClient := &mockConnectorClient{}
	grpcClient.On("DeleteProject", mock.Anything, mock.Anything).
		Return(nil, errors.New("not found"))

	svc := NewProjectService(&mockRepo{}, grpcClient, testLogger())
	err := svc.Delete(context.Background(), 7)

	assert.Error(t, err)
}

func TestProjectService_Update(t *testing.T) {
	grpcClient := &mockConnectorClient{}
	grpcClient.On("UpdateProject", mock.Anything, mock.Anything).
		Return(&pb.UpdateProjectResponse{Status: "ok", Project: "ABC"}, nil)

	svc := NewProjectService(&mockRepo{}, grpcClient, testLogger())
	err := svc.Update(context.Background(), "ABC")

	require.NoError(t, err)
	grpcClient.AssertExpectations(t)
}

func TestProjectService_Update_Error(t *testing.T) {
	grpcClient := &mockConnectorClient{}
	grpcClient.On("UpdateProject", mock.Anything, mock.Anything).
		Return(nil, errors.New("connector down"))

	svc := NewProjectService(&mockRepo{}, grpcClient, testLogger())
	err := svc.Update(context.Background(), "ABC")

	assert.Error(t, err)
}

func TestProjectService_GetStat_Aggregates(t *testing.T) {
	repo := &mockRepo{}
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	project := &models.Project{JiraID: 10, Key: "PRJ", Name: "Project"}
	issues := []models.Issue{
		{JiraID: 1, Status: "Open", CreatedAt: base},
		{JiraID: 2, Status: "Closed", CreatedAt: base, ClosedAt: ptrTime(base.AddDate(0, 0, 4))},
		{JiraID: 3, Status: "Closed", CreatedAt: base, ClosedAt: ptrTime(base.AddDate(0, 0, 6))},
		{JiraID: 4, Status: "In Progress", CreatedAt: base},
		{JiraID: 5, Status: "Resolved", CreatedAt: base},
		{JiraID: 6, Status: "Reopened", CreatedAt: base},
	}
	repo.On("GetByID", mock.Anything, int64(10)).Return(project, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(10)).Return(issues, nil)

	svc := NewProjectService(repo, &mockConnectorClient{}, testLogger())
	stat, err := svc.GetStat(context.Background(), 10)

	require.NoError(t, err)
	assert.Equal(t, int64(10), stat.ID)
	assert.Equal(t, "PRJ", stat.Key)
	assert.Equal(t, 6, stat.AllIssuesCount)
	assert.Equal(t, 1, stat.OpenIssuesCount)
	assert.Equal(t, 2, stat.CloseIssuesCount)
	assert.Equal(t, 1, stat.ProgressIssuesCount)
	assert.Equal(t, 1, stat.ResolvedIssuesCount)
	assert.Equal(t, 1, stat.ReopenedIssuesCount)
	assert.Equal(t, 5.0, stat.AverageTime)
	repo.AssertExpectations(t)
}

func TestProjectService_GetStat_NotFound(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByID", mock.Anything, int64(99)).Return(nil, nil)

	svc := NewProjectService(repo, &mockConnectorClient{}, testLogger())
	_, err := svc.GetStat(context.Background(), 99)

	assert.ErrorIs(t, err, ErrProjectNotFound)
	repo.AssertNotCalled(t, "GetIssuesByProject", mock.Anything, mock.Anything)
}

func TestProjectService_GetStat_LookupError(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByID", mock.Anything, int64(10)).Return(nil, errors.New("db down"))

	svc := NewProjectService(repo, &mockConnectorClient{}, testLogger())
	_, err := svc.GetStat(context.Background(), 10)

	assert.Error(t, err)
}

func TestProjectService_GetStat_IssuesError(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByID", mock.Anything, int64(10)).Return(&models.Project{JiraID: 10}, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(10)).Return(nil, errors.New("query failed"))

	svc := NewProjectService(repo, &mockConnectorClient{}, testLogger())
	_, err := svc.GetStat(context.Background(), 10)

	assert.Error(t, err)
}
