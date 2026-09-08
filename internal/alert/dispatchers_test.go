package alert_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"watchtower/internal/alert"
	"watchtower/internal/checker"
)

func TestWebhookDispatcher(t *testing.T) {
	received := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		if r.Header.Get("X-Custom") != "test-val" {
			t.Errorf("missing expected custom header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := alert.NewWebhookDispatcher(server.URL, map[string]string{"X-Custom": "test-val"})
	if d.Name() != "webhook" {
		t.Errorf("expected name 'webhook', got %s", d.Name())
	}

	err := d.Send(context.Background(), alert.AlertEvent{
		TargetName: "web",
		Status:     checker.StatusDown,
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected webhook send error: %v", err)
	}
	if !received {
		t.Fatal("webhook was not received by server")
	}
}

func TestSlackDispatcher(t *testing.T) {
	received := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := alert.NewSlackDispatcher(server.URL, "#monitoring")
	if d.Name() != "slack" {
		t.Errorf("expected name 'slack', got %s", d.Name())
	}

	// Outage
	err := d.Send(context.Background(), alert.AlertEvent{
		TargetName: "api",
		Status:     checker.StatusDown,
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected slack send error: %v", err)
	}

	// Recovery
	err = d.Send(context.Background(), alert.AlertEvent{
		TargetName: "api",
		Status:     checker.StatusUp,
		IsRecovery: true,
		Timestamp:  time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected slack recovery send error: %v", err)
	}
	if !received {
		t.Fatal("slack webhook was not received by server")
	}
}
