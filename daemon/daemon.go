package daemon

import (
	"context"
	"time"
)

type ScheduledJobFunc func(ctx context.Context) error

type Daemon interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type TimerDaemon struct {
	job    ScheduledJobFunc
	period time.Duration
	stop   chan struct{}
}

var _ Daemon = (*TimerDaemon)(nil) // Ensure TimerDaemon implements Daemon

func NewTimerDaemon(job ScheduledJobFunc, period time.Duration) *TimerDaemon {
	return &TimerDaemon{
		job:    job,
		period: period,
		stop:   make(chan struct{}),
	}
}

func (t *TimerDaemon) Start(ctx context.Context) error {
	panic("implement me")
}

func (t *TimerDaemon) Stop(ctx context.Context) error {
	panic("implement me")
}
