package connector

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"hse-2026-golang-project/internal/config"
	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/internal/models"
	pb "hse-2026-golang-project/internal/proto/connector"
)

type GRPCServer struct {
	pb.UnimplementedConnectorServiceServer
	storage    *db.Storage
	client     *JiraClient
	cfg        config.ProgramSettings
	log        *logrus.Logger
	producer   sarama.SyncProducer
	kafkaTopic string
}

func NewGRPCServer(storage *db.Storage, client *JiraClient, cfg config.ProgramSettings, log *logrus.Logger, producer sarama.SyncProducer, topic string) *GRPCServer {
	return &GRPCServer{
		storage:    storage,
		client:     client,
		cfg:        cfg,
		log:        log,
		producer:   producer,
		kafkaTopic: topic,
	}
}

func (s *GRPCServer) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	h, err := s.storage.HealthCheck(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "health check failed: %v", err)
	}

	return &pb.HealthResponse{
		MasterUp:          h.MasterUp,
		ReplicaUp:         h.ReplicaUp,
		MasterInRecovery:  h.MasterRecovery,
		ReplicaInRecovery: h.ReplicaRecovery,
	}, nil
}

func (s *GRPCServer) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	limit := int(req.Limit)
	page := int(req.Page)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	allProjects, err := s.client.GetProjects(ctx)
	if err != nil {
		s.log.WithError(err).Error("failed to get projects")
		return nil, status.Errorf(codes.Unavailable, "jira unavaliable: %v", err)
	}

	search := strings.ToLower(req.Search)
	var filtered []JiraProject
	for _, p := range allProjects {
		if search == "" || strings.Contains(strings.ToLower(p.Name), search) || strings.Contains(strings.ToLower(p.Key), search) {
			filtered = append(filtered, p)
		}
	}

	total := len(filtered)
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	if page < 1 {
		page = 1
	}

	start := (page - 1) * limit
	if start < 0 {
		start = 0
	}
	end := start + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	var dtos []*pb.ProjectDTO
	for _, p := range filtered[start:end] {
		dtos = append(dtos, &pb.ProjectDTO{Key: p.Key, Name: p.Name, Url: p.Self})
	}

	return &pb.GetProjectsResponse{
		Projects: dtos,
		PageInfo: &pb.PageInfo{
			CurrentPage:   int32(page),
			ProjectsCount: int32(total),
			TotalPages:    int32(totalPages),
		},
	}, nil
}

func (s *GRPCServer) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	if req.ProjectKey == "" {
		return nil, status.Error(codes.InvalidArgument, "project_key is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	s.log.WithField("project", req.ProjectKey).Info("UpdateProject started")

	projects, err := s.client.GetProjects(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "jira unavailable: %v", err)
	}

	var jiraProject *JiraProject
	for i := range projects {
		if strings.EqualFold(projects[i].Key, req.ProjectKey) {
			jiraProject = &projects[i]
			break
		}
	}

	if jiraProject == nil {
		return nil, status.Errorf(codes.NotFound, "project %q not found in jira", req.ProjectKey)
	}

	projectID := hashUsername(jiraProject.ID)

	if _, err := s.storage.UpsertProject(ctx, models.Project{
		JiraID: projectID,
		Key:    jiraProject.Key,
		Name:   jiraProject.Name,
		URL:    jiraProject.Self,
	}); err != nil {
		s.log.WithError(err).Error("failed to upsert project")
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	if err := LoadProject(ctx, s.storage, s.client, req.ProjectKey, projectID, s.cfg, s.log); err != nil {
		s.log.WithError(err).Error("LoadProject failed")
		return nil, status.Errorf(codes.Internal, "load failed: %v", err)
	}

	s.log.Info("Sending notification to Kafka...")

	kafkaMessage := map[string]interface{}{
		"event":     "project_updated",
		"project":   req.ProjectKey,
		"status":    "success",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	msgBytes, _ := json.Marshal(kafkaMessage)
	msg := &sarama.ProducerMessage{
		Topic: s.kafkaTopic,
		Value: sarama.ByteEncoder(msgBytes),
	}

	if _, _, err := s.producer.SendMessage(msg); err != nil {
		s.log.WithError(err).Warn("Failed to send message to Kafka, but ETL succeeded")
	} else {
		s.log.Info("Successfully pushed ETL result to Kafka!")
	}

	s.log.WithField("project", req.ProjectKey).Info("UpdateProject completed")

	return &pb.UpdateProjectResponse{
		Status:  "ok",
		Project: req.ProjectKey,
	}, nil
}

func (s *GRPCServer) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	if req.ProjectId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid project_id")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.storage.DeleteProject(ctx, req.ProjectId); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "project %d not found", req.ProjectId)
		}
		s.log.WithError(err).Error("delete project failed")
		return nil, status.Errorf(codes.Internal, "db error: %v", err)
	}

	return &pb.DeleteProjectResponse{Status: "ok"}, nil
}
