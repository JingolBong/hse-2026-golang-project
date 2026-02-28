package connector

import (
	"context"
	"hash/fnv"
	"time"

	"hse-2026-golang-project/internal/config"
	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/internal/models"

	"github.com/sirupsen/logrus"
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
}
func transformUser(u User) models.Author                                                        {}
func transformIssue(ji JiraIssue, projectID int64) (models.Issue, []models.StatusChange, error) {}
func parseJiraTime(s string) (time.Time, error) {
	return time.Parse(jiraTimeLayout, s)
}
func hashUsername(username string) int64 {
	h := fnv.New64a()
	h.Write([]byte(username))
	return int64(h.Sum64() & 0x7FFFFFFFFFFFFFFF)
}
