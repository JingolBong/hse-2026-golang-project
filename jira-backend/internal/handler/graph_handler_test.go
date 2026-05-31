package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/jira-backend/internal/service"
)

func TestGraphHandler_Make(t *testing.T) {
	t.Run("missing project -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.Make(rr, httptest.NewRequest(http.MethodPost, "/api/v1/graph", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid task -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "xx"}))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("empty project -> 422", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsEmpty", mock.Anything, "ABC").Return(true, nil)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "1"}))
		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("IsEmpty project not found -> 404", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsEmpty", mock.Anything, "ABC").Return(false, service.ErrProjectNotFound)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "1"}))
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("IsEmpty error -> 500", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsEmpty", mock.Anything, "ABC").Return(false, errors.New("boom"))

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "1"}))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("project not found -> 404", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsEmpty", mock.Anything, "ABC").Return(false, nil)
		svc.On("Make", mock.Anything, "ABC", 1).Return(service.ErrProjectNotFound)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "1"}))
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("unsupported task -> 400", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsEmpty", mock.Anything, "ABC").Return(false, nil)
		svc.On("Make", mock.Anything, "ABC", 9).Return(service.ErrUnsupportedTask)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "9"}))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsEmpty", mock.Anything, "ABC").Return(false, nil)
		svc.On("Make", mock.Anything, "ABC", 1).Return(errors.New("boom"))

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "1"}))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsEmpty", mock.Anything, "ABC").Return(false, nil)
		svc.On("Make", mock.Anything, "ABC", 1).Return(nil)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Make(rr, requestWithVars(http.MethodPost, "/", map[string]string{"project": "ABC", "task": "1"}))
		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestGraphHandler_Get(t *testing.T) {
	t.Run("missing project -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.Get(rr, httptest.NewRequest(http.MethodGet, "/api/v1/graph?task=1", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid task -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.Get(rr, httptest.NewRequest(http.MethodGet, "/api/v1/graph?project=ABC", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("not found -> 404", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("Get", mock.Anything, "ABC", 3).Return(nil, service.ErrProjectNotFound)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Get(rr, httptest.NewRequest(http.MethodGet, "/api/v1/graph?project=ABC&task=3", nil))
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("unsupported task -> 400", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("Get", mock.Anything, "ABC", 8).Return(nil, service.ErrUnsupportedTask)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Get(rr, httptest.NewRequest(http.MethodGet, "/api/v1/graph?project=ABC&task=8", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("Get", mock.Anything, "ABC", 3).Return(nil, errors.New("boom"))

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Get(rr, httptest.NewRequest(http.MethodGet, "/api/v1/graph?project=ABC&task=3", nil))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("Get", mock.Anything, "ABC", 3).Return(map[string]int{"x": 1}, nil)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Get(rr, httptest.NewRequest(http.MethodGet, "/api/v1/graph?project=ABC&task=3", nil))
		require.Equal(t, http.StatusOK, rr.Code)
		assert.True(t, decodeEnvelope(t, rr).Status)
		svc.AssertExpectations(t)
	})
}

func TestGraphHandler_Compare(t *testing.T) {
	t.Run("invalid task -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.Compare(rr, requestWithVars(http.MethodGet, "/api/v1/compare?project=A", map[string]string{"task": "xx"}))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("no projects -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.Compare(rr, requestWithVars(http.MethodGet, "/api/v1/compare", map[string]string{"task": "1"}))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("unsupported task -> 400", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("Compare", mock.Anything, []string{"A", "B"}, 2).Return(nil, service.ErrUnsupportedTask)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Compare(rr, requestWithVars(http.MethodGet, "/api/v1/compare?project=A,B", map[string]string{"task": "2"}))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("internal error -> 500", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("Compare", mock.Anything, []string{"A"}, 1).Return(nil, errors.New("boom"))

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Compare(rr, requestWithVars(http.MethodGet, "/api/v1/compare?project=A", map[string]string{"task": "1"}))
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("Compare", mock.Anything, []string{"A", "B"}, 1).Return(map[string]int{"x": 1}, nil)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.Compare(rr, requestWithVars(http.MethodGet, "/api/v1/compare?project=A,B", map[string]string{"task": "1"}))
		require.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestGraphHandler_IsAnalyzed(t *testing.T) {
	t.Run("missing project -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.IsAnalyzed(rr, httptest.NewRequest(http.MethodGet, "/api/v1/isAnalyzed", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("IsAnalyzed", "ABC").Return(true)

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.IsAnalyzed(rr, httptest.NewRequest(http.MethodGet, "/api/v1/isAnalyzed?project=ABC", nil))

		require.Equal(t, http.StatusOK, rr.Code)
		env := decodeEnvelope(t, rr)
		data, ok := env.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, true, data["isAnalyzed"])
		svc.AssertExpectations(t)
	})
}

func TestGraphHandler_DeleteGraphs(t *testing.T) {
	t.Run("missing project -> 400", func(t *testing.T) {
		h := NewGraphHandler(&fakeGraphService{})
		rr := httptest.NewRecorder()
		h.DeleteGraphs(rr, httptest.NewRequest(http.MethodDelete, "/api/v1/graphs", nil))
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("success drops analyzed", func(t *testing.T) {
		svc := &fakeGraphService{}
		svc.On("DropAnalyzed", "ABC").Return()

		h := NewGraphHandler(svc)
		rr := httptest.NewRecorder()
		h.DeleteGraphs(rr, httptest.NewRequest(http.MethodDelete, "/api/v1/graphs?project=ABC", nil))

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})
}
