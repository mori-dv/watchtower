package checker_test

import (
	"context"
	"net"
	"testing"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/config"
)

func TestTCPChecker_Success(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	// Accept in background
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	c := checker.NewTCPChecker()
	target := config.Target{
		Name:    "test-tcp",
		Type:    "tcp",
		Address: ln.Addr().String(),
		Timeout: 2 * time.Second,
	}

	res := c.Check(context.Background(), target)
	if res.Status != checker.StatusUp {
		t.Fatalf("expected StatusUp, got %s (error: %s)", res.Status, res.Error)
	}
	if res.Latency <= 0 {
		t.Errorf("expected positive latency, got %v", res.Latency)
	}
}

func TestTCPChecker_InvalidAddress(t *testing.T) {
	c := checker.NewTCPChecker()
	target := config.Target{
		Name:    "test-tcp-invalid",
		Type:    "tcp",
		Address: "127.0.0.1", // Missing port
		Timeout: 1 * time.Second,
	}

	res := c.Check(context.Background(), target)
	if res.Status != checker.StatusDown {
		t.Errorf("expected StatusDown for address missing port, got %s", res.Status)
	}
}

func TestTCPChecker_ConnectionRefused(t *testing.T) {
	c := checker.NewTCPChecker()
	target := config.Target{
		Name:    "test-tcp-refused",
		Type:    "tcp",
		Address: "127.0.0.1:54322",
		Timeout: 500 * time.Millisecond,
	}

	res := c.Check(context.Background(), target)
	if res.Status != checker.StatusDown {
		t.Errorf("expected StatusDown for closed port, got %s", res.Status)
	}
}
