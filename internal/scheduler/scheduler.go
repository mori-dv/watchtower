package scheduler

import (
	"context"
	"reflect"
	"sync"
	"time"

	"watchtower/internal/config"
	"watchtower/internal/logging"
	"watchtower/internal/metrics"
	"watchtower/internal/worker"
)

type targetRunner struct {
	target config.Target
	cancel context.CancelFunc
}

// Scheduler coordinates periodic probe scheduling across dynamic targets.
type Scheduler struct {
	mu      sync.RWMutex
	runners map[string]*targetRunner
	jobs    chan<- worker.Job
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewScheduler(jobs chan<- worker.Job) *Scheduler {
	return &Scheduler{
		runners: make(map[string]*targetRunner),
		jobs:    jobs,
	}
}

// Start initiates scheduling for the provided targets under parent context.
func (s *Scheduler) Start(parentCtx context.Context, targets []config.Target) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ctx, s.cancel = context.WithCancel(parentCtx)
	for _, target := range targets {
		s.startRunnerLocked(target)
	}
}

// UpdateTargets dynamically synchronizes active runners with the new target list.
// New targets are started, removed targets are stopped, and modified targets are refreshed.
func (s *Scheduler) UpdateTargets(newTargets []config.Target) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx == nil || s.ctx.Err() != nil {
		return
	}

	newMap := make(map[string]config.Target, len(newTargets))
	for _, t := range newTargets {
		newMap[t.Name] = t
	}

	// Stop removed targets or targets with changed configs
	for name, runner := range s.runners {
		newT, exists := newMap[name]
		if !exists {
			logging.Logger.Info().Str("target", name).Msg("removing target from scheduler")
			runner.cancel()
			delete(s.runners, name)
			continue
		}

		if !reflect.DeepEqual(runner.target, newT) {
			logging.Logger.Info().Str("target", name).Msg("updating target configuration in scheduler")
			runner.cancel()
			delete(s.runners, name)
			s.startRunnerLocked(newT)
		}
	}

	// Start newly added targets
	for name, newT := range newMap {
		if _, exists := s.runners[name]; !exists {
			logging.Logger.Info().Str("target", name).Msg("adding new target to scheduler")
			s.startRunnerLocked(newT)
		}
	}
}

func (s *Scheduler) startRunnerLocked(target config.Target) {
	runnerCtx, cancel := context.WithCancel(s.ctx)
	runner := &targetRunner{
		target: target,
		cancel: cancel,
	}
	s.runners[target.Name] = runner

	s.wg.Add(1)
	go func(t config.Target, ctx context.Context) {
		defer s.wg.Done()

		// Initial immediate probe
		s.dispatch(ctx, t)

		ticker := time.NewTicker(t.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.dispatch(ctx, t)
			}
		}
	}(target, runnerCtx)
}

func (s *Scheduler) dispatch(ctx context.Context, target config.Target) {
	select {
	case <-ctx.Done():
		return
	case s.jobs <- worker.Job{Target: target}:
		metrics.QueueSize.Set(float64(len(s.jobs)))
	default:
		// Queue full: drop job and record metric instead of deadlocking the scheduler ticker
		metrics.JobsDroppedTotal.WithLabelValues(target.Name).Inc()
		logging.Logger.Warn().
			Str("target", target.Name).
			Msg("worker job queue full, dropped probe execution")
	}
}

// Stop shuts down all running target tickers and waits for routines to exit.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	s.wg.Wait()
}


