package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	httpPort := os.Getenv("MOCK_HTTP_PORT")
	if httpPort == "" {
		httpPort = ":8081"
	} else if httpPort[0] != ':' {
		httpPort = ":" + httpPort
	}

	tcpPort := os.Getenv("MOCK_TCP_PORT")
	if tcpPort == "" {
		tcpPort = ":9000"
	} else if tcpPort[0] != ':' {
		tcpPort = ":" + tcpPort
	}

	var flakyCounter int64

	mux := http.NewServeMux()

	// 1. Healthy endpoint: instant 200 OK
	mux.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","service":"mock-server"}`))
	})

	// 2. Flaky endpoint: fails 2 times out of 3 to test retry backoff
	mux.HandleFunc("/flaky", func(w http.ResponseWriter, r *http.Request) {
		cnt := atomic.AddInt64(&flakyCounter, 1)
		w.Header().Set("Content-Type", "application/json")
		if cnt%3 == 0 {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"status":"recovered","attempt":%d}`, cnt)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, `{"status":"error","attempt":%d}`, cnt)
		}
	})

	// 3. Slow endpoint: artificial 1.2s delay to test latency measurements
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"slow_ok","delay_ms":1200}`))
	})

	// 4. Down endpoint: persistent 503 outage
	mux.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"outage","error":"database connection pool exhausted"}`))
	})

	// Start TCP echo listener for TCP probes
	go startTCPEchoListener(tcpPort)

	log.Printf("Mock HTTP server listening on %s", httpPort)
	if err := http.ListenAndServe(httpPort, mux); err != nil {
		log.Fatalf("Mock HTTP server failed: %v", err)
	}
}

func startTCPEchoListener(port string) {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("Warning: failed to start TCP mock listener on %s: %v", port, err)
		return
	}
	defer func() { _ = ln.Close() }()
	log.Printf("Mock TCP listener listening on %s", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go func(c net.Conn) {
			defer func() { _ = c.Close() }()
			_ = c.SetDeadline(time.Now().Add(5 * time.Second))
			_, _ = io.Copy(c, c)
		}(conn)
	}
}
