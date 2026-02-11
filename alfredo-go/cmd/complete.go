package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete [task-id]",
	Short: "Mark a task as completed",
	Long:  `Mark a task as completed in Todoist using the task ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]

		if taskID == "" {
			fmt.Fprintf(os.Stderr, "Error: task ID is required\n")
			os.Exit(1)
		}

		err := taskService.CompleteTask(taskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error completing task: %v\n", err)
			fmt.Println("❌ server error\ncheck debugger")
			os.Exit(1)
		}

		fmt.Println("🎯 task completed!\nWell done 💪")
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
