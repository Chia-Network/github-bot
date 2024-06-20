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

var notifyStaleCmd = &cobra.Command{
	Use:   "notify-stale",
	Short: "Sends a Keybase message to a channel, alerting that a community PR has not been updated in 7 days",
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
			"stale_pr_status",
		)

		if err != nil {
			log.Printf("[ERROR] Could not initialize mysql connection: %s", err.Error())
			return
		}
		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		sendMsgDuration := 24 * time.Hour // Define the sendMsgDuration
		ctx := context.Background()
		for {
			log.Println("Checking for community PRs that have no update in the last 7 days")
			listPendingPRs, err := github2.CheckStalePRs(ctx, client, cfg)
			if err != nil {
				log.Printf("The following error occurred while obtaining a list of stale PRs: %s", err)
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
					log.Printf("Storing data in db")
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						log.Printf("Error storing PR data: %v", err)
						continue
					}
					shouldSendMessage = true
				} else if time.Since(prInfo.LastMessageSent) > sendMsgDuration {
					// 24 hours has elapsed since the last message was issued, update the record and send a message
					log.Printf("Updating last_message_sent time in db")
					err := datastore.StorePRData(pr.Repo, int64(pr.PRNumber))
					if err != nil {
						log.Printf("Error updating PR data: %v", err)
						continue
					}
					shouldSendMessage = true
				}

				if shouldSendMessage {
					message := fmt.Sprintf("The following pull request is waiting for approval for CI checks to run: %s", pr.URL)
					log.Printf("Sending message via keybase")
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
	rootCmd.AddCommand(notifyStaleCmd)
}
