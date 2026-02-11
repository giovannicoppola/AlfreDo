package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"alfredo-go/pkg/alfred"

	"github.com/spf13/cobra"
)

var editparseCmd = &cobra.Command{
	Use:   "editparse [input]",
	Short: "Parse task edit input with autocomplete",
	Long:  `Parse user input for editing an existing task, providing label/project autocomplete and task preview.`,
	Args:               cobra.ExactArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		taskID := os.Getenv("myTaskID")

		// Check if we need to create a label first
		mySource := os.Getenv("mySource")
		if mySource == "createLabel" {
			myNewLabel := os.Getenv("myNewLabel")
			if myNewLabel != "" {
				if err := taskService.CreateLabel(myNewLabel); err != nil {
					fmt.Fprintf(os.Stderr, "Error creating label: %v\n", err)
				}
			}
		}

		output, err := taskService.ParseNewTask(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing input: %v\n", err)
			// Show friendly Alfred message instead of silent exit
			errOutput := &alfred.Output{Items: []alfred.OutputItem{{
				Title:    "Downloading your Todoist data...",
				Subtitle: "Press Enter to retry",
				Arg:      input,
				Icon:     &alfred.Icon{Path: "icons/loading.png"},
			}}}
			if errJSON, e := json.Marshal(errOutput); e == nil {
				fmt.Println(string(errJSON))
			}
			os.Exit(0)
		}

		// Inject myTaskID into all output items and update subtitle for edit mode
		for i := range output.Items {
			if output.Items[i].Variables == nil {
				output.Items[i].Variables = make(map[string]any)
			}
			output.Items[i].Variables["myTaskID"] = taskID

			// Update subtitle to say "edit" instead of "create" for preview items
			if output.Items[i].Subtitle != "" {
				output.Items[i].Subtitle = replaceLastCreate(output.Items[i].Subtitle)
			}
		}

		jsonOutput, err := output.Marshal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	},
}

func replaceLastCreate(s string) string {
	const old = "⇧↩️ to create"
	const new_ = "⇧↩️ to edit"
	i := len(s) - len(old)
	if i >= 0 && s[i:] == old {
		return s[:i] + new_
	}
	return s
}

func init() {
	rootCmd.AddCommand(editparseCmd)
}
