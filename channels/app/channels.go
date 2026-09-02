package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/shared/mlog"
	"youtrack_backend/channels/store"
)

// YouTrackChannels يحتوي وينسّق جميع خدمات نطاق التطبيق (Users, Projects, Issues, Admin, Subscriptions, Search)
// مطابقاً لهيكل Mattermost YouTrackChannels.
type YouTrackChannels struct {
	server               *YouTrackServer
	dndTaskMut           sync.Mutex
	dndTask              *model.ScheduledTask
	interruptQuitChan    chan struct{}
	scheduledPostMut     sync.Mutex
	scheduledPostTask    *model.ScheduledTask
	startOnce            sync.Once
	stopOnce             sync.Once
}

func (ch *YouTrackChannels) Start() error {
	if ch == nil {
		return nil
	}

	ch.startOnce.Do(func() {
		if ch.interruptQuitChan == nil {
			ch.interruptQuitChan = make(chan struct{})
		}

		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			defer signal.Stop(interruptChan)
			select {
			case <-interruptChan:
				if err := ch.Stop(); err != nil {
					if ch.server != nil && ch.server.Log() != nil {
						ch.server.Log().Warn("Error stopping channels", mlog.Err(err))
					}
				}
				if ch.server != nil && ch.server.Server != nil {
					ch.server.Server.SetKeepAlivesEnabled(false)
					shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					if err := ch.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
						if ch.server.Log() != nil {
							ch.server.Log().Warn("Error during graceful shutdown", mlog.Err(err))
						}
					}
				}
			case <-ch.interruptQuitChan:
				return
			}
		}()
	})

	return nil
}

func (ch *YouTrackChannels) Stop() error {
	if ch == nil {
		return nil
	}

	var stopErr error
	ch.stopOnce.Do(func() {
		ch.dndTaskMut.Lock()
		if ch.dndTask != nil {
			ch.dndTask.Cancel()
		}
		ch.dndTaskMut.Unlock()

		select {
		case <-ch.interruptQuitChan:
		default:
			if ch.interruptQuitChan != nil {
				close(ch.interruptQuitChan)
			}
		}
	})
	return stopErr
}

// NewChannels ينشئ كائن Channels جديد مرتبط بخدمة المنصة PlatformService.
func NewChannels(server *YouTrackServer) *YouTrackChannels {
	return &YouTrackChannels{
		server:            server,
		interruptQuitChan: make(chan struct{}),
	}
}

// Platform يعيد مرجع خدمة المنصة PlatformService.
func (ch *YouTrackChannels) Platform() *platform.YouTrackPlatformService {
	if ch == nil || ch.server == nil {
		return nil
	}
	return ch.server.platform
}

// Store يعيد الوصول المباشر لمخزن البيانات عبر خدمة المنصة.
func (ch *YouTrackChannels) Store() store.Store {
	if ch == nil || ch.server == nil {
		return nil
	}
	return ch.server.Store()
}

// Config يعيد إعدادات النظام عبر خدمة المنصة.
func (ch *YouTrackChannels) Config() *model.ServerConfig {
	if ch == nil || ch.server == nil {
		return nil
	}
	return ch.server.Config()
}

// JWTSecret يعيد المفتاح السري للتوكنات.
func (ch *YouTrackChannels) JWTSecret() string {
	if ch == nil || ch.server == nil {
		return ""
	}
	return ch.server.JWTSecret()
}
