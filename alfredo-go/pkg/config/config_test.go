package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test with no environment variable
	os.Unsetenv("TOKEN")
	os.Unsetenv("TODOIST_TOKEN")

	config := LoadConfig()
	if config.Token != "" {
		t.Errorf("Expected empty token, got %s", config.Token)
	}

	// Test with TOKEN environment variable
	expectedToken := "test-token-123"
	os.Setenv("TOKEN", expectedToken)

	config = LoadConfig()
	if config.Token != expectedToken {
		t.Errorf("Expected token %s, got %s", expectedToken, config.Token)
	}

	// Clean up
	os.Unsetenv("TOKEN")
}

func TestGetToken(t *testing.T) {
	expectedToken := "test-token-456"
	config := &Config{Token: expectedToken}

	if config.GetToken() != expectedToken {
		t.Errorf("Expected token %s, got %s", expectedToken, config.GetToken())
	}
}
