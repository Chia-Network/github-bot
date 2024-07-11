package cmd

import (
	"context"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/go-modules/pkg/slogs"

	"github.com/chia-network/github-bot/internal/config"
	github2 "github.com/chia-network/github-bot/internal/github"
)

var notifyUnsignedCommitsCmd = &cobra.Command{
	Use:   "notify-unsigned",
	Short: "Provides a comment to the Pull Request author that unsigned commits are present",
	Run: func(cmd *cobra.Command, args []string) {
		slogs.Init("info")
		cfg, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			slogs.Logr.Fatal("Error loading config", "error", err)
		}
		client := github.NewClient(nil).WithAuthToken(cfg.GithubToken)

		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		ctx := context.Background()

		for {
			slogs.Logr.Info("Checking for community PRs that are waiting for CI to run")
			listPendingPRs, err := github2.CheckUnsignedCommits(ctx, client, cfg)
			if err != nil {
				slogs.Logr.Error("Error obtaining a list of pending PRs", "error", err)
				time.Sleep(loopDuration)
				continue
			}

			for _, pr := range listPendingPRs {
				err = github2.CheckAndComment(ctx, client, pr.Owner, pr.Repo, pr.PRNumber)
				slogs.Logr.Info("Found PR with not signed commits", "PR", pr.URL)
				if err != nil {
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
	rootCmd.AddCommand(notifyUnsignedCommitsCmd)
}
