package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

type ProjectStat struct {
	ID                  int64
	Key                 string
	Name                string
	AllIssuesCount      int
	OpenIssuesCount     int
	CloseIssuesCount    int
	ReopenedIssuesCount int
	ResolvedIssuesCount int
	ProgressIssuesCount int
	AverageTime         float64
	AverageIssuesCount  string
}

func (s *ProjectService) GetStat(ctx context.Context, jiraID int64) (*ProjectStat, error) {
	project, err := s.repo.GetByID(ctx, jiraID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	issues, err := s.repo.GetIssuesByProject(ctx, jiraID)
	if err != nil {
		return nil, err
	}

	stat := &ProjectStat{
		ID:             project.JiraID,
		Key:            project.Key,
		Name:           project.Name,
		AllIssuesCount: len(issues),
	}

	var (
		closedDaysSum float64
		closedCount   int
		minCreated    time.Time
		maxCreated    time.Time
	)

	for _, issue := range issues {
		switch classifyStatus(issue.Status) {
		case "open":
			stat.OpenIssuesCount++
		case "close":
			stat.CloseIssuesCount++
		case "reopen":
			stat.ReopenedIssuesCount++
		case "resolve":
			stat.ResolvedIssuesCount++
		case "progress":
			stat.ProgressIssuesCount++
		}

		if issue.ClosedAt != nil {
			closedDaysSum += issue.ClosedAt.Sub(issue.CreatedAt).Hours() / 24
			closedCount++
		}

		if minCreated.IsZero() || issue.CreatedAt.Before(minCreated) {
			minCreated = issue.CreatedAt
		}
		if issue.CreatedAt.After(maxCreated) {
			maxCreated = issue.CreatedAt
		}
	}

	if closedCount > 0 {
		stat.AverageTime = math.Round(closedDaysSum/float64(closedCount)*10) / 10
	}
	stat.AverageIssuesCount = averagePerMonth(len(issues), minCreated, maxCreated)

	return stat, nil
}

func classifyStatus(status string) string {
	s := strings.ToLower(status)
	switch {
	case strings.Contains(s, "closed"):
		return "close"
	case strings.Contains(s, "reopened"):
		return "reopen"
	case strings.Contains(s, "resolved"):
		return "resolve"
	case strings.Contains(s, "progress"):
		return "progress"
	case strings.Contains(s, "open"):
		return "open"
	default:
		return ""
	}
}

func averagePerMonth(total int, first, last time.Time) string {
	if total == 0 || first.IsZero() {
		return "0.0"
	}
	months := last.Sub(first).Hours() / 24 / 30
	if months < 1 {
		months = 1
	}
	return fmt.Sprintf("%.1f", float64(total)/months)
}
