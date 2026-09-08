package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"watchtower/internal/api"
	"watchtower/internal/checker"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	api.HealthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestReadyHandler_SuccessAndFailure(t *testing.T) {
	// Ready success
	readyHandler := api.ReadyHandler(func(ctx context.Context) error {
		return nil
	})
	rec := httptest.NewRecorder()
	readyHandler(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for ready, got %d", rec.Code)
	}

	// Ready failure
	notReadyHandler := api.ReadyHandler(func(ctx context.Context) error {
		return errors.New("db down")
	})
	recErr := httptest.NewRecorder()
	notReadyHandler(recErr, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if recErr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for unready, got %d", recErr.Code)
	}
}

func TestStatusRegistry(t *testing.T) {
	reg := api.NewStatusRegistry()

	days := 45
	reg.Update(checker.Result{
		TargetName:    "prod-web",
		TargetType:    "http",
		TargetAddress: "https://web.internal",
		Status:        checker.StatusUp,
		StatusCode:    200,
		LatencyMs:     24.5,
		SSLExpiryDays: &days,
		CheckedAt:     time.Now(),
	}, 0)

	rec := httptest.NewRecorder()
	reg.Handler()(rec, httptest.NewRequest(http.MethodGet, "/api/v1/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from status handler, got %d", rec.Code)
	}
}

func TestReloadHandler(t *testing.T) {
	called := false
	reloadFn := func() error {
		called = true
		return nil
	}

	h := api.ReloadHandler(reloadFn)

	// GET method should be rejected
	recGet := httptest.NewRecorder()
	h(recGet, httptest.NewRequest(http.MethodGet, "/reload", nil))
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET /reload, got %d", recGet.Code)
	}

	// POST method should succeed
	recPost := httptest.NewRecorder()
	h(recPost, httptest.NewRequest(http.MethodPost, "/reload", nil))
	if recPost.Code != http.StatusOK {
		t.Errorf("expected 200 for POST /reload, got %d", recPost.Code)
	}
	if !called {
		t.Errorf("expected reloadFn to be invoked")
	}
}
