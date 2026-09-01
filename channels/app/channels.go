package app

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/shared/mlog"
	"youtrack_backend/channels/store"
)

// ServerChannels يحتوي وينسّق جميع خدمات نطاق التطبيق (Users, Projects, Issues, Admin, Subscriptions, Search)
// مطابقاً لهيكل Mattermost ServerChannels.
type ServerChannels struct {
	srv               *YouTrackServer
	dndTaskMut        sync.Mutex
	dndTask           *model.ScheduledTask
	interruptQuitChan chan struct{}
	scheduledPostMut  sync.Mutex
	scheduledPostTask *model.ScheduledTask
}

func (ch *ServerChannels) Start() error {
	// ctx := request.EmptyContext(ch.srv.Log())
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-interruptChan:
			if err := ch.Stop(); err != nil {
				ch.srv.Log().Warn("Error stopping channels", mlog.Err(err))
			}
			os.Exit(1)
		case <-ch.interruptQuitChan:
			return
		}
	}()
	return nil
}

func (ch *ServerChannels) Stop() error {
	ch.dndTaskMut.Lock()
	if ch.dndTask != nil {
		ch.dndTask.Cancel()
	}
	ch.dndTaskMut.Unlock()

	close(ch.interruptQuitChan)

	return nil
}

// NewChannels ينشئ كائن Channels جديد مرتبط بخدمة المنصة PlatformService.
func NewChannels(server *YouTrackServer) *ServerChannels {
	return &ServerChannels{
		srv: server,
	}
}

// Platform يعيد مرجع خدمة المنصة PlatformService.
func (ch *ServerChannels) Platform() *platform.YouTrackPlatformService {
	return ch.srv.platform
}

// Store يعيد الوصول المباشر لمخزن البيانات عبر خدمة المنصة.
func (ch *ServerChannels) Store() store.Store {
	return ch.srv.Store()
}

// Config يعيد إعدادات النظام عبر خدمة المنصة.
func (ch *ServerChannels) Config() *model.ServerConfig {
	return ch.srv.Config()
}

// JWTSecret يعيد المفتاح السري للتوكنات.
func (ch *ServerChannels) JWTSecret() string {
	return ch.srv.JWTSecret()
}
