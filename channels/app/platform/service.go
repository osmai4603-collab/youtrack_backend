package platform

import (
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/shared/mlog"
	"youtrack_backend/channels/store"
)

// YouTrackPlatformService هي الخدمة المسؤولة عن مهام البنية التحتية والمنصة غير المرتبطة مباشرة بالكيانات
// (مثل الوصول لقاعدة البيانات، الإعدادات، التخزين المؤقت، دورة الحياة) مطابقاً لـ Mattermost YouTrackPlatformService.
type YouTrackPlatformService struct {
	store     store.Store
	config    *model.ServerConfig
	jwtSecret string
	logger    *mlog.Logger
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

// Start يبدأ أي خدمات خلفية تابعة للمنصة.
func (ps *YouTrackPlatformService) Start() error {
	return nil
}

// Shutdown يغلق خدمات وموارد المنصة بأمان.
func (ps *YouTrackPlatformService) Shutdown() error {
	if closer, ok := ps.store.(interface{ Close() }); ok {
		closer.Close()
	}
	return nil
}
