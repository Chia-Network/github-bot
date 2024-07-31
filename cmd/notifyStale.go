package cmd

import (
	"context"
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

var notifyStaleCmd = &cobra.Command{
	Use:   "notify-stale",
	Short: "Sends a Keybase message to a channel, alerting that a community PR has not been updated in 7 days",
	Run: func(cmd *cobra.Command, args []string) {
		slogs.Init("info")
		cfg, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			slogs.Logr.Fatal("Error loading config", "error", err)
		}
		client := github.NewClient(nil).WithAuthToken(cfg.GithubToken)
		webhookURL := "https://alert-receiver.chiaops.com/devrel"

		datastore, err := database.NewDatastore(
			viper.GetString("db-host"),
			viper.GetUint16("db-port"),
			viper.GetString("db-user"),
			viper.GetString("db-pass"),
			viper.GetString("db-name"),
			"stale_pr_status",
		)

		if err != nil {
			slogs.Logr.Error("Could not initialize mysql connection", "error", err)
			return
		}
		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		sendMsgDuration := 24 * time.Hour // Define the sendMsgDuration
		ctx := context.Background()
		for {
			slogs.Logr.Info("Checking for community PRs that have no update in the last 7 days")
			listPendingPRs, err := github2.CheckStalePRs(ctx, client, cfg)
			if err != nil {
				slogs.Logr.Error("Error checking PR info in database", "error", err)
				time.Sleep(loopDuration)
				continue
			}

			for _, pr := range listPendingPRs {
				prInfo, err := datastore.GetPRData(pr.Repo, int64(pr.PRNumber))
				if err != nil {
					slogs.Logr.Error("Error checking PR info in database", "error", err)
					continue
				}

				if prInfo != nil && prInfo.SuppressMessages {
					slogs.Logr.Info("Skipping message for PR due to suppress_messages flag", "repository", pr.Repo, "PR", int64(pr.PRNumber))
					continue
				}

				shouldSendMessage := false
				if prInfo == nil {
					// New PR, record it and send a message
					slogs.Logr.Info("Storing data in db", "repository", pr.Repo, "PR", int64(pr.PRNumber))
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						slogs.Logr.Error("Error storing PR data", "error", err)
						continue
					}
					shouldSendMessage = true
				} else if time.Since(prInfo.LastMessageSent) > sendMsgDuration {
					// 24 hours has elapsed since the last message was issued, update the record and send a message
					slogs.Logr.Info("Updating last_message_sent time in db", "repository", pr.Repo, "PR", int64(pr.PRNumber))
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						slogs.Logr.Error("Error updating PR data", "error", err)
						continue
					}
					shouldSendMessage = true
				}

				if shouldSendMessage {
					status := "message"
					title := "The following pull request has no activity from a Chia team member in the last 7 days"
					description := pr.URL
					slogs.Logr.Info("Sending message via keybase for", "repository", pr.Repo, "PR", int64(pr.PRNumber))
					message := keybase.NewMessage(status, title, description)
					if err := message.SendKeybaseMsg(webhookURL); err != nil {
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
	rootCmd.AddCommand(notifyStaleCmd)
}
