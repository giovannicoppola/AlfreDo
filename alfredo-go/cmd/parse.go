package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"alfredo-go/pkg/alfred"

	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse [input]",
	Short: "Parse new task input with autocomplete",
	Long:  `Parse user input for new task creation, providing label/project autocomplete and task preview.`,
	Args:               cobra.ExactArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

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

		jsonOutput, err := output.Marshal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}
