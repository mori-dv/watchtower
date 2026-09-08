package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

var startTime = time.Now()

type HealthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status: "ok",
		Uptime: time.Since(startTime).Truncate(time.Second).String(),
	})
}

// ReadyHandler verifies critical dependencies (e.g. storage) before reporting readiness.
func ReadyHandler(checkReady func(ctx context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if checkReady != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := checkReady(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status": "not_ready",
					"error":  err.Error(),
				})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ready",
		})
	}
}

