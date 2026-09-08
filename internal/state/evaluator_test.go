package state_test

import (
	"context"
	"testing"

	"watchtower/internal/checker"
	"watchtower/internal/state"
	"watchtower/internal/storage"
)

func TestEvaluator_FailureEscalationAndRecovery(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	defer func() { _ = store.Close() }()

	eval := state.NewEvaluator(store)

	failResult := checker.Result{
		TargetName: "api-service",
		TargetType: "http",
		Status:     checker.StatusDown,
	}
	upResult := checker.Result{
		TargetName: "api-service",
		TargetType: "http",
		Status:     checker.StatusUp,
	}

	// 1st failure -> DEGRADED (prevents jitter false alarms)
	s1, err := eval.Evaluate(ctx, failResult)
	if err != nil || s1.Status != checker.StatusDegraded || s1.ConsecutiveFailures != 1 {
		t.Fatalf("attempt 1: expected DEGRADED with 1 failure, got %s (%d)", s1.Status, s1.ConsecutiveFailures)
	}
	if s1.IsOutage {
		t.Errorf("attempt 1 should not flag full outage")
	}

	// 2nd failure -> still DEGRADED
	s2, err := eval.Evaluate(ctx, failResult)
	if err != nil || s2.Status != checker.StatusDegraded || s2.ConsecutiveFailures != 2 {
		t.Fatalf("attempt 2: expected DEGRADED with 2 failures, got %s (%d)", s2.Status, s2.ConsecutiveFailures)
	}

	// 3rd failure -> confirmed DOWN outage
	s3, err := eval.Evaluate(ctx, failResult)
	if err != nil || s3.Status != checker.StatusDown || s3.ConsecutiveFailures != 3 {
		t.Fatalf("attempt 3: expected DOWN with 3 failures, got %s (%d)", s3.Status, s3.ConsecutiveFailures)
	}
	if !s3.IsOutage {
		t.Errorf("attempt 3 should trigger IsOutage=true")
	}

	// 4th failure -> still DOWN, but not a new outage transition
	s4, _ := eval.Evaluate(ctx, failResult)
	if s4.Status != checker.StatusDown || s4.IsOutage {
		t.Errorf("attempt 4: expected ongoing DOWN without new outage flag")
	}

	// Recovery -> UP, IsRecovery=true
	sUp, err := eval.Evaluate(ctx, upResult)
	if err != nil || sUp.Status != checker.StatusUp {
		t.Fatalf("recovery: expected StatusUp, got %s", sUp.Status)
	}
	if !sUp.IsRecovery {
		t.Errorf("recovery: expected IsRecovery=true")
	}
	if sUp.ConsecutiveFailures != 0 {
		t.Errorf("recovery: expected 0 consecutive failures, got %d", sUp.ConsecutiveFailures)
	}
}
