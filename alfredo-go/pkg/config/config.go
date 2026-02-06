package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Token        string
	ShowGoals    bool
	PartialMatch bool
	RefreshRate  int    // days between auto-refresh
	TaskOpen     string // "app" or "browser"
	DataFolder   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	token := os.Getenv("TOKEN")
	if token == "" {
		token = os.Getenv("TODOIST_TOKEN")
	}

	dataFolder := os.Getenv("alfred_workflow_data")
	if dataFolder == "" {
		dataFolder = os.Getenv("DATA_FOLDER")
	}

	showGoals := envIntBool("SHOW_GOALS", true)
	partialMatch := envIntBool("PARTIAL_MATCH", true)
	refreshRate := envInt("RefreshRate", 1)
	taskOpen := os.Getenv("taskOpen")
	if taskOpen == "" {
		taskOpen = "browser"
	}

	if dataFolder != "" {
		os.MkdirAll(dataFolder, 0755)
	}

	return &Config{
		Token:        token,
		ShowGoals:    showGoals,
		PartialMatch: partialMatch,
		RefreshRate:  refreshRate,
		TaskOpen:     taskOpen,
		DataFolder:   dataFolder,
	}
}

// GetToken returns the Todoist API token
func (c *Config) GetToken() string {
	return c.Token
}

func envIntBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n == 1
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}
