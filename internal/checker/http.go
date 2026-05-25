package checker

import (
	"context"
	"net/http"
	"time"

	"watchtower/internal/config"
)

type HTTPChecker struct {
	client *http.Client
}

func NewHTTPChecker() *HTTPChecker {
	return &HTTPChecker{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (h *HTTPChecker) Check(
	ctx context.Context,
	target config.Target,
) Result {

	start := time.Now()

	result := Result{
		TargetName: target.Name,
		TargetType: target.Type,
		CheckedAt:  time.Now(),
		Status:     StatusDown,
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		target.Address,
		nil,
	)

	if err != nil {
		result.Error = err.Error()
		return result
	}
	resp, err := h.client.Do(req)

	latency := time.Since(start)
	result.Latency = latency
		if err != nil {
		result.Error = err.Error()
		return result
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Status = StatusUp
		return result
	}

	result.Error = resp.Status
	return result
}