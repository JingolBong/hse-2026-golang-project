package connector

import (
	"net/http"

	"hse-2026-golang-project/internal/config"
	"hse-2026-golang-project/internal/db"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	storage *db.Storage
	client  *JiraClient
	cfg     config.ProgramSettings
	log     *logrus.Logger
}

func NewHandler(storage *db.Storage, client *JiraClient, cfg config.ProgramSettings, log *logrus.Logger) http.Handler {
	h := &Handler{
		storage: storage,
		client:  client,
		cfg:     cfg,
		log:     log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/projects", h.handleGetProjects)
	mux.HandleFunc("/updateProject", h.handleUpdateProject)
	mux.HandleFunc("/health", h.handleHealth)

	return mux
}
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request)
func (h *Handler) handleUpdateProject(w http.ResponseWriter, r *http.Request)
func (h *Handler) handleGetProjects(w http.ResponseWriter, r *http.Request)
