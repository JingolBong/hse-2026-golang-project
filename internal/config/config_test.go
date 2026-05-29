package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleConfig = `
jira:
  program:
    jiraUrl: "http://jira.local"
    threadCount: 4
    issueInOneRequest: 50
    minTimeSleep: 100
    maxTimeSleep: 2000
    port: 8001
  writeDB:
    user: master
    password: secret
    host: db-master
    port: 5432
    database: testdb
    sslmode: disable
  readDB:
    user: replica
    password: secret
    host: db-replica
    port: 5433
    database: testdb
    sslmode: disable
kafka:
  brokers:
    - kafka:9092
  topic: events
server:
  port: 8000
`

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(body), 0o644))
	return dir
}

func TestLoadConfig_ParsesAllSections(t *testing.T) {
	viper.Reset()
	dir := writeConfig(t, sampleConfig)

	cfg, err := LoadConfig(dir)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "http://jira.local", cfg.Jira.Program.JiraURL)
	assert.Equal(t, 4, cfg.Jira.Program.ThreadCount)
	assert.Equal(t, 50, cfg.Jira.Program.IssueInOneRequest)
	assert.Equal(t, 8001, cfg.Jira.Program.Port)

	assert.Equal(t, "db-master", cfg.Jira.WriteDB.Host)
	assert.Equal(t, 5432, cfg.Jira.WriteDB.Port)
	assert.Equal(t, "disable", cfg.Jira.WriteDB.SSLMode)
	assert.Equal(t, "db-replica", cfg.Jira.ReadDB.Host)
	assert.Equal(t, 5433, cfg.Jira.ReadDB.Port)

	require.Len(t, cfg.Kafka.Brokers, 1)
	assert.Equal(t, "kafka:9092", cfg.Kafka.Brokers[0])
	assert.Equal(t, "events", cfg.Kafka.Topic)
	assert.Equal(t, 8000, cfg.Server.Port)
}

func TestLoadConfig_MissingFileReturnsZeroConfig(t *testing.T) {
	viper.Reset()
	// Empty dir: ReadInConfig fails but LoadConfig only logs a warning and
	// returns a zero-valued config without error.
	cfg, err := LoadConfig(t.TempDir())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 0, cfg.Server.Port)
	assert.Empty(t, cfg.Jira.Program.JiraURL)
}
