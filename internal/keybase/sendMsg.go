package keybase

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/chia-network/go-modules/pkg/slogs"
)

// SendKeybaseMsg sends a message to a specified Keybase channel
func SendKeybaseMsg(message string) error {
	webhookURL := os.Getenv("KEYBASE_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("KEYBASE_WEBHOOK_URL environment variable is not set")
	}

	resp, err := http.Post(webhookURL, "text/plain", bytes.NewBufferString(message))
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
