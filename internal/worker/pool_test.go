package worker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/config"
	"watchtower/internal/worker"
)

func TestPool_ExecuteAndDrain(t *testing.T) {
	var requestCount int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan worker.Job, 10)
	results := make(chan checker.Result, 10)

	pool := worker.NewPool()
	pool.Start(ctx, 3, jobs, results)

	totalJobs := 5
	for i := 0; i < totalJobs; i++ {
		jobs <- worker.Job{
			Target: config.Target{
				Name:     "pool-test",
				Type:     "http",
				Address:  server.URL,
				Timeout:  1 * time.Second,
				Interval: 5 * time.Second,
			},
		}
	}

	// Close jobs to initiate drain
	close(jobs)
	pool.Wait()
	close(results)

	received := 0
	for res := range results {
		if res.Status != checker.StatusUp {
			t.Errorf("expected StatusUp, got %s", res.Status)
		}
		received++
	}

	if received != totalJobs {
		t.Fatalf("expected %d results, got %d", totalJobs, received)
	}
	if atomic.LoadInt64(&requestCount) != int64(totalJobs) {
		t.Errorf("expected %d server requests, got %d", totalJobs, requestCount)
	}
}

func TestPool_RetriesOnFailure(t *testing.T) {
	var attempts int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cnt := atomic.AddInt64(&attempts, 1)
		if cnt <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan worker.Job, 1)
	results := make(chan checker.Result, 1)

	pool := worker.NewPool()
	pool.Start(ctx, 1, jobs, results)

	jobs <- worker.Job{
		Target: config.Target{
			Name:     "retry-target",
			Type:     "http",
			Address:  server.URL,
			Timeout:  1 * time.Second,
			Interval: 5 * time.Second,
			Retries:  2,
		},
	}

	close(jobs)
	pool.Wait()
	close(results)

	res := <-results
	if res.Status != checker.StatusUp {
		t.Fatalf("expected target to recover to StatusUp after 2 retries, got %s (attempts: %d)", res.Status, res.Attempts)
	}
	if res.Attempts != 3 {
		t.Errorf("expected 3 total attempts, got %d", res.Attempts)
	}
}
