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

type WebhookDispatcher struct {
	url     string
	headers map[string]string
	client  *http.Client
}

func NewWebhookDispatcher(url string, headers map[string]string) *WebhookDispatcher {
	return &WebhookDispatcher{
		url:     url,
		headers: headers,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *WebhookDispatcher) Name() string {
	return "webhook"
}

func (w *WebhookDispatcher) Send(ctx context.Context, event AlertEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Watchtower-AlertManager/2.0")
	for k, v := range w.headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook responded with non-2xx status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}
