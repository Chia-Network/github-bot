package keybase

import (
	"fmt"
	"log"
	"os"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
	"github.com/keybase/managed-bots/base"
	"golang.org/x/sync/errgroup"
)

// Options holds configuration for the bot server.
type Options struct {
	*base.Options
	Channels string
}

// NewOptions creates and returns a new Options instance.
func NewOptions() *Options {
	return &Options{
		Options: base.NewOptions(),
	}
}

// BotServer encapsulates the server and its options.
type BotServer struct {
	*base.Server
	opts Options
	kbc  *kbchat.API
}

// NewBotServer initializes a new bot server with the provided options.
func NewBotServer(opts Options) *BotServer {
	username, isUserSet := os.LookupEnv("KEYBASE_USERNAME")
	paperkey, isPaperSet := os.LookupEnv("KEYBASE_PAPERKEY")

	if !isUserSet || !isPaperSet {
		log.Printf("Check the KEYBASE_USERNAME or KEYBASE_PAPERKEY environment variables. One or both are unset.")
		return nil
	}

	runOptions := kbchat.RunOptions{
		KeybaseLocation: opts.KeybaseLocation,
		HomeDir:         opts.Home,
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
		Server: base.NewServer("ChiaGithubBot", opts.Announcement, opts.AWSOpts, opts.MultiDSN, opts.ReadSelf, runOptions),
		kbc:    kbc,
		opts:   opts,
	}
}

// makeAdvertisement creates advertisement commands for the bot.
func (s *BotServer) makeAdvertisement() kbchat.Advertisement {
	cmds := []chat1.UserBotCommandInput{
		{
			Name:        "chiagithubbot",
			Description: "Send Github Alerts to Keybase Channels",
		},
		base.GetFeedbackCommandAdvertisement(s.kbc.GetUsername()),
	}
	return kbchat.Advertisement{
		Alias: "ChiaGithubBot",
		Advertisements: []chat1.AdvertiseCommandAPIParam{
			{
				Typ:      "public",
				Commands: cmds,
			},
		},
	}
}

// Go starts the bot server and its components.
func (s *BotServer) Go() (err error) {
	if s.kbc, err = s.Start(s.opts.ErrReportConv); err != nil {
		return err
	}

	debugConfig := base.NewChatDebugOutputConfig(s.kbc, s.opts.ErrReportConv)
	stats, err := base.NewStatsRegistry(debugConfig, s.opts.StathatEZKey)
	if err != nil {
		s.Debug("unable to create stats: %v", err)
		return err
	}
	stats = stats.SetPrefix(s.Name())
	httpSrv := NewHTTPSrv(stats, debugConfig, s.kbc) // Ensure this function is defined
	handler := NewHandler(stats, s.kbc, debugConfig) // Ensure this function is correctly implemented
	eg := &errgroup.Group{}
	s.GoWithRecover(eg, func() error { return s.Listen(handler) })
	s.GoWithRecover(eg, func() error { return httpSrv.Listen() }) // Adapt based on the actual HTTP server implementation
	s.GoWithRecover(eg, func() error { return s.HandleSignals(httpSrv, stats) })
	s.GoWithRecover(eg, func() error { return s.AnnounceAndAdvertise(s.makeAdvertisement(), "Help message here") })
	if err := eg.Wait(); err != nil {
		s.Debug("wait error: %s", err)
		return err
	}
	return nil
}
