package model

import (
	"testing"
	"time"
)

func TestScheduledTaskCancelIsIdempotent(t *testing.T) {
	task := CreateRecurringTask("test", func() {}, time.Hour)
	task.Cancel()
	task.Cancel()
}

func TestUnstartedScheduledTaskCanBeCanceled(t *testing.T) {
	task := NewRecurringTask("unstarted", func() {}, time.Hour)
	task.Cancel()
	task.Start()
	task.Cancel()
}
