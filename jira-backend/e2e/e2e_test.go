//go:build integration

// End-to-end test of the Go stack (no frontend):
//
//	REST (jira-backend router) -> gRPC (connector over bufconn) -> Postgres
//	(testcontainers) + a stubbed Jira REST API (httptest). Kafka is replaced
//	by a no-op producer.
//
// Run with: go test -tags=integration ./test/e2e/   (requires Docker)
package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/IBM/sarama"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"hse-2026-golang-project/internal/config"
	"hse-2026-golang-project/internal/db"
	connector "hse-2026-golang-project/internal/jira"
	pb "hse-2026-golang-project/internal/proto/connector"
	"hse-2026-golang-project/jira-backend/internal/app"
	"hse-2026-golang-project/jira-backend/internal/handler"
	"hse-2026-golang-project/jira-backend/internal/repository"
	"hse-2026-golang-project/jira-backend/internal/service"
)

const migrationFile = "../../migrations/01_create_tables.sql"

// --- no-op Kafka producer (the REST flow doesn't assert on Kafka here) ---

type noopProducer struct{}

func (noopProducer) SendMessage(*sarama.ProducerMessage) (int32, int64, error) { return 0, 0, nil }
func (noopProducer) SendMessages([]*sarama.ProducerMessage) error              { return nil }
func (noopProducer) Close() error                                              { return nil }
func (noopProducer) TxnStatus() sarama.ProducerTxnStatusFlag                   { return 0 }
func (noopProducer) IsTransactional() bool                                     { return false }
func (noopProducer) BeginTxn() error                                           { return nil }
func (noopProducer) CommitTxn() error                                          { return nil }
func (noopProducer) AbortTxn() error                                           { return nil }
func (noopProducer) AddOffsetsToTxn(map[string][]*sarama.PartitionOffsetMetadata, string) error {
	return nil
}
func (noopProducer) AddMessageToTxn(*sarama.ConsumerMessage, string, *string) error { return nil }

// --- stub Jira REST API ---

func stubJira(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/api/2/project", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"10000","key":"PRJ","name":"Project","self":"http://stub/PRJ"}]`))
	})
	mux.HandleFunc("/rest/api/2/search", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"startAt":0,"maxResults":50,"total":1,
			"issues":[{
				"id":"5001","key":"PRJ-1",
				"fields":{
					"summary":"Bug","status":{"name":"Open"},"priority":{"name":"Major"},
					"created":"2025-01-01T00:00:00.000+0000"
				}
			}]
		}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func testLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	return l
}

// --- infrastructure: Postgres via testcontainers ---

func startPostgres(t *testing.T) *db.Storage {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"postgres:13.8",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, ctr)

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	sqlDB, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.Eventually(t, func() bool { return sqlDB.Ping() == nil }, 30*time.Second, time.Second)

	schema, err := os.ReadFile(migrationFile)
	require.NoError(t, err)
	_, err = sqlDB.Exec(string(schema))
	require.NoError(t, err)

	return db.NewStorage(sqlDB, sqlDB)
}

// --- connector gRPC server over bufconn -> returns a real client ---

func startConnector(t *testing.T, storage *db.Storage, jiraURL string) pb.ConnectorServiceClient {
	t.Helper()

	cfg := config.ProgramSettings{
		JiraURL:           jiraURL,
		ThreadCount:       2,
		IssueInOneRequest: 50,
		MinTimeSleep:      1,
		MaxTimeSleep:      4,
	}
	jiraClient := connector.NewJiraClient(cfg, testLogger())
	srv := connector.NewGRPCServer(storage, jiraClient, cfg, testLogger(), noopProducer{}, "topic")

	lis := bufconn.Listen(1024 * 1024)
	gsrv := grpc.NewServer()
	pb.RegisterConnectorServiceServer(gsrv, srv)
	go func() { _ = gsrv.Serve(lis) }()
	t.Cleanup(gsrv.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return pb.NewConnectorServiceClient(conn)
}

// --- the real jira-backend HTTP server ---

func startBackend(t *testing.T, storage *db.Storage, grpcClient pb.ConnectorServiceClient) *httptest.Server {
	t.Helper()
	log := testLogger()

	repo := repository.NewProjectRepository(storage)
	projectH := handler.NewProjectHandler(service.NewProjectService(repo, grpcClient, log), log)
	issueH := handler.NewIssueHandler(service.NewIssueService(repo))
	graphH := handler.NewGraphHandler(service.NewGraphService(repo))

	router := app.NewRouter(projectH, issueH, graphH, "*")
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

// --- response helpers ---

type envelope struct {
	Data     json.RawMessage `json:"data"`
	Status   bool            `json:"status"`
	PageInfo *struct {
		ProjectsCount int `json:"projectsCount"`
	} `json:"pageInfo"`
}

type projectView struct {
	Existence bool   `json:"Existence"`
	Id        int64  `json:"Id"`
	Key       string `json:"Key"`
}

func itoa(v int64) string { return strconv.FormatInt(v, 10) }

func getJSON(t *testing.T, client *http.Client, url string) envelope {
	t.Helper()
	resp, err := client.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "GET %s", url)
	var e envelope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&e))
	return e
}

func TestE2E_UpdateReadDeleteFlow(t *testing.T) {
	storage := startPostgres(t)
	jira := stubJira(t)
	grpcClient := startConnector(t, storage, jira.URL)
	backend := startBackend(t, storage, grpcClient)
	client := backend.Client()

	// 1. Trigger ETL: REST -> gRPC UpdateProject -> stub Jira -> Postgres.
	resp, err := client.Post(backend.URL+"/api/v1/projects/PRJ/update", "application/json", nil)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "update should succeed")

	// 2. The project is now persisted and visible via REST.
	env := getJSON(t, client, backend.URL+"/api/v1/myprojects")
	var projects []projectView
	require.NoError(t, json.Unmarshal(env.Data, &projects))
	require.Len(t, projects, 1)
	assert.Equal(t, "PRJ", projects[0].Key)
	projectID := projects[0].Id
	require.NotZero(t, projectID)

	// 3. Stats reflect the single loaded issue.
	statEnv := getJSON(t, client, backend.URL+"/api/v1/myprojects/"+itoa(projectID)+"/stat")
	var stat struct {
		AllIssuesCount  int `json:"allIssuesCount"`
		OpenIssuesCount int `json:"openIssuesCount"`
	}
	require.NoError(t, json.Unmarshal(statEnv.Data, &stat))
	assert.Equal(t, 1, stat.AllIssuesCount)
	assert.Equal(t, 1, stat.OpenIssuesCount)

	// 4. Issues endpoint returns the loaded issue.
	issuesEnv := getJSON(t, client, backend.URL+"/api/v1/projects/PRJ/issues")
	var issues []map[string]interface{}
	require.NoError(t, json.Unmarshal(issuesEnv.Data, &issues))
	require.Len(t, issues, 1)

	// 5. Catalog (gRPC GetProjects from stub Jira) marks PRJ as already saved.
	catEnv := getJSON(t, client, backend.URL+"/api/v1/projects?page=1&limit=10")
	var catalog []projectView
	require.NoError(t, json.Unmarshal(catEnv.Data, &catalog))
	require.Len(t, catalog, 1)
	assert.True(t, catalog[0].Existence, "PRJ should be flagged as existing")

	// 6. Delete: REST -> gRPC DeleteProject -> Postgres cascade.
	delReq, _ := http.NewRequest(http.MethodDelete, backend.URL+"/api/v1/projects/"+itoa(projectID), nil)
	delResp, err := client.Do(delReq)
	require.NoError(t, err)
	delResp.Body.Close()
	require.Equal(t, http.StatusOK, delResp.StatusCode)

	// 7. Project is gone.
	afterEnv := getJSON(t, client, backend.URL+"/api/v1/myprojects")
	var after []projectView
	require.NoError(t, json.Unmarshal(afterEnv.Data, &after))
	assert.Empty(t, after)
}
