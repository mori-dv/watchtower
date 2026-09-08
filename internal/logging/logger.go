package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"watchtower/internal/checker"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

// Init configures the global Zerolog logger based on level and output format.
func Init(level string, format string) {
	var output io.Writer = os.Stdout
	if strings.ToLower(format) == "console" || strings.ToLower(format) == "pretty" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	zerologLevel := zerolog.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		zerologLevel = zerolog.DebugLevel
	case "info":
		zerologLevel = zerolog.InfoLevel
	case "warn", "warning":
		zerologLevel = zerolog.WarnLevel
	case "error":
		zerologLevel = zerolog.ErrorLevel
	}

	Logger = zerolog.New(output).
		Level(zerologLevel).
		With().
		Timestamp().
		Caller().
		Logger()
}

// LogProbeResult captures structured operational telemetry for each probe completion.
func LogProbeResult(result checker.Result, consecutiveFailures int) {
	event := Logger.Info()

	// Categorize severity
	if result.Status == checker.StatusDown {
		event = Logger.Error()
	} else if result.Status == checker.StatusDegraded || result.Latency > 2*time.Second {
		event = Logger.Warn()
	}

	event.
		Str("target", result.TargetName).
		Str("type", result.TargetType).
		Str("status", string(result.Status)).
		Float64("latency_ms", result.LatencyMs).
		Int("attempts", result.Attempts).
		Int("consecutive_failures", consecutiveFailures)

	if result.StatusCode > 0 {
		event.Int("status_code", result.StatusCode)
	}
	if result.SSLExpiryDays != nil {
		event.Int("ssl_expiry_days", *result.SSLExpiryDays)
	}
	if result.Error != "" {
		event.Str("error", result.Error)
	}

	if result.Latency > 2*time.Second {
		event.Msg("probe latency spike detected")
	} else {
		event.Msg("probe completed")
	}
}

