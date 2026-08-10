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

func TestLoadConfig_RefreshRateClamped(t *testing.T) {
	tests := []struct {
		env  string
		want int
	}{
		{"", 1},    // unset falls back to the default
		{"0", 1},   // 0 would refresh on every keystroke (issue #41)
		{"-3", 1},  // negative is meaningless
		{"abc", 1}, // unparseable falls back to the default
		{"7", 7},   // valid values are kept
	}

	for _, tt := range tests {
		if tt.env == "" {
			os.Unsetenv("RefreshRate")
		} else {
			os.Setenv("RefreshRate", tt.env)
		}

		if got := LoadConfig().RefreshRate; got != tt.want {
			t.Errorf("RefreshRate=%q: got %d, want %d", tt.env, got, tt.want)
		}
	}

	os.Unsetenv("RefreshRate")
}

func TestGetToken(t *testing.T) {
	expectedToken := "test-token-456"
	config := &Config{Token: expectedToken}

	if config.GetToken() != expectedToken {
		t.Errorf("Expected token %s, got %s", expectedToken, config.GetToken())
	}
}
