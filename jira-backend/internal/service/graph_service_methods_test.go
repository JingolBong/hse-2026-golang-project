package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/models"
)

// These tests cover the routing / guard logic of GraphService methods (task
// validation, project resolution, empty handling, repo orchestration). The
// underlying histogram math is covered by graph_service_test.go.

func sampleIssues(base time.Time) []models.Issue {
	return []models.Issue{
		{JiraID: 1, Status: "Open", Priority: "Major", CreatedAt: base, TimeSpent: ptrI32(3600)},
		{JiraID: 2, Status: "Closed", Priority: "Blocker", CreatedAt: base, ClosedAt: ptrTime(base.AddDate(0, 0, 3))},
	}
}

func TestGraphService_Get_InvalidTask(t *testing.T) {
	svc := NewGraphService(&mockRepo{})
	for _, task := range []int{0, 7, -1} {
		_, err := svc.Get(context.Background(), "KEY", task)
		assert.ErrorIs(t, err, ErrUnsupportedTask)
	}
}

func TestGraphService_Get_NotFound(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByKey", mock.Anything, "KEY").Return(nil, nil)
	repo.On("GetByName", mock.Anything, "KEY").Return(nil, nil)

	svc := NewGraphService(repo)
	_, err := svc.Get(context.Background(), "KEY", 5)

	assert.ErrorIs(t, err, ErrProjectNotFound)
	repo.AssertExpectations(t)
}

func TestGraphService_Get_EmptyIssues(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(1)).Return([]models.Issue{}, nil)

	svc := NewGraphService(repo)
	got, err := svc.Get(context.Background(), "KEY", 5)

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGraphService_Get_Task5_NoStatusChanges(t *testing.T) {
	repo := &mockRepo{}
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(1)).Return(sampleIssues(base), nil)

	svc := NewGraphService(repo)
	got, err := svc.Get(context.Background(), "KEY", 5)

	require.NoError(t, err)
	assert.NotNil(t, got)
	// task 5 (priority histogram) must not touch status changes.
	repo.AssertNotCalled(t, "GetStatusChangesByProject", mock.Anything, mock.Anything)
}

func TestGraphService_Get_Task1_UsesStatusChanges(t *testing.T) {
	repo := &mockRepo{}
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(1)).Return(sampleIssues(base), nil)
	repo.On("GetStatusChangesByProject", mock.Anything, int64(1)).Return([]models.StatusChange{}, nil)

	svc := NewGraphService(repo)
	got, err := svc.Get(context.Background(), "KEY", 1)

	require.NoError(t, err)
	assert.NotNil(t, got)
	repo.AssertExpectations(t)
}

func TestGraphService_Get_AllTasksDispatch(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// task -> whether it reads status changes (tasks 1 and 2 do).
	tasks := map[int]bool{1: true, 2: true, 3: false, 4: false, 5: false, 6: false}

	for task, needsChanges := range tasks {
		t.Run("task", func(t *testing.T) {
			repo := &mockRepo{}
			repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)
			repo.On("GetIssuesByProject", mock.Anything, int64(1)).Return(sampleIssues(base), nil)
			if needsChanges {
				repo.On("GetStatusChangesByProject", mock.Anything, int64(1)).
					Return([]models.StatusChange{}, nil)
			}

			svc := NewGraphService(repo)
			got, err := svc.Get(context.Background(), "KEY", task)

			require.NoError(t, err)
			assert.NotNil(t, got, "task %d should produce a result", task)
			repo.AssertExpectations(t)
		})
	}
}

func TestGraphService_Get_ResolvesByName(t *testing.T) {
	repo := &mockRepo{}
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Not found by key, found by name.
	repo.On("GetByKey", mock.Anything, "My Project").Return(nil, nil)
	repo.On("GetByName", mock.Anything, "My Project").Return(&models.Project{JiraID: 2}, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(2)).Return(sampleIssues(base), nil)

	svc := NewGraphService(repo)
	got, err := svc.Get(context.Background(), "My Project", 5)

	require.NoError(t, err)
	assert.NotNil(t, got)
	repo.AssertExpectations(t)
}

func TestGraphService_Make(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)

	svc := NewGraphService(repo)

	require.NoError(t, svc.Make(context.Background(), "KEY", 3))
	assert.True(t, svc.IsAnalyzed("KEY"))

	svc.DropAnalyzed("KEY")
	assert.False(t, svc.IsAnalyzed("KEY"))
}

func TestGraphService_Make_InvalidTask(t *testing.T) {
	svc := NewGraphService(&mockRepo{})
	err := svc.Make(context.Background(), "KEY", 99)
	assert.ErrorIs(t, err, ErrUnsupportedTask)
	assert.False(t, svc.IsAnalyzed("KEY"))
}

func TestGraphService_Make_NotFound(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByKey", mock.Anything, "KEY").Return(nil, nil)
	repo.On("GetByName", mock.Anything, "KEY").Return(nil, nil)

	svc := NewGraphService(repo)
	err := svc.Make(context.Background(), "KEY", 1)
	assert.ErrorIs(t, err, ErrProjectNotFound)
	assert.False(t, svc.IsAnalyzed("KEY"))
}

func TestGraphService_IsEmpty(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)
		repo.On("GetIssuesByProject", mock.Anything, int64(1)).Return([]models.Issue{}, nil)

		svc := NewGraphService(repo)
		empty, err := svc.IsEmpty(context.Background(), "KEY")
		require.NoError(t, err)
		assert.True(t, empty)
	})

	t.Run("non-empty", func(t *testing.T) {
		repo := &mockRepo{}
		base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)
		repo.On("GetIssuesByProject", mock.Anything, int64(1)).Return(sampleIssues(base), nil)

		svc := NewGraphService(repo)
		empty, err := svc.IsEmpty(context.Background(), "KEY")
		require.NoError(t, err)
		assert.False(t, empty)
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockRepo{}
		repo.On("GetByKey", mock.Anything, "KEY").Return(nil, nil)
		repo.On("GetByName", mock.Anything, "KEY").Return(nil, nil)

		svc := NewGraphService(repo)
		_, err := svc.IsEmpty(context.Background(), "KEY")
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})
}

func TestGraphService_Compare_UnsupportedTask(t *testing.T) {
	svc := NewGraphService(&mockRepo{})
	_, err := svc.Compare(context.Background(), []string{"KEY"}, 2)
	assert.ErrorIs(t, err, ErrUnsupportedTask)
}

func TestGraphService_Compare_EmptyKeys(t *testing.T) {
	svc := NewGraphService(&mockRepo{})
	got, err := svc.Compare(context.Background(), nil, 1)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGraphService_Compare_NoProjectsResolved(t *testing.T) {
	repo := &mockRepo{}
	repo.On("GetByKey", mock.Anything, "GONE").Return(nil, nil)
	repo.On("GetByName", mock.Anything, "GONE").Return(nil, nil)

	svc := NewGraphService(repo)
	got, err := svc.Compare(context.Background(), []string{"GONE"}, 1)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGraphService_Compare_Success(t *testing.T) {
	repo := &mockRepo{}
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	repo.On("GetByKey", mock.Anything, "KEY").Return(&models.Project{JiraID: 1}, nil)
	repo.On("GetIssuesByProject", mock.Anything, int64(1)).Return(sampleIssues(base), nil)
	repo.On("GetStatusChangesByProject", mock.Anything, int64(1)).Return([]models.StatusChange{}, nil)

	svc := NewGraphService(repo)
	got, err := svc.Compare(context.Background(), []string{"KEY"}, 1)

	require.NoError(t, err)
	assert.NotNil(t, got)
	repo.AssertExpectations(t)
}
