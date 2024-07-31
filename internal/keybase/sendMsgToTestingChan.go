package keybase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/chia-network/go-modules/pkg/slogs"
)

// Annotation represents the annotations in the alert
type TestingAnnotation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Alert represents a single alert
type TestingAlert struct {
	Status      string            `json:"status"`
	Annotations TestingAnnotation `json:"annotations"`
}

// TestingWebhookMessage represents the message to be sent to the Keybase webhook
type TestingWebhookMessage struct {
	Alerts []TestingAlert `json:"alerts"`
}

var clientTesting *http.Client

func init() {
	clientTesting = &http.Client{}
}

// NewMessageTesting creates and returns an instance of the WebhookMessage struct
func NewMessageTesting(status, title, description string) TestingWebhookMessage {
	alert := TestingAlert{
		Status: status,
		Annotations: TestingAnnotation{
			Title:       title,
			Description: description,
		},
	}
	return TestingWebhookMessage{
		Alerts: []TestingAlert{alert},
	}
}

// SendKeybaseTestingMsg sends a message to a specified Keybase channel
func (msg *TestingWebhookMessage) SendKeybaseTestingMsg() error {
	webhookURL := os.Getenv("TESTING_KEYBASE_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("TESTING_KEYBASE_WEBHOOK_URL environment variable is not set")
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		slogs.Logr.Error("Error converting message to JSON", "error", err)
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		slogs.Logr.Error("Error creating webhook HTTP request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := clientTesting.Do(req)
	if err != nil {
		slogs.Logr.Error("Error sending message", "error", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slogs.Logr.Error("Error closing keybase webhook response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		slogs.Logr.Error("Keybase webhook returned error", "status", resp.Status)
		return fmt.Errorf("received error response: %s", resp.Status)
	}

	slogs.Logr.Info("Message successfully sent")
	return nil
}
