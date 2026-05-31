package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

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

func timeoutFromSeconds(seconds, def int) time.Duration {
	if seconds <= 0 {
		seconds = def
	}
	return time.Duration(seconds) * time.Second
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

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
	graphService := service.NewGraphService(repo, logger)

	bindAddress := cfg.Server.BindAddress
	if bindAddress == "" {
		bindAddress = "0.0.0.0"
	}
	bindPort := cfg.Server.Port
	if bindPort == 0 {
		bindPort = 8000
	}
	linkBuilder := handler.NewLinkBuilder(bindAddress, bindPort)

	projectHandler := handler.NewProjectHandler(projectService, logger, linkBuilder)
	issueHandler := handler.NewIssueHandler(issueService, linkBuilder)
	graphHandler := handler.NewGraphHandler(graphService, linkBuilder)
	systemHandler := handler.NewSystemHandler(linkBuilder)

	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:4200"
	}

	resourceTimeout := timeoutFromSeconds(cfg.Server.ResourseTimeout, 5)
	analyticsTimeout := timeoutFromSeconds(cfg.Server.AnalyticsTimeout, 15)
	router := app.NewRouter(projectHandler, issueHandler, graphHandler, systemHandler, corsOrigin, logger, resourceTimeout, analyticsTimeout)

	address := fmt.Sprintf("%s:%d", bindAddress, bindPort)
	serverTimeout := maxDuration(resourceTimeout, analyticsTimeout)
	server := &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: resourceTimeout,
		ReadTimeout:       serverTimeout,
		WriteTimeout:      serverTimeout,
		IdleTimeout:       60 * time.Second,
	}

	logger.Printf("Server started on %s", address)
	logger.Fatal(server.ListenAndServe())
}
