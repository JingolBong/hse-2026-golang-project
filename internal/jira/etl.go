package connector

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"hse-2026-golang-project/internal/config"
	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/internal/models"
)

type pageTask struct {
	startAt int
	pageNum int
}

type pageResult struct {
	issues  []models.Issue
	changes []models.StatusChange
	authors map[int64]models.Author
}

func LoadProject(
	ctx context.Context,
	storage *db.Storage,
	client *JiraClient,
	projectKey string,
	projectID int64,
	cfg config.ProgramSettings,
	log *logrus.Logger,
) error {
	firstPage, err := client.FetchIssuesPage(ctx, projectKey, 0, 1)
	if err != nil {
		return fmt.Errorf("fetch total count failed: %w", err)
	}

	total := firstPage.Total
	log.WithFields(logrus.Fields{"project": projectKey, "total": total}).Info("Starting ETL")

	if total == 0 {
		return nil
	}

	pageSize := cfg.IssueInOneRequest
	totalPages := (total + pageSize - 1) / pageSize
	threadCount := cfg.ThreadCount

	tasks := make(chan pageTask, totalPages)
	results := make(chan pageResult, totalPages)

	for page := 0; page < totalPages; page++ {
		tasks <- pageTask{startAt: page * pageSize, pageNum: page}
	}
	close(tasks)

	eg, egCtx := errgroup.WithContext(ctx)

	for i := 0; i < threadCount; i++ {
		workerID := i
		eg.Go(func() error {
			for task := range tasks {
				if egCtx.Err() != nil {
					return egCtx.Err()
				}

				resp, err := client.FetchIssuesPage(egCtx, projectKey, task.startAt, pageSize)
				if err != nil {
					return fmt.Errorf("worker %d failed on page %d: %w", workerID, task.pageNum, err)
				}

				r := pageResult{authors: make(map[int64]models.Author)}

				for _, ji := range resp.Issues {
					issue, changes, err := transformIssue(ji, projectID)
					if err != nil {
						return fmt.Errorf("transform issue %s failed: %w", ji.Key, err)
					}

					if ji.Fields.Creator != nil {
						a := transformUser(*ji.Fields.Creator)
						r.authors[a.JiraID] = a
					}
					if ji.Fields.Assignee != nil {
						a := transformUser(*ji.Fields.Assignee)
						r.authors[a.JiraID] = a
					}

					r.issues = append(r.issues, issue)
					r.changes = append(r.changes, changes...)
				}
				results <- r
			}
			return nil
		})
	}

	go func() {
		eg.Wait()
		close(results)
	}()

	allAuthors := make(map[int64]models.Author)
	var allIssues []models.Issue
	var allChanges []models.StatusChange

	for r := range results {
		for id, a := range r.authors {
			allAuthors[id] = a
		}
		allIssues = append(allIssues, r.issues...)
		allChanges = append(allChanges, r.changes...)
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("ETL aborted due to fetch error: %w", err)
	}

	log.WithFields(logrus.Fields{"project": projectKey, "issues": len(allIssues)}).Info("Fetch done, writing to DB")

	for _, author := range allAuthors {
		if _, err := storage.UpsertAuthor(ctx, author); err != nil {
			return fmt.Errorf("upsert author failed: %w", err)
		}
	}

	const batchSize = 500
	for i := 0; i < len(allIssues); i += batchSize {
		end := i + batchSize
		if end > len(allIssues) {
			end = len(allIssues)
		}
		if err := storage.UpsertIssuesBatch(ctx, allIssues[i:end]); err != nil {
			return fmt.Errorf("upsert issues batch failed: %w", err)
		}
	}

	const changeBatchSize = 2000
	for i := 0; i < len(allChanges); i += changeBatchSize {
		end := i + changeBatchSize
		if end > len(allChanges) {
			end = len(allChanges)
		}
		if err := storage.InsertStatusChangesBatch(ctx, allChanges[i:end]); err != nil {
			return fmt.Errorf("insert status changes batch failed: %w", err)
		}
	}

	return nil
}
func transformUser(u User) models.Author {
	id := hashUsername(u.Name)
	var email *string
	if u.EmailAddress != "" {
		e := u.EmailAddress
		email = &e
	}
	return models.Author{
		JiraID:   id,
		Username: u.Name,
		Email:    email,
	}
}

func transformIssue(ji JiraIssue, projectID int64) (models.Issue, []models.StatusChange, error) {
	jiraID, err := strconv.ParseInt(ji.ID, 10, 64)
	if err != nil {
		return models.Issue{}, nil, fmt.Errorf("parse issue id %q: %w", ji.ID, err)
	}

	createdAt, err := parseJiraTime(ji.Fields.Created)
	if err != nil {
		return models.Issue{}, nil, fmt.Errorf("parse created for %s: %w", ji.Key, err)
	}

	issue := models.Issue{
		JiraID:    jiraID,
		ProjectID: projectID,
		Key:       ji.Key,
		Summary:   ji.Fields.Summary,
		Status:    ji.Fields.Status.Name,
		Priority:  ji.Fields.Priority.Name,
		CreatedAt: createdAt,
	}

	if ji.Fields.TimeSpent != nil {
		issue.TimeSpent = ji.Fields.TimeSpent
	}
	if ji.Fields.Updated != "" {
		if t, err := parseJiraTime(ji.Fields.Updated); err == nil {
			issue.UpdatedAt = &t
		}
	}
	if ji.Fields.ResolutionDate != nil {
		if t, err := parseJiraTime(*ji.Fields.ResolutionDate); err == nil {
			issue.ClosedAt = &t
		}
	}
	if ji.Fields.Creator != nil {
		id := hashUsername(ji.Fields.Creator.Name)
		issue.CreatorID = &id
	}
	if ji.Fields.Assignee != nil {
		id := hashUsername(ji.Fields.Assignee.Name)
		issue.AssigneeID = &id
	}

	var changes []models.StatusChange
	for _, h := range ji.ChangeLog.Histories {
		for _, item := range h.Items {
			if item.Field != "status" {
				continue
			}
			changeTime, err := parseJiraTime(h.Created)
			if err != nil {
				continue
			}
			sc := models.StatusChange{
				IssueID:    jiraID,
				ChangeTime: changeTime,
			}
			if item.FromString != "" {
				s := item.FromString
				sc.OldStatus = &s
			}
			if item.ToString != "" {
				s := item.ToString
				sc.NewStatus = &s
			}
			changes = append(changes, sc)
		}
	}

	return issue, changes, nil
}

func parseJiraTime(s string) (time.Time, error) {
	return time.Parse(jiraTimeLayout, s)
}

func hashUsername(username string) int64 {
	h := fnv.New64a()
	h.Write([]byte(username))
	return int64(h.Sum64() & 0x7FFFFFFFFFFFFFFF)
}
