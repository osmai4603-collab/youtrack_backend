package app

import (
	"net/http"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/store"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
	"github.com/pkg/errors"
)

// YouTrackServer يمثل الكائن الدائم لدورة حياة الخادم وشبكة الاتصال ومخزن البيانات،
// ويحتوي على PlatformService و Channels وموجهات HTTP مطابقاً لمعمارية Mattermost.
type YouTrackServer struct {
	platform    *platform.YouTrackPlatformService
	ch          *ServerChannels
	RootRouter  *mux.Router
	LocalRouter *mux.Router
	Router      *mux.Router
	Server      *http.Server
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
	server.ch = NewChannels(server)

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, errors.Wrap(err, "failed to apply option")
		}
	}

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

// JWTSecret يعيد المفتاح السري للتوكنات من خلال خدمة المنصة.
func (s *YouTrackServer) JWTSecret() string {
	if s.platform != nil {
		return s.platform.JWTSecret()
	}
	return ""
}

// NewWithChannels ينشئ كائن App مرتبطاً بـ Channels.
func NewWithChannels(ch *ServerChannels) *YouTrackApp {
	if ch == nil {
		ch = NewChannels(nil)
	}
	return &YouTrackApp{
		channels: ch,
	}
}

func (s *YouTrackServer) Start() error {
	if err := s.Channels().Start(); err != nil {
		return errors.Wrap(err, "Unable to start channels")
	}
	s.Server = &http.Server{
		Addr: ":8090",
		// Handler:      api.NewServerRouter(s),
		// ReadTimeout:  time.Duration(*s.platform.Config().ServiceSettings.ReadTimeout) * time.Second,
		// WriteTimeout: time.Duration(*s.platform.Config().ServiceSettings.WriteTimeout) * time.Second,
		// IdleTimeout:  time.Duration(*s.platform.Config().ServiceSettings.IdleTimeout) * time.Second,
		// ErrorLog:     errStdLog,
	}

	return nil
}

func (s *YouTrackServer) Channels() *ServerChannels {
	return s.ch
}
