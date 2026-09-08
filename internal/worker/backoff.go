package worker

import (
	"context"
	"math"
	"math/rand/v2"
	"time"
)

// BackoffStrategy computes backoff delays with jitter to avoid thundering herd problems.
type BackoffStrategy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Factor    float64
	Jitter    float64
}

// DefaultBackoff provides production-tuned backoff parameters for network probes.
func DefaultBackoff() *BackoffStrategy {
	return &BackoffStrategy{
		BaseDelay: 200 * time.Millisecond,
		MaxDelay:  3 * time.Second,
		Factor:    2.0,
		Jitter:    0.2, // +/- 20% random jitter
	}
}

// Calculate computes the backoff duration for a given attempt (0-indexed).
func (b *BackoffStrategy) Calculate(attempt int) time.Duration {
	if attempt <= 0 {
		return b.BaseDelay
	}

	multiplier := math.Pow(b.Factor, float64(attempt))
	delaySec := b.BaseDelay.Seconds() * multiplier

	if delaySec > b.MaxDelay.Seconds() {
		delaySec = b.MaxDelay.Seconds()
	}

	// Apply jitter: delay * (1 + random(-jitter, +jitter))
	jitterFactor := 1.0 + (rand.Float64()*2-1)*b.Jitter
	finalDelay := delaySec * jitterFactor

	if finalDelay < 0.05 {
		finalDelay = 0.05
	}

	return time.Duration(finalDelay * float64(time.Second))
}

// Sleep pauses execution for the calculated backoff period, respecting context cancellation.
func (b *BackoffStrategy) Sleep(ctx context.Context, attempt int) error {
	delay := b.Calculate(attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
