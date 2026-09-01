package app

import (
	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/store"
)

// YouTrackApp هو كائن منطق الأعمال المعياري المنشأ لكل طلب، ويحمل مؤشراً إلى Server و Channels.
type YouTrackApp struct {
	channels *ServerChannels
}

func New(options ...AppOption) *YouTrackApp {
	app := &YouTrackApp{}

	for _, option := range options {
		option(app)
	}

	return app
}

// Srv يعيد الخادم المرتبط.
func (a *YouTrackApp) Srv() *YouTrackServer { return a.channels.srv }

// Channels يعيد منسق خدمات النطاق.
func (a *YouTrackApp) Channels() *ServerChannels { return a.channels }

// Platform يعيد خدمة المنصة.
func (a *YouTrackApp) Platform() *platform.YouTrackPlatformService {
	return a.channels.Platform()
}

// Store يعيد الوصول إلى مستودعات البيانات عبر Channels و PlatformService.
func (a *YouTrackApp) Store() store.Store {
	return a.channels.Store()
}

// JWTSecret يعيد المفتاح السري للتوكنات.
func (a *YouTrackApp) JWTSecret() string {
	return a.channels.JWTSecret()
}
