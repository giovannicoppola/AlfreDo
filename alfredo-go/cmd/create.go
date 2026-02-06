package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [input]",
	Short: "Create a new task",
	Long:  `Create a new Todoist task using variables passed from Alfred workflow.`,
	Args:               cobra.ExactArgs(1),
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		taskText := os.Getenv("myTaskText")
		taskLabels := os.Getenv("myTagString")
		taskProjectID := os.Getenv("myProjectID")
		taskSectionID := os.Getenv("mySectionID")
		myDueDate := os.Getenv("myDueDate")
		myPriorityStr := os.Getenv("myPriority")

		priority := 1
		if myPriorityStr != "" {
			if p, err := strconv.Atoi(myPriorityStr); err == nil {
				priority = p
			}
		}

		if taskText == "" {
			taskText = args[0]
		}

		err := taskService.CreateTask(taskText, taskLabels, taskProjectID, taskSectionID, myDueDate, priority)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating task: %v\n", err)
			fmt.Println("❌ server error\ncheck debugger")
			os.Exit(1)
		}

		fmt.Println("🎯 task created!\nWell done.")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
