package keybase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/chia-network/go-modules/pkg/slogs"
)

// WebhookMessage represents the message to be sent to the Keybase webhook.
type WebhookMessage struct {
	Status      string `json:"status"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// NewMessage creates and returns an instance of the WebhookMessage struct
func NewMessage(status string, title string, description string) WebhookMessage {
	return WebhookMessage{
		Status:      status,
		Title:       title,
		Description: description,
	}
}

// SendKeybaseMsg sends a message to a specified Keybase channel
func (msg *WebhookMessage) SendKeybaseMsg() error {
	webhookURL := os.Getenv("KEYBASE_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("KEYBASE_WEBHOOK_URL environment variable is not set")
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slogs.Logr.Error("Error converting string to json", "error", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		slogs.Logr.Error("Error sending message", "error", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slogs.Logr.Error("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received error response: %s", resp.Status)
	}

	slogs.Logr.Info("Message successfully sent")
	return nil
}
