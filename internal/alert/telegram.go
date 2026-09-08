package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TelegramDispatcher struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegramDispatcher(botToken, chatID string) *TelegramDispatcher {
	return &TelegramDispatcher{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (t *TelegramDispatcher) Name() string {
	return "telegram"
}

type TelegramPayload struct {
	ChatId    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func (t *TelegramDispatcher) Send(ctx context.Context, event AlertEvent) error {
	var message string
	if event.IsRecovery {
		message = fmt.Sprintf("✅ *[RECOVERED]* Target `%s` is back *UP*!\n*Type:* `%s`\n*Latency:* `%s`\n*Time:* `%s`",
			event.TargetName, event.TargetType, event.Latency, event.Timestamp.Format(time.RFC3339))
	} else if event.Status == "DOWN" {
		message = fmt.Sprintf("🚨 *[OUTAGE ALERT]* Target `%s` is *DOWN*!\n*Type:* `%s`\n*Consecutive Failures:* `%d`\n*Error:* `%s`\n*Time:* `%s`",
			event.TargetName, event.TargetType, event.ConsecutiveFailures, event.Error, event.Timestamp.Format(time.RFC3339))
	} else {
		message = fmt.Sprintf("⚠️ *[DEGRADED]* Target `%s` is failing checks!\n*Type:* `%s`\n*Consecutive Failures:* `%d`\n*Error:* `%s`",
			event.TargetName, event.TargetType, event.ConsecutiveFailures, event.Error)
	}

	return t.sendRaw(ctx, message)
}

func (t *TelegramDispatcher) sendRaw(ctx context.Context, message string) error {
	payload := TelegramPayload{
		ChatId:    t.chatID,
		Text:      message,
		ParseMode: "Markdown",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("telegram api error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SendTelegramAlert provides backwards compatibility for direct invocations.
func SendTelegramAlert(botToken, chatId, message string) error {
	d := NewTelegramDispatcher(botToken, chatId)
	return d.sendRaw(context.Background(), message)
}