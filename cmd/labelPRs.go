package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// labelPRsCmd represents the labelPRs command
var labelPRsCmd = &cobra.Command{
	Use:   "label-prs",
	Short: "Adds community and internal labels to pull requests in designated repos",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("labelPRs called")
	},
}

func init() {
	rootCmd.AddCommand(labelPRsCmd)
}
