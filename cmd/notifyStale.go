package cmd

import (
	"context"
	"fmt"
	"os"
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
			slogs.Logger.Info("Error loading config", "error", err)
			os.Exit(1)
		}
		client := github.NewClient(nil).WithAuthToken(cfg.GithubToken)

		datastore, err := database.NewDatastore(
			viper.GetString("db-host"),
			viper.GetUint16("db-port"),
			viper.GetString("db-user"),
			viper.GetString("db-pass"),
			viper.GetString("db-name"),
			"stale_pr_status",
		)

		if err != nil {
			slogs.Logger.Error("Could not initialize mysql connection", "error", err)
			return
		}
		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		sendMsgDuration := 24 * time.Hour // Define the sendMsgDuration
		ctx := context.Background()
		for {
			slogs.Logger.Info("Checking for community PRs that have no update in the last 7 days")
			listPendingPRs, err := github2.CheckStalePRs(ctx, client, cfg)
			if err != nil {
				slogs.Logger.Error("Error checking PR info in database", "error", err)
				time.Sleep(loopDuration)
				continue
			}

			for _, pr := range listPendingPRs {
				prInfo, err := datastore.GetPRData(pr.Repo, int64(pr.PRNumber))
				if err != nil {
					slogs.Logger.Error("Error checking PR info in database", "error", err)
					continue
				}

				shouldSendMessage := false
				if prInfo == nil {
					// New PR, record it and send a message
					slogs.Logger.Info("Storing data in db")
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						slogs.Logger.Error("Error storing PR data", "error", err)
						continue
					}
					shouldSendMessage = true
				} else if time.Since(prInfo.LastMessageSent) > sendMsgDuration {
					// 24 hours has elapsed since the last message was issued, update the record and send a message
					slogs.Logger.Info("Updating last_message_sent time in db")
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						slogs.Logger.Error("Error updating PR data", "error", err)
						continue
					}
					shouldSendMessage = true
				}

				if shouldSendMessage {
					message := fmt.Sprintf("The following pull request is waiting for approval for CI checks to run: %s", pr.URL)
					slogs.Logger.Info("Sending message via keybase")
					if err := keybase.SendKeybaseMsg(message); err != nil {
						slogs.Logger.Error("Failed to send message", "error", err)
					} else {
						slogs.Logger.Info("Message sent for PR", "URL", pr.URL)
					}
				}
			}

			if !loop {
				break
			}

			slogs.Logger.Info("Waiting for next iteration", "duration", loopDuration.String())
			time.Sleep(loopDuration)
		}
	},
}

func init() {
	rootCmd.AddCommand(notifyStaleCmd)
}
