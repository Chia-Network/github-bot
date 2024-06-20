package keybase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// WebhookMessage represents the message to be sent to the Keybase webhook.
type WebhookMessage struct {
	Message string `json:"message"`
}

// SendKeybaseMsg sends a message to a specified Keybase channel
func SendKeybaseMsg(message string) error {
	webhookURL := os.Getenv("KEYBASE_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("KEYBASE_WEBHOOK_URL environment variable is not set")
	}

	payload := WebhookMessage{
		Message: message,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slogs.Logger.Error("Error converting data to a JSON string", "error", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		slogs.Logger.Error("Error sending message", "error", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slogs.Logger.Error("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received error response: %s", resp.Status)
	}

	slogs.Logger.Info("Message successfully sent")
	return nil
}
