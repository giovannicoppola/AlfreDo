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
	DueLang      string // language for Todoist NLP dates (e.g., "en", "de")
	TaskStamp    string // template for task description (supports {timestamp} placeholder)
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

	dueLang := detectSystemLanguage()

	if dataFolder != "" {
		os.MkdirAll(dataFolder, 0755)
	}

	taskStamp := os.Getenv("TASK_STAMP")

	return &Config{
		Token:        token,
		ShowGoals:    showGoals,
		PartialMatch: partialMatch,
		RefreshRate:  refreshRate,
		TaskOpen:     taskOpen,
		DataFolder:   dataFolder,
		DueLang:      dueLang,
		TaskStamp:    taskStamp,
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

// detectSystemLanguage returns a Todoist-supported language code from the system locale.
// Todoist supports: da, de, en, es, fi, fr, it, ja, ko, nl, pl, pt, ru, sv, tr, zh.
func detectSystemLanguage() string {
	supported := map[string]bool{
		"da": true, "de": true, "en": true, "es": true, "fi": true,
		"fr": true, "it": true, "ja": true, "ko": true, "nl": true,
		"pl": true, "pt": true, "ru": true, "sv": true, "tr": true, "zh": true,
	}

	// Check LANG, then LC_ALL, then LANGUAGE
	for _, key := range []string{"LANG", "LC_ALL", "LANGUAGE"} {
		val := os.Getenv(key)
		if val == "" || val == "C" || val == "POSIX" {
			continue
		}
		// Extract language code: "en_US.UTF-8" → "en", "de_DE" → "de"
		lang := val
		if i := indexByte(lang, '_'); i > 0 {
			lang = lang[:i]
		} else if i := indexByte(lang, '.'); i > 0 {
			lang = lang[:i]
		} else if i := indexByte(lang, '-'); i > 0 {
			lang = lang[:i]
		}
		if supported[lang] {
			return lang
		}
	}

	return "en"
}

func indexByte(s string, c byte) int {
	for i := range s {
		if s[i] == c {
			return i
		}
	}
	return -1
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
