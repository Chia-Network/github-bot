package cmd

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/github-bot/internal/config"
	github2 "github.com/chia-network/github-bot/internal/github"
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

		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		ctx := context.Background()
		var listPendingPRs []string
		for {
			log.Println("Checking for community PRs that are waiting for CI to run")
			listPendingPRs, err = github2.CheckForPendingCI(ctx, client, cfg)
			if err != nil {
				log.Printf("The following error occurred while obtaining a list of pending PRs: %s", err)
				time.Sleep(loopDuration)
				continue
			}
			log.Printf("Pending PRs ready for CI: %v\n", listPendingPRs)

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
