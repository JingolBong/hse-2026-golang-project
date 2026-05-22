package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"hse-2026-golang-project/jira-backend/internal/service"
)

type IssueHandler struct {
	service *service.IssueService
}

func NewIssueHandler(s *service.IssueService) *IssueHandler {
	return &IssueHandler{service: s}
}

func (h *IssueHandler) GetByProject(w http.ResponseWriter, r *http.Request) {
	key := projectKeyFromRequest(r)
	if key == "" {
		writeError(w, http.StatusBadRequest, "project key is required")
		return
	}

	data, err := h.service.GetByProjectKey(r.Context(), key)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load issues")
		return
	}

	if err := writeData(w, http.StatusOK, data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func projectKeyFromRequest(r *http.Request) string {
	key := strings.TrimSpace(r.URL.Query().Get("project"))
	if key != "" {
		return key
	}

	return strings.TrimSpace(mux.Vars(r)["project"])
}
