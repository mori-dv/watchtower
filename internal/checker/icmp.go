package checker

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"watchtower/internal/config"
)

type ICMPChecker struct{}

func NewICMPChecker() *ICMPChecker {
	return &ICMPChecker{}
}

func (c *ICMPChecker) Check(
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

	host := target.Address
	// Strip port or protocol if mistakenly provided
	if strings.Contains(host, "://") {
		parts := strings.Split(host, "://")
		if len(parts) > 1 {
			host = parts[1]
		}
	}
	if strings.Contains(host, "/") {
		parts := strings.Split(host, "/")
		host = parts[0]
	}
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err == nil {
			host = h
		}
	}

	// First attempt: unprivileged UDP ICMP or raw ping
	rtt, err := pingSystem(ctx, host, target.Timeout)
	if err == nil {
		result.Status = StatusUp
		result.Latency = rtt
		result.LatencyMs = float64(rtt.Microseconds()) / 1000.0
		return result
	}

	result.Error = fmt.Sprintf("icmp ping failed: %v", err)
	return result
}

// pingSystem executes the system ping binary which is standard across Linux, macOS, and container environments.
// It handles platform-specific flags (e.g. -c 1, -W timeout vs -W ms on macOS).
func pingSystem(ctx context.Context, host string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()

	var cmd *exec.Cmd
	timeoutSec := int(timeout.Seconds())
	if timeoutSec < 1 {
		timeoutSec = 1
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: -c 1 count, -W timeout in milliseconds
		ms := strconv.Itoa(int(timeout.Milliseconds()))
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", ms, host)
	case "windows":
		// Windows: -n 1 count, -w timeout in milliseconds
		ms := strconv.Itoa(int(timeout.Milliseconds()))
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", ms, host)
	default:
		// Linux (iputils): -c 1 count, -W timeout in seconds
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", strconv.Itoa(timeoutSec), host)
	}

	out, err := cmd.CombinedOutput()
	latency := time.Since(start)

	if err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		return 0, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}

	// Try to extract exact RTT from ping output if available
	parsedRtt := parsePingOutput(string(out))
	if parsedRtt > 0 {
		return parsedRtt, nil
	}

	return latency, nil
}

// parsePingOutput extracts the round-trip time from standard ping output.
func parsePingOutput(output string) time.Duration {
	for _, line := range strings.Split(output, "\n") {
		// Linux/macOS: time=12.3 ms or round-trip min/avg/max/stddev = 10.1/12.3/...
		if strings.Contains(line, "time=") {
			idx := strings.Index(line, "time=")
			sub := line[idx+5:]
			fields := strings.Fields(sub)
			if len(fields) > 0 {
				val := strings.TrimSuffix(fields[0], "ms")
				if ms, err := strconv.ParseFloat(val, 64); err == nil {
					return time.Duration(ms * float64(time.Millisecond))
				}
			}
		}
		if strings.Contains(line, "min/avg/max") || strings.Contains(line, "rtt min/avg/max") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				stats := strings.Split(strings.TrimSpace(parts[1]), "/")
				if len(stats) >= 2 {
					if avgMs, err := strconv.ParseFloat(stats[1], 64); err == nil {
						return time.Duration(avgMs * float64(time.Millisecond))
					}
				}
			}
		}
	}
	return 0
}
