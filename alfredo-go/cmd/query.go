package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"alfredo-go/pkg/alfred"

	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [today|due|all] [search]",
	Short: "Query tasks with filtering and autocomplete",
	Long: `Query tasks from Todoist with mode-based filtering and search support.

Modes:
  today    - Tasks due today
  due      - Overdue tasks
  all      - All active tasks
  deadline - Tasks with deadlines, sorted by closest deadline

Search supports @label, #project, and text filtering.`,
	Args:                  cobra.RangeArgs(1, 2),
	DisableFlagParsing:    true,
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]
		search := ""
		if len(args) > 1 {
			search = args[1]
		}

		if mode != "today" && mode != "due" && mode != "all" && mode != "deadline" {
			fmt.Fprintf(os.Stderr, "Error: mode must be 'today', 'due', 'all', or 'deadline'\n")
			os.Exit(1)
		}

		output, err := taskService.QueryTasks(mode, search)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying tasks: %v\n", err)
			// Show friendly Alfred message instead of silent exit
			errOutput := &alfred.Output{Items: []alfred.OutputItem{{
				Title:    "Downloading your Todoist data...",
				Subtitle: "Press Enter to retry",
				Arg:      "",
				Icon:     &alfred.Icon{Path: "icons/loading.png"},
			}}}
			if errJSON, e := json.Marshal(errOutput); e == nil {
				fmt.Println(string(errJSON))
			}
			os.Exit(0)
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
	rootCmd.AddCommand(queryCmd)
}
