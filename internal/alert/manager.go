package alert

import (
	"context"
	"fmt"
	"time"

	"watchtower/internal/checker"
	"watchtower/internal/storage"
)

type Manager struct {
	store    *storage.RedisStore
	botToken string
	chatID   string
	cooldown time.Duration
}

func NewManager(
	store *storage.RedisStore,
	botToken string,
	chatID string,
	cooldown time.Duration,
) *Manager {

	return &Manager{
		store: store,

		botToken: botToken,
		chatID:   chatID,

		cooldown: cooldown,
	}
}

func (m *Manager) Handle(
	ctx context.Context,
	result checker.Result,
) error {
	if result.Status != checker.StatusDown {
		return nil
	}

	cooldownKey := "cooldown:" + result.TargetName
	exists, err := m.store.Client().
		Exists(ctx, cooldownKey).
		Result()

	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	message := fmt.Sprintf(
		"🚨 %s is DOWN",
		result.TargetName,
	)
	if err := SendTelegramAlert(
		m.botToken,
		m.chatID,
		message,
	); err != nil {
		return err
	}
	return m.store.Client().
		Set(
			ctx,
			cooldownKey,
			"1",
			m.cooldown,
		).
		Err()
}
