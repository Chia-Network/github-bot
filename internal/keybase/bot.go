package keybase

import (
	"fmt"
	"log"
	"os"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

// Options holds configuration for the bot server.
type Options struct {
	KeybaseLocation string
	HomeDir         string
}

// NewOptions creates and returns a new Options instance.
func NewOptions(keybaseLocation, homeDir string) *Options {
	return &Options{
		KeybaseLocation: keybaseLocation,
		HomeDir:         homeDir,
	}
}

// BotServer encapsulates the server and its options.
type BotServer struct {
	kbc  *kbchat.API
	opts *Options
}

// NewBotServer initializes a new bot server with the provided options.
func NewBotServer(opts *Options) *BotServer {
	username, isUserSet := os.LookupEnv("KEYBASE_USERNAME")
	paperkey, isPaperSet := os.LookupEnv("KEYBASE_PAPERKEY")

	if !isUserSet || !isPaperSet {
		log.Printf("Check the KEYBASE_USERNAME or KEYBASE_PAPERKEY environment variables. One or both are unset.")
		return nil
	}

	runOptions := kbchat.RunOptions{
		KeybaseLocation: opts.KeybaseLocation,
		HomeDir:         opts.HomeDir,
		Oneshot: &kbchat.OneshotOptions{
			Username: username,
			PaperKey: paperkey,
		},
	}

	kbc, err := kbchat.Start(runOptions)
	if err != nil {
		fmt.Printf("Error starting Keybase chat: %v\n", err)
		os.Exit(1)
	}

	return &BotServer{
		kbc:  kbc,
		opts: opts,
	}
}

// Go starts the bot server and its components.
func (s *BotServer) Go() error {
	//
	return nil
}
