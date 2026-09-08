package alert_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"watchtower/internal/alert"
	"watchtower/internal/checker"
	"watchtower/internal/storage"
)

type mockDispatcher struct {
	mu     sync.Mutex
	events []alert.AlertEvent
}

func (m *mockDispatcher) Name() string {
	return "mock"
}

func (m *mockDispatcher) Send(_ context.Context, event alert.AlertEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockDispatcher) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func TestAlertManager_OutageDebounceAndRecovery(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	defer func() { _ = store.Close() }()

	mock := &mockDispatcher{}
	mgr := alert.NewManager(store, 1*time.Second, mock)

	result := checker.Result{
		TargetName:    "prod-db",
		TargetType:    "tcp",
		TargetAddress: "db.internal:5432",
		Status:        checker.StatusDown,
		Latency:       50 * time.Millisecond,
		Error:         "connection timeout",
	}

	// 1. Initial Outage alert
	err := mgr.Handle(ctx, result, checker.StatusDown, checker.StatusDegraded, 3, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.count() != 1 {
		t.Fatalf("expected 1 outage alert, got %d", mock.count())
	}
	if mock.events[0].IsRecovery {
		t.Errorf("expected outage alert, got recovery")
	}

	// 2. Immediate subsequent failure -> should be debounced by cooldown
	err = mgr.Handle(ctx, result, checker.StatusDown, checker.StatusDown, 4, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.count() != 1 {
		t.Fatalf("expected alert to be debounced, count remained %d", mock.count())
	}

	// 3. Recovery alert
	upResult := checker.Result{
		TargetName:    "prod-db",
		TargetType:    "tcp",
		TargetAddress: "db.internal:5432",
		Status:        checker.StatusUp,
		Latency:       10 * time.Millisecond,
	}
	err = mgr.Handle(ctx, upResult, checker.StatusUp, checker.StatusDown, 0, true)
	if err != nil {
		t.Fatalf("unexpected error on recovery: %v", err)
	}
	if mock.count() != 2 {
		t.Fatalf("expected 2 total alerts (1 outage + 1 recovery), got %d", mock.count())
	}
	if !mock.events[1].IsRecovery {
		t.Errorf("second alert should be recovery alert")
	}
}
