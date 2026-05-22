package service

import (
	"testing"
	"time"

	"hse-2026-golang-project/internal/models"
)

func ptrTime(t time.Time) *time.Time { return &t }
func ptrI32(v int32) *int32          { return &v }

func TestPriorityHistogram(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	issues := []models.Issue{
		{JiraID: 1, Priority: "Blocker", CreatedAt: base},
		{JiraID: 2, Priority: "major", CreatedAt: base, ClosedAt: ptrTime(base.AddDate(0, 0, 2))},
		{JiraID: 3, Priority: "Major", CreatedAt: base, ClosedAt: ptrTime(base.AddDate(0, 0, 5))},
		{JiraID: 4, Priority: "Unknown", CreatedAt: base},
	}

	all := priorityHistogram(issues, false).(*histogram)
	if all.Count["Blocker"] != 1 || all.Count["Major"] != 2 {
		t.Fatalf("all priorities: got %+v", all.Count)
	}
	if all.Count["Trivial"] != 0 {
		t.Fatalf("expected zero Trivial, got %d", all.Count["Trivial"])
	}

	closed := priorityHistogram(issues, true).(*histogram)
	if closed.Count["Major"] != 2 || closed.Count["Blocker"] != 0 {
		t.Fatalf("closed priorities: got %+v", closed.Count)
	}
}

func TestComplexityHistogram(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	issues := []models.Issue{
		{JiraID: 1, CreatedAt: base, TimeSpent: ptrI32(1800)},
		{JiraID: 2, CreatedAt: base, TimeSpent: ptrI32(7200)},
		{JiraID: 3, CreatedAt: base, TimeSpent: ptrI32(90000)},
		{JiraID: 4, CreatedAt: base},
	}
	h := complexityHistogram(issues).(*histogram)
	if h.Count["0-1h"] != 1 || h.Count["1-4h"] != 1 || h.Count[">24h"] != 1 {
		t.Fatalf("complexity: got %+v", h.Count)
	}
}

func TestActivityByDate(t *testing.T) {
	d1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	issues := []models.Issue{
		{JiraID: 1, CreatedAt: d1},
		{JiraID: 2, CreatedAt: d1, ClosedAt: ptrTime(d2)},
		{JiraID: 3, CreatedAt: d2},
	}
	a := activityByDate(issues).(*activity)
	dates := a.Categories["all"]
	last := dates[len(dates)-1]
	if a.Open[last] != 3 {
		t.Fatalf("cumulative open on %s = %d, want 3", last, a.Open[last])
	}
	if a.Close[last] != 1 {
		t.Fatalf("cumulative close on %s = %d, want 1", last, a.Close[last])
	}
	if a.Open[dates[0]] != 2 {
		t.Fatalf("open on first day = %d, want 2", a.Open[dates[0]])
	}
}

func TestOpenStateCountsFromTimeline(t *testing.T) {
	created := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	movedToProgress := created.AddDate(0, 0, 5)
	issues := []models.Issue{
		{JiraID: 1, Status: "In Progress", CreatedAt: created},
	}
	open := "Open"
	prog := "In Progress"
	changes := []models.StatusChange{
		{IssueID: 1, OldStatus: &open, NewStatus: &prog, ChangeTime: movedToProgress},
	}

	counts := openStateCounts(issues, changes)
	if counts["3-7d"] != 1 {
		t.Fatalf("open-state counts: got %+v, want 3-7d=1", counts)
	}

	dist := stateDistribution(issues, changes).(*stateDist)
	if dist.Open[">7d"] != 0 || dist.Open["3-7d"] != 1 {
		t.Fatalf("state dist open: got %+v", dist.Open)
	}
	if dist.Progress[">7d"] != 1 {
		t.Fatalf("state dist progress: got %+v", dist.Progress)
	}
}

func TestEmptyReturnsNil(t *testing.T) {
	if got := priorityHistogram(nil, false); got != nil {
		t.Fatalf("expected nil for no matching issues, got %+v", got)
	}
	if got := activityByDate(nil); got != nil {
		t.Fatalf("expected nil activity for no issues, got %+v", got)
	}
}

func TestBuckets(t *testing.T) {
	cases := []struct {
		days float64
		want string
	}{
		{0.5, "0-1d"}, {1, "0-1d"}, {2, "1-3d"}, {7, "3-7d"}, {10, "7-14d"}, {31, ">30d"},
	}
	for _, c := range cases {
		if got := dayBucket6(c.days); got != c.want {
			t.Errorf("dayBucket6(%v) = %q, want %q", c.days, got, c.want)
		}
	}
}
