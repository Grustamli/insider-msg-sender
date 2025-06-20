package daemon_test

import (
	"context"
	"github.com/grustamli/insider-msg-sender/daemon"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestTimerDaemon_ExecutesJobAtInterval(t *testing.T) {
	// use a very short period so the test runs quickly
	period := 20 * time.Millisecond

	// counter for how many times our job was invoked
	var count int32

	// our scheduled job just increments the counter
	job := func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	// a no-op logger to satisfy the API
	logger := zerolog.New(io.Discard).With().Timestamp().Logger()

	td := daemon.NewTimerDaemon("test-job", job, period, &logger)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	if err := td.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	// wait until context times out
	<-ctx.Done()

	// stop should be idempotent
	if err := td.Stop(context.Background()); err != nil {
		t.Errorf("Stop returned error: %v", err)
	}

	got := atomic.LoadInt32(&count)
	// we expect around 8–12 invocations in 200ms at 20ms intervals
	if got < 5 {
		t.Errorf("expected at least 5 job runs, got %d", got)
	}
}

func TestTimerDaemon_StopPreventsFurtherRuns(t *testing.T) {
	period := 10 * time.Millisecond
	var count int32

	job := func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	logger := zerolog.New(io.Discard).With().Timestamp().Logger()
	td := daemon.NewTimerDaemon("stop-test", job, period, &logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := td.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	// let it run a few times
	time.Sleep(50 * time.Millisecond)

	// capture count so far
	before := atomic.LoadInt32(&count)

	// now stop
	if err := td.Stop(context.Background()); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}

	// clear counter and wait to see if any more runs occur
	atomic.StoreInt32(&count, 0)
	time.Sleep(50 * time.Millisecond)
	after := atomic.LoadInt32(&count)

	if before < 1 {
		t.Errorf("expected at least 1 run before stop, got %d", before)
	}
	if after != 0 {
		t.Errorf("expected 0 runs after stop, got %d", after)
	}
}

func TestTimerDaemon_MultipleStartCallsDoNothing(t *testing.T) {
	period := 50 * time.Millisecond
	var count int32

	job := func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	logger := zerolog.New(io.Discard).With().Timestamp().Logger()
	td := daemon.NewTimerDaemon("idempotent-start", job, period, &logger)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	// call Start twice in a row
	if err := td.Start(ctx); err != nil {
		t.Fatalf("first Start error: %v", err)
	}
	if err := td.Start(ctx); err != nil {
		t.Fatalf("second Start error: %v", err)
	}

	<-ctx.Done()
	// roughly 2–3 invocations at 50ms over 120ms
	if got := atomic.LoadInt32(&count); got < 2 {
		t.Errorf("expected at least 2 runs, got %d", got)
	}

	// clean up
	_ = td.Stop(context.Background())
}
