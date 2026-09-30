package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"youtrack_backend/channels/app"
)

// Routes يحمل هيكل الموجهات الهرمية الفرعية للنظام، مطابقاً لمعمارية Mattermost api4.Routes.
type Routes struct {
	Root    *chi.Mux // '/'
	APIRoot *chi.Mux // '/api'
	HubRoot *chi.Mux // '/hub/api/rest'

	Users *chi.Mux // '/api/users'
	User  *chi.Mux // '/api/users/{id}'

	Projects *chi.Mux // '/api/projects'
	Project  *chi.Mux // '/api/projects/{id}'

	Issues *chi.Mux // '/api/issues'
	Issue  *chi.Mux // '/api/issues/{id}'

	Admin *chi.Mux // '/api/admin'
	Auth  *chi.Mux // '/api/auth'
	Inbox *chi.Mux // '/api/inbox'
}

// API يمثل طبقة الـ API المركزية التي تربط الموجهات بالخادم.
type API struct {
	srv        *app.YouTrackServer
	BaseRoutes *Routes
}

// Init يبني طبقة الـ API ويهيئ جميع مسارات المكونات الفرعية.
func Init(srv *app.YouTrackServer) *API {
	api := &API{
		srv:        srv,
		BaseRoutes: &Routes{},
	}

	api.BaseRoutes.Root = srv.Router
	api.BaseRoutes.APIRoot = chi.NewRouter()
	api.BaseRoutes.HubRoot = chi.NewRouter()

	api.BaseRoutes.Users = chi.NewRouter()
	api.BaseRoutes.Projects = chi.NewRouter()
	api.BaseRoutes.Issues = chi.NewRouter()
	api.BaseRoutes.Admin = chi.NewRouter()
	api.BaseRoutes.Auth = chi.NewRouter()
	api.BaseRoutes.Inbox = chi.NewRouter()

	api.BaseRoutes.APIRoot.Mount("/users", api.BaseRoutes.Users)
	api.BaseRoutes.APIRoot.Mount("/projects", api.BaseRoutes.Projects)
	api.BaseRoutes.APIRoot.Mount("/issues", api.BaseRoutes.Issues)
	api.BaseRoutes.APIRoot.Mount("/admin", api.BaseRoutes.Admin)
	api.BaseRoutes.APIRoot.Mount("/auth", api.BaseRoutes.Auth)
	api.BaseRoutes.APIRoot.Mount("/inbox", api.BaseRoutes.Inbox)

	srv.Router.Mount("/api", api.BaseRoutes.APIRoot)
	srv.Router.Mount("/hub/api/rest", api.BaseRoutes.HubRoot)

	api.InitHealth()
	api.InitConfig()
	api.InitFeature()
	api.InitAuth()
	api.InitUser()
	api.InitProject()
	api.InitIssue()
	api.InitAdmin()
	api.InitInbox()
	api.InitSavedQuery()
	api.InitSearch()
	api.InitSecuritySearch()
	api.InitSubscription()
	api.InitHub()

	return api
}

// newApp ينشئ كائن App مربوطاً بخادم هذا الـ API (مع Store و Config و JWT).
func (api *API) newApp() *app.YouTrackApp {
	return app.New(app.ServerConnector(api.srv.Channels()))
}

// APIHandler ينشئ مساراً عاماً بدون مصادقة.
func (api *API) APIHandler(h HandlerFunc) http.Handler {
	return &Handler{
		App:            app.New(app.ServerConnector(api.srv.Channels())),
		HandleFunc:     h,
		RequireSession: false,
	}
}

// APISessionRequired ينشئ مساراً محمياً يتطلب جلسة وتوكن صالح.
func (api *API) APISessionRequired(h HandlerFunc) http.Handler {
	return &Handler{
		App:            app.New(app.ServerConnector(api.srv.Channels())),
		HandleFunc:     h,
		RequireSession: true,
	}
}

// APISessionRequiredWithPermission ينشئ مساراً محمياً يتطلب جلسة صالحة وصلاحية محددة.
func (api *API) APISessionRequiredWithPermission(permission string, h HandlerFunc) http.Handler {
	return &Handler{
		App:               app.New(app.ServerConnector(api.srv.Channels())),
		HandleFunc:        h,
		RequireSession:    true,
		RequirePermission: permission,
	}
}

// APISessionRequiredTrustRequester keeps the requester identity available to handlers.
func (api *API) APISessionRequiredTrustRequester(h HandlerFunc) http.Handler {
	return &Handler{
		App:            app.New(app.ServerConnector(api.srv.Channels())),
		HandleFunc:     h,
		RequireSession: true,
		TrustRequester: true,
	}
}

// APISessionRequiredDisableWhenBusy is reserved for handlers that may be disabled while busy.
func (api *API) APISessionRequiredDisableWhenBusy(h HandlerFunc) http.Handler {
	return api.APISessionRequired(h)
}
