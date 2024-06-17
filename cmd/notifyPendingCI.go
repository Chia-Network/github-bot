package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/github-bot/internal/config"
	"github.com/chia-network/github-bot/internal/database"
	github2 "github.com/chia-network/github-bot/internal/github"
	"github.com/chia-network/github-bot/internal/keybase"
)

var notifyPendingCICmd = &cobra.Command{
	Use:   "notify-pendingci",
	Short: "Sends a Keybase message to a channel, alerting that a community PR is ready for CI to run",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			log.Fatalf("error loading config: %s\n", err.Error())
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
			log.Printf("[ERROR] Could not initialize mysql connection: %s", err.Error())
			return
		}

		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		ctx := context.Background()

		sendMsgDuration := 24 * time.Hour

		for {
			log.Println("Checking for community PRs that are waiting for CI to run")
			listPendingPRs, err := github2.CheckForPendingCI(ctx, client, cfg)
			if err != nil {
				log.Printf("The following error occurred while obtaining a list of pending PRs: %s", err)
				time.Sleep(loopDuration)
				continue
			}

			for _, pr := range listPendingPRs {
				prInfo, err := datastore.GetPRData(pr.Repo, int64(pr.PRNumber))
				if err != nil {
					log.Printf("Error checking PR info in database: %v", err)
					continue
				}

				shouldSendMessage := false
				if prInfo == nil {
					// New PR, record it and send a message
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						log.Printf("Error storing PR data: %v", err)
						continue
					}
					shouldSendMessage = true
				} else if time.Since(prInfo.LastMessageSent) > sendMsgDuration {
					// 24 hours has elapsed since the last message was issues, update the record and send a message
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						log.Printf("Error updating PR data: %v", err)
						continue
					}
					shouldSendMessage = true
				}

				if shouldSendMessage {
					message := fmt.Sprintf("The following pull request is waiting for approval for CI checks to run: %s", pr.URL)
					if err := keybase.SendKeybaseMsg(message); err != nil {
						log.Printf("Failed to send message: %s", err)
					} else {
						log.Printf("Message sent for PR: %s", pr.URL)
					}
				}
			}

			if !loop {
				break
			}
			log.Printf("Waiting %s for next iteration\n", loopDuration.String())
			time.Sleep(loopDuration)
		}
	},
}

func init() {
	rootCmd.AddCommand(notifyPendingCICmd)
}
