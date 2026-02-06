package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Force refresh the local cache",
	Long:  `Force a full refresh of the cached Todoist data.`,
	Run: func(cmd *cobra.Command, args []string) {
		output, err := taskService.ForceRebuild()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rebuilding cache: %v\n", err)
			os.Exit(1)
		}

		jsonOutput, err := output.Marshal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	},
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
}
