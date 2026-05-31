package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "hse-2026-golang-project/internal/proto/connector"

	"hse-2026-golang-project/internal/config"
	"hse-2026-golang-project/internal/db"
	applog "hse-2026-golang-project/internal/logger"
	"hse-2026-golang-project/jira-backend/internal/app"
	"hse-2026-golang-project/jira-backend/internal/handler"
	"hse-2026-golang-project/jira-backend/internal/repository"
	"hse-2026-golang-project/jira-backend/internal/service"
)

func main() {
	cfg, err := config.LoadConfig("configs")
	if err != nil {
		applog.New(applog.Options{Service: "backend"}).Fatalf("load config: %v", err)
	}

	logger := applog.New(applog.Options{
		Level:   cfg.Log.Level,
		Service: "backend",
		ToFile:  cfg.Log.ToFile,
	})

	dsn := "postgres://pguser:pgpwd@postgres-master:5432/testdb?sslmode=disable"

	writeDB, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatalf("open master db: %v", err)
	}
	readDB, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Fatalf("open replica db: %v", err)
	}

	storage := db.NewStorage(writeDB, readDB)
	storage.SetLogger(logger)
	defer func() {
		if err := storage.Close(); err != nil {
			logger.WithError(err).Warn("close db connections")
		}
	}()

	repo := repository.NewProjectRepository(storage)

	connectorAddress := "connector:8001"
	conn, err := grpc.NewClient(connectorAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("create grpc connection: %v", err)
	}
	defer conn.Close()
	grpcClient := pb.NewConnectorServiceClient(conn)

	projectService := service.NewProjectService(repo, grpcClient, logger)
	issueService := service.NewIssueService(repo)
	graphService := service.NewGraphService(repo)

	projectHandler := handler.NewProjectHandler(projectService, logger)
	issueHandler := handler.NewIssueHandler(issueService)
	graphHandler := handler.NewGraphHandler(graphService)

	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:4200"
	}

	router := app.NewRouter(projectHandler, issueHandler, graphHandler, corsOrigin, logger)

	logger.Info("Server started on :8000")
	logger.Fatal(http.ListenAndServe(":8000", router))
}
