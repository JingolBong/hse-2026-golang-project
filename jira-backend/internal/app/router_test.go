package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"hse-2026-golang-project/jira-backend/internal/handler"
)

func TestCORS_Preflight(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := cors("http://localhost:4200")(next)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil))

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.False(t, nextCalled, "preflight must short-circuit before the handler")
	assert.Equal(t, "http://localhost:4200", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "DELETE")
}

func TestCORS_PassesThrough(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusTeapot)
	})

	handler := cors("http://example.com")(next)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil))

	assert.True(t, nextCalled, "non-preflight request must reach the handler")
	assert.Equal(t, http.StatusTeapot, rr.Code)
	assert.Equal(t, "http://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
}

// Regression: a CORS preflight (OPTIONS) to a method-specific route must be
// answered by the cors middleware (204 + headers), not rejected with 405 by
// mux's method matcher before the middleware runs. The handlers are never
// invoked for a preflight, so nil services are fine here.
func TestRouter_CORSPreflightOnMethodSpecificRoute(t *testing.T) {
	log := logrus.New()
	log.SetOutput(io.Discard)

	router := NewRouter(
		handler.NewProjectHandler(nil, log),
		handler.NewIssueHandler(nil),
		handler.NewGraphHandler(nil),
		handler.NewSystemHandler(),
		"http://localhost:4200",
		log,
	)

	for _, tc := range []struct {
		name   string
		method string
		path   string
	}{
		{"POST updateProject", http.MethodPost, "/api/v1/connector/updateProject?project=AAR"},
		{"DELETE project", http.MethodDelete, "/api/v1/projects/7"},
		{"POST graph make", http.MethodPost, "/api/v1/graph/make/1?project=AAR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, tc.path, nil)
			req.Header.Set("Origin", "http://localhost:4200")
			req.Header.Set("Access-Control-Request-Method", tc.method)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusNoContent, rr.Code, "preflight must not be 405")
			assert.Equal(t, "http://localhost:4200", rr.Header().Get("Access-Control-Allow-Origin"))
			assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), tc.method)
		})
	}
}
