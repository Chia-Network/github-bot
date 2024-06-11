package keybase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type WebhookMessage struct {
	Message string `json:"message"`
}

// SendKeybaseMessage sends a message to a specified Keybase channel
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received error response: %s", resp.Status)
	}

	log.Printf("Message successfully sent")
	return nil
}
