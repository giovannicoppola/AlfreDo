package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	Token string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	token := os.Getenv("TOKEN")
	if token == "" {
		// Fallback to checking other common environment variable names
		token = os.Getenv("TODOIST_TOKEN")
	}

	return &Config{
		Token: token,
	}
}

// GetToken returns the Todoist API token
func (c *Config) GetToken() string {
	return c.Token
}
