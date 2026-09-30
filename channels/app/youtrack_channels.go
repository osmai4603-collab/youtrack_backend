package app

import (
	"context"
	"sync"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/store"
)

// YouTrackChannels يحتوي وينسّق جميع خدمات نطاق التطبيق (Users, Projects, Issues, Admin, Subscriptions, Search)
// مطابقاً لهيكل Mattermost YouTrackChannels.
type YouTrackChannels struct {
	server            *YouTrackServer
	interruptQuitChan chan struct{}
	scheduledTaskMu   sync.Mutex
	scheduledTasks    []*model.ScheduledTask
	scheduledStopped  bool
	startOnce         sync.Once
	stopOnce          sync.Once
}

func (ch *YouTrackChannels) Start() error {
	if ch == nil {
		return nil
	}

	ch.startOnce.Do(func() {
		if ch.interruptQuitChan == nil {
			ch.interruptQuitChan = make(chan struct{})
		}
	})

	return nil
}

func (ch *YouTrackChannels) Stop() error {
	if ch == nil {
		return nil
	}

	var stopErr error
	ch.stopOnce.Do(func() {
		ch.scheduledTaskMu.Lock()
		ch.scheduledStopped = true
		tasks := append([]*model.ScheduledTask(nil), ch.scheduledTasks...)
		ch.scheduledTasks = nil
		ch.scheduledTaskMu.Unlock()
		for _, task := range tasks {
			task.Cancel()
		}

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

// RegisterScheduledTask transfers ownership of a scheduled task to Channels.
// Stop cancels every registered task before it returns.
func (ch *YouTrackChannels) RegisterScheduledTask(task *model.ScheduledTask) bool {
	if ch == nil || task == nil {
		return false
	}
	ch.scheduledTaskMu.Lock()
	if ch.scheduledStopped {
		ch.scheduledTaskMu.Unlock()
		task.Cancel()
		return false
	}
	ch.scheduledTasks = append(ch.scheduledTasks, task)
	ch.scheduledTaskMu.Unlock()
	return true
}

// StartScheduledTask registers a task before starting its goroutine.
func (ch *YouTrackChannels) StartScheduledTask(task *model.ScheduledTask) bool {
	if !ch.RegisterScheduledTask(task) {
		return false
	}
	task.Start()
	return true
}

// StartWorker starts a channels worker under the platform lifecycle owner.
func (ch *YouTrackChannels) StartWorker(worker func(context.Context)) bool {
	if ch == nil || ch.Platform() == nil {
		return false
	}
	return ch.Platform().GoContext(worker)
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
