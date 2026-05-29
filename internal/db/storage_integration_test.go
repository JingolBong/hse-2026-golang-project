//go:build integration

// Package db integration tests. Run with: go test -tags=integration ./internal/db/
// Requires a working Docker daemon (testcontainers spins up a throwaway Postgres).
package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"hse-2026-golang-project/internal/models"
)

const migrationFile = "../../migrations/01_create_tables.sql"

// setupTestDB spins up a Postgres container, applies the schema migration and
// returns a Storage wired to it (same DB used for both write and read roles).
func setupTestDB(t *testing.T) *Storage {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"postgres:13.8",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ctr)

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.Eventually(t, func() bool { return db.Ping() == nil }, 30*time.Second, time.Second)

	schema, err := os.ReadFile(migrationFile)
	require.NoError(t, err)
	_, err = db.Exec(string(schema))
	require.NoError(t, err)

	// Storage takes separate write/read handles; in tests both point at the
	// same instance, which exercises the read-with-fallback path too.
	return NewStorage(db, db)
}

func strPtr(s string) *string { return &s }

func seedProject(t *testing.T, s *Storage, jiraID int64, key string) {
	t.Helper()
	_, err := s.UpsertProject(context.Background(), models.Project{
		JiraID: jiraID, Key: key, Name: key + " name", URL: "http://x/" + key,
	})
	require.NoError(t, err)
}

func TestStorage_ProjectUpsertAndGet(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	seedProject(t, s, 1, "ABC")

	byID, err := s.GetProjectByJiraID(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, byID)
	assert.Equal(t, "ABC", byID.Key)

	byKey, err := s.GetProjectByKey(ctx, "ABC")
	require.NoError(t, err)
	require.NotNil(t, byKey)
	assert.Equal(t, int64(1), byKey.JiraID)

	byName, err := s.GetProjectByName(ctx, "ABC name")
	require.NoError(t, err)
	require.NotNil(t, byName)

	// Upsert with same jira_id updates the row in place.
	_, err = s.UpsertProject(ctx, models.Project{JiraID: 1, Key: "ABC", Name: "Renamed", URL: "u"})
	require.NoError(t, err)
	updated, _ := s.GetProjectByJiraID(ctx, 1)
	assert.Equal(t, "Renamed", updated.Name)
}

func TestStorage_GetProject_NotFoundReturnsNil(t *testing.T) {
	s := setupTestDB(t)
	got, err := s.GetProjectByJiraID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, got, "missing project must be (nil, nil), not an error")
}

func TestStorage_GetAllProjects(t *testing.T) {
	s := setupTestDB(t)
	seedProject(t, s, 1, "ABC")
	seedProject(t, s, 2, "DEF")

	all, err := s.GetAllProjects(context.Background())
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestStorage_AuthorUpsertAndGet(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()

	_, err := s.UpsertAuthor(ctx, models.Author{JiraID: 10, Username: "alice", Email: strPtr("a@x.com")})
	require.NoError(t, err)
	_, err = s.UpsertAuthor(ctx, models.Author{JiraID: 11, Username: "bob"}) // nil email
	require.NoError(t, err)

	alice, err := s.GetAuthorByJiraID(ctx, 10)
	require.NoError(t, err)
	require.NotNil(t, alice)
	require.NotNil(t, alice.Email)
	assert.Equal(t, "a@x.com", *alice.Email)

	bob, err := s.GetAuthorByJiraID(ctx, 11)
	require.NoError(t, err)
	require.NotNil(t, bob)
	assert.Nil(t, bob.Email, "absent email must scan as nil")
}

func TestStorage_IssuesBatchAndGet(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()
	seedProject(t, s, 1, "ABC")

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	spent := int32(3600)
	issues := []models.Issue{
		{JiraID: 100, ProjectID: 1, Key: "ABC-1", Summary: "first", Status: "Open", Priority: "Major", CreatedAt: base.AddDate(0, 0, 2)},
		{JiraID: 101, ProjectID: 1, Key: "ABC-2", Summary: "second", Status: "Closed", Priority: "Minor", CreatedAt: base, ClosedAt: &base, TimeSpent: &spent},
	}
	require.NoError(t, s.UpsertIssuesBatch(ctx, issues))

	got, err := s.GetIssuesByProject(ctx, 1)
	require.NoError(t, err)
	require.Len(t, got, 2)
	// Ordered by created_time ASC -> ABC-2 (earlier) first.
	assert.Equal(t, "ABC-2", got[0].Key)
	assert.Equal(t, "ABC-1", got[1].Key)

	require.NotNil(t, got[0].ClosedAt)
	require.NotNil(t, got[0].TimeSpent)
	assert.Equal(t, int32(3600), *got[0].TimeSpent)
	assert.Nil(t, got[1].ClosedAt, "open issue must have nil closed_time")

	// Empty batch is a no-op.
	require.NoError(t, s.UpsertIssuesBatch(ctx, nil))
}

func TestStorage_StatusChangesBatchAndDedup(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()
	seedProject(t, s, 1, "ABC")
	require.NoError(t, s.UpsertIssuesBatch(ctx, []models.Issue{
		{JiraID: 100, ProjectID: 1, Key: "ABC-1", Status: "Open", Priority: "Major", CreatedAt: time.Now()},
	}))

	ct := time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC)
	changes := []models.StatusChange{
		{IssueID: 100, OldStatus: strPtr("Open"), NewStatus: strPtr("In Progress"), ChangeTime: ct},
	}
	require.NoError(t, s.InsertStatusChangesBatch(ctx, changes))
	// Re-inserting the same change is ignored (ON CONFLICT DO NOTHING).
	require.NoError(t, s.InsertStatusChangesBatch(ctx, changes))

	byProject, err := s.GetStatusChangesByProject(ctx, 1)
	require.NoError(t, err)
	require.Len(t, byProject, 1, "duplicate status change must not be inserted twice")
	assert.Equal(t, "In Progress", *byProject[0].NewStatus)

	byIssue, err := s.GetStatusChangesByIssue(ctx, 100)
	require.NoError(t, err)
	require.Len(t, byIssue, 1)
}

func TestStorage_DeleteProjectCascade(t *testing.T) {
	s := setupTestDB(t)
	ctx := context.Background()
	seedProject(t, s, 1, "ABC")
	require.NoError(t, s.UpsertIssuesBatch(ctx, []models.Issue{
		{JiraID: 100, ProjectID: 1, Key: "ABC-1", Status: "Open", Priority: "Major", CreatedAt: time.Now()},
	}))
	require.NoError(t, s.InsertStatusChangesBatch(ctx, []models.StatusChange{
		{IssueID: 100, NewStatus: strPtr("Open"), ChangeTime: time.Now()},
	}))

	require.NoError(t, s.DeleteProject(ctx, 1))

	// Project, its issues and their status changes are all gone.
	proj, _ := s.GetProjectByJiraID(ctx, 1)
	assert.Nil(t, proj)
	issues, _ := s.GetIssuesByProject(ctx, 1)
	assert.Empty(t, issues)
	changes, _ := s.GetStatusChangesByIssue(ctx, 100)
	assert.Empty(t, changes)

	// Deleting a missing project surfaces ErrNotFound.
	err := s.DeleteProject(ctx, 1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestStorage_HealthCheck(t *testing.T) {
	s := setupTestDB(t)
	h, err := s.HealthCheck(context.Background())
	require.NoError(t, err)
	assert.True(t, h.MasterUp)
	assert.True(t, h.ReplicaUp)
	assert.False(t, h.MasterRecovery, "a standalone primary is not in recovery")
}
