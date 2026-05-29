package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
