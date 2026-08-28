package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/usecase"
	"youtrack_backend/pkg/fields"
)

type APIHandler struct {
	notifUC    *usecase.NotificationUseCase
	searchUC   *usecase.SearchUseCase
	agileUC    *usecase.AgileUseCase
	ttUC       *usecase.TimeTrackingUseCase
	ppUC       *usecase.ProjectPeopleUseCase
	vcsUC      *usecase.VCSUseCase
	appUC      *usecase.AppUseCase
	hubUC      *usecase.HubUseCase
	configUC   *usecase.ConfigUseCase
	userUC     *usecase.UserUseCase
	adminUC    *usecase.AdminUseCase
}

func NewAPIHandler(
	notifUC *usecase.NotificationUseCase,
	searchUC *usecase.SearchUseCase,
	agileUC *usecase.AgileUseCase,
	ttUC *usecase.TimeTrackingUseCase,
	ppUC *usecase.ProjectPeopleUseCase,
	vcsUC *usecase.VCSUseCase,
	appUC *usecase.AppUseCase,
	hubUC *usecase.HubUseCase,
	configUC *usecase.ConfigUseCase,
	userUC *usecase.UserUseCase,
	adminUC *usecase.AdminUseCase,
) *APIHandler {
	return &APIHandler{
		notifUC:  notifUC,
		searchUC: searchUC,
		agileUC:  agileUC,
		ttUC:     ttUC,
		ppUC:     ppUC,
		vcsUC:    vcsUC,
		appUC:    appUC,
		hubUC:    hubUC,
		configUC: configUC,
		userUC:   userUC,
		adminUC:  adminUC,
	}
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func getFields(r *http.Request) *fields.FieldNode {
	fq := r.URL.Query().Get("fields")
	if fq == "" {
		return nil
	}
	return fields.Parse(fq)
}

func filterAndWrite(w http.ResponseWriter, r *http.Request, data any) {
	tree := getFields(r)
	if tree == nil {
		writeJSON(w, data)
		return
	}
	filtered := fields.Filter(data, tree)
	writeJSON(w, filtered)
}

func getPathParam(r *http.Request, name string) string {
	return r.PathValue(name)
}

func getQueryInt(r *http.Request, name string, def int) int {
	s := r.URL.Query().Get(name)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// ─────────────────────── Inbox ───────────────────────

func (h *APIHandler) GetInboxFolders(w http.ResponseWriter, r *http.Request) {
	folders, err := h.notifUC.GetInboxFolders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, folders)
}

func (h *APIHandler) GetInboxThreads(w http.ResponseWriter, r *http.Request) {
	folderID := getPathParam(r, "folderID")
	if folderID == "" {
		folderID = "direct"
	}
	threads, err := h.notifUC.GetThreads(folderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, threads)
}

// ─────────────────────── Search ───────────────────────

func (h *APIHandler) SearchAssist(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	result, err := h.searchUC.SearchAssist(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, result)
}

// ─────────────────────── Agile ───────────────────────

func (h *APIHandler) GetAgileUserProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.agileUC.GetUserProfile()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, profile)
}

func (h *APIHandler) GetBoardExtensions(w http.ResponseWriter, r *http.Request) {
	boardID := getPathParam(r, "boardID")
	ext, err := h.agileUC.GetBoardExtensions(boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ext == nil {
		ext = &domain.Extensions{Type: "Extensions"}
	}
	filterAndWrite(w, r, ext)
}

// ─────────────────────── Time Tracking ───────────────────────

func (h *APIHandler) GetAttributePrototypes(w http.ResponseWriter, r *http.Request) {
	prototypes, err := h.ttUC.GetAttributePrototypes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if prototypes == nil {
		prototypes = []*domain.AttributePrototype{}
	}
	filterAndWrite(w, r, prototypes)
}

func (h *APIHandler) GetAttributePrototype(w http.ResponseWriter, r *http.Request) {
	id := getPathParam(r, "id")
	prototype, err := h.ttUC.GetAttributePrototype(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if prototype == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	filterAndWrite(w, r, prototype)
}

func (h *APIHandler) GetBoardTimeTrackingData(w http.ResponseWriter, r *http.Request) {
	boardID := getPathParam(r, "boardID")
	data, err := h.ttUC.GetBoardTimeTrackingData(boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if data == nil {
		data = &domain.BoardTimeTrackingData{Type: "BoardTimeTrackingData"}
	}
	filterAndWrite(w, r, data)
}

// ─────────────────────── Project People & Dashboard ───────────────────────

func (h *APIHandler) GetProjectPeople(w http.ResponseWriter, r *http.Request) {
	projectID := getPathParam(r, "projectID")
	query := r.URL.Query().Get("transitiveRolesQuery")
	people, err := h.ppUC.GetProjectPeople(projectID, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if people == nil {
		people = &domain.ProjectPeople{Type: "ProjectPeople"}
	}
	filterAndWrite(w, r, people)
}

func (h *APIHandler) GetProjectDashboard(w http.ResponseWriter, r *http.Request) {
	projectID := getPathParam(r, "projectID")
	dashboard, err := h.ppUC.GetProjectDashboard(projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if dashboard == nil {
		dashboard = &domain.ProjectDashboard{Widgets: []*domain.DashboardWidgetEmbedding{}, Type: "ProjectDashboard"}
	}
	filterAndWrite(w, r, dashboard)
}

// ─────────────────────── Project Custom Fields ───────────────────────

func (h *APIHandler) GetProjectCustomFields(w http.ResponseWriter, r *http.Request) {
	result := []any{}
	filterAndWrite(w, r, result)
}

// ─────────────────────── Project Time Tracking Settings ───────────────────────

func (h *APIHandler) GetProjectTimeTrackingSettings(w http.ResponseWriter, r *http.Request) {
	result := map[string]any{
		"enabled":       false,
		"workItemTypes": []any{},
		"attributes":    []any{},
		"$type":         "ProjectTimeTrackingSettings",
	}
	filterAndWrite(w, r, result)
}

// ─────────────────────── Custom Field Aggregated Users ───────────────────────

func (h *APIHandler) GetAggregatedUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userUC.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if users == nil {
		users = []*domain.User{}
	}
	filterAndWrite(w, r, users)
}

// ─────────────────────── VCS ───────────────────────

func (h *APIHandler) GetVCSServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.vcsUC.GetVCSServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if servers == nil {
		servers = []*domain.VCSServer{}
	}
	filterAndWrite(w, r, servers)
}

// ─────────────────────── App/Services ───────────────────────

func (h *APIHandler) GetServicesPage(w http.ResponseWriter, r *http.Request) {
	page, err := h.appUC.GetServicesPage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, page)
}

// ─────────────────────── Hub ───────────────────────

func (h *APIHandler) GetHubUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.hubUC.GetHubUser()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, user)
}

// ─────────────────────── Misc Endpoints ───────────────────────

func (h *APIHandler) GetNotificationSupplement(w http.ResponseWriter, r *http.Request) {
	supplement := map[string]any{
		"preview": map[string]any{
			"issueId": "",
			"$type":   "NotificationPreview",
		},
		"$type": "NotificationSupplement",
	}
	filterAndWrite(w, r, supplement)
}

func (h *APIHandler) GetQuestionnaireProfile(w http.ResponseWriter, r *http.Request) {
	profile := map[string]any{
		"showSurvey":               false,
		"showPmfSurvey":            false,
		"demoEligibilityTimestamp": nil,
		"$type":                    "QuestionnaireUserProfile",
	}
	filterAndWrite(w, r, profile)
}

func (h *APIHandler) GetPublicSettings(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := "default"
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		id = path[idx+1:]
	}
	result := map[string]any{
		"id":    id,
		"$type": "public",
	}
	filterAndWrite(w, r, result)
}

func (h *APIHandler) GetIssueListSubscription(w http.ResponseWriter, r *http.Request) {
	sub := map[string]any{
		"ticket": "",
		"$type":  "IssueListSubscriptionBean",
	}
	filterAndWrite(w, r, sub)
}

func (h *APIHandler) GetPermissionHolders(w http.ResponseWriter, r *http.Request) {
	filterAndWrite(w, r, []any{})
}

func (h *APIHandler) GetAllowedForeignKeyTypes(w http.ResponseWriter, r *http.Request) {
	filterAndWrite(w, r, []any{})
}

func (h *APIHandler) GetCustomFieldSettings(w http.ResponseWriter, r *http.Request) {
	filterAndWrite(w, r, []any{})
}

func (h *APIHandler) GetDigests(w http.ResponseWriter, r *http.Request) {
	filterAndWrite(w, r, []any{})
}

func (h *APIHandler) GetApps(w http.ResponseWriter, r *http.Request) {
	filterAndWrite(w, r, []any{})
}

func (h *APIHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.adminUC.GetRoles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, roles)
}

func (h *APIHandler) GetIssueLinkTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.adminUC.GetIssueLinkTypes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, types)
}

func (h *APIHandler) GetFilterFields(w http.ResponseWriter, r *http.Request) {
	filterAndWrite(w, r, []any{})
}

func (h *APIHandler) GetBundles(w http.ResponseWriter, r *http.Request) {
	filterAndWrite(w, r, []any{})
}

func (h *APIHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.adminUC.GetProjects()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, projects)
}

func (h *APIHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userUC.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, users)
}

func (h *APIHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userID := getPathParam(r, "userID")
	user, err := h.userUC.GetUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	filterAndWrite(w, r, user)
}

func (h *APIHandler) GetDashboardWidgets(w http.ResponseWriter, r *http.Request) {
	widgets, err := h.adminUC.GetWidgets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	filterAndWrite(w, r, widgets)
}
