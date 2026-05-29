package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/service"
)

func testLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}

// ---- fake services (testify-based) ----

type fakeProjectService struct{ mock.Mock }

var _ projectService = (*fakeProjectService)(nil)

func (m *fakeProjectService) GetAll(ctx context.Context) ([]models.Project, error) {
	args := m.Called(ctx)
	var p []models.Project
	if v := args.Get(0); v != nil {
		p = v.([]models.Project)
	}
	return p, args.Error(1)
}

func (m *fakeProjectService) GetCatalog(ctx context.Context, page, limit int, search string) (*service.CatalogResult, error) {
	args := m.Called(ctx, page, limit, search)
	var r *service.CatalogResult
	if v := args.Get(0); v != nil {
		r = v.(*service.CatalogResult)
	}
	return r, args.Error(1)
}

func (m *fakeProjectService) GetStat(ctx context.Context, id int64) (*service.ProjectStat, error) {
	args := m.Called(ctx, id)
	var s *service.ProjectStat
	if v := args.Get(0); v != nil {
		s = v.(*service.ProjectStat)
	}
	return s, args.Error(1)
}

func (m *fakeProjectService) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *fakeProjectService) Update(ctx context.Context, key string) error {
	return m.Called(ctx, key).Error(0)
}

type fakeIssueService struct{ mock.Mock }

var _ issueService = (*fakeIssueService)(nil)

func (m *fakeIssueService) GetByProjectKey(ctx context.Context, key string) ([]models.Issue, error) {
	args := m.Called(ctx, key)
	var i []models.Issue
	if v := args.Get(0); v != nil {
		i = v.([]models.Issue)
	}
	return i, args.Error(1)
}

type fakeGraphService struct{ mock.Mock }

var _ graphService = (*fakeGraphService)(nil)

func (m *fakeGraphService) Make(ctx context.Context, project string, task int) error {
	return m.Called(ctx, project, task).Error(0)
}

func (m *fakeGraphService) Get(ctx context.Context, project string, task int) (interface{}, error) {
	args := m.Called(ctx, project, task)
	return args.Get(0), args.Error(1)
}

func (m *fakeGraphService) Compare(ctx context.Context, keys []string, task int) (interface{}, error) {
	args := m.Called(ctx, keys, task)
	return args.Get(0), args.Error(1)
}

func (m *fakeGraphService) IsAnalyzed(project string) bool {
	return m.Called(project).Bool(0)
}

func (m *fakeGraphService) IsEmpty(ctx context.Context, project string) (bool, error) {
	args := m.Called(ctx, project)
	return args.Bool(0), args.Error(1)
}

func (m *fakeGraphService) DropAnalyzed(project string) {
	m.Called(project)
}

// ---- shared test helpers ----

// requestWithVars builds a request and injects gorilla/mux path variables so a
// handler can be exercised in isolation, without a full router.
func requestWithVars(method, target string, vars map[string]string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	if len(vars) > 0 {
		req = mux.SetURLVars(req, vars)
	}
	return req
}

func decodeEnvelope(t *testing.T, rr *httptest.ResponseRecorder) envelope {
	t.Helper()
	var e envelope
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &e))
	return e
}
