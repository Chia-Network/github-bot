package cmd

import (
	"log"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/github-bot/internal/config"
	github2 "github.com/chia-network/github-bot/internal/github"
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

		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		var listPendingPRs []string
		for {
			log.Println("Checking for community PRs that have no update in the last 7 days")
			_, err = github2.CheckStalePRs(client, cfg.InternalTeam, cfg.CheckStalePending)
			if err != nil {
				log.Printf("The following error occurred while obtaining a list of stale PRs: %s", err)
				continue
			}
			log.Printf("Stale PRs: %v\n", listPendingPRs)
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
