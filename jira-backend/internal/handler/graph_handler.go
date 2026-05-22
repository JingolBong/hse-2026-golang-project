package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"hse-2026-golang-project/jira-backend/internal/service"
)

type GraphHandler struct {
	service *service.GraphService
}

func NewGraphHandler(s *service.GraphService) *GraphHandler {
	return &GraphHandler{service: s}
}

func (h *GraphHandler) Make(w http.ResponseWriter, r *http.Request) {
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, http.StatusBadRequest, "project key is required")
		return
	}

	task, err := parseTask(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task")
		return
	}

	err = h.service.Make(r.Context(), project, task)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if errors.Is(err, service.ErrUnsupportedTask) {
		writeError(w, http.StatusBadRequest, "unsupported task")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare graph data")
		return
	}

	if err := writeData(w, http.StatusOK, nil); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) Get(w http.ResponseWriter, r *http.Request) {
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, http.StatusBadRequest, "project key is required")
		return
	}

	task, err := parseTask(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task")
		return
	}

	data, err := h.service.Get(r.Context(), project, task)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if errors.Is(err, service.ErrUnsupportedTask) {
		writeError(w, http.StatusBadRequest, "unsupported task")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load graph data")
		return
	}

	if err := writeData(w, http.StatusOK, data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) Compare(w http.ResponseWriter, r *http.Request) {
	task, err := strconv.Atoi(r.URL.Query().Get("task"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task")
		return
	}

	keys := r.URL.Query()["projects"]
	if len(keys) == 0 {
		writeError(w, http.StatusBadRequest, "projects are required")
		return
	}

	data, err := h.service.Compare(r.Context(), keys, task)
	if errors.Is(err, service.ErrUnsupportedTask) {
		writeError(w, http.StatusBadRequest, "unsupported task")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load compare data")
		return
	}

	if err := writeData(w, http.StatusOK, data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) IsAnalyzed(w http.ResponseWriter, r *http.Request) {
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, http.StatusBadRequest, "project key is required")
		return
	}

	if err := writeData(w, http.StatusOK, map[string]bool{"isAnalyzed": h.service.IsAnalyzed(project)}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) IsEmpty(w http.ResponseWriter, r *http.Request) {
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, http.StatusBadRequest, "project key is required")
		return
	}

	empty, err := h.service.IsEmpty(r.Context(), project)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to check project")
		return
	}

	if err := writeData(w, http.StatusOK, map[string]bool{"isEmpty": empty}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *GraphHandler) DeleteGraphs(w http.ResponseWriter, r *http.Request) {
	project := projectKeyFromRequest(r)
	if project == "" {
		writeError(w, http.StatusBadRequest, "project key is required")
		return
	}

	h.service.DropAnalyzed(project)

	if err := writeData(w, http.StatusOK, nil); err != nil {
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
