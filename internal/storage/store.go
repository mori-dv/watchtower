package storage

import (
	"context"
	"time"

	"watchtower/internal/checker"
)

// Store defines operations for tracking probe failure state and alert cooldowns.
type Store interface {
	// IncrFailures increments the failure counter for a target and returns the new count.
	IncrFailures(ctx context.Context, target string) (int64, error)

	// ResetFailures resets the failure counter for a target back to 0.
	ResetFailures(ctx context.Context, target string) error

	// GetFailures returns the current consecutive failure count for a target.
	GetFailures(ctx context.Context, target string) (int64, error)

	// SetCooldown sets an expiration cooldown window for an alert key.
	SetCooldown(ctx context.Context, key string, ttl time.Duration) error

	// IsOnCooldown checks whether an active cooldown exists for the key.
	IsOnCooldown(ctx context.Context, key string) (bool, error)

	// SetLastStatus persists the last evaluated status of a target.
	SetLastStatus(ctx context.Context, target string, status checker.Status) error

	// GetLastStatus retrieves the previously stored status of a target.
	GetLastStatus(ctx context.Context, target string) (checker.Status, error)

	// Close cleans up any open connections.
	Close() error
}
