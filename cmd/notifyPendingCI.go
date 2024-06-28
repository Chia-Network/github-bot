package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/github-bot/internal/config"
	"github.com/chia-network/github-bot/internal/database"
	github2 "github.com/chia-network/github-bot/internal/github"
	"github.com/chia-network/github-bot/internal/keybase"

	"github.com/chia-network/go-modules/pkg/slogs"
)

var notifyPendingCICmd = &cobra.Command{
	Use:   "notify-pendingci",
	Short: "Sends a Keybase message to a channel, alerting that a community PR is ready for CI to run",
	Run: func(cmd *cobra.Command, args []string) {
		slogs.Init("info")
		cfg, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			slogs.Logr.Fatal("Error loading config", "error", err)
		}
		client := github.NewClient(nil).WithAuthToken(cfg.GithubToken)

		datastore, err := database.NewDatastore(
			viper.GetString("db-host"),
			viper.GetUint16("db-port"),
			viper.GetString("db-user"),
			viper.GetString("db-pass"),
			viper.GetString("db-name"),
			"pending_ci_status",
		)

		if err != nil {
			slogs.Logr.Error("Could not initialize mysql connection", "error", err)
			return
		}

		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		ctx := context.Background()

		sendMsgDuration := 24 * time.Hour

		for {
			slogs.Logr.Info("Checking for community PRs that are waiting for CI to run")
			listPendingPRs, err := github2.CheckForPendingCI(ctx, client, cfg)
			if err != nil {
				slogs.Logr.Error("Error obtaining a list of pending PRs", "error", err)
				time.Sleep(loopDuration)
				continue
			}

			for _, pr := range listPendingPRs {
				prInfo, err := datastore.GetPRData(pr.Repo, int64(pr.PRNumber))
				if err != nil {
					slogs.Logr.Error("Error checking PR info in database", "error", err)
					continue
				}

				shouldSendMessage := false
				if prInfo == nil {
					// New PR, record it and send a message
					slogs.Logr.Info("Storing data in db")
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						slogs.Logr.Error("Error storing PR data", "error", err)
						continue
					}
					shouldSendMessage = true
				} else if time.Since(prInfo.LastMessageSent) > sendMsgDuration {
					// 24 hours has elapsed since the last message was issued, update the record and send a message
					slogs.Logr.Info("Updating last_message_sent time in db")
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						slogs.Logr.Error("Error updating PR data", "error", err)
						continue
					}
					shouldSendMessage = true
				}

				if shouldSendMessage {
					status := ":rotating_light:"
					title := "Pending/Failed Checks"
					description := fmt.Sprintf("The following pull request is either waiting for approval for CI checks to run or has failed checks: %s", pr.URL)
					slogs.Logr.Info("Sending message via keybase")
					message := keybase.NewMessage(status, title, description)
					if err := message.SendKeybaseMsg(); err != nil {
						slogs.Logr.Error("Failed to send message", "error", err)
						time.Sleep(15 * time.Second) // This is to prevent "error response: 429 Too Many Requests""
					} else {
						slogs.Logr.Info("Message sent for PR", "URL", pr.URL)
					}
				}
			}

			if !loop {
				break
			}
			slogs.Logr.Info("Waiting for next iteration", "duration", loopDuration.String())
			time.Sleep(loopDuration)
		}
	},
}

func init() {
	rootCmd.AddCommand(notifyPendingCICmd)
}
