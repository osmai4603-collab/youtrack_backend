package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"youtrack_backend/internal/config"
	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/handler"
	"youtrack_backend/internal/middleware"
	"youtrack_backend/internal/repository/postgres"
	"youtrack_backend/internal/usecase"
	"youtrack_backend/pkg/database"
)

func main() {
	cfg := config.LoadConfig()

	// الاتصال بقاعدة بيانات PostgreSQL
	var db *sql.DB
	var userRepo domain.UserRepository
	var issueRepo domain.IssueRepository
	var adminRepo domain.AdminRepository
	var configRepo domain.ConfigRepository
	var notifRepo domain.NotificationRepository
	var searchRepo domain.SearchRepository
	var agileRepo domain.AgileRepository
	var ttRepo domain.TimeTrackingRepository
	var ppRepo domain.ProjectPeopleRepository
	var vcsRepo domain.VCSRepository
	var appRepo domain.AppRepository
	var hubRepo domain.HubRepository

	conn, err := database.NewPostgresConnection(cfg.DSN())
	if err != nil {
		log.Printf("Warning: Could not connect to PostgreSQL database (%v). Running with in-memory fallbacks.", err)
	} else {
		db = conn
		defer db.Close()
		userRepo = postgres.NewUserRepository(db)
		issueRepo = postgres.NewIssueRepository(db)
		adminRepo = postgres.NewAdminRepository(db)
		configRepo = postgres.NewConfigRepository(db)
		notifRepo = postgres.NewNotificationRepository(db)
		searchRepo = postgres.NewSearchRepository(db)
		agileRepo = postgres.NewAgileRepository(db)
		ttRepo = postgres.NewTimeTrackingRepository(db)
		ppRepo = postgres.NewProjectPeopleRepository(db)
		vcsRepo = postgres.NewVCSRepository(db)
		appRepo = postgres.NewAppRepository(db)
		hubRepo = postgres.NewHubRepository(db)
	}

	mux := http.NewServeMux()

	// UseCases
	userUC := usecase.NewUserUseCase(userRepo)
	issueUC := usecase.NewIssueUseCase(issueRepo)
	adminUC := usecase.NewAdminUseCase(adminRepo)
	configUC := usecase.NewConfigUseCase(configRepo)
	notifUC := usecase.NewNotificationUseCase(notifRepo)
	searchUC := usecase.NewSearchUseCase(searchRepo)
	agileUC := usecase.NewAgileUseCase(agileRepo)
	ttUC := usecase.NewTimeTrackingUseCase(ttRepo)
	ppUC := usecase.NewProjectPeopleUseCase(ppRepo)
	vcsUC := usecase.NewVCSUseCase(vcsRepo)
	appUC := usecase.NewAppUseCase(appRepo)
	hubUC := usecase.NewHubUseCase(hubRepo)

	// Handlers مع حقن الـ UseCases
	healthHandler := handler.NewHealthHandler()
	configHandler := handler.NewConfigHandler(configUC)
	userHandler := handler.NewUserHandler(userUC)
	permissionHandler := handler.NewPermissionHandler(configUC)
	issueHandler := handler.NewIssueHandler(issueUC)
	adminHandler := handler.NewAdminHandler(adminUC)
	apiHandler := handler.NewAPIHandler(notifUC, searchUC, agileUC, ttUC, ppUC, vcsUC, appUC, hubUC, configUC, userUC, adminUC)
	jwtSecret := "youtrack-secret-jwt-key"
	authHandler := handler.NewAuthHandler(userUC, jwtSecret)
	wsHandler := handler.NewWebSocketHandler()
	attachmentHandler := handler.NewAttachmentHandler("./uploads")

	// Health Routes
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/v1/health", healthHandler.Check)

	// Auth Routes
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/register", authHandler.Register)

	// Real-time WebSocket Route
	mux.HandleFunc("GET /api/events", wsHandler.ServeWS)

	// Attachments & Files Routes
	mux.HandleFunc("POST /api/issues/{issueID}/attachments", attachmentHandler.UploadAttachment)
	mux.HandleFunc("GET /api/files/{fileName}", attachmentHandler.GetFile)

	// YouTrack Core & Config Routes
	mux.HandleFunc("GET /api/config", configHandler.GetConfig)

	// Users & Profiles Routes
	mux.Handle("GET /api/users/me", middleware.JWTAuth(jwtSecret)(http.HandlerFunc(userHandler.GetCurrentUser)))
	mux.HandleFunc("GET /api/users/me/profiles/grazie", userHandler.GetGrazieProfile)
	mux.HandleFunc("GET /api/users/me/profiles/questionnaire", apiHandler.GetQuestionnaireProfile)

	// Permissions Routes
	mux.HandleFunc("GET /api/permissions/cache", permissionHandler.GetPermissionsCache)
	mux.HandleFunc("GET /api/admin/permissionHolders", apiHandler.GetPermissionHolders)

	// Issues, Links & Activities Routes
	mux.HandleFunc("GET /api/issues/{issueID}", issueHandler.GetByID)
	mux.HandleFunc("POST /api/issues", issueHandler.CreateIssue)
	mux.HandleFunc("GET /api/issueLinkTypes", apiHandler.GetIssueLinkTypes)
	mux.HandleFunc("GET /api/issues/{issueID}/activitiesPage", issueHandler.GetActivitiesPage)
	mux.HandleFunc("GET /api/issues/{issueID}/listSubscription", apiHandler.GetIssueListSubscription)

	// Admin Settings & Widgets Routes
	mux.HandleFunc("GET /api/admin/timeTrackingSettings/workTimeSettings", adminHandler.GetWorkTimeSettings)
	mux.HandleFunc("GET /api/admin/globalSettings", adminHandler.GetGlobalSettings)
	mux.HandleFunc("GET /api/admin/widgets/general", adminHandler.GetGeneralWidgets)
	mux.HandleFunc("GET /api/admin/allowedForeignKeyTypes", apiHandler.GetAllowedForeignKeyTypes)
	mux.HandleFunc("GET /api/admin/customFieldSettings", apiHandler.GetCustomFieldSettings)
	mux.HandleFunc("GET /api/admin/customFieldSettings/bundles/user/{bundleID}/aggregatedUsers", apiHandler.GetAggregatedUsers)
	mux.HandleFunc("GET /api/admin/digests", apiHandler.GetDigests)
	mux.HandleFunc("GET /api/admin/notificationSupplement", apiHandler.GetNotificationSupplement)
	mux.HandleFunc("GET /api/admin/projects/{projectID}/people", apiHandler.GetProjectPeople)
	mux.HandleFunc("GET /api/admin/projects/{projectID}/dashboard", apiHandler.GetProjectDashboard)
	mux.HandleFunc("GET /api/admin/projects/{projectID}/customFields", apiHandler.GetProjectCustomFields)
	mux.HandleFunc("GET /api/admin/projects/{projectID}/timeTrackingSettings", apiHandler.GetProjectTimeTrackingSettings)
	mux.HandleFunc("GET /api/admin/integrations/vcsHostingServers", apiHandler.GetVCSServers)

	// Time Tracking Routes
	mux.HandleFunc("GET /api/admin/timeTrackingSettings/attributePrototypes", apiHandler.GetAttributePrototypes)
	mux.HandleFunc("GET /api/admin/timeTrackingSettings/attributePrototypes/{id}", apiHandler.GetAttributePrototype)

	// Search Routes
	mux.HandleFunc("GET /api/searchAssist", apiHandler.SearchAssist)

	// Agile Routes
	mux.HandleFunc("GET /api/agileUserProfile", apiHandler.GetAgileUserProfile)
	mux.HandleFunc("GET /api/agile/{boardID}/extensions", apiHandler.GetBoardExtensions)

	// Board Time Tracking Routes
	mux.HandleFunc("GET /api/timeTracking/boards/{boardID}", apiHandler.GetBoardTimeTrackingData)

	// Inbox Routes
	mux.HandleFunc("GET /api/inbox/folders", apiHandler.GetInboxFolders)
	mux.HandleFunc("GET /api/inbox/folders/{folderID}/threads", apiHandler.GetInboxThreads)

	// Users Routes
	mux.HandleFunc("GET /api/users", apiHandler.GetUsers)
	mux.HandleFunc("GET /api/users/{userID}", apiHandler.GetUserByID)

	// Hub Routes
	mux.HandleFunc("GET /hub/api/rest/users/me", apiHandler.GetHubUser)
	mux.HandleFunc("GET /hub/api/rest/settings/public", apiHandler.GetPublicSettings)

	// Services & Apps Routes
	mux.HandleFunc("GET /api/services/page", apiHandler.GetServicesPage)
	mux.HandleFunc("GET /api/apps", apiHandler.GetApps)

	// Roles & Permissions Routes
	mux.HandleFunc("GET /api/admin/roles", apiHandler.GetRoles)
	mux.HandleFunc("GET /api/filterFields", apiHandler.GetFilterFields)
	mux.HandleFunc("GET /api/bundles", apiHandler.GetBundles)
	mux.HandleFunc("GET /api/admin/projects", apiHandler.GetProjects)
	mux.HandleFunc("GET /api/widgets", apiHandler.GetDashboardWidgets)

	// Middlewares Pipeline: RequestID -> Recoverer -> CORS -> Logger -> Mux
	corsMiddleware := middleware.CORS()
	handlerWithMiddleware := middleware.RequestID(
		middleware.Recoverer(
			corsMiddleware(
				middleware.Logger(mux),
			),
		),
	)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handlerWithMiddleware,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// تشغيل الخادم في Goroutine منفصلة للتعامل مع الإيقاف السلس (Graceful Shutdown)
	go func() {
		log.Printf("Server starting on port %s in %s mode...", cfg.Port, cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// الاستماع لإشارات الإيقاف (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited successfully")
}
