package alert

import (
	"context"
	"errors"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/logging"
	"watchtower/internal/metrics"
	"watchtower/internal/storage"
)

type Manager struct {
	store       storage.Store
	dispatchers []Dispatcher
	cooldown    time.Duration
}

func NewManager(
	store storage.Store,
	cooldown time.Duration,
	dispatchers ...Dispatcher,
) *Manager {
	return &Manager{
		store:       store,
		dispatchers: dispatchers,
		cooldown:    cooldown,
	}
}

func (m *Manager) Register(d Dispatcher) {
	m.dispatchers = append(m.dispatchers, d)
}

// Handle processes evaluated probe results, triggering alerts on confirmed outages and recoveries
// while debouncing repetitive alerts according to the configured cooldown.
func (m *Manager) Handle(
	ctx context.Context,
	result checker.Result,
	status checker.Status,
	prevStatus checker.Status,
	consecutiveFailures int,
	isRecovery bool,
) error {
	if len(m.dispatchers) == 0 {
		return nil
	}

	// 1. RECOVERY ALERT: previous status was DOWN, now UP
	if isRecovery && prevStatus == checker.StatusDown {
		event := AlertEvent{
			TargetName:          result.TargetName,
			TargetType:          result.TargetType,
			TargetAddress:       result.TargetAddress,
			Status:              status,
			PreviousStatus:      prevStatus,
			IsRecovery:          true,
			Latency:             result.Latency,
			ConsecutiveFailures: 0,
			Error:               "",
			Timestamp:           time.Now(),
		}

		// Clear outage cooldown on recovery
		_ = m.store.SetCooldown(ctx, "down:"+result.TargetName, 0)
		return m.dispatchAll(ctx, event)
	}

	// 2. OUTAGE ALERT: target is DOWN
	if status == checker.StatusDown {
		cooldownKey := "down:" + result.TargetName
		onCooldown, err := m.store.IsOnCooldown(ctx, cooldownKey)
		if err != nil {
			logging.Logger.Warn().Err(err).Str("target", result.TargetName).Msg("failed to check alert cooldown")
		}

		if onCooldown {
			// Suppress alert storm
			return nil
		}

		event := AlertEvent{
			TargetName:          result.TargetName,
			TargetType:          result.TargetType,
			TargetAddress:       result.TargetAddress,
			Status:              status,
			PreviousStatus:      prevStatus,
			IsRecovery:          false,
			Latency:             result.Latency,
			ConsecutiveFailures: consecutiveFailures,
			Error:               result.Error,
			Timestamp:           time.Now(),
		}

		// Set cooldown window
		if err := m.store.SetCooldown(ctx, cooldownKey, m.cooldown); err != nil {
			logging.Logger.Warn().Err(err).Str("target", result.TargetName).Msg("failed to set alert cooldown")
		}

		return m.dispatchAll(ctx, event)
	}

	return nil
}

func (m *Manager) dispatchAll(ctx context.Context, event AlertEvent) error {
	var errs []error
	for _, d := range m.dispatchers {
		if err := d.Send(ctx, event); err != nil {
			metrics.AlertsSentTotal.WithLabelValues(d.Name(), "failure").Inc()
			logging.Logger.Error().
				Err(err).
				Str("dispatcher", d.Name()).
				Str("target", event.TargetName).
				Msg("failed to send alert notification")
			errs = append(errs, err)
		} else {
			metrics.AlertsSentTotal.WithLabelValues(d.Name(), "success").Inc()
			logging.Logger.Info().
				Str("dispatcher", d.Name()).
				Str("target", event.TargetName).
				Bool("recovery", event.IsRecovery).
				Msg("alert notification sent")
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

