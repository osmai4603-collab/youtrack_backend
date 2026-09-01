package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"youtrack_backend/channels/app"
)

// NewRouter يبني راوتر التطبيق ويربط طبقة الـ API مع الـ middleware العام.
func NewRouter(a *app.YouTrackApp, jwtSecret string) http.Handler {
	srv := a.Srv()
	if srv == nil {
		srv = app.New().Srv()
	} else {
		if srv.Router == nil {
			rootRouter := mux.NewRouter()
			srv.RootRouter = rootRouter
			srv.Router = rootRouter
		}
		if srv.JWTSecret() == "" && srv.Channels().Platform() != nil {
			srv.Channels().Platform().SetJWTSecret(jwtSecret)
		}
	}

	Init(srv)
	return CORS()(Logger(RequestID(Recoverer(srv.Router))))
}

// NewServerRouter يبني الراوتر مباشرة من كائن Server.
func NewServerRouter(srv *app.YouTrackServer) http.Handler {
	if srv.Router == nil {
		rootRouter := mux.NewRouter()
		srv.RootRouter = rootRouter
		srv.Router = rootRouter
	}
	Init(srv)
	return CORS()(Logger(RequestID(Recoverer(srv.Router))))
}
