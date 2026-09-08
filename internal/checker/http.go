package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"watchtower/internal/config"
)

type HTTPChecker struct {
	defaultClient  *http.Client
	insecureClient *http.Client
}

func NewHTTPChecker() *HTTPChecker {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	insecureTransport := transport.Clone()
	insecureTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 - intentional testing option

	return &HTTPChecker{
		defaultClient: &http.Client{
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
		insecureClient: &http.Client{
			Transport: insecureTransport,
		},
	}
}

func (h *HTTPChecker) Check(
	ctx context.Context,
	target config.Target,
) Result {
	address := target.Address
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "https://" + address
	}

	method := target.Method
	if method == "" {
		method = http.MethodGet
	}

	result := Result{
		TargetName:    target.Name,
		TargetType:    target.Type,
		TargetAddress: address,
		CheckedAt:     time.Now(),
		Status:        StatusDown,
	}

	req, err := http.NewRequestWithContext(ctx, method, address, nil)
	if err != nil {
		result.Error = fmt.Sprintf("invalid request: %v", err)
		return result
	}

	// Set headers
	req.Header.Set("User-Agent", "Watchtower/2.0 (+https://github.com/watchtower)")
	for k, v := range target.Headers {
		req.Header.Set(k, v)
	}

	client := h.defaultClient
	if target.InsecureSkipVerify {
		client = h.insecureClient
	}

	start := time.Now()
	resp, err := client.Do(req)
	result.Latency = time.Since(start)
	result.LatencyMs = float64(result.Latency.Microseconds()) / 1000.0

	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Always drain and close response body to reuse TCP connections and avoid file descriptor leaks
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
	}()

	result.StatusCode = resp.StatusCode

	// Inspect SSL/TLS certificate if HTTPS
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		daysRemaining := int(time.Until(cert.NotAfter).Hours() / 24)
		result.SSLExpiryDays = &daysRemaining
	}

	// Verify status code
	if target.ExpectedStatus > 0 {
		if resp.StatusCode == target.ExpectedStatus {
			result.Status = StatusUp
		} else {
			result.Status = StatusDown
			result.Error = fmt.Sprintf("expected status %d, got %d", target.ExpectedStatus, resp.StatusCode)
		}
		return result
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Status = StatusUp
		return result
	}

	result.Status = StatusDown
	result.Error = fmt.Sprintf("HTTP %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	return result
}