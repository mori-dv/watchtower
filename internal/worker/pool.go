package worker

import (
	"context"
	"sync"

	"watchtower/internal/checker"
	"watchtower/internal/config"
	"watchtower/internal/metrics"
)

type Job struct {
	Target config.Target
}

type Pool struct {
	engine  *checker.Engine
	backoff *BackoffStrategy
	wg      sync.WaitGroup
}

func NewPool() *Pool {
	return &Pool{
		engine:  checker.NewEngine(),
		backoff: DefaultBackoff(),
	}
}

// Start launches worker goroutines and tracks them in a sync.WaitGroup.
func (p *Pool) Start(
	ctx context.Context,
	workerCount int,
	jobs <-chan Job,
	results chan<- checker.Result,
) {
	for i := 0; i < workerCount; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.worker(ctx, jobs, results)
		}()
	}
}

// Wait blocks until all active worker goroutines finish executing in-flight jobs.
func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) worker(
	ctx context.Context,
	jobs <-chan Job,
	results chan<- checker.Result,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			metrics.ActiveWorkers.Inc()
			result := p.executeJob(ctx, job)
			metrics.ActiveWorkers.Dec()

			// Safely emit result without deadlocking if context terminates
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// executeJob executes the probe with per-attempt timeout and exponential backoff.
func (p *Pool) executeJob(ctx context.Context, job Job) checker.Result {
	var result checker.Result
	retries := job.Target.Retries
	if retries < 0 {
		retries = 0
	}

	for attempt := 0; attempt <= retries; attempt++ {
		// Strict per-attempt timeout to prevent runaway checks
		checkCtx, cancel := context.WithTimeout(ctx, job.Target.Timeout)
		result = p.engine.Check(checkCtx, job.Target)
		cancel()

		result.Attempts = attempt + 1

		if result.Status == checker.StatusUp {
			return result
		}

		// If more retries remain, apply exponential backoff with jitter
		if attempt < retries {
			if err := p.backoff.Sleep(ctx, attempt); err != nil {
				// Context cancelled during backoff sleep
				result.Error = "probe interrupted: " + err.Error()
				return result
			}
		}
	}

	return result
}

