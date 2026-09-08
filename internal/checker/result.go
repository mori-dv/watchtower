package checker

import "time"


// Result represents the outcome of a single probe execution.
type Result struct {
	TargetName    string        `json:"target_name"`
	TargetType    string        `json:"target_type"`
	TargetAddress string        `json:"target_address"`
	Status        Status        `json:"status"`
	StatusCode    int           `json:"status_code,omitempty"`
	Latency       time.Duration `json:"latency"`
	LatencyMs     float64       `json:"latency_ms"`
	SSLExpiryDays *int          `json:"ssl_expiry_days,omitempty"`
	Attempts      int           `json:"attempts"`
	Error         string        `json:"error,omitempty"`
	CheckedAt     time.Time     `json:"checked_at"`
}