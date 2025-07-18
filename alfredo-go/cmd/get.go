package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [mode]",
	Short: "Get tasks from Todoist",
	Long: `Get tasks from Todoist API. 
	
Mode can be:
- today: Get only tasks due today
- overdue: Get all overdue tasks (including today)`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]

		if mode != "today" && mode != "overdue" {
			fmt.Fprintf(os.Stderr, "Error: mode must be 'today' or 'overdue'\n")
			os.Exit(1)
		}

		output, err := taskService.GetTasksOutput(mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting tasks: %v\n", err)
			os.Exit(1)
		}

		jsonOutput, err := json.Marshal(output)
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
