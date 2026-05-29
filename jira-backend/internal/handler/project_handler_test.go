package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/service"
)

func TestProjectHandler_GetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("GetAll", mock.Anything).
			Return([]models.Project{{JiraID: 1, Key: "A", Name: "Alpha"}}, nil)

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.GetAll(rr, httptest.NewRequest(http.MethodGet, "/api/v1/myprojects", nil))

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		env := decodeEnvelope(t, rr)
		assert.True(t, env.Status)
	})

	t.Run("service error -> 500", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("GetAll", mock.Anything).Return(nil, errors.New("db down"))

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.GetAll(rr, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.False(t, decodeEnvelope(t, rr).Status)
	})
}

func TestProjectHandler_GetCatalog_Defaults(t *testing.T) {
	svc := &fakeProjectService{}
	// No query params -> defaults page=1, limit=10, search="".
	svc.On("GetCatalog", mock.Anything, 1, 10, "").
		Return(&service.CatalogResult{CurrentPage: 1, PageCount: 1, TotalCount: 0}, nil)

	h := NewProjectHandler(svc, testLogger())
	rr := httptest.NewRecorder()
	h.GetCatalog(rr, httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil))

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestProjectHandler_GetCatalog_QueryParams(t *testing.T) {
	svc := &fakeProjectService{}
	svc.On("GetCatalog", mock.Anything, 3, 5, "foo").
		Return(&service.CatalogResult{CurrentPage: 3, PageCount: 10, TotalCount: 47}, nil)

	h := NewProjectHandler(svc, testLogger())
	rr := httptest.NewRecorder()
	h.GetCatalog(rr, httptest.NewRequest(http.MethodGet, "/api/v1/projects?page=3&limit=5&search=foo", nil))

	require.Equal(t, http.StatusOK, rr.Code)
	env := decodeEnvelope(t, rr)
	require.NotNil(t, env.PageInfo)
	assert.Equal(t, 3, env.PageInfo.CurrentPage)
	assert.Equal(t, 47, env.PageInfo.ProjectsCount)
	svc.AssertExpectations(t)
}

func TestProjectHandler_GetCatalog_Error(t *testing.T) {
	svc := &fakeProjectService{}
	svc.On("GetCatalog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("connector down"))

	h := NewProjectHandler(svc, testLogger())
	rr := httptest.NewRecorder()
	h.GetCatalog(rr, httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestProjectHandler_Stat(t *testing.T) {
	t.Run("invalid id -> 400", func(t *testing.T) {
		h := NewProjectHandler(&fakeProjectService{}, testLogger())
		rr := httptest.NewRecorder()
		h.Stat(rr, requestWithVars(http.MethodGet, "/", map[string]string{"id": "abc"}))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found -> 404", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("GetStat", mock.Anything, int64(5)).Return(nil, service.ErrProjectNotFound)

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Stat(rr, requestWithVars(http.MethodGet, "/", map[string]string{"id": "5"}))
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("success maps fields", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("GetStat", mock.Anything, int64(10)).Return(&service.ProjectStat{
			ID: 10, Key: "PRJ", Name: "Project",
			AllIssuesCount: 6, OpenIssuesCount: 1, CloseIssuesCount: 2,
			AverageTime: 5.0, AverageIssuesCount: "6.0",
		}, nil)

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Stat(rr, requestWithVars(http.MethodGet, "/", map[string]string{"id": "10"}))

		require.Equal(t, http.StatusOK, rr.Code)
		var resp struct {
			Data   statView `json:"data"`
			Status bool     `json:"status"`
		}
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
		assert.True(t, resp.Status)
		assert.Equal(t, int64(10), resp.Data.Id)
		assert.Equal(t, 6, resp.Data.AllIssuesCount)
		assert.Equal(t, 5.0, resp.Data.AverageTime)
		assert.Equal(t, "6.0", resp.Data.AverageIssuesCount)
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("GetStat", mock.Anything, int64(10)).Return(nil, errors.New("boom"))

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Stat(rr, requestWithVars(http.MethodGet, "/", map[string]string{"id": "10"}))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestProjectHandler_Delete(t *testing.T) {
	t.Run("invalid id -> 400", func(t *testing.T) {
		h := NewProjectHandler(&fakeProjectService{}, testLogger())
		rr := httptest.NewRecorder()
		h.Delete(rr, requestWithVars(http.MethodDelete, "/", map[string]string{"id": "xx"}))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found -> 404", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("Delete", mock.Anything, int64(3)).Return(db.ErrNotFound)

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Delete(rr, requestWithVars(http.MethodDelete, "/", map[string]string{"id": "3"}))
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("Delete", mock.Anything, int64(3)).Return(nil)

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Delete(rr, requestWithVars(http.MethodDelete, "/", map[string]string{"id": "3"}))
		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("Delete", mock.Anything, int64(3)).Return(errors.New("boom"))

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Delete(rr, requestWithVars(http.MethodDelete, "/", map[string]string{"id": "3"}))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestProjectHandler_Update(t *testing.T) {
	t.Run("missing key -> 400", func(t *testing.T) {
		h := NewProjectHandler(&fakeProjectService{}, testLogger())
		rr := httptest.NewRecorder()
		h.Update(rr, httptest.NewRequest(http.MethodPost, "/api/v1/projects//update", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("Update", mock.Anything, "ABC").Return(nil)

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Update(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC"}))

		require.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error -> 500", func(t *testing.T) {
		svc := &fakeProjectService{}
		svc.On("Update", mock.Anything, "ABC").Return(errors.New("connector down"))

		h := NewProjectHandler(svc, testLogger())
		rr := httptest.NewRecorder()
		h.Update(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC"}))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
