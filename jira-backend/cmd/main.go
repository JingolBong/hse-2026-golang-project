package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	connector "hse-2026-golang-project/internal/jira"
	pb "hse-2026-golang-project/internal/proto/connector"

	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/jira-backend/internal/app"
	"hse-2026-golang-project/jira-backend/internal/handler"
	"hse-2026-golang-project/jira-backend/internal/repository"
	"hse-2026-golang-project/jira-backend/internal/service"
)

func main() {
	logger := connector.NewLogger()

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
	defer func() {
		if err := storage.Close(); err != nil {
			logger.Printf("close db connections: %v", err)
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

	router := app.NewRouter(projectHandler, issueHandler, graphHandler, corsOrigin)

	logger.Println("Server started on :8000")
	logger.Fatal(http.ListenAndServe(":8000", router))
}
