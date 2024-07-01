package keybase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/chia-network/go-modules/pkg/slogs"
)

// Annotation represents the annotations in the alert
type Annotation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Alert represents a single alert
type Alert struct {
	Status      string     `json:"status"`
	Annotations Annotation `json:"annotations"`
}

// WebhookMessage represents the message to be sent to the Keybase webhook
type WebhookMessage struct {
	Alerts []Alert `json:"alerts"`
}

// NewMessage creates and returns an instance of the WebhookMessage struct
func NewMessage(status string, title string, description string) WebhookMessage {
	alert := Alert{
		Status: status,
		Annotations: Annotation{
			Title:       title,
			Description: description,
		},
	}
	return WebhookMessage{
		Alerts: []Alert{alert},
	}
}

// SendKeybaseMsg sends a message to a specified Keybase channel
func (msg *WebhookMessage) SendKeybaseMsg() error {
	webhookURL := os.Getenv("KEYBASE_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("KEYBASE_WEBHOOK_URL environment variable is not set")
	}

	authToken := os.Getenv("WEBHOOK_AUTH_SECRET_TOKEN")
	if authToken == "" {
		return fmt.Errorf("AUTH_TOKEN environment variable is not set")
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slogs.Logr.Error("Error converting string to json", "error", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		slogs.Logr.Error("Error creating HTTP request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := client.Do(req)
	if err != nil {
		slogs.Logr.Error("Error sending message", "error", err)
		return err
	}

	if resp.StatusCode == http.StatusOK {
		slogs.Logr.Info("Message successfully sent")
		return nil
	} else {

	}

	slogs.Logr.Info("Message successfully sent")
	return nil
}
