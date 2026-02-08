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
		myDueString := os.Getenv("myDueString")
		myDueLang := os.Getenv("myDueLang")
		myDeadline := os.Getenv("myDeadline")
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

		// If we have a natural language due string, prefer it over coded date
		if myDueString != "" {
			myDueDate = ""
		}

		// Determine deadline lang from config
		deadlineLang := ""
		if myDeadline != "" {
			deadlineLang = cfg.DueLang
			if deadlineLang == "" {
				deadlineLang = "en"
			}
		}

		err := taskService.CreateTask(taskText, taskLabels, taskProjectID, taskSectionID, myDueDate, myDueString, myDueLang, priority, myDeadline, deadlineLang)
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
