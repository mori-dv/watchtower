package state

import (
	"context"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/storage"
)

type EvaluatedState struct {
	Status              checker.Status
	ConsecutiveFailures int
}

type Evaluator struct {
	store *storage.RedisStore
}

func NewEvaluator(
	store *storage.RedisStore,
) *Evaluator {

	return &Evaluator{
		store: store,
	}
}

func (e *Evaluator) Evaluate(
	ctx context.Context,
	result checker.Result,
) (EvaluatedState, error) {

	failureKey := "failures:" + result.TargetName

	// SUCCESS
	if result.Status == checker.StatusUp {

		e.store.Client().Del(
			ctx,
			failureKey,
		)

		return EvaluatedState{
			Status:              checker.StatusUp,
			ConsecutiveFailures: 0,
		}, nil
	}

	// FAILURE
	count, err := e.store.Client().
		Incr(ctx, failureKey).
		Result()

	if err != nil {
		return EvaluatedState{}, err
	}

	e.store.Client().Expire(
		ctx,
		failureKey,
		time.Hour,
	)

	// DEGRADED
	if count < 3 {

		return EvaluatedState{
			Status:              checker.StatusDegraded,
			ConsecutiveFailures: int(count),
		}, nil
	}

	// DOWN
	return EvaluatedState{
		Status:              checker.StatusDown,
		ConsecutiveFailures: int(count),
	}, nil
}