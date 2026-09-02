package app

import (
	"context"
	"net/http"
	"sync"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/shared/mlog"
	"youtrack_backend/channels/store"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// YouTrackServer يمثل الكائن الدائم لدورة حياة الخادم وشبكة الاتصال ومخزن البيانات،
// ويحتوي على PlatformService و Channels وموجهات HTTP مطابقاً لمعمارية Mattermost.
type YouTrackServer struct {
	platform     *platform.YouTrackPlatformService
	ch           *YouTrackChannels
	RootRouter   *mux.Router
	LocalRouter  *mux.Router
	Router       *mux.Router
	Server       *http.Server
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownOnce sync.Once
	readyMu      sync.RWMutex
	ready        bool
}

func (s *YouTrackServer) Log() *mlog.Logger {
	return s.platform.Logger()
}

// NewServer ينشئ نسخة جديدة من Server ويهيئ PlatformService و Channels والموجهات الأساسية.
func NewServer(options ...Option) (*YouTrackServer, error) {
	server := &YouTrackServer{
		RootRouter: mux.NewRouter(),
		Router:     mux.NewRouter(),
	}

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, errors.Wrap(err, "failed to apply option")
		}
	}
	if server.platform == nil {
		server.platform, _ = platform.New()
	}

	server.ch = NewChannels(server)

	return server, nil
}

// NewServerWithOptions ينشئ نسخة جديدة من Server مع تخصيص PlatformService.
func NewServerWithOptions(ps *platform.YouTrackPlatformService) *YouTrackServer {
	if ps == nil {
		ps, _ = platform.New()
	}
	rootRouter := mux.NewRouter()

	server := &YouTrackServer{
		platform:   ps,
		RootRouter: rootRouter,
		Router:     rootRouter,
	}
	server.ch = NewChannels(server)

	return server
}

// Store يعيد الوصول إلى مخزن البيانات من خلال خدمة المنصة.
func (s *YouTrackServer) Store() store.Store {
	if s.platform != nil {
		return s.platform.Store()
	}
	return nil
}

// Config يعيد إعدادات النظام من خلال خدمة المنصة.
func (s *YouTrackServer) Config() *model.ServerConfig {
	if s.platform != nil {
		return s.platform.Config()
	}
	return nil
}

// Platform يعيد خدمة المنصة المرتبطة بالخادم.
func (s *YouTrackServer) Platform() *platform.YouTrackPlatformService {
	return s.platform
}

// JWTSecret يعيد المفتاح السري للتوكنات من خلال خدمة المنصة.
func (s *YouTrackServer) JWTSecret() string {
	if s.platform != nil {
		return s.platform.JWTSecret()
	}
	return ""
}

// NewWithChannels ينشئ كائن App مرتبطاً بـ Channels.
func NewWithChannels(ch *YouTrackChannels) *YouTrackApp {
	if ch == nil {
		ch = NewChannels(nil)
	}
	return &YouTrackApp{
		channels: ch,
	}
}

// Validate checks that the server and its platform are in a valid startup state.
// This is the explicit preflight phase of the lifecycle contract: configuration and
// dependency checks must pass before Start can begin runtime initialization.
func (s *YouTrackServer) Validate() error {
	if s == nil {
		return errors.New("server is nil")
	}
	if s.platform == nil {
		ps, err := platform.New()
		if err != nil {
			return errors.Wrap(err, "failed to initialize platform")
		}
		s.platform = ps
	}
	if err := s.platform.Validate(); err != nil {
		return errors.Wrap(err, "invalid platform configuration")
	}
	if s.RootRouter == nil {
		s.RootRouter = mux.NewRouter()
	}
	if s.Router == nil {
		s.Router = s.RootRouter
	}
	if s.LocalRouter == nil {
		s.LocalRouter = mux.NewRouter()
		s.LocalRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"healthy"}`))
		}).Methods(http.MethodGet)
	}
	if s.ch == nil {
		s.ch = NewChannels(s)
	}
	if s.ctx == nil {
		s.ctx, s.cancel = context.WithCancel(context.Background())
	}
	if s.Server == nil {
		addr := ":8090"
		if s.Config() != nil && s.Config().ServerPort != "" {
			addr = ":" + s.Config().ServerPort
		}
		s.Server = &http.Server{
			Addr:    addr,
			Handler: s.RootRouter,
		}
	}
	return nil
}

func (s *YouTrackServer) Start() error {
	if err := s.Validate(); err != nil {
		return err
	}
	if err := s.platform.Start(); err != nil {
		return errors.Wrap(err, "failed to start platform")
	}
	if err := s.Channels().Start(); err != nil {
		return errors.Wrap(err, "Unable to start channels")
	}
	s.setReady(true)
	return nil
}

func (s *YouTrackServer) Run(ctx context.Context) error {
	if err := s.Start(); err != nil {
		return err
	}
	if s.Server == nil {
		return errors.New("server is not initialized")
	}
	if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	if ctx != nil {
		<-ctx.Done()
	}
	return nil
}

func (s *YouTrackServer) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}

	var shutdownErr error
	s.shutdownOnce.Do(func() {
		s.setReady(false)
		if s.ch != nil {
			if err := s.ch.Stop(); err != nil {
				shutdownErr = errors.Wrap(err, "failed to stop channels")
				return
			}
		}
		if s.Server != nil {
			if ctx == nil {
				ctx = context.Background()
			}
			if err := s.Server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				shutdownErr = errors.Wrap(err, "failed to shutdown server")
				return
			}
		}
		if s.platform != nil {
			if err := s.platform.Shutdown(); err != nil {
				shutdownErr = errors.Wrap(err, "failed to shutdown platform")
				return
			}
		}
		if s.cancel != nil {
			s.cancel()
		}
	})

	return shutdownErr
}

func (s *YouTrackServer) IsReady() bool {
	if s == nil {
		return false
	}
	s.readyMu.RLock()
	defer s.readyMu.RUnlock()
	return s.ready
}

func (s *YouTrackServer) setReady(v bool) {
	if s == nil {
		return
	}
	s.readyMu.Lock()
	s.ready = v
	s.readyMu.Unlock()
}

func (s *YouTrackServer) Channels() *YouTrackChannels {
	return s.ch
}
