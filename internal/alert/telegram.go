package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TelegramPayload struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
}

func SendTelegramAlert(botToken, chatId, message string) error {
	payload := TelegramPayload{
		ChatId: chatId,
		Text:   message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}