package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"watchtower/internal/config"
)

type TCPChecker struct{}

func NewTCPChecker() *TCPChecker {
	return &TCPChecker{}
}

func (t *TCPChecker) Check(
	ctx context.Context,
	target config.Target,
) Result {
	result := Result{
		TargetName:    target.Name,
		TargetType:    target.Type,
		TargetAddress: target.Address,
		CheckedAt:     time.Now(),
		Status:        StatusDown,
	}

	address := target.Address
	// Validate host:port format
	if !strings.Contains(address, ":") {
		result.Error = fmt.Sprintf("invalid TCP address %q: missing port (e.g. host:port)", address)
		return result
	}

	start := time.Now()
	dialer := net.Dialer{}

	// If SSL check is enabled or port is 443 / 8443, perform a TLS dial
	if (target.CheckSSL != nil && *target.CheckSSL) || strings.HasSuffix(address, ":443") || strings.HasSuffix(address, ":8443") {
		tlsDialer := tls.Dialer{
			NetDialer: &dialer,
			Config: &tls.Config{
				InsecureSkipVerify: target.InsecureSkipVerify, // #nosec G402
			},
		}

		conn, err := tlsDialer.DialContext(ctx, "tcp", address)
		result.Latency = time.Since(start)
		result.LatencyMs = float64(result.Latency.Microseconds()) / 1000.0

		if err != nil {
			result.Error = err.Error()
			return result
		}
		defer func() { _ = conn.Close() }()

		if tlsConn, ok := conn.(*tls.Conn); ok {
			state := tlsConn.ConnectionState()
			if len(state.PeerCertificates) > 0 {
				cert := state.PeerCertificates[0]
				daysRemaining := int(time.Until(cert.NotAfter).Hours() / 24)
				result.SSLExpiryDays = &daysRemaining
			}
		}

		result.Status = StatusUp
		return result
	}

	// Plain TCP dial
	conn, err := dialer.DialContext(ctx, "tcp", address)
	result.Latency = time.Since(start)
	result.LatencyMs = float64(result.Latency.Microseconds()) / 1000.0

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer func() { _ = conn.Close() }()

	result.Status = StatusUp
	return result
}