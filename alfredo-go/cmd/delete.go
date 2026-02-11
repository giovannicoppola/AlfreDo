package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [task-id]",
	Short: "Delete a task",
	Long:  `Delete a task from Todoist using the task ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]

		if taskID == "" {
			fmt.Fprintf(os.Stderr, "Error: task ID is required\n")
			os.Exit(1)
		}

		err := taskService.DeleteTask(taskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
			fmt.Println("❌ server error\ncheck debugger")
			os.Exit(1)
		}

		fmt.Println("🗑️ task deleted!\nGoodbye task.")
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
