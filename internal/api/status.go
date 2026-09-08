package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"watchtower/internal/checker"
)

// TargetSnapshot captures live operational status for an individual target.
type TargetSnapshot struct {
	Name                string         `json:"name"`
	Type                string         `json:"type"`
	Address             string         `json:"address"`
	Status              checker.Status `json:"status"`
	StatusCode          int            `json:"status_code,omitempty"`
	LatencyMs           float64        `json:"latency_ms"`
	SSLExpiryDays       *int           `json:"ssl_expiry_days,omitempty"`
	ConsecutiveFailures int            `json:"consecutive_failures"`
	Error               string         `json:"error,omitempty"`
	LastChecked         time.Time      `json:"last_checked"`
}

// StatusRegistry maintains the latest probe snapshots in-memory for the status API.
type StatusRegistry struct {
	mu        sync.RWMutex
	snapshots map[string]TargetSnapshot
}

func NewStatusRegistry() *StatusRegistry {
	return &StatusRegistry{
		snapshots: make(map[string]TargetSnapshot),
	}
}

func (s *StatusRegistry) Update(res checker.Result, consecutiveFailures int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots[res.TargetName] = TargetSnapshot{
		Name:                res.TargetName,
		Type:                res.TargetType,
		Address:             res.TargetAddress,
		Status:              res.Status,
		StatusCode:          res.StatusCode,
		LatencyMs:           res.LatencyMs,
		SSLExpiryDays:       res.SSLExpiryDays,
		ConsecutiveFailures: consecutiveFailures,
		Error:               res.Error,
		LastChecked:         res.CheckedAt,
	}
}

func (s *StatusRegistry) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		list := make([]TargetSnapshot, 0, len(s.snapshots))
		for _, snap := range s.snapshots {
			list = append(list, snap)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   len(list),
			"targets": list,
		})
	}
}

// ReloadHandler provides an HTTP endpoint (POST /reload) to trigger configuration reload.
func ReloadHandler(reloadFn func() error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed - use POST", http.StatusMethodNotAllowed)
			return
		}

		if err := reloadFn(); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "configuration reloaded successfully",
		})
	}
}