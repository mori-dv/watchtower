package scheduler

import (
	"context"
	"time"

	"watchtower/internal/config"
	"watchtower/internal/metrics"
	"watchtower/internal/worker"
)

type Scheduler struct {
	targets []config.Target
	jobs    chan worker.Job
}
func NewScheduler(
	targets []config.Target,
	jobs chan worker.Job,
) *Scheduler {

	return &Scheduler{
		targets: targets,
		jobs:    jobs,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
		for _, target := range s.targets {

		go func(target config.Target) {
			ticker := time.NewTicker(target.Interval)

			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					metrics.QueueSize.Set(float64(len(s.jobs)))
					s.jobs <- worker.Job{
						Target: target,
					}

				}
			}
		}(target)
	}
}

