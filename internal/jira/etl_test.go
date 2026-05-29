package connector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashUsername(t *testing.T) {
	a := hashUsername("alice")
	b := hashUsername("alice")
	assert.Equal(t, a, b, "same input must hash identically")
	assert.GreaterOrEqual(t, a, int64(0), "hash must be non-negative")
	assert.NotEqual(t, hashUsername("alice"), hashUsername("bob"))
}

func TestParseJiraTime(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := parseJiraTime("2025-01-02T15:04:05.000+0000")
		require.NoError(t, err)
		assert.Equal(t, 2025, got.Year())
		assert.Equal(t, time.January, got.Month())
		assert.Equal(t, 2, got.Day())
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseJiraTime("not-a-date")
		assert.Error(t, err)
	})
}

func TestTransformUser(t *testing.T) {
	t.Run("with email", func(t *testing.T) {
		a := transformUser(User{Name: "alice", EmailAddress: "alice@example.com"})
		assert.Equal(t, hashUsername("alice"), a.JiraID)
		assert.Equal(t, "alice", a.Username)
		require.NotNil(t, a.Email)
		assert.Equal(t, "alice@example.com", *a.Email)
	})

	t.Run("empty email -> nil", func(t *testing.T) {
		a := transformUser(User{Name: "bob"})
		assert.Nil(t, a.Email)
	})
}

func sampleJiraIssue() JiraIssue {
	timeSpent := int32(3600)
	resolution := "2025-01-05T10:00:00.000+0000"
	return JiraIssue{
		ID:  "1001",
		Key: "ABC-1",
		Fields: IssueFields{
			Summary:        "Fix bug",
			Status:         Status{Name: "Closed"},
			Priority:       Priority{Name: "Major"},
			Created:        "2025-01-01T09:00:00.000+0000",
			Updated:        "2025-01-04T09:00:00.000+0000",
			ResolutionDate: &resolution,
			TimeSpent:      &timeSpent,
			Creator:        &User{Name: "alice"},
			Assignee:       &User{Name: "bob"},
		},
		ChangeLog: ChangeLog{
			Histories: []History{
				{
					Created: "2025-01-02T12:00:00.000+0000",
					Items: []HistoryItem{
						{Field: "status", FromString: "Open", ToString: "In Progress"},
						{Field: "assignee", FromString: "x", ToString: "y"}, // ignored
					},
				},
			},
		},
	}
}

func TestTransformIssue_FullMapping(t *testing.T) {
	issue, changes, err := transformIssue(sampleJiraIssue(), 77)
	require.NoError(t, err)

	assert.Equal(t, int64(1001), issue.JiraID)
	assert.Equal(t, int64(77), issue.ProjectID)
	assert.Equal(t, "ABC-1", issue.Key)
	assert.Equal(t, "Fix bug", issue.Summary)
	assert.Equal(t, "Closed", issue.Status)
	assert.Equal(t, "Major", issue.Priority)

	require.NotNil(t, issue.TimeSpent)
	assert.Equal(t, int32(3600), *issue.TimeSpent)
	require.NotNil(t, issue.UpdatedAt)
	require.NotNil(t, issue.ClosedAt)
	require.NotNil(t, issue.CreatorID)
	assert.Equal(t, hashUsername("alice"), *issue.CreatorID)
	require.NotNil(t, issue.AssigneeID)
	assert.Equal(t, hashUsername("bob"), *issue.AssigneeID)

	require.Len(t, changes, 1)
	assert.Equal(t, int64(1001), changes[0].IssueID)
	require.NotNil(t, changes[0].OldStatus)
	assert.Equal(t, "Open", *changes[0].OldStatus)
	require.NotNil(t, changes[0].NewStatus)
	assert.Equal(t, "In Progress", *changes[0].NewStatus)
}

func TestTransformIssue_OptionalFieldsNil(t *testing.T) {
	ji := JiraIssue{
		ID:  "5",
		Key: "X-5",
		Fields: IssueFields{
			Summary:  "minimal",
			Status:   Status{Name: "Open"},
			Priority: Priority{Name: "Minor"},
			Created:  "2025-01-01T00:00:00.000+0000",
		},
	}
	issue, changes, err := transformIssue(ji, 1)
	require.NoError(t, err)
	assert.Nil(t, issue.TimeSpent)
	assert.Nil(t, issue.UpdatedAt)
	assert.Nil(t, issue.ClosedAt)
	assert.Nil(t, issue.CreatorID)
	assert.Nil(t, issue.AssigneeID)
	assert.Empty(t, changes)
}

func TestTransformIssue_Errors(t *testing.T) {
	t.Run("bad id", func(t *testing.T) {
		ji := sampleJiraIssue()
		ji.ID = "not-numeric"
		_, _, err := transformIssue(ji, 1)
		assert.Error(t, err)
	})

	t.Run("bad created", func(t *testing.T) {
		ji := sampleJiraIssue()
		ji.Fields.Created = "bad-date"
		_, _, err := transformIssue(ji, 1)
		assert.Error(t, err)
	})
}

func TestTransformIssue_SkipsUnparsableChangeTime(t *testing.T) {
	ji := JiraIssue{
		ID:     "9",
		Key:    "X-9",
		Fields: IssueFields{Created: "2025-01-01T00:00:00.000+0000", Status: Status{Name: "Open"}},
		ChangeLog: ChangeLog{Histories: []History{
			{Created: "bad-time", Items: []HistoryItem{{Field: "status", ToString: "Done"}}},
		}},
	}
	_, changes, err := transformIssue(ji, 1)
	require.NoError(t, err)
	assert.Empty(t, changes, "history with unparsable time must be skipped")
}
