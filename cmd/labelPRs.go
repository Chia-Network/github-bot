package cmd

import (
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/github-bot/internal/config"
	"github.com/chia-network/github-bot/internal/label"
	log "github.com/chia-network/github-bot/internal/logger"
)

// labelPRsCmd represents the labelPRs command
var labelPRsCmd = &cobra.Command{
	Use:   "label-prs",
	Short: "Adds community and internal labels to pull requests in designated repos",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig(viper.GetString("config"))
		if err != nil {
			log.Logger.Error("Error loading config", "error", err)
		}
		client := github.NewClient(nil).WithAuthToken(cfg.GithubToken)

		loop := viper.GetBool("loop")
		loopDuration := viper.GetDuration("loop-time")
		for {
			log.Logger.Info("Labeling Pull Requests")
			err = label.PullRequests(client, cfg)
			if err != nil {
				log.Logger.Error("Error labeling pull requests", "error", err)
			}

			if !loop {
				break
			}

			log.Logger.Info("Waiting for next iteration", "duration", loopDuration.String())
			time.Sleep(loopDuration)
		}
	},
}

func init() {
	rootCmd.AddCommand(labelPRsCmd)
}
