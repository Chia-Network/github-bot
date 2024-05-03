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
	Short: "Sends a Keybase message to those ",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			log.Fatalf("error loading config: %s\n", err.Error())
		}
		client := github.NewClient(nil).WithAuthToken(cfg.GithubToken)

		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		for {
			log.Println("Checking for community PRs that have no update in the last 7 days")
			err = github2.CheckStalePRs(client, cfg.InternalTeam, cfg.LabelConfig)
			if err != nil {
				log.Fatalln(err.Error())
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
