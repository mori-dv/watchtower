package alert

import (
	"context"
	"time"

	"watchtower/internal/checker"
)

// AlertEvent captures the contextual metadata needed to build alert notifications.
type AlertEvent struct {
	TargetName          string         `json:"target_name"`
	TargetType          string         `json:"target_type"`
	TargetAddress       string         `json:"target_address"`
	Status              checker.Status `json:"status"`
	PreviousStatus      checker.Status `json:"previous_status"`
	IsRecovery          bool           `json:"is_recovery"`
	Latency             time.Duration  `json:"latency"`
	ConsecutiveFailures int            `json:"consecutive_failures"`
	Error               string         `json:"error,omitempty"`
	Timestamp           time.Time      `json:"timestamp"`
}

// Dispatcher represents a pluggable alert transport (Webhook, Slack, Telegram, etc.).
type Dispatcher interface {
	Name() string
	Send(ctx context.Context, event AlertEvent) error
}
