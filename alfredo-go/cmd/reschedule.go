package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var finddateCmd = &cobra.Command{
	Use:                "finddate [taskID] [input]",
	Short:              "Show reschedule date menu",
	Long:               `Show a date picker menu for rescheduling a task.`,
	Args:               cobra.RangeArgs(1, 2),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		customDays := ""
		if len(args) > 1 {
			customDays = args[1]
		}

		taskContent := os.Getenv("myTaskContent")

		output := taskService.BuildRescheduleMenu(customDays, taskContent)

		jsonOutput, err := output.Marshal()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(jsonOutput))
	},
}

var rescheduleCmd = &cobra.Command{
	Use:                "reschedule [date]",
	Short:              "Reschedule a task",
	Long:               `Reschedule a task to a new date. Reads task ID from myTaskID environment variable.`,
	Args:               cobra.ExactArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		dateInput := args[0]
		taskID := os.Getenv("myTaskID")

		if taskID == "" {
			fmt.Fprintf(os.Stderr, "Error: myTaskID environment variable is required\n")
			os.Exit(1)
		}

		err := taskService.RescheduleTask(taskID, dateInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rescheduling task: %v\n", err)
			fmt.Println("❌ server error\ncheck debugger")
			os.Exit(1)
		}

		fmt.Println("🎯 task rescheduled!\nGet to work!😅")
	},
}

func init() {
	rootCmd.AddCommand(finddateCmd)
	rootCmd.AddCommand(rescheduleCmd)
}
