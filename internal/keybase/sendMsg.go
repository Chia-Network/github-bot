package keybase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("Error converting data to a JSON string: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received error response: %s", resp.Status)
	}

	log.Printf("Message successfully sent")
	return nil
}