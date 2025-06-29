// Package daemon provides a simple timer-based daemon implementation
// that periodically executes a user-defined job function.
package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// ScheduledJobFunc defines the signature for functions executed by the daemon.
// The provided context should be used to observe cancellation.
type ScheduledJobFunc func(ctx context.Context) error

// Daemon represents a long-running background process that can be started and stopped.
type Daemon interface {
	// Start begins execution of the daemon's job at the configured interval.
	// If already running, Start does nothing.
	// Returns an error only on configuration or startup failures.
	Start(ctx context.Context) error

	// Stop signals the daemon to cease executing its job.
	// If not running, Stop does nothing.
	// Returns an error only on shutdown failures.
	Stop(ctx context.Context) error
}

// TimerDaemon runs a ScheduledJobFunc at a fixed period using time.Ticker.
// It logs start/stop events and job execution via zerolog.Logger.
type TimerDaemon struct {
	jobName string           // descriptive name for logging
	job     ScheduledJobFunc // function to execute periodically
	period  time.Duration    // interval between job executions
	stop    chan struct{}    // channel to signal stop
	logger  *zerolog.Logger  // logger for lifecycle and job events
	running bool             // indicates if the daemon is active
	mu      sync.Mutex       // protects running and stop fields
}

// Ensure TimerDaemon implements the Daemon interface.
var _ Daemon = (*TimerDaemon)(nil)

// NewTimerDaemon constructs a new TimerDaemon with the given job, period, and logger.
// jobName is used in log messages to identify this daemon instance.
func NewTimerDaemon(jobName string, job ScheduledJobFunc, period time.Duration, logger *zerolog.Logger) *TimerDaemon {
	return &TimerDaemon{
		jobName: jobName,
		job:     job,
		period:  period,
		stop:    make(chan struct{}),
		logger:  logger,
	}
}

// Start begins the periodic execution of the daemon's job.
// It spawns a goroutine to run the job loop and logs the start event.
// Subsequent calls to Start while running have no effect.
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

// Stop signals the daemon to stop and resets its internal state.
// It logs the stop event. If not running, Stop returns immediately.
func (t *TimerDaemon) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.running {
		// Already stopped, nothing to do
		return nil
	}
	// signal the background loop to exit
	close(t.stop)
	// prepare channel for potential future restarts
	t.stop = make(chan struct{})
	t.logger.Debug().Msgf("Stopped daemon for: %s", t.jobName)
	return nil
}

// runJob contains the main loop that triggers the job at each tick.
// It listens for context cancellation or stop signals to exit cleanly.
func (t *TimerDaemon) runJob(ctx context.Context) {
	// ensure running flag is cleared when this goroutine exits
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
			// context canceled, exit
			return
		case <-t.stop:
			// explicit stop signal, exit
			return
		case <-ticker.C:
			// trigger the job asynchronously to avoid blocking
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
