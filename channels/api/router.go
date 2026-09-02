package api

import (
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	"youtrack_backend/channels/app"
)

// NewRouter يبني راوتر التطبيق ويربط طبقة الـ API مع الـ middleware العام.
func NewRouter(a *app.YouTrackApp, jwtSecret string) http.Handler {
	srv := a.Server()
	if srv == nil {
		srv = app.New().Server()
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
	return app.NewRateLimiter(120, 60).Middleware(CORS()(Logger(RequestID(Recoverer(VersionedAPI(srv.Router))))))
}

// NewServerRouter يبني الراوتر مباشرة من كائن Server.
func NewServerRouter(srv *app.YouTrackServer) http.Handler {
	if srv.Router == nil {
		rootRouter := mux.NewRouter()
		srv.RootRouter = rootRouter
		srv.Router = rootRouter
	}
	Init(srv)
	return app.NewRateLimiter(120, 60).Middleware(CORS()(Logger(RequestID(Recoverer(VersionedAPI(srv.Router))))))
}

// NewLocalRouter exposes only health checks for local process supervision.
func NewLocalRouter(srv *app.YouTrackServer) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	}).Methods("GET")
	if srv != nil {
		srv.LocalRouter = router
	}
	return router
}

// StartLocalAPI starts the unauthenticated supervision API on a Unix socket.
func StartLocalAPI(srv *app.YouTrackServer, socketPath string) (net.Listener, *http.Server, error) {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o750); err != nil {
		return nil, nil, err
	}
	_ = os.Remove(socketPath)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, nil, err
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		return nil, nil, err
	}
	server := &http.Server{Handler: NewLocalRouter(srv)}
	return listener, server, nil
}
