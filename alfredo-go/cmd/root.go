package cmd

import (
	"alfredo-go/internal/service"
	"alfredo-go/pkg/config"
	"alfredo-go/pkg/todoist"

	"github.com/spf13/cobra"
)

var (
	cfg           *config.Config
	todoistClient *todoist.Client
	taskService   *service.TaskService
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "alfredo-go",
	Short: "AlfreDo - A Todoist workflow application",
	Long: `AlfreDo is a command-line application for managing Todoist tasks.
It provides commands to get tasks, complete tasks, and manage your workflow.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config and ENV variables
func initConfig() {
	cfg = config.LoadConfig()
	todoistClient = todoist.NewClient(cfg.GetToken())
	taskService = service.NewTaskService(todoistClient)
}
