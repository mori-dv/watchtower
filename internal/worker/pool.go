package worker

import (
	"context"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/config"
	"watchtower/internal/metrics"
)


type Job struct {
	Target config.Target
}

type Pool struct {
	httpChecker *checker.HTTPChecker
	tcpChecker  *checker.TCPChecker
}

func NewPool() *Pool {
	return &Pool{
		httpChecker: checker.NewHTTPChecker(),
		tcpChecker:  checker.NewTCPChecker(),
	}
}

func (p *Pool) Start(
	ctx context.Context,
	workerCount int,
	jobs <-chan Job,
	results chan<- checker.Result,
) {
	for range workerCount {
		go p.worker(ctx, jobs, results)
	}
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
			
			checkCtx, cancel := context.WithTimeout(
				ctx,
				job.Target.Timeout,
			)
			var result checker.Result
			for i := 0; i <= job.Target.Retries; i++ {
				switch job.Target.Type {

				case "http":
					result = p.httpChecker.Check(
						checkCtx,
						job.Target,
					)

				case "tcp":
					result = p.tcpChecker.Check(
						checkCtx,
						job.Target,
					)
				}
				
				if result.Status == checker.StatusUp {
					break
				}
				time.Sleep(time.Second)
			}
			cancel()
			results <- result

			metrics.ActiveWorkers.Dec()
		}
	}
}
