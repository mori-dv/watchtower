package checker

import (
	"context"
	"fmt"
	"time"

	"watchtower/internal/config"
)

type Checker interface {
	Check(ctx context.Context, target config.Target) Result
}

// Engine dispatches targets to their respective protocol checkers.
type Engine struct {
	httpChecker *HTTPChecker
	tcpChecker  *TCPChecker
	icmpChecker *ICMPChecker
}

func NewEngine() *Engine {
	return &Engine{
		httpChecker: NewHTTPChecker(),
		tcpChecker:  NewTCPChecker(),
		icmpChecker: NewICMPChecker(),
	}
}

func (e *Engine) Check(ctx context.Context, target config.Target) Result {
	switch target.Type {
	case "http":
		return e.httpChecker.Check(ctx, target)
	case "tcp":
		return e.tcpChecker.Check(ctx, target)
	case "icmp":
		return e.icmpChecker.Check(ctx, target)
	default:
		return Result{
			TargetName:    target.Name,
			TargetType:    target.Type,
			TargetAddress: target.Address,
			Status:        StatusDown,
			Error:         fmt.Sprintf("unknown checker type: %q", target.Type),
			CheckedAt:     time.Now(),
		}
	}
}