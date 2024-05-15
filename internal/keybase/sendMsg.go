package keybase

import (
	"log"
	"os"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

// SendKeybaseMessage sends a message to a specified Keybase channel
func SendKeybaseMsg(message string) error {
	var channel = chat1.ConvIDStr(os.Getenv("GITHUBBOT_CHANNEL"))
	options := kbchat.RunOptions{
		Oneshot: &kbchat.OneshotOptions{
			Username: os.Getenv("KEYBASE_USERNAME"),
			PaperKey: os.Getenv("KEYBASE_PAPERKEY"),
		},
	}

	api, err := kbchat.Start(options)
	if err != nil {
		log.Printf("Error starting Keybase chat: %v", err)
		return err
	}

	_, err = api.SendMessageByConvID(channel, message)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
		return err
	}

	log.Printf("Message successfully sent")
	return nil
}
