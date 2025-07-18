package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Get completion statistics from Todoist",
	Long:  `Get completion statistics from Todoist including daily and weekly goals and progress.`,
	Run: func(cmd *cobra.Command, args []string) {
		stats, err := taskService.GetStats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting stats: %v\n", err)
			os.Exit(1)
		}

		jsonOutput, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
