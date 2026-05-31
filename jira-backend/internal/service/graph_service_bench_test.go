package service

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"hse-2026-golang-project/internal/models"
)

// genIssues builds a deterministic, reproducible dataset of n issues (plus a
// proportional set of status changes) for the analytics benchmarks. Values are
// spread across the branches the analytics functions actually distinguish:
//   - Status cycles Open/In Progress/Resolved/Reopened/Closed (see classifyStatus)
//   - Priority cycles Blocker/Critical/Major/Minor/Trivial (see matchPriority)
//   - CreatedAt is spread across a year; ~half the issues are closed
//   - TimeSpent varies to hit different hourBucket ranges
//   - ~1/4 of issues carry a status-change chain (exercises stateDaysByIssue)
//
// No rand: the same n always produces the same data, so ns/op numbers are stable.
func genIssues(n int) ([]models.Issue, []models.StatusChange) {
	statuses := []string{"Open", "In Progress", "Resolved", "Reopened", "Closed"}
	priorities := []string{"Blocker", "Critical", "Major", "Minor", "Trivial"}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	issues := make([]models.Issue, 0, n)
	changes := make([]models.StatusChange, 0, n)

	for i := 0; i < n; i++ {
		id := int64(i + 1)
		created := base.AddDate(0, 0, i%365)

		iss := models.Issue{
			JiraID:    id,
			ProjectID: 1,
			Key:       "P-" + strconv.Itoa(i),
			Summary:   "issue summary",
			Status:    statuses[i%len(statuses)],
			Priority:  priorities[i%len(priorities)],
			CreatedAt: created,
		}
		if i%2 == 0 {
			cl := created.AddDate(0, 0, 1+i%30)
			iss.ClosedAt = &cl
		}
		if i%3 != 0 {
			ts := int32(1800 * (1 + i%16))
			iss.TimeSpent = &ts
		}
		issues = append(issues, iss)

		if i%4 == 0 {
			openS, progressS, resolveS := "Open", "In Progress", "Resolved"
			changes = append(changes,
				models.StatusChange{
					IssueID: id, OldStatus: &openS, NewStatus: &progressS,
					ChangeTime: created.AddDate(0, 0, 1),
				},
				models.StatusChange{
					IssueID: id, OldStatus: &progressS, NewStatus: &resolveS,
					ChangeTime: created.AddDate(0, 0, 4),
				},
			)
		}
	}
	return issues, changes
}

var benchSizes = []int{100, 1000, 10000}

func BenchmarkStateDaysByIssue(b *testing.B) {
	for _, n := range benchSizes {
		issues, changes := genIssues(n)
		b.Run(fmt.Sprintf("issues=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = stateDaysByIssue(issues, changes)
			}
		})
	}
}

func BenchmarkStateDistribution(b *testing.B) {
	for _, n := range benchSizes {
		issues, changes := genIssues(n)
		b.Run(fmt.Sprintf("issues=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = stateDistribution(issues, changes)
			}
		})
	}
}

func BenchmarkActivityByDate(b *testing.B) {
	for _, n := range benchSizes {
		issues, _ := genIssues(n)
		b.Run(fmt.Sprintf("issues=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = activityByDate(issues)
			}
		})
	}
}

func BenchmarkComplexityHistogram(b *testing.B) {
	for _, n := range benchSizes {
		issues, _ := genIssues(n)
		b.Run(fmt.Sprintf("issues=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = complexityHistogram(issues)
			}
		})
	}
}

func BenchmarkPriorityHistogram(b *testing.B) {
	for _, n := range benchSizes {
		issues, _ := genIssues(n)
		for _, closedOnly := range []bool{false, true} {
			b.Run(fmt.Sprintf("issues=%d/closedOnly=%v", n, closedOnly), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = priorityHistogram(issues, closedOnly)
				}
			})
		}
	}
}
