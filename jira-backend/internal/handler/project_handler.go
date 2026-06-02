package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"hse-2026-golang-project/internal/db"
	"hse-2026-golang-project/jira-backend/internal/reqid"
	"hse-2026-golang-project/jira-backend/internal/service"
)

const defaultPageLimit = 10

type ProjectHandler struct {
	service projectService
	log     *logrus.Logger
	links   *LinkBuilder
}

func NewProjectHandler(s projectService, log *logrus.Logger, builders ...*LinkBuilder) *ProjectHandler {
	return &ProjectHandler{
		service: s,
		log:     log,
		links:   linkBuilderOrDefault(builders),
	}
}

func (h *ProjectHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	links := h.links.ResourceLinks(r)
	data, err := h.service.GetAll(r.Context())
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to load projects", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, projectsFromModels(r, data, h.links), links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	links := h.links.ConnectorLinks(r)
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", defaultPageLimit)
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	h.log.WithFields(logrus.Fields{
		"request_id": reqid.FromContext(r.Context()),
		"page":       page,
		"limit":      limit,
		"search":     search,
	}).Debug("GetCatalog handler")

	result, err := h.service.GetCatalog(r.Context(), page, limit, search)
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to load catalog", links)
		return
	}

	pageInfo := &PageInfo{
		CurrentPage:   result.CurrentPage,
		PageCount:     result.PageCount,
		ProjectsCount: result.TotalCount,
	}
	if err := writeDataPaged(w, r, http.StatusOK, projectsFromCatalog(r, result.Projects, h.links), pageInfo, links); err != nil {
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
	links := h.links.ResourceLinks(r)
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid project id", links)
		return
	}

	stat, err := h.service.GetStat(r.Context(), id)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, r, http.StatusNotFound, "project not found", links)
		return
	}
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to load project stat", links)
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
	if err := writeData(w, r, http.StatusOK, view, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	links := h.links.ResourceLinks(r)
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id <= 0 {
		writeError(w, r, http.StatusBadRequest, "invalid project id", links)
		return
	}

	err = h.service.Delete(r.Context(), id)
	if errors.Is(err, db.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, "project not found", links)
		return
	}
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to delete project", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, nil, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	links := h.links.ConnectorLinks(r)
	key := projectKeyFromRequest(r)
	if key == "" {
		writeError(w, r, http.StatusBadRequest, "[project key is required]", links)
		return
	}

	h.log.WithFields(logrus.Fields{
		"request_id": reqid.FromContext(r.Context()),
		"project":    key,
	}).Debug("Update handler")

	if err := h.service.Update(r.Context(), key); err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to update project", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, map[string]string{"project": key}, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
