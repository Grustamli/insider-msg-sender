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
	jobName string
	job     ScheduledJobFunc
	period  time.Duration
	stop    chan struct{}
	logger  *zerolog.Logger
	running bool
	mu      sync.Mutex
}

var _ Daemon = (*TimerDaemon)(nil) // Ensure TimerDaemon implements Daemon

func NewTimerDaemon(jobName string, job ScheduledJobFunc, period time.Duration, logger *zerolog.Logger) *TimerDaemon {
	return &TimerDaemon{
		jobName: jobName,
		job:     job,
		period:  period,
		stop:    make(chan struct{}),
		logger:  logger,
	}
}

func (t *TimerDaemon) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return nil
	}

	t.logger.Debug().Msgf("Starting daemon for: %s", t.jobName)
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
	t.logger.Debug().Msgf("Stopped daemon for: %s", t.jobName)
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
				t.logger.Debug().Msgf("running job: %s", t.jobName)
				if err := t.job(ctx); err != nil {
					t.logger.Error().Err(err).Msgf("job failed: %s", t.jobName)
				}
				t.logger.Debug().Msgf("finished job: %s", t.jobName)
			}()
		}
	}
}
