package connector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"hse-2026-golang-project/internal/config"
	pb "hse-2026-golang-project/internal/proto/connector"
)

const fiveProjectsJSON = `[
	{"id":"1","key":"ABC","name":"Alpha","self":"http://x/1"},
	{"id":"2","key":"BCD","name":"Beta","self":"http://x/2"},
	{"id":"3","key":"CDE","name":"Gamma","self":"http://x/3"},
	{"id":"4","key":"DEF","name":"Delta","self":"http://x/4"},
	{"id":"5","key":"EFG","name":"Epsilon","self":"http://x/5"}
]`

func newProjectsServer(t *testing.T, body string, statusCode int) *GRPCServer {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	cfg := config.ProgramSettings{JiraURL: srv.URL, MinTimeSleep: 1, MaxTimeSleep: 4}
	client := NewJiraClient(cfg, discardLogger())
	return NewGRPCServer(nil, client, cfg, discardLogger(), nil, "topic")
}

func TestGRPC_GetProjects_Pagination(t *testing.T) {
	s := newProjectsServer(t, fiveProjectsJSON, http.StatusOK)

	resp, err := s.GetProjects(context.Background(), &pb.GetProjectsRequest{Page: 1, Limit: 2})
	require.NoError(t, err)

	require.Len(t, resp.Projects, 2)
	assert.Equal(t, "ABC", resp.Projects[0].Key)
	assert.Equal(t, "BCD", resp.Projects[1].Key)
	assert.Equal(t, "http://x/1", resp.Projects[0].Url)

	require.NotNil(t, resp.PageInfo)
	assert.Equal(t, int32(1), resp.PageInfo.CurrentPage)
	assert.Equal(t, int32(5), resp.PageInfo.ProjectsCount)
	assert.Equal(t, int32(3), resp.PageInfo.TotalPages) // ceil(5/2)
}

func TestGRPC_GetProjects_LastPagePartial(t *testing.T) {
	s := newProjectsServer(t, fiveProjectsJSON, http.StatusOK)

	resp, err := s.GetProjects(context.Background(), &pb.GetProjectsRequest{Page: 3, Limit: 2})
	require.NoError(t, err)
	require.Len(t, resp.Projects, 1) // only the 5th item remains
	assert.Equal(t, "EFG", resp.Projects[0].Key)
}

func TestGRPC_GetProjects_PageBeyondRange(t *testing.T) {
	s := newProjectsServer(t, fiveProjectsJSON, http.StatusOK)

	resp, err := s.GetProjects(context.Background(), &pb.GetProjectsRequest{Page: 99, Limit: 2})
	require.NoError(t, err)
	assert.Empty(t, resp.Projects)
	assert.Equal(t, int32(5), resp.PageInfo.ProjectsCount)
}

func TestGRPC_GetProjects_Search(t *testing.T) {
	s := newProjectsServer(t, fiveProjectsJSON, http.StatusOK)

	// "el" matches "Delta" (DEF) by name; nothing else by name or key.
	resp, err := s.GetProjects(context.Background(), &pb.GetProjectsRequest{Page: 1, Limit: 10, Search: "el"})
	require.NoError(t, err)
	require.Len(t, resp.Projects, 1)
	assert.Equal(t, "DEF", resp.Projects[0].Key)
	assert.Equal(t, int32(1), resp.PageInfo.ProjectsCount)
}

func TestGRPC_GetProjects_SearchByKeyCaseInsensitive(t *testing.T) {
	s := newProjectsServer(t, fiveProjectsJSON, http.StatusOK)

	resp, err := s.GetProjects(context.Background(), &pb.GetProjectsRequest{Page: 1, Limit: 10, Search: "abc"})
	require.NoError(t, err)
	require.Len(t, resp.Projects, 1)
	assert.Equal(t, "ABC", resp.Projects[0].Key)
}

func TestGRPC_GetProjects_ConnectorUnavailable(t *testing.T) {
	s := newProjectsServer(t, `boom`, http.StatusInternalServerError)

	_, err := s.GetProjects(context.Background(), &pb.GetProjectsRequest{Page: 1, Limit: 2})
	require.Error(t, err)
	assert.Equal(t, codes.Unavailable, status.Code(err))
}

func TestGRPC_UpdateProject_RequiresKey(t *testing.T) {
	s := NewGRPCServer(nil, nil, config.ProgramSettings{}, discardLogger(), nil, "topic")

	_, err := s.UpdateProject(context.Background(), &pb.UpdateProjectRequest{ProjectKey: ""})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGRPC_DeleteProject_InvalidID(t *testing.T) {
	s := NewGRPCServer(nil, nil, config.ProgramSettings{}, discardLogger(), nil, "topic")

	for _, id := range []int64{0, -5} {
		_, err := s.DeleteProject(context.Background(), &pb.DeleteProjectRequest{ProjectId: id})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	}
}
