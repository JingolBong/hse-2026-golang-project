package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"hse-2026-golang-project/jira-backend/internal/service"
)

type GraphHandler struct {
	service graphService
	links   *LinkBuilder
}

func NewGraphHandler(s graphService, builders ...*LinkBuilder) *GraphHandler {
	return &GraphHandler{
		service: s,
		links:   linkBuilderOrDefault(builders),
	}
}

func (h *GraphHandler) Make(w http.ResponseWriter, r *http.Request) {
	links := h.links.AnalyticsLinks(r)
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, r, http.StatusBadRequest, "project key is required", links)
		return
	}

	task, err := parseTask(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid task", links)
		return
	}

	empty, err := h.service.IsEmpty(r.Context(), project)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, r, http.StatusNotFound, "project not found", links)
		return
	}
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to check project data", links)
		return
	}
	if empty {
		writeError(w, r, http.StatusUnprocessableEntity, "project is empty", links)
		return
	}

	err = h.service.Make(r.Context(), project, task)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, r, http.StatusNotFound, "project not found", links)
		return
	}
	if errors.Is(err, service.ErrUnsupportedTask) {
		writeError(w, r, http.StatusBadRequest, "unsupported task", links)
		return
	}
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to prepare graph data", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, nil, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) Get(w http.ResponseWriter, r *http.Request) {
	links := h.links.AnalyticsLinks(r)
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, r, http.StatusBadRequest, "project key is required", links)
		return
	}

	task, err := parseTask(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid task", links)
		return
	}

	data, err := h.service.Get(r.Context(), project, task)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, r, http.StatusNotFound, "project not found", links)
		return
	}
	if errors.Is(err, service.ErrUnsupportedTask) {
		writeError(w, r, http.StatusBadRequest, "unsupported task", links)
		return
	}
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to load graph data", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, data, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) Compare(w http.ResponseWriter, r *http.Request) {
	links := h.links.AnalyticsLinks(r)
	task, err := parseTask(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid task", links)
		return
	}

	projectsStr := r.URL.Query().Get("project")
	if projectsStr == "" {
		writeError(w, r, http.StatusBadRequest, "projects are required", links)
		return
	}
	keys := strings.Split(projectsStr, ",")

	data, err := h.service.Compare(r.Context(), keys, task)
	if errors.Is(err, service.ErrUnsupportedTask) {
		writeError(w, r, http.StatusBadRequest, "unsupported task", links)
		return
	}
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to load compare data", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, data, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) IsAnalyzed(w http.ResponseWriter, r *http.Request) {
	links := h.links.AnalyticsLinks(r)
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, r, http.StatusBadRequest, "project key is required", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, map[string]bool{"isAnalyzed": h.service.IsAnalyzed(project)}, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) DeleteGraphs(w http.ResponseWriter, r *http.Request) {
	links := h.links.AnalyticsLinks(r)
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, r, http.StatusBadRequest, "project key is required", links)
		return
	}

	h.service.DropAnalyzed(project)

	if err := writeData(w, r, http.StatusOK, nil, links); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func parseTask(r *http.Request) (int, error) {
	raw := mux.Vars(r)["task"]
	if raw == "" {
		raw = r.URL.Query().Get("task")
	}
	return strconv.Atoi(raw)
}
