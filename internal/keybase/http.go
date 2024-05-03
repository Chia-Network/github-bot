package keybase

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/managed-bots/base"
	"golang.org/x/oauth2"
)

type HTTPSrv struct {
	*base.OAuthHTTPSrv

	kbc     *kbchat.API
	db      *DB
	handler *Handler
	atr     *ghinstallation.AppsTransport
	secret  string
}

func NewHTTPSrv(stats *base.StatsRegistry, kbc *kbchat.API, debugConfig *base.ChatDebugOutputConfig, db *DB, handler *Handler,
	oauthConfig *oauth2.Config, atr *ghinstallation.AppsTransport, secret string) *HTTPSrv {
	h := &HTTPSrv{
		kbc:     kbc,
		db:      db,
		handler: handler,
		atr:     atr,
		secret:  secret,
	}
	h.OAuthHTTPSrv = base.NewOAuthHTTPSrv(stats, kbc, debugConfig, oauthConfig, h.db, h.handler.HandleAuth,
		"githubbot", base.Images["logo"], "/githubbot")
	http.HandleFunc("/githubbot", h.handleHealthCheck)
	http.HandleFunc("/githubbot/webhook", h.handleWebhook)
	return h
}
