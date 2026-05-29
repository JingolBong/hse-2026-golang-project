package connector

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hse-2026-golang-project/internal/config"
)

func discardLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}

func newTestClient(baseURL string) *JiraClient {
	return NewJiraClient(config.ProgramSettings{
		JiraURL:      baseURL,
		MinTimeSleep: 1,
		MaxTimeSleep: 4,
	}, discardLogger())
}

func TestGetProjects_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/project", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"1","key":"ABC","name":"Alpha","self":"http://x/1"}]`))
	}))
	defer srv.Close()

	got, err := newTestClient(srv.URL).GetProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "ABC", got[0].Key)
	assert.Equal(t, "Alpha", got[0].Name)
}

func TestFetchIssuesPage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/search", r.URL.Path)
		assert.Equal(t, "project=ABC", r.URL.Query().Get("jql"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"startAt":0,"maxResults":50,"total":2,"issues":[]}`))
	}))
	defer srv.Close()

	got, err := newTestClient(srv.URL).FetchIssuesPage(context.Background(), "ABC", 0, 50)
	require.NoError(t, err)
	assert.Equal(t, 2, got.Total)
}

func TestDoRequest_RetriesThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 on first try
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	_, err := newTestClient(srv.URL).GetProjects(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls), "should retry once after 503")
}

func TestDoRequest_RetriesOn429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	_, err := newTestClient(srv.URL).GetProjects(context.Background())
	require.NoError(t, err)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&calls), int32(2))
}

func TestDoRequest_RetriesExhausted(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError) // always 500
	}))
	defer srv.Close()

	_, err := newTestClient(srv.URL).GetProjects(context.Background())
	require.Error(t, err)
	assert.Greater(t, atomic.LoadInt32(&calls), int32(1), "must have retried before failing")
}

func TestDoRequest_NonRetryableStatus(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound) // 404 is not retried
	}))
	defer srv.Close()

	_, err := newTestClient(srv.URL).GetProjects(context.Background())
	require.Error(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls), "4xx (non-429) must not be retried")
}

func TestDoRequest_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	_, err := newTestClient(srv.URL).GetProjects(context.Background())
	assert.Error(t, err)
}

func TestDoRequest_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before the call

	_, err := newTestClient(srv.URL).GetProjects(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}
