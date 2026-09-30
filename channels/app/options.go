package app

import (
	"errors"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/store/retrylayer"
	"youtrack_backend/channels/store/sqlstore"
	"youtrack_backend/channels/store/timerlayer"
)

type Option func(s *YouTrackServer) error

func ensurePlatform(s *YouTrackServer) error {
	if s == nil {
		return nil
	}
	if s.platform != nil {
		return nil
	}
	ps, err := platform.New()
	if err != nil {
		return err
	}
	s.platform = ps
	return nil
}

type (
	AppOption        func(a *YouTrackApp)
	AppOptionCreator func() []AppOption
)

// ServerConnector يربط App بـ Channels ليحصل على وصول للـ Store و Config و Server.
// هذا هو الخيار الأساسي الذي يجب تمريره عند إنشاء App في كل طلب HTTP.
func ServerConnector(ch *YouTrackChannels) AppOption {
	return func(a *YouTrackApp) {
		a.channels = ch
	}
}

// ── Server Options ──

// WithPlatformService يعيّن PlatformService للخادم.
func WithPlatformService(ps *platform.YouTrackPlatformService) Option {
	return func(s *YouTrackServer) error {
		s.platform = ps
		return nil
	}
}

// WithStore ينشئ PlatformService مع Store محدد.
func WithStore(dsn string) Option {
	return func(s *YouTrackServer) error {
		if err := ensurePlatform(s); err != nil {
			return err
		}

		st, err := sqlstore.New(dsn)
		if err != nil {
			return err
		}
		if err := st.RunMigrations(); err != nil {
			st.Close()
			return err
		}

		retryStore := retrylayer.New(st)
		timerStore := timerlayer.New(retryStore)
		s.platform.SetStore(timerStore)
		return nil
	}
}

// WithConfig يعيّن إعدادات الخادم.
func WithConfig(cfg *model.ServerConfig) Option {
	return func(s *YouTrackServer) error {
		if err := ensurePlatform(s); err != nil {
			return err
		}
		s.platform.SetConfig(cfg)
		return nil
	}
}

// WithJWTSecret يعيّن المفتاح السري للتوكنات.
func WithJWTSecret(secret string) Option {
	return func(s *YouTrackServer) error {
		if err := ensurePlatform(s); err != nil {
			return err
		}
		s.platform.SetJWTSecret(secret)
		return nil
	}
}

// WithListenerFactory injects listener creation for startup testing or custom transports.
func WithListenerFactory(factory ListenerFactory) Option {
	return func(s *YouTrackServer) error {
		if factory == nil {
			return errors.New("listener factory is nil")
		}
		s.listenerFunc = factory
		return nil
	}
}
