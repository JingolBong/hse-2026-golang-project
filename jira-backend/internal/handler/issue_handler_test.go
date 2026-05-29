package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/service"
)

func TestIssueHandler_GetByProject(t *testing.T) {
	t.Run("missing project key -> 400", func(t *testing.T) {
		h := NewIssueHandler(&fakeIssueService{})
		rr := httptest.NewRecorder()
		h.GetByProject(rr, httptest.NewRequest(http.MethodGet, "/api/v1/issues", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("project not found -> 404", func(t *testing.T) {
		svc := &fakeIssueService{}
		svc.On("GetByProjectKey", mock.Anything, "ABC").Return(nil, service.ErrProjectNotFound)

		h := NewIssueHandler(svc)
		rr := httptest.NewRecorder()
		h.GetByProject(rr, httptest.NewRequest(http.MethodGet, "/api/v1/issues?project=ABC", nil))
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		svc := &fakeIssueService{}
		svc.On("GetByProjectKey", mock.Anything, "ABC").Return(nil, errors.New("db down"))

		h := NewIssueHandler(svc)
		rr := httptest.NewRecorder()
		h.GetByProject(rr, httptest.NewRequest(http.MethodGet, "/api/v1/issues?project=ABC", nil))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("success (key from path var)", func(t *testing.T) {
		svc := &fakeIssueService{}
		svc.On("GetByProjectKey", mock.Anything, "ABC").
			Return([]models.Issue{{JiraID: 1}, {JiraID: 2}}, nil)

		h := NewIssueHandler(svc)
		rr := httptest.NewRecorder()
		h.GetByProject(rr, requestWithVars(http.MethodGet, "/", map[string]string{"project": "ABC"}))

		require.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, decodeEnvelope(t, rr).Status)
		svc.AssertExpectations(t)
	})
}
