package checker

import (
	"context"
	"net"
	"time"
	"log"

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

	start := time.Now()

	result := Result{
		TargetName: target.Name,
		TargetType: target.Type,
		CheckedAt:  time.Now(),
		Status:     StatusDown,
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(
		ctx,
		"tcp",
		target.Address,
	)

	result.Latency = time.Since(start)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("connection was not closed and failed")
		}
	}()
	result.Status = StatusUp

	return result
}