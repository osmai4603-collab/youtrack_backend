package platform

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/shared/mlog"
	"youtrack_backend/channels/store"
)

// YouTrackPlatformService هي الخدمة المسؤولة عن مهام البنية التحتية والمنصة غير المرتبطة مباشرة بالكيانات
// (مثل الوصول لقاعدة البيانات، الإعدادات، التخزين المؤقت، دورة الحياة) مطابقاً لـ Mattermost YouTrackPlatformService.
type YouTrackPlatformService struct {
	store       store.Store
	config      *model.ServerConfig
	jwtSecret   string
	logger      *mlog.Logger
	startOnce   sync.Once
	shutdownOnce sync.Once
}

func (ps *YouTrackPlatformService) Logger() *mlog.Logger {
	return ps.logger
}

// Option تحدد دوال التكوين المرن لـ PlatformService.
type Option func(*YouTrackPlatformService)

// ServiceOptionStore يحدد مخزن البيانات للـ PlatformService.
func ServiceOptionStore(s store.Store) Option {
	return func(ps *YouTrackPlatformService) {
		ps.store = s
	}
}

// ServiceOptionConfig يحدد إعدادات الخادم للـ PlatformService.
func ServiceOptionConfig(cfg *model.ServerConfig) Option {
	return func(ps *YouTrackPlatformService) {
		ps.config = cfg
	}
}

// ServiceOptionJWTSecret يحدد المفتاح السري للتوكنات.
func ServiceOptionJWTSecret(secret string) Option {
	return func(ps *YouTrackPlatformService) {
		ps.jwtSecret = secret
	}
}

// New ينشئ نسخة جديدة من PlatformService مع تطبيق الخيارات المحددة.
func New(options ...Option) (*YouTrackPlatformService, error) {
	ps := &YouTrackPlatformService{}
	for _, opt := range options {
		opt(ps)
	}
	return ps, nil
}

// Store يعيد مخزن البيانات الرئيسي للخدمة.
func (ps *YouTrackPlatformService) Store() store.Store {
	return ps.store
}

// SetStore يتيح تغيير أو تعيين مخزن البيانات.
func (ps *YouTrackPlatformService) SetStore(s store.Store) {
	ps.store = s
}

// Config يعيد إعدادات الخادم.
func (ps *YouTrackPlatformService) Config() *model.ServerConfig {
	return ps.config
}

// SetConfig يتيح تعيين إعدادات الخادم.
func (ps *YouTrackPlatformService) SetConfig(cfg *model.ServerConfig) {
	ps.config = cfg
}

// JWTSecret يعيد المفتاح السري للتوكنات.
func (ps *YouTrackPlatformService) JWTSecret() string {
	return ps.jwtSecret
}

// SetJWTSecret يتيح ضبط المفتاح السري.
func (ps *YouTrackPlatformService) SetJWTSecret(secret string) {
	ps.jwtSecret = secret
}

func (ps *YouTrackPlatformService) Validate() error {
	if ps == nil {
		return errors.New("platform service is nil")
	}
	if ps.config == nil {
		return errors.New("platform config is required")
	}
	if ps.config.ServerPort == "" {
		return errors.New("server port is required")
	}
	if ps.config.AppEnv == "" {
		return errors.New("app env is required")
	}
	if ps.jwtSecret == "" {
		return errors.New("jwt secret is required")
	}
	if ps.store != nil {
		if err := ps.store.Ready(context.Background()); err != nil {
			return fmt.Errorf("database store is not ready: %w", err)
		}
	}
	return nil
}

// Start يبدأ أي خدمات خلفية تابعة للمنصة.
func (ps *YouTrackPlatformService) Start() error {
	if ps == nil {
		return nil
	}
	if err := ps.Validate(); err != nil {
		return err
	}
	ps.startOnce.Do(func() {
		if ps.logger == nil {
			logger, err := mlog.NewLogger()
			if err == nil {
				ps.logger = logger
			}
		}
	})
	return nil
}

// Shutdown يغلق خدمات وموارد المنصة بأمان.
func (ps *YouTrackPlatformService) Shutdown() error {
	if ps == nil {
		return nil
	}

	var shutdownErr error
	ps.shutdownOnce.Do(func() {
		if closer, ok := ps.store.(interface{ Close() }); ok {
			closer.Close()
		}
	})
	return shutdownErr
}
