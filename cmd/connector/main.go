package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"hse-2026-golang-project/internal/config"
	"hse-2026-golang-project/internal/db"
	connector "hse-2026-golang-project/internal/jira"
	pb "hse-2026-golang-project/internal/proto/connector"
)

func main() {
	logger := connector.NewLogger()
	logger.Println("Starting Jira Connector Service...")

	cfg, err := config.LoadConfig("configs")
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Printf("Connecting to Write DB at %s:%d...", cfg.Jira.WriteDB.Host, cfg.Jira.WriteDB.Port)
	writeDB, err := config.NewDB(ctx, cfg.Jira.WriteDB)
	if err != nil {
		logger.Fatalf("Write DB connection failed: %v", err)
	}
	defer writeDB.Close()
	logger.Println("Write DB connected successfully!")

	logger.Printf("Connecting to Read DB at %s:%d...", cfg.Jira.ReadDB.Host, cfg.Jira.ReadDB.Port)
	readDB, err := config.NewDB(ctx, cfg.Jira.ReadDB)
	if err != nil {
		logger.Fatalf("Read DB connection failed: %v", err)
	}
	defer readDB.Close()
	logger.Println("Read DB connected successfully!")

	logger.Printf("Initializing Kafka Producer (Brokers: %v, Topic: %s)...", cfg.Kafka.Brokers, cfg.Kafka.Topic)
	kafkaCfg := sarama.NewConfig()
	kafkaCfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, kafkaCfg)
	if err != nil {
		logger.Fatalf("Failed to start Sarama producer: %v", err)
	}
	defer producer.Close()

	storage := db.NewStorage(writeDB, readDB)
	jiraClient := connector.NewJiraClient(cfg.Jira.Program, logger)

	port := fmt.Sprintf(":%d", cfg.Jira.Program.Port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()

	connectorService := connector.NewGRPCServer(storage, jiraClient, cfg.Jira.Program, logger, producer, cfg.Kafka.Topic)
	pb.RegisterConnectorServiceServer(grpcServer, connectorService)

	reflection.Register(grpcServer)

	go func() {
		logger.Printf("gRPC server listening on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	logger.Println("Service is up and running!")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down service...")
	grpcServer.GracefulStop()
	logger.Println("Done.")
}
