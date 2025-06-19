package daemon

import (
	"context"
	"github.com/rs/zerolog"
	"sync"
	"time"
)

type ScheduledJobFunc func(ctx context.Context) error

type Daemon interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type TimerDaemon struct {
	job     ScheduledJobFunc
	period  time.Duration
	stop    chan struct{}
	logger  *zerolog.Logger
	running bool
	mu      sync.Mutex
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
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return nil
	}

	t.running = true

	go t.runJob(ctx)

	return nil
}

func (t *TimerDaemon) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.running {
		return nil // Already stopped
	}
	close(t.stop)                // Signal the daemon to stop
	t.stop = make(chan struct{}) // Reset the channel for future starts
	return nil
}

func (t *TimerDaemon) runJob(ctx context.Context) {
	defer func() {
		t.mu.Lock()
		t.running = false
		t.mu.Unlock()
	}()

	ticker := time.NewTicker(t.period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.stop:
			return
		case <-ticker.C:
			// Execute the job, but don't let it block the daemon
			go func() {
				if err := t.job(ctx); err != nil {
					t.logger.Error().Err(err).Msg("job failed")
				}
			}()
		}
	}
}
