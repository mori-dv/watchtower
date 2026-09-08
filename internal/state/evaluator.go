package state

import (
	"context"

	"watchtower/internal/checker"
	"watchtower/internal/storage"
)

// EvaluatedState holds the refined health status and transition flags after evaluating a check.
type EvaluatedState struct {
	Status              checker.Status `json:"status"`
	PreviousStatus      checker.Status `json:"previous_status"`
	ConsecutiveFailures int            `json:"consecutive_failures"`
	Transitioned        bool           `json:"transitioned"`
	IsRecovery          bool           `json:"is_recovery"`
	IsOutage            bool           `json:"is_outage"`
}

type Evaluator struct {
	store storage.Store
}

func NewEvaluator(store storage.Store) *Evaluator {
	return &Evaluator{
		store: store,
	}
}

// Evaluate processes a probe result against consecutive failure thresholds and previous status.
func (e *Evaluator) Evaluate(
	ctx context.Context,
	result checker.Result,
) (EvaluatedState, error) {
	prevStatus, _ := e.store.GetLastStatus(ctx, result.TargetName)

	// SUCCESS
	if result.Status == checker.StatusUp {
		_ = e.store.ResetFailures(ctx, result.TargetName)
		_ = e.store.SetLastStatus(ctx, result.TargetName, checker.StatusUp)

		isRecovery := prevStatus == checker.StatusDown || prevStatus == checker.StatusDegraded
		transitioned := prevStatus != "" && prevStatus != checker.StatusUp

		return EvaluatedState{
			Status:              checker.StatusUp,
			PreviousStatus:      prevStatus,
			ConsecutiveFailures: 0,
			Transitioned:        transitioned,
			IsRecovery:          isRecovery,
			IsOutage:            false,
		}, nil
	}

	// FAILURE
	count, err := e.store.IncrFailures(ctx, result.TargetName)
	if err != nil {
		// Fall back gracefully with count 1
		count = 1
	}

	var newStatus checker.Status
	// 1-2 consecutive failures = DEGRADED (prevents flapping from momentary network jitter)
	// 3+ consecutive failures = DOWN (confirmed outage)
	if count < 3 {
		newStatus = checker.StatusDegraded
	} else {
		newStatus = checker.StatusDown
	}

	isOutage := newStatus == checker.StatusDown && prevStatus != checker.StatusDown
	transitioned := prevStatus != newStatus
	_ = e.store.SetLastStatus(ctx, result.TargetName, newStatus)

	return EvaluatedState{
		Status:              newStatus,
		PreviousStatus:      prevStatus,
		ConsecutiveFailures: int(count),
		Transitioned:        transitioned,
		IsRecovery:          false,
		IsOutage:            isOutage,
	}, nil
}