package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/models"
)

func TestIssueService_GetByProjectKey_Success(t *testing.T) {
	repo := &mockRepo{}
	project := &models.Project{JiraID: 42, Key: "ABC"}
	issues := []models.Issue{{JiraID: 1}, {JiraID: 2}}

	repo.On("GetByKey", mock.Anything, "ABC").Return(project, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(42)).Return(issues, nil)

	svc := NewIssueService(repo)
	got, err := svc.GetByProjectKey(context.Background(), "ABC")

	require.NoError(t, err)
	assert.Equal(t, issues, got)
	repo.AssertExpectations(t)
}

func TestIssueService_GetByProjectKey_NotFound(t *testing.T) {
	repo := &mockRepo{}
	// Project resolves to nil -> service must return ErrProjectNotFound and
	// must NOT query issues. GetIssuesByProject is intentionally not mocked,
	// so an unexpected call would panic and fail the test.
	repo.On("GetByKey", mock.Anything, "NOPE").Return(nil, nil)

	svc := NewIssueService(repo)
	got, err := svc.GetByProjectKey(context.Background(), "NOPE")

	assert.ErrorIs(t, err, ErrProjectNotFound)
	assert.Nil(t, got)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "GetIssuesByProject", mock.Anything, mock.Anything)
}

func TestIssueService_GetByProjectKey_LookupError(t *testing.T) {
	repo := &mockRepo{}
	wantErr := errors.New("db down")
	repo.On("GetByKey", mock.Anything, "ABC").Return(nil, wantErr)

	svc := NewIssueService(repo)
	_, err := svc.GetByProjectKey(context.Background(), "ABC")

	assert.ErrorIs(t, err, wantErr)
	repo.AssertExpectations(t)
}

func TestIssueService_GetByProjectKey_IssuesError(t *testing.T) {
	repo := &mockRepo{}
	project := &models.Project{JiraID: 42, Key: "ABC"}
	wantErr := errors.New("query failed")

	repo.On("GetByKey", mock.Anything, "ABC").Return(project, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(42)).Return(nil, wantErr)

	svc := NewIssueService(repo)
	_, err := svc.GetByProjectKey(context.Background(), "ABC")

	assert.ErrorIs(t, err, wantErr)
	repo.AssertExpectations(t)
}
