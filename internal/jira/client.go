package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"hse-2026-golang-project/internal/config"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const jiraTimeLayout = "2006-01-02T15:04:05.999-0700"

type JiraClient struct {
	baseURL    string
	httpClient *http.Client
	minSleep   time.Duration
	maxSleep   time.Duration
	log        *logrus.Logger
}

func NewJiraClient(cfg config.ProgramSettings, log *logrus.Logger) *JiraClient {
	return &JiraClient{
		baseURL: cfg.JiraURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		minSleep: time.Duration(cfg.MinTimeSleep) * time.Millisecond,
		maxSleep: time.Duration(cfg.MaxTimeSleep) * time.Millisecond,
		log:      log,
	}
}

func sleepContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func (c *JiraClient) doRequest(ctx context.Context, url string, target interface{}) error {
	wait := c.minSleep
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(request)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if wait > c.maxSleep {
				return fmt.Errorf("network error after retries: %w", err)
			}
			c.log.WithField(logrus.Fields{
				"url":     url,
				"wait_ms": wait.Milliseconds(),
				"error":   err.Error(),
			}).Warn("network error, retrying")

			if err := sleepContext(ctx, wait); err != nil {
				return err
			}
			wait *= 2
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			resp.Body.Close()
			if wait > c.maxSleep {
				return fmt.Errorf("jira returned %d after all retries for %s", resp.StatusCode, url)
			}
			c.log.WithField(logrus.Fields{
				"url":         url,
				"wait_ms":     wait.Milliseconds(),
				"status_code": resp.StatusCode,
			}).Warn("jira rate limit or server error, retrying")

			if err := sleepContext(ctx, wait); err != nil {
				return err
			}
			wait *= 2
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
		}

		err = json.NewDecoder(resp.Body).Decode(target)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("decode response from %s: %w", url, err)
		}
		return nil
	}
}

func (c *JiraClient) GetProjects(ctx context.Context) ([]JiraProject, error) {
	url := c.baseURL + "/rest/api/2/project"

	c.log.WithField("url", url).Info("fetching projects from jira")
	var projects []JiraProject
	if err := c.doRequest(ctx, url, &projects); err != nil {
		return nil, fmt.Errorf("fetch projects: %w", err)
	}

	c.log.WithField("count", len(projects)).Info("projects fetched")
	return projects, nil
}

func (c *JiraClient) FetchIssuesPage(
	ctx context.Context,
	projectKey string,
	startAt, maxResults int,
) (*SearchResponse, error) {
	url := fmt.Sprintf(
		"%s/rest/api/2/search?jql=project=%s&startAt=%d&maxResults=%d&expand=changelog",
		c.baseURL, projectKey, startAt, maxResults,
	)
	c.log.WithFields(logrus.Fields{
		"project":   projectKey,
		"start_at":  startAt,
		"page_size": maxResults,
	}).Info("fetching issues page from jira")

	var result SearchResponse
	if err := c.doRequest(ctx, url, &result); err != nil {
		return nil, fmt.Errorf("fetch issues page: %w", err)
	}

	return &result, nil
}
