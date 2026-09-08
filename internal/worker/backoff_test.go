package worker_test

import (
	"context"
	"testing"
	"time"

	"watchtower/internal/worker"
)

func TestBackoffStrategy_Calculate(t *testing.T) {
	strategy := &worker.BackoffStrategy{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  1 * time.Second,
		Factor:    2.0,
		Jitter:    0.1, // 10%
	}

	d0 := strategy.Calculate(0)
	if d0 < 80*time.Millisecond || d0 > 120*time.Millisecond {
		t.Errorf("unexpected delay for attempt 0: %v", d0)
	}

	d1 := strategy.Calculate(1)
	if d1 < 160*time.Millisecond || d1 > 240*time.Millisecond {
		t.Errorf("unexpected delay for attempt 1: %v", d1)
	}

	// High attempt should cap at MaxDelay +/- jitter
	d10 := strategy.Calculate(10)
	if d10 > 1200*time.Millisecond {
		t.Errorf("delay exceeded max backoff cap: %v", d10)
	}
}

func TestBackoffStrategy_ContextCancellation(t *testing.T) {
	strategy := worker.DefaultBackoff()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	start := time.Now()
	err := strategy.Sleep(ctx, 3)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error on cancelled context, got nil")
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("sleep did not exit promptly on cancellation: took %v", elapsed)
	}
}
