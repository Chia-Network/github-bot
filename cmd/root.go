package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "github-bot",
	Short: "GitHub Bot is our do-it-all bot to help manage GitHub",
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	var (
		cfgFile  string
		loop     bool
		loopTime time.Duration
	)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yml", "config file to load")
	rootCmd.PersistentFlags().BoolVar(&loop, "loop", false, "Use this var to periodically check on a loop")
	rootCmd.PersistentFlags().DurationVar(&loopTime, "loop-time", 1*time.Hour, "The amount of time to wait between each iteration of the loop")

	cobra.CheckErr(viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")))
	cobra.CheckErr(viper.BindPFlag("loop", rootCmd.PersistentFlags().Lookup("loop")))
	cobra.CheckErr(viper.BindPFlag("loop-time", rootCmd.PersistentFlags().Lookup("loop-time")))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".github-bot" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".github-bot")

	viper.SetEnvPrefix("GITHUB_BOT")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_, err := fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		if err != nil {
			return
		}
	}
}
