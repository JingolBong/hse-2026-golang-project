package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"hse-2026-golang-project/jira-backend/internal/service"
)

type IssueHandler struct {
	service issueService
	links   *LinkBuilder
}

func NewIssueHandler(s issueService, builders ...*LinkBuilder) *IssueHandler {
	return &IssueHandler{
		service: s,
		links:   linkBuilderOrDefault(builders),
	}
}

func (h *IssueHandler) GetByProject(w http.ResponseWriter, r *http.Request) {
	links := h.links.ResourceLinks(r)
	key := projectKeyFromRequest(r)
	if key == "" {
		writeError(w, r, http.StatusBadRequest, "project key is required", links)
		return
	}

	data, err := h.service.GetByProjectKey(r.Context(), key)
	if errors.Is(err, service.ErrProjectNotFound) {
		writeError(w, r, http.StatusNotFound, "project not found", links)
		return
	}
	if err != nil {
		if isTimeoutError(err) {
			writeRequestTimeout(w, r, links)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "failed to load issues", links)
		return
	}

	if err := writeData(w, r, http.StatusOK, data, links); err != nil {
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
