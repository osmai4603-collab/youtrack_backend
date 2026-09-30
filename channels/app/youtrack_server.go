package app

import (
	"context"
	stderrors "errors"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/shared/mlog"
	"youtrack_backend/channels/store"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

type LifecycleState string

// ListenerFactory creates a network listener for the server lifecycle.
// It is injectable so startup bind failures can be tested without fixed ports.
type ListenerFactory func(network, address string) (net.Listener, error)

const shutdownStageTimeout = 10 * time.Second

const (
	LifecycleCreated  LifecycleState = "created"
	LifecycleStarting LifecycleState = "starting"
	LifecycleRunning  LifecycleState = "running"
	LifecycleStopping LifecycleState = "stopping"
	LifecycleStopped  LifecycleState = "stopped"
	LifecycleFailed   LifecycleState = "failed"
)

// YouTrackServer يمثل الكائن الدائم لدورة حياة الخادم وشبكة الاتصال ومخزن البيانات،
// ويحتوي على PlatformService و Channels وموجهات HTTP مطابقاً لمعمارية Mattermost.
type YouTrackServer struct {
	platform      *platform.YouTrackPlatformService
	ch            *YouTrackChannels
	RootRouter    *chi.Mux
	LocalRouter   *chi.Mux
	Router        *chi.Mux
	Server        *http.Server
	listenerFunc  ListenerFactory
	listener      net.Listener
	localServer   *http.Server
	localListener net.Listener
	localSocket   string
	localServeErr chan error
	ctx           context.Context
	cancel        context.CancelFunc
	lifecycleMu   sync.Mutex
	runMu         sync.Mutex
	runStarted    bool
	shutdownOnce  sync.Once
	shutdownErr   error
	readyMu       sync.RWMutex
	ready         bool
	stateMu       sync.RWMutex
	state         LifecycleState
}

func (s *YouTrackServer) Log() *mlog.Logger {
	return s.platform.Logger()
}

// NewServer ينشئ نسخة جديدة من Server ويهيئ PlatformService و Channels والموجهات الأساسية.
func NewServer(options ...Option) (*YouTrackServer, error) {
	server := &YouTrackServer{
		RootRouter:   chi.NewRouter(),
		Router:       chi.NewRouter(),
		listenerFunc: net.Listen,
		state:        LifecycleCreated,
	}

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, errors.Wrap(err, "failed to apply option")
		}
	}
	if server.platform == nil {
		ps, err := platform.New()
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize platform")
		}
		server.platform = ps
	}
	server.ch = NewChannels(server)

	return server, nil
}

// NewServerWithOptions ينشئ نسخة جديدة من Server مع تخصيص PlatformService.
func NewServerWithOptions(ps *platform.YouTrackPlatformService) (*YouTrackServer, error) {
	if ps == nil {
		var err error
		ps, err = platform.New()
		if err != nil {
			return nil, errors.Wrap(err, "failed to initialize platform")
		}
	}
	rootRouter := chi.NewRouter()

	server := &YouTrackServer{
		platform:     ps,
		RootRouter:   rootRouter,
		Router:       rootRouter,
		listenerFunc: net.Listen,
		state:        LifecycleCreated,
	}
	server.ch = NewChannels(server)

	return server, nil
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
	if s.listenerFunc == nil {
		s.listenerFunc = net.Listen
	}
	if err := s.platform.Validate(); err != nil {
		return errors.Wrap(err, "invalid platform configuration")
	}
	if s.RootRouter == nil {
		s.RootRouter = chi.NewRouter()
	}
	if s.Router == nil {
		s.Router = s.RootRouter
	}
	if s.LocalRouter == nil {
		s.LocalRouter = chi.NewRouter()
		readyHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if !s.IsReady() {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(`{"status":"unavailable","ready":false}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"healthy","ready":true}`))
		}
		s.LocalRouter.Get("/health", readyHandler)
		s.LocalRouter.Get("/ready", readyHandler)
		s.LocalRouter.Get("/live", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"live"}`))
		})
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
	} else if s.Server.Handler == nil {
		s.Server.Handler = s.RootRouter
	}
	return nil
}

func (s *YouTrackServer) Start() error {
	if s == nil {
		return errors.New("server is nil")
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	s.stateMu.Lock()
	switch s.state {
	case LifecycleStarting, LifecycleRunning:
		s.stateMu.Unlock()
		return nil
	case LifecycleFailed, LifecycleStopped, LifecycleStopping:
		state := s.state
		s.stateMu.Unlock()
		return errors.Errorf("server cannot start from %s state", state)
	default:
		s.state = LifecycleStarting
		s.stateMu.Unlock()
	}

	if err := s.Validate(); err != nil {
		s.setState(LifecycleFailed)
		return err
	}
	if err := s.platform.Start(); err != nil {
		s.setState(LifecycleFailed)
		return errors.Wrap(err, "failed to start platform")
	}
	if err := s.Channels().Start(); err != nil {
		s.setState(LifecycleFailed)
		startupErr := errors.Wrap(err, "Unable to start channels")
		return stderrors.Join(startupErr, s.rollbackStartup())
	}
	return nil
}

func (s *YouTrackServer) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.runMu.Lock()
	if s.runStarted {
		s.runMu.Unlock()
		return errors.New("server run has already started")
	}
	s.runStarted = true
	s.runMu.Unlock()

	if err := s.Start(); err != nil {
		return err
	}
	if s.Server == nil {
		startupErr := errors.New("server is not initialized")
		s.setState(LifecycleFailed)
		return stderrors.Join(startupErr, s.rollbackStartup())
	}
	listener, err := s.listenerFunc("tcp", s.Server.Addr)
	if err != nil {
		startupErr := errors.Wrap(err, "failed to bind server listener")
		s.setState(LifecycleFailed)
		return stderrors.Join(startupErr, s.rollbackStartup())
	}
	s.listener = listener
	s.setReady(true)
	s.setState(LifecycleRunning)
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.Server.Serve(listener)
	}()
	s.lifecycleMu.Lock()
	localServer := s.localServer
	localListener := s.localListener
	s.lifecycleMu.Unlock()
	if localServer != nil && localListener != nil {
		s.localServeErr = make(chan error, 1)
		go func() {
			s.localServeErr <- localServer.Serve(localListener)
		}()
	}

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return s.shutdownWithCause(err)
		}
	case err := <-s.localServeErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return s.shutdownWithCause(errors.Wrap(err, "local server failed"))
		}
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
	return nil
}

func (s *YouTrackServer) Shutdown(ctx context.Context) error {
	if s == nil {
		return nil
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()

	s.shutdownOnce.Do(func() {
		shutdownCtx, cancel := context.WithTimeout(normalizeShutdownContext(ctx), shutdownStageTimeout)
		defer cancel()
		s.setState(LifecycleStopping)
		s.setReady(false)
		if s.cancel != nil {
			s.cancel()
		}
		var shutdownErrors []error
		if s.Server != nil {
			err := s.Server.Shutdown(shutdownCtx)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to shutdown server"))
			}
		}
		if s.localServer != nil {
			err := s.localServer.Shutdown(shutdownCtx)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to shutdown local server"))
			}
		}
		if s.listener != nil {
			if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to close server listener"))
			}
			s.listener = nil
		}
		if s.ch != nil {
			if err := s.ch.Stop(); err != nil {
				shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to stop channels"))
			}
		}
		if s.platform != nil {
			err := s.platform.ShutdownContext(shutdownCtx)
			if err != nil {
				shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to shutdown platform"))
			}
			if logger := s.platform.Logger(); logger != nil {
				if err := logger.FlushWithTimeout(shutdownCtx); err != nil {
					shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to flush logger"))
				}
			}
		}
		if s.localListener != nil {
			if err := s.localListener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to close local listener"))
			}
			s.localListener = nil
		}
		if s.localSocket != "" {
			if err := os.Remove(s.localSocket); err != nil && !os.IsNotExist(err) {
				shutdownErrors = append(shutdownErrors, errors.Wrap(err, "failed to remove local socket"))
			}
			s.localSocket = ""
		}
		s.localServer = nil
		s.shutdownErr = stderrors.Join(shutdownErrors...)
		if s.shutdownErr != nil {
			s.setState(LifecycleFailed)
		} else {
			s.setState(LifecycleStopped)
		}
	})

	return s.shutdownErr
}

// SetHTTPServer assigns the fully configured HTTP server before startup.
// YouTrackServer owns serving and shutdown after this point.
func (s *YouTrackServer) SetHTTPServer(server *http.Server) error {
	if s == nil {
		return errors.New("server is nil")
	}
	if server == nil {
		return errors.New("http server is required")
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.LifecycleState() != LifecycleCreated {
		return errors.Errorf("http server cannot be changed from %s state", s.LifecycleState())
	}
	s.Server = server
	return nil
}

func (s *YouTrackServer) shutdownStageContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(normalizeShutdownContext(parent), shutdownStageTimeout)
}

func normalizeShutdownContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (s *YouTrackServer) rollbackStartup() error {
	rollbackCtx, cancel := context.WithTimeout(context.Background(), shutdownStageTimeout)
	defer cancel()
	var rollbackErrors []error
	if s.localServer != nil {
		if err := s.localServer.Shutdown(rollbackCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			rollbackErrors = append(rollbackErrors, errors.Wrap(err, "failed to rollback local server"))
		}
	}
	if s.localListener != nil {
		if err := s.localListener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			rollbackErrors = append(rollbackErrors, errors.Wrap(err, "failed to rollback local listener"))
		}
		s.localListener = nil
	}
	if s.localSocket != "" {
		if err := os.Remove(s.localSocket); err != nil && !os.IsNotExist(err) {
			rollbackErrors = append(rollbackErrors, errors.Wrap(err, "failed to rollback local socket"))
		}
		s.localSocket = ""
	}
	s.localServer = nil
	if s.ch != nil {
		if err := s.ch.Stop(); err != nil {
			rollbackErrors = append(rollbackErrors, errors.Wrap(err, "failed to rollback channels"))
		}
	}
	if s.platform != nil {
		if err := s.platform.ShutdownContext(rollbackCtx); err != nil {
			rollbackErrors = append(rollbackErrors, errors.Wrap(err, "failed to rollback platform"))
		}
	}
	return stderrors.Join(rollbackErrors...)
}

func (s *YouTrackServer) shutdownWithCause(cause error) error {
	shutdownErr := s.Shutdown(context.Background())
	return stderrors.Join(cause, shutdownErr)
}

// SetLocalServer attaches an optional local supervision server to the lifecycle.
func (s *YouTrackServer) SetLocalServer(listener net.Listener, server *http.Server) error {
	if s == nil {
		return errors.New("server is nil")
	}
	if listener == nil || server == nil {
		return errors.New("local listener and server are required")
	}
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.LifecycleState() != LifecycleCreated {
		return errors.Errorf("local server cannot be changed from %s state", s.LifecycleState())
	}
	if s.localListener != nil || s.localServer != nil {
		return errors.New("local server is already configured")
	}
	s.localListener = listener
	s.localServer = server
	if addr := listener.Addr(); addr != nil {
		s.localSocket = addr.String()
	}
	return nil
}

func (s *YouTrackServer) LifecycleState() LifecycleState {
	if s == nil {
		return LifecycleFailed
	}
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()
	if s.state == "" {
		return LifecycleCreated
	}
	return s.state
}

func (s *YouTrackServer) setState(state LifecycleState) {
	if s == nil {
		return
	}
	s.stateMu.Lock()
	s.state = state
	s.stateMu.Unlock()
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
