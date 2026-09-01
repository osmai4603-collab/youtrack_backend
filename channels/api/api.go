package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"youtrack_backend/channels/app"
)

// Routes يحمل هيكل الموجهات الهرمية الفرعية للنظام، مطابقاً لمعمارية Mattermost api4.Routes.
type Routes struct {
	Root    *mux.Router // '/'
	APIRoot *mux.Router // '/api'
	HubRoot *mux.Router // '/hub/api/rest'

	Users *mux.Router // '/api/users'
	User  *mux.Router // '/api/users/{id}'

	Projects *mux.Router // '/api/projects'
	Project  *mux.Router // '/api/projects/{id}'

	Issues *mux.Router // '/api/issues'
	Issue  *mux.Router // '/api/issues/{id}'

	Admin *mux.Router // '/api/admin'
	Auth  *mux.Router // '/api/auth'
	Inbox *mux.Router // '/api/inbox'
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
	api.BaseRoutes.APIRoot = srv.Router.PathPrefix("/api").Subrouter()
	api.BaseRoutes.HubRoot = srv.Router.PathPrefix("/hub/api/rest").Subrouter()

	api.BaseRoutes.Users = api.BaseRoutes.APIRoot.PathPrefix("/users").Subrouter()
	api.BaseRoutes.Projects = api.BaseRoutes.APIRoot.PathPrefix("/projects").Subrouter()
	api.BaseRoutes.Issues = api.BaseRoutes.APIRoot.PathPrefix("/issues").Subrouter()
	api.BaseRoutes.Admin = api.BaseRoutes.APIRoot.PathPrefix("/admin").Subrouter()
	api.BaseRoutes.Auth = api.BaseRoutes.APIRoot.PathPrefix("/auth").Subrouter()
	api.BaseRoutes.Inbox = api.BaseRoutes.APIRoot.PathPrefix("/inbox").Subrouter()

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

// APIHandler ينشئ مساراً عاماً بدون مصادقة.
func (api *API) APIHandler(h HandlerFunc) http.Handler {
	return &Handler{
		App:            app.New(),
		HandleFunc:     h,
		RequireSession: false,
	}
}

// APISessionRequired ينشئ مساراً محمياً يتطلب جلسة وتوكن صالح.
func (api *API) APISessionRequired(h HandlerFunc) http.Handler {
	return &Handler{
		App:            app.New(),
		HandleFunc:     h,
		RequireSession: true,
	}
}
