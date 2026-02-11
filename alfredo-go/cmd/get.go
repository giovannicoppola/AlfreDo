package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// getCmd is kept for backward compatibility; prefer using "query" instead
var getCmd = &cobra.Command{
	Use:   "get [mode]",
	Short: "Get tasks from Todoist (legacy, use 'query' instead)",
	Long: `Get tasks from Todoist API.

Mode can be:
- today: Get only tasks due today
- overdue: Get all overdue tasks (maps to "due" mode)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]

		// Map legacy "overdue" to new "due" mode
		if mode == "overdue" {
			mode = "due"
		}

		if mode != "today" && mode != "due" {
			fmt.Fprintf(os.Stderr, "Error: mode must be 'today' or 'overdue'\n")
			os.Exit(1)
		}

		output, err := taskService.QueryTasks(mode, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting tasks: %v\n", err)
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
	rootCmd.AddCommand(getCmd)
}
