// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"fmt"
	"sync"
	"time"
)

type TaskFunc func()

type ScheduledTask struct {
	Name                 string        `json:"name"`
	Interval             time.Duration `json:"interval"`
	Recurring            bool          `json:"recurring"`
	function             func()
	cancel               chan struct{}
	cancelled            chan struct{}
	cancelOnce           sync.Once
	cancelledOnce        sync.Once
	startOnce            sync.Once
	stateMu              sync.Mutex
	started              bool
	fromNextIntervalTime bool
}

func CreateTask(name string, function TaskFunc, timeToExecution time.Duration) *ScheduledTask {
	return createTask(name, function, timeToExecution, false, false)
}

func CreateRecurringTask(name string, function TaskFunc, interval time.Duration) *ScheduledTask {
	return createTask(name, function, interval, true, false)
}

func CreateRecurringTaskFromNextIntervalTime(name string, function TaskFunc, interval time.Duration) *ScheduledTask {
	return createTask(name, function, interval, true, true)
}

func createTask(name string, function TaskFunc, interval time.Duration, recurring bool, fromNextIntervalTime bool) *ScheduledTask {
	task := newTask(name, function, interval, recurring, fromNextIntervalTime)
	task.Start()
	return task
}

func NewTask(name string, function TaskFunc, timeToExecution time.Duration) *ScheduledTask {
	return newTask(name, function, timeToExecution, false, false)
}

func NewRecurringTask(name string, function TaskFunc, interval time.Duration) *ScheduledTask {
	return newTask(name, function, interval, true, false)
}

func NewRecurringTaskFromNextIntervalTime(name string, function TaskFunc, interval time.Duration) *ScheduledTask {
	return newTask(name, function, interval, true, true)
}

func newTask(name string, function TaskFunc, interval time.Duration, recurring bool, fromNextIntervalTime bool) *ScheduledTask {
	task := &ScheduledTask{
		Name:                 name,
		Interval:             interval,
		Recurring:            recurring,
		function:             function,
		cancel:               make(chan struct{}),
		cancelled:            make(chan struct{}),
		fromNextIntervalTime: fromNextIntervalTime,
	}

	return task
}

func (task *ScheduledTask) Start() {
	if task == nil {
		return
	}
	task.startOnce.Do(func() {
		task.stateMu.Lock()
		task.started = true
		task.stateMu.Unlock()
		go func() {
			defer task.cancelledOnce.Do(func() { close(task.cancelled) })

			var firstTick <-chan time.Time
			var ticker *time.Ticker

			if task.fromNextIntervalTime {
				currTime := time.Now()
				first := currTime.Truncate(task.Interval)
				if first.Before(currTime) {
					first = first.Add(task.Interval)
				}
				firstTick = time.After(time.Until(first))
				ticker = &time.Ticker{C: nil}
			} else {
				firstTick = nil
				ticker = time.NewTicker(task.Interval)
			}
			defer func() {
				ticker.Stop()
			}()

			for {
				select {
				case <-firstTick:
					ticker = time.NewTicker(task.Interval)
					task.function()
				case <-ticker.C:
					task.function()
				case <-task.cancel:
					return
				}

				if !task.Recurring {
					break
				}
			}
		}()
	})
}

func (task *ScheduledTask) Cancel() {
	if task == nil {
		return
	}
	task.cancelOnce.Do(func() { close(task.cancel) })
	task.stateMu.Lock()
	started := task.started
	task.stateMu.Unlock()
	if !started {
		task.cancelledOnce.Do(func() { close(task.cancelled) })
		return
	}
	<-task.cancelled
}

func (task *ScheduledTask) String() string {
	return fmt.Sprintf(
		"%s\nInterval: %s\nRecurring: %t\n",
		task.Name,
		task.Interval.String(),
		task.Recurring,
	)
}
