package checker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/config"
)

func TestHTTPChecker_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	c := checker.NewHTTPChecker()
	target := config.Target{
		Name:     "test-http",
		Type:     "http",
		Address:  server.URL,
		Timeout:  2 * time.Second,
		Interval: 5 * time.Second,
	}

	res := c.Check(context.Background(), target)

	if res.Status != checker.StatusUp {
		t.Fatalf("expected StatusUp, got %s (error: %s)", res.Status, res.Error)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200, got %d", res.StatusCode)
	}
	if res.Latency <= 0 {
		t.Errorf("expected positive latency, got %v", res.Latency)
	}
}

func TestHTTPChecker_ExpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	c := checker.NewHTTPChecker()

	// Matching expected status
	targetMatch := config.Target{
		Name:           "teapot",
		Type:           "http",
		Address:        server.URL,
		ExpectedStatus: http.StatusTeapot,
		Timeout:        2 * time.Second,
	}
	resMatch := c.Check(context.Background(), targetMatch)
	if resMatch.Status != checker.StatusUp {
		t.Errorf("expected StatusUp for matching expected status, got %s", resMatch.Status)
	}

	// Mismatched expected status
	targetMismatch := config.Target{
		Name:           "teapot-mismatch",
		Type:           "http",
		Address:        server.URL,
		ExpectedStatus: http.StatusOK,
		Timeout:        2 * time.Second,
	}
	resMismatch := c.Check(context.Background(), targetMismatch)
	if resMismatch.Status != checker.StatusDown {
		t.Errorf("expected StatusDown for mismatched expected status, got %s", resMismatch.Status)
	}
}

func TestHTTPChecker_TLSCertificate(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer tlsServer.Close()

	c := checker.NewHTTPChecker()
	target := config.Target{
		Name:               "test-tls",
		Type:               "http",
		Address:            tlsServer.URL,
		Timeout:            2 * time.Second,
		InsecureSkipVerify: true,
	}

	res := c.Check(context.Background(), target)
	if res.Status != checker.StatusUp {
		t.Fatalf("expected StatusUp, got %s (error: %s)", res.Status, res.Error)
	}
	if res.SSLExpiryDays == nil {
		t.Fatalf("expected SSLExpiryDays to be populated for HTTPS, got nil")
	}
	if *res.SSLExpiryDays <= 0 {
		t.Errorf("expected positive SSL expiry days, got %d", *res.SSLExpiryDays)
	}
}

func TestHTTPChecker_ServerDown(t *testing.T) {
	c := checker.NewHTTPChecker()
	target := config.Target{
		Name:    "unreachable",
		Type:    "http",
		Address: "http://127.0.0.1:54321", // unused port
		Timeout: 500 * time.Millisecond,
	}

	res := c.Check(context.Background(), target)
	if res.Status != checker.StatusDown {
		t.Errorf("expected StatusDown, got %s", res.Status)
	}
	if res.Error == "" {
		t.Errorf("expected error string to be populated")
	}
}
