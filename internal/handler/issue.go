package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/usecase"
)

type IssueHandler struct {
	issueUC *usecase.IssueUseCase
}

func NewIssueHandler(issueUC *usecase.IssueUseCase) *IssueHandler {
	return &IssueHandler{
		issueUC: issueUC,
	}
}

// GetIssueLinkTypes يعالج طلبات /api/issueLinkTypes
func (h *IssueHandler) GetIssueLinkTypes(w http.ResponseWriter, r *http.Request) {
	linkTypes, err := h.issueUC.GetLinkTypes()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(linkTypes)
}

// GetActivitiesPage يعالج طلبات /api/issues/{issueID}/activitiesPage
func (h *IssueHandler) GetActivitiesPage(w http.ResponseWriter, r *http.Request) {
	issueID := r.PathValue("issueID")
	if issueID == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			issueID = parts[3]
		}
	}

	page, err := h.issueUC.GetActivitiesPage(issueID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(page)
}

// GetByID يعالج طلبات /api/issues/{issueID}
func (h *IssueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	issueID := r.PathValue("issueID")
	if issueID == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 {
			issueID = parts[3]
		}
	}

	issue, err := h.issueUC.GetByID(issueID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(issue)
}

// CreateIssue يعالج طلبات إنشاء تذكرة جديدة POST /api/issues
func (h *IssueHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	var issue domain.Issue
	if err := json.NewDecoder(r.Body).Decode(&issue); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	created, err := h.issueUC.CreateIssue(&issue)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}


