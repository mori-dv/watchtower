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

type SlackDispatcher struct {
	webhookURL string
	channel    string
	client     *http.Client
}

func NewSlackDispatcher(webhookURL, channel string) *SlackDispatcher {
	return &SlackDispatcher{
		webhookURL: webhookURL,
		channel:    channel,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SlackDispatcher) Name() string {
	return "slack"
}

type slackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackAttachment struct {
	Color     string                 `json:"color"`
	Title     string                 `json:"title"`
	Text      string                 `json:"text,omitempty"`
	Fields    []slackAttachmentField `json:"fields"`
	Timestamp int64                  `json:"ts"`
}

type slackPayload struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username"`
	IconEmoji   string            `json:"icon_emoji"`
	Attachments []slackAttachment `json:"attachments"`
}

func (s *SlackDispatcher) Send(ctx context.Context, event AlertEvent) error {
	var color, title, emoji string
	if event.IsRecovery {
		color = "#2eb886" // green
		title = fmt.Sprintf("✅ [RECOVERED] %s is UP", event.TargetName)
		emoji = ":white_check_mark:"
	} else if event.Status == "DOWN" {
		color = "#a30200" // red
		title = fmt.Sprintf("🚨 [OUTAGE] %s is DOWN", event.TargetName)
		emoji = ":rotating_light:"
	} else {
		color = "#daa038" // yellow
		title = fmt.Sprintf("⚠️ [DEGRADED] %s is experiencing issues", event.TargetName)
		emoji = ":warning:"
	}

	fields := []slackAttachmentField{
		{Title: "Target", Value: event.TargetName, Short: true},
		{Title: "Type", Value: event.TargetType, Short: true},
		{Title: "Latency", Value: event.Latency.String(), Short: true},
		{Title: "Consecutive Failures", Value: fmt.Sprintf("%d", event.ConsecutiveFailures), Short: true},
	}

	if event.Error != "" {
		fields = append(fields, slackAttachmentField{
			Title: "Error",
			Value: event.Error,
			Short: false,
		})
	}

	payload := slackPayload{
		Channel:   s.channel,
		Username:  "Watchtower",
		IconEmoji: emoji,
		Attachments: []slackAttachment{
			{
				Color:     color,
				Title:     title,
				Fields:    fields,
				Timestamp: event.Timestamp.Unix(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack webhook post failed: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("slack returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
