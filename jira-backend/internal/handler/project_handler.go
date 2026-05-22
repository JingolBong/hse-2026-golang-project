package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/jira-backend/internal/service"
)

const defaultPageLimit = 10

type ProjectHandler struct {
	service *service.ProjectService
	log     *logrus.Logger
}

func NewProjectHandler(s *service.ProjectService, log *logrus.Logger) *ProjectHandler {
	return &ProjectHandler{
		service: s,
		log: log,
	}
}

func (h *ProjectHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load projects")
		return
	}

	if err := writeData(w, http.StatusOK, projectsFromModels(data)); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", defaultPageLimit)
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	result, err := h.service.GetCatalog(r.Context(), page, limit, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load catalog")
		return
	}

	pageInfo := &PageInfo{
		CurrentPage:   result.CurrentPage,
		PageCount:     result.PageCount,
		ProjectsCount: result.TotalCount,
	}
	if err := writeDataPaged(w, http.StatusOK, projectsFromCatalog(result.Projects), pageInfo); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func queryInt(r *http.Request, key string, def int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || v < 1 {
		return def
	}
	return v
}

type statView struct {
	Id                  int64   `json:"Id"`
	Key                 string  `json:"Key"`
	Name                string  `json:"Name"`
	AllIssuesCount      int     `json:"allIssuesCount"`
	OpenIssuesCount     int     `json:"openIssuesCount"`
	CloseIssuesCount    int     `json:"closeIssuesCount"`
	ReopenedIssuesCount int     `json:"reopenedIssuesCount"`
	ResolvedIssuesCount int     `json:"resolvedIssuesCount"`
	ProgressIssuesCount int     `json:"progressIssuesCount"`
	AverageTime         float64 `json:"averageTime"`
	AverageIssuesCount  string  `json:"averageIssuesCount"`
}

func (h *ProjectHandler) Stat(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	stat, err := h.service.GetStat(r.Context(), id)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load project stat")
		return
	}

	view := statView{
		Id:                  stat.ID,
		Key:                 stat.Key,
		Name:                stat.Name,
		AllIssuesCount:      stat.AllIssuesCount,
		OpenIssuesCount:     stat.OpenIssuesCount,
		CloseIssuesCount:    stat.CloseIssuesCount,
		ReopenedIssuesCount: stat.ReopenedIssuesCount,
		ResolvedIssuesCount: stat.ResolvedIssuesCount,
		ProgressIssuesCount: stat.ProgressIssuesCount,
		AverageTime:         stat.AverageTime,
		AverageIssuesCount:  stat.AverageIssuesCount,
	}
	if err := writeData(w, http.StatusOK, view); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	err = h.service.Delete(r.Context(), id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete project")
		return
	}

	if err := writeData(w, http.StatusOK, nil); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	key := projectKeyFromRequest(r)
	if key == "" {
		writeError(w, http.StatusBadRequest, "[project key is required]")
		return
	}

	if err := h.service.Update(r.Context(), key); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update project")
		return
	}

	if err := writeData(w, http.StatusOK, map[string]string{"project": key}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
