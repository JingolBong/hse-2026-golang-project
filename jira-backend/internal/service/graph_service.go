package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"hse-2026-golang-project/internal/models"
	"hse-2026-golang-project/jira-backend/internal/repository"
)

var ErrUnsupportedTask = errors.New("unsupported graph task")

var (
	dayBins6     = []string{"0-1d", "1-3d", "3-7d", "7-14d", "14-30d", ">30d"}
	dayBins4     = []string{"0-1d", "1-3d", "3-7d", ">7d"}
	hourBins     = []string{"0-1h", "1-4h", "4-8h", "8-24h", ">24h"}
	priorityBins = []string{"Blocker", "Critical", "Major", "Minor", "Trivial"}
)

type GraphService struct {
	repo     *repository.ProjectRepository
	mu       sync.RWMutex
	analyzed map[string]bool
}

func NewGraphService(repo *repository.ProjectRepository) *GraphService {
	return &GraphService{
		repo:     repo,
		analyzed: make(map[string]bool),
	}
}

type histogram struct {
	Categories []string       `json:"categories"`
	Count      map[string]int `json:"count"`
}

type stateDist struct {
	Categories map[string][]string `json:"categories"`
	Open       map[string]int      `json:"open"`
	Resolve    map[string]int      `json:"resolve"`
	Progress   map[string]int      `json:"progress"`
	Reopen     map[string]int      `json:"reopen"`
}

type activity struct {
	Categories map[string][]string `json:"categories"`
	Open       map[string]int      `json:"open"`
	Close      map[string]int      `json:"close"`
}

type compareHistogram struct {
	Categories []string         `json:"categories"`
	Count      map[string][]int `json:"count"`
}

func validTask(task int) bool { return task >= 1 && task <= 6 }

func resolveProject(ctx context.Context, repo *repository.ProjectRepository, ref string) (*models.Project, error) {
	p, err := repo.GetByKey(ctx, ref)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}
	return repo.GetByName(ctx, ref)
}

func (s *GraphService) Make(ctx context.Context, projectKey string, task int) error {
	if !validTask(task) {
		return ErrUnsupportedTask
	}

	project, err := resolveProject(ctx, s.repo, projectKey)
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	s.mu.Lock()
	s.analyzed[projectKey] = true
	s.mu.Unlock()

	return nil
}

func (s *GraphService) IsEmpty(ctx context.Context, ref string) (bool, error) {
	project, err := resolveProject(ctx, s.repo, ref)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, ErrProjectNotFound
	}

	issues, err := s.repo.GetIssuesByProject(ctx, project.JiraID)
	if err != nil {
		return false, err
	}
	return len(issues) == 0, nil
}

func (s *GraphService) Get(ctx context.Context, projectKey string, task int) (interface{}, error) {
	if !validTask(task) {
		return nil, ErrUnsupportedTask
	}

	project, err := resolveProject(ctx, s.repo, projectKey)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	issues, err := s.repo.GetIssuesByProject(ctx, project.JiraID)
	if err != nil {
		return nil, err
	}
	if len(issues) == 0 {
		return nil, nil
	}

	switch task {
	case 1:
		changes, err := s.repo.GetStatusChangesByProject(ctx, project.JiraID)
		if err != nil {
			return nil, err
		}
		return openStateHistogram(issues, changes), nil
	case 2:
		changes, err := s.repo.GetStatusChangesByProject(ctx, project.JiraID)
		if err != nil {
			return nil, err
		}
		return stateDistribution(issues, changes), nil
	case 3:
		return activityByDate(issues), nil
	case 4:
		return complexityHistogram(issues), nil
	case 5:
		return priorityHistogram(issues, false), nil
	case 6:
		return priorityHistogram(issues, true), nil
	default:
		return nil, ErrUnsupportedTask
	}
}

func (s *GraphService) Compare(ctx context.Context, keys []string, task int) (interface{}, error) {
	if task != 1 {
		return nil, ErrUnsupportedTask
	}
	if len(keys) == 0 {
		return nil, nil
	}

	result := &compareHistogram{
		Categories: dayBins6,
		Count:      make(map[string][]int, len(dayBins6)),
	}
	for _, cat := range dayBins6 {
		result.Count[cat] = make([]int, len(keys))
	}

	any := false
	for j, key := range keys {
		project, err := resolveProject(ctx, s.repo, key)
		if err != nil {
			return nil, err
		}
		if project == nil {
			continue
		}
		issues, err := s.repo.GetIssuesByProject(ctx, project.JiraID)
		if err != nil {
			return nil, err
		}
		if len(issues) == 0 {
			continue
		}
		changes, err := s.repo.GetStatusChangesByProject(ctx, project.JiraID)
		if err != nil {
			return nil, err
		}
		counts := openStateCounts(issues, changes)
		for _, cat := range dayBins6 {
			result.Count[cat][j] = counts[cat]
		}
		any = true
	}

	if !any {
		return nil, nil
	}
	return result, nil
}

func (s *GraphService) IsAnalyzed(projectKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.analyzed[projectKey]
}

func (s *GraphService) DropAnalyzed(projectKey string) {
	s.mu.Lock()
	delete(s.analyzed, projectKey)
	s.mu.Unlock()
}

func stateDaysByIssue(issues []models.Issue, changes []models.StatusChange) map[int64]map[string]float64 {
	byIssue := make(map[int64][]models.StatusChange)
	for _, c := range changes {
		byIssue[c.IssueID] = append(byIssue[c.IssueID], c)
	}

	now := time.Now()
	result := make(map[int64]map[string]float64, len(issues))

	for _, issue := range issues {
		ch := byIssue[issue.JiraID]
		days := make(map[string]float64, 5)

		current := issue.Status
		if len(ch) > 0 && ch[0].OldStatus != nil {
			current = *ch[0].OldStatus
		}
		start := issue.CreatedAt

		add := func(status string, until time.Time) {
			d := until.Sub(start).Hours() / 24
			if d > 0 {
				days[classifyStatus(status)] += d
			}
		}

		for _, c := range ch {
			if c.NewStatus == nil {
				continue
			}
			add(current, c.ChangeTime)
			current = *c.NewStatus
			start = c.ChangeTime
		}

		end := now
		if issue.ClosedAt != nil {
			end = *issue.ClosedAt
		}
		add(current, end)

		result[issue.JiraID] = days
	}

	return result
}

func openStateCounts(issues []models.Issue, changes []models.StatusChange) map[string]int {
	stateDays := stateDaysByIssue(issues, changes)
	counts := newZeroCounts(dayBins6)
	for _, issue := range issues {
		if d := stateDays[issue.JiraID]["open"]; d > 0 {
			counts[dayBucket6(d)]++
		}
	}
	return counts
}

func openStateHistogram(issues []models.Issue, changes []models.StatusChange) interface{} {
	counts := openStateCounts(issues, changes)
	if total(counts) == 0 {
		return nil
	}
	return &histogram{Categories: dayBins6, Count: counts}
}

func stateDistribution(issues []models.Issue, changes []models.StatusChange) interface{} {
	stateDays := stateDaysByIssue(issues, changes)

	open := newZeroCounts(dayBins4)
	resolve := newZeroCounts(dayBins4)
	progress := newZeroCounts(dayBins4)
	reopen := newZeroCounts(dayBins4)

	for _, issue := range issues {
		d := stateDays[issue.JiraID]
		if v := d["open"]; v > 0 {
			open[dayBucket4(v)]++
		}
		if v := d["resolve"]; v > 0 {
			resolve[dayBucket4(v)]++
		}
		if v := d["progress"]; v > 0 {
			progress[dayBucket4(v)]++
		}
		if v := d["reopen"]; v > 0 {
			reopen[dayBucket4(v)]++
		}
	}

	if total(open)+total(resolve)+total(progress)+total(reopen) == 0 {
		return nil
	}

	return &stateDist{
		Categories: map[string][]string{
			"open": dayBins4, "resolve": dayBins4, "progress": dayBins4, "reopen": dayBins4,
		},
		Open: open, Resolve: resolve, Progress: progress, Reopen: reopen,
	}
}

func activityByDate(issues []models.Issue) interface{} {
	const layout = "2006-01-02"

	createdByDate := make(map[string]int)
	closedByDate := make(map[string]int)
	dateSet := make(map[string]struct{})

	for _, issue := range issues {
		cd := issue.CreatedAt.Format(layout)
		createdByDate[cd]++
		dateSet[cd] = struct{}{}
		if issue.ClosedAt != nil {
			cl := issue.ClosedAt.Format(layout)
			closedByDate[cl]++
			dateSet[cl] = struct{}{}
		}
	}

	if len(dateSet) == 0 {
		return nil
	}

	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	openCum := make(map[string]int, len(dates))
	closeCum := make(map[string]int, len(dates))
	runOpen, runClose := 0, 0
	for _, d := range dates {
		runOpen += createdByDate[d]
		runClose += closedByDate[d]
		openCum[d] = runOpen
		closeCum[d] = runClose
	}

	return &activity{
		Categories: map[string][]string{"all": dates},
		Open:       openCum,
		Close:      closeCum,
	}
}

func complexityHistogram(issues []models.Issue) interface{} {
	counts := newZeroCounts(hourBins)
	for _, issue := range issues {
		if issue.TimeSpent == nil || *issue.TimeSpent <= 0 {
			continue
		}
		hours := float64(*issue.TimeSpent) / 3600
		counts[hourBucket(hours)]++
	}
	if total(counts) == 0 {
		return nil
	}
	return &histogram{Categories: hourBins, Count: counts}
}

func priorityHistogram(issues []models.Issue, closedOnly bool) interface{} {
	counts := newZeroCounts(priorityBins)
	for _, issue := range issues {
		if closedOnly && issue.ClosedAt == nil {
			continue
		}
		if cat, ok := matchPriority(issue.Priority); ok {
			counts[cat]++
		}
	}
	if total(counts) == 0 {
		return nil
	}
	return &histogram{Categories: priorityBins, Count: counts}
}

func dayBucket6(days float64) string {
	switch {
	case days <= 1:
		return "0-1d"
	case days <= 3:
		return "1-3d"
	case days <= 7:
		return "3-7d"
	case days <= 14:
		return "7-14d"
	case days <= 30:
		return "14-30d"
	default:
		return ">30d"
	}
}

func dayBucket4(days float64) string {
	switch {
	case days <= 1:
		return "0-1d"
	case days <= 3:
		return "1-3d"
	case days <= 7:
		return "3-7d"
	default:
		return ">7d"
	}
}

func hourBucket(hours float64) string {
	switch {
	case hours <= 1:
		return "0-1h"
	case hours <= 4:
		return "1-4h"
	case hours <= 8:
		return "4-8h"
	case hours <= 24:
		return "8-24h"
	default:
		return ">24h"
	}
}

func matchPriority(priority string) (string, bool) {
	for _, p := range priorityBins {
		if strings.EqualFold(priority, p) {
			return p, true
		}
	}
	return "", false
}

func newZeroCounts(bins []string) map[string]int {
	m := make(map[string]int, len(bins))
	for _, b := range bins {
		m[b] = 0
	}
	return m
}

func total(counts map[string]int) int {
	sum := 0
	for _, v := range counts {
		sum += v
	}
	return sum
}
