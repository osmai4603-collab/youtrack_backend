package api

import (
	"net/http"

	"youtrack_backend/internal/app"
)

// NewRouter يبني راوتر التطبيق ويسجّل كل المسارات مع الـ middleware.
func NewRouter(a *app.App, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	health := NewHealthHandler()
	auth := NewAuthHandler(a, jwtSecret)
	user := NewUserHandler(a)
	project := NewProjectHandler(a)
	issue := NewIssueHandler(a)
	config := NewConfigHandler(a)
	admin := NewAdminHandler(a)
	inbox := NewInboxHandler(a)
	feature := NewFeatureHandler(a)
	savedQueries := NewSavedQueriesHandler(a)
	search := NewSearchHandler(a)
	subscription := NewSubscriptionHandler(a)
	hub := NewHubHandler(a)

	// مسارات عامة (بدون مصادقة)
	mux.HandleFunc("GET /health", health.Check)
	mux.HandleFunc("GET /api/config", config.Get)
	mux.HandleFunc("GET /static/features-en_US.json", feature.Get)
	mux.HandleFunc("POST /api/auth/login", auth.Login)
	mux.HandleFunc("POST /api/auth/register", auth.Register)

	// المسارات المحمية JWT
	protected := JWTAuth(jwtSecret)

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /api/users/me", user.GetMe)
	protectedMux.HandleFunc("GET /api/users/me/profiles/grazie", user.GetGrazieProfile)
	protectedMux.HandleFunc("GET /api/users/me/profiles/general", user.GetGeneralProfile)
	protectedMux.HandleFunc("GET /api/users/me/profiles/questionnaire", user.GetQuestionnaireProfile)
	protectedMux.HandleFunc("GET /api/users/me/recent/issues", user.GetRecentIssues)
	protectedMux.HandleFunc("GET /api/users/me/recent/articles", user.GetRecentArticles)
	protectedMux.HandleFunc("GET /api/users", user.List)
	protectedMux.HandleFunc("GET /api/users/{id}", user.GetByID)

	protectedMux.HandleFunc("GET /hub/api/rest/users/me", user.GetHubMe)

	protectedMux.HandleFunc("GET /api/inbox/folders", user.GetInboxFolders)
	protectedMux.HandleFunc("GET /api/inbox/threads", inbox.GetThreads)

	protectedMux.HandleFunc("GET /api/savedQueries", savedQueries.List)

	protectedMux.HandleFunc("POST /api/search/assist", search.GetAssist)

	protectedMux.HandleFunc("GET /hub/api/rest/services", hub.GetServices)
	protectedMux.HandleFunc("GET /api/services", hub.GetServices)

	protectedMux.HandleFunc("GET /api/issueListSubscription", subscription.SubscribeIssueList)
	protectedMux.HandleFunc("POST /api/issueListSubscription", subscription.SubscribeIssueList)

	protectedMux.HandleFunc("GET /api/projects", project.List)
	protectedMux.HandleFunc("GET /api/projects/{id}", project.GetByID)
	protectedMux.HandleFunc("GET /api/admin/projects", project.List)
	protectedMux.HandleFunc("GET /api/admin/projects/{id}", project.GetByID)
	protectedMux.HandleFunc("GET /api/admin/projects/{id}/dashboard", admin.ProjectDashboard)

	protectedMux.HandleFunc("GET /api/issues", issue.List)
	protectedMux.HandleFunc("GET /api/sortedIssues", issue.GetSortedIssues)
	protectedMux.HandleFunc("GET /api/issuesGetter/count", issue.Count)
	protectedMux.HandleFunc("GET /api/issuesGetter", issue.Getter)
	protectedMux.HandleFunc("GET /api/issues/{id}", issue.GetByID)
	protectedMux.HandleFunc("GET /api/issues/{id}/comments", issue.Comments)
	protectedMux.HandleFunc("POST /api/issuesGetter/count", issue.Count)
	protectedMux.HandleFunc("POST /api/issuesGetter", issue.Getter)
	protectedMux.HandleFunc("POST /api/issues", issue.Create)

	protectedMux.HandleFunc("GET /api/roles", admin.Roles)
	protectedMux.HandleFunc("GET /api/permissions", admin.Permissions)
	protectedMux.HandleFunc("GET /api/permissions/cache", admin.PermissionsCache)
	protectedMux.HandleFunc("GET /api/admin/globalSettings", admin.GlobalSettings)
	protectedMux.HandleFunc("GET /api/admin/widgets/general", admin.Widgets)
	protectedMux.HandleFunc("GET /api/admin/organizations", admin.Organizations)

	protectedHandler := protected(protectedMux)
	mux.Handle("/api/", protectedHandler)
	mux.Handle("/hub/", protectedHandler)

	// تكديس الـ middleware العام
	return CORS()(Logger(RequestID(Recoverer(mux))))
}
