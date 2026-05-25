package checker

import (
	"context"

	"watchtower/internal/config"
)

type Checker interface {
	Check(ctx context.Context, target config.Target) Result
}