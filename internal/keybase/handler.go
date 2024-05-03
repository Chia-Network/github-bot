package keybase

import (
	"log"
	"os"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/managed-bots/base"
)

type Handler struct {
	*base.DebugOutput
	stats           *base.StatsRegistry
	kbc             *kbchat.API
	allowedChannels map[string]string
}

// Ensure Handler implements the base.Handler interface.
var _ base.Handler = (*Handler)(nil)

// NewHandler initializes and returns a new instance of Handler.
func NewHandler(stats *base.StatsRegistry, kbc *kbchat.API, debugConfig *base.ChatDebugOutputConfig) *Handler {
	// Load and check environment variables once during handler initialization.
	ipsChannel := os.Getenv("IPS_CHANNEL")
	securityChannel := os.Getenv("SECURITY_CHANNEL")

	if ipsChannel == "" || securityChannel == "" {
		log.Fatal("IPS_CHANNEL or SECURITY_CHANNEL environment variables are not set.")
	}

	// Mapping convID to respective email recipients
	allowedChannels := map[string]string{
		os.Getenv("IPS_CHANNEL"):            "ips@chia.net",
		os.Getenv("SECURITY_CHANNEL"):       "security@chia.net",
		os.Getenv("TEAMBOTTESTING_CHANNEL"): "security@chia.net",
	}

	return &Handler{
		DebugOutput:     base.NewDebugOutput("Handler", debugConfig),
		stats:           stats.SetPrefix("Handler"),
		kbc:             kbc,
		allowedChannels: allowedChannels, // Correct assignment of allowedChannels map
	}
}
