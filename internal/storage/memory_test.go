package storage_test

import (
	"context"
	"testing"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/storage"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryStore()
	defer func() { _ = store.Close() }()

	// Failure count increments
	cnt, err := store.IncrFailures(ctx, "targetA")
	if err != nil || cnt != 1 {
		t.Fatalf("expected count 1, got %d (err: %v)", cnt, err)
	}
	cnt, _ = store.IncrFailures(ctx, "targetA")
	if cnt != 2 {
		t.Fatalf("expected count 2, got %d", cnt)
	}

	// Reset failures
	_ = store.ResetFailures(ctx, "targetA")
	cnt, _ = store.GetFailures(ctx, "targetA")
	if cnt != 0 {
		t.Fatalf("expected count 0 after reset, got %d", cnt)
	}

	// Last Status
	_ = store.SetLastStatus(ctx, "targetA", checker.StatusDown)
	status, err := store.GetLastStatus(ctx, "targetA")
	if err != nil || status != checker.StatusDown {
		t.Fatalf("expected StatusDown, got %s (err: %v)", status, err)
	}

	// Cooldown with expiration
	_ = store.SetCooldown(ctx, "down:targetA", 100*time.Millisecond)
	onCooldown, _ := store.IsOnCooldown(ctx, "down:targetA")
	if !onCooldown {
		t.Fatal("expected active cooldown")
	}

	time.Sleep(150 * time.Millisecond)
	onCooldownAfter, _ := store.IsOnCooldown(ctx, "down:targetA")
	if onCooldownAfter {
		t.Fatal("expected cooldown to have expired")
	}
}
