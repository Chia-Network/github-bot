package cmd

import (
	"fmt"
	"strconv"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/chia-network/github-bot/internal/database"
)

var updateSuppressCmd = &cobra.Command{
	Use:   "update-suppress",
	Short: "Update the suppress_messages flag for a specific PR",
	Run: func(cmd *cobra.Command, args []string) {
		slogs.Init("info") // Initialize logging

		repo := viper.GetString("repo")
		prNumber := viper.GetInt64("pr-number")
		table := viper.GetString("table")
		suppress := viper.GetString("suppress-message")

		slogs.Logr.Info("Received parameters", "repo", repo, "pr-number", prNumber, "table", table, "suppress-message", suppress)

		if repo == "" || prNumber == 0 || table == "" || suppress == "" {
			slogs.Logr.Error("Missing required flags", "repo", repo, "pr-number", prNumber, "table", table, "suppress-message", suppress)
			return
		}

		suppressBool, err := strconv.ParseBool(suppress)
		if err != nil {
			slogs.Logr.Error("Invalid value for suppress-message", "suppress-message", suppress, "error", err)
			return
		}

		slogs.Logr.Info("Connecting to the database")
		datastore, err := database.NewDatastore(
			viper.GetString("db-host"),
			viper.GetUint16("db-port"),
			viper.GetString("db-user"),
			viper.GetString("db-pass"),
			viper.GetString("db-name"),
			table,
		)
		if err != nil {
			slogs.Logr.Error("Could not initialize MySQL connection", "error", err)
			return
		}

		slogs.Logr.Info("Updating suppress_messages flag", "repo", repo, "pr-number", prNumber, "suppress", suppressBool)
		err = datastore.UpdateSuppressMessages(repo, prNumber, suppressBool)
		if err != nil {
			action := "suppressing"
			if !suppressBool {
				action = "unsuppressing"
			}
			slogs.Logr.Error(fmt.Sprintf("Error %s messages for PR", action), "repo", repo, "pr-number", prNumber, "error", err)
			return
		}

		action := "Messages suppressed"
		if !suppressBool {
			action = "Messages unsuppressed"
		}
		slogs.Logr.Info(fmt.Sprintf("%s for PR", action), "repository", repo, "PR", prNumber)
	},
}

func init() {
	rootCmd.AddCommand(updateSuppressCmd)
	updateSuppressCmd.Flags().String("repo", "", "Repository name")
	updateSuppressCmd.Flags().Int64("pr-number", 0, "PR number")
	updateSuppressCmd.Flags().String("table", "", "Database table name")
	updateSuppressCmd.Flags().String("suppress-message", "", "Set to true to suppress messages, false to unsuppress")

	cobra.CheckErr(viper.BindPFlag("repo", updateSuppressCmd.Flags().Lookup("repo")))
	cobra.CheckErr(viper.BindPFlag("pr-number", updateSuppressCmd.Flags().Lookup("pr-number")))
	cobra.CheckErr(viper.BindPFlag("table", updateSuppressCmd.Flags().Lookup("table")))
	cobra.CheckErr(viper.BindPFlag("suppress-message", updateSuppressCmd.Flags().Lookup("suppress-message")))
}
