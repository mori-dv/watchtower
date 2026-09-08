package scheduler_test

import (
	"context"
	"testing"
	"time"

	"watchtower/internal/config"
	"watchtower/internal/scheduler"
	"watchtower/internal/worker"
)

func TestScheduler_ImmediateProbeAndDynamicReload(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan worker.Job, 20)
	sched := scheduler.NewScheduler(jobs)

	target1 := config.Target{
		Name:     "target-1",
		Type:     "http",
		Address:  "http://example.com/1",
		Interval: 200 * time.Millisecond,
		Timeout:  1 * time.Second,
	}

	sched.Start(ctx, []config.Target{target1})

	// Check that target1 produces an immediate probe without waiting for interval
	select {
	case j := <-jobs:
		if j.Target.Name != "target-1" {
			t.Errorf("expected job for target-1, got %s", j.Target.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for immediate initial probe")
	}

	// Dynamic target reload: replace target-1 with target-2
	target2 := config.Target{
		Name:     "target-2",
		Type:     "http",
		Address:  "http://example.com/2",
		Interval: 200 * time.Millisecond,
		Timeout:  1 * time.Second,
	}

	sched.UpdateTargets([]config.Target{target2})

	// Target-2 should now emit an immediate probe
	select {
	case j := <-jobs:
		if j.Target.Name != "target-2" {
			t.Errorf("expected job for newly added target-2, got %s", j.Target.Name)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for target-2 initial probe")
	}

	// Stop scheduler cleanly
	sched.Stop()
}
