package parser

import (
	"strings"
	"testing"
	"time"
)

func TestParseDueString(t *testing.T) {
	tests := []struct {
		name      string
		dueStr    string
		wantDate  bool
		wantMenu  bool
	}{
		{"days", "7d", true, false},
		{"days with time", "7d13:30", true, false},
		{"weeks", "2w", true, false},
		{"months", "3m", true, false},
		{"absolute date", "2025-03-15", true, false},
		{"absolute datetime", "2025-03-15T14:00", true, false},
		{"empty shows menu", "", false, true},
		{"partial shows menu", "abc", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, items, needsMenu := ParseDueString(tt.dueStr, "test input", "en")
			if tt.wantDate && result == "" {
				t.Errorf("expected a resolved date for %q, got empty", tt.dueStr)
			}
			if tt.wantMenu && !needsMenu {
				t.Errorf("expected menu for %q, got resolved: %q", tt.dueStr, result)
			}
			if !tt.wantMenu && needsMenu {
				t.Errorf("expected resolved date for %q, got menu with %d items", tt.dueStr, len(items))
			}
		})
	}
}

func TestParseDueStringDays(t *testing.T) {
	result, _, _ := ParseDueString("7d", "", "en")
	expected := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	if result != expected {
		t.Errorf("ParseDueString(7d) = %q, want %q", result, expected)
	}
}

func TestParseDueStringDaysWithTime(t *testing.T) {
	result, _, _ := ParseDueString("7d13:30", "", "en")
	expected := time.Now().AddDate(0, 0, 7).Format("2006-01-02") + "T13:30"
	if result != expected {
		t.Errorf("ParseDueString(7d13:30) = %q, want %q", result, expected)
	}
}

func TestParseDueStringWeeks(t *testing.T) {
	result, _, _ := ParseDueString("2w", "", "en")
	expected := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	if result != expected {
		t.Errorf("ParseDueString(2w) = %q, want %q", result, expected)
	}
}

func TestParseDueStringAbsolute(t *testing.T) {
	result, _, _ := ParseDueString("2025-03-15", "", "en")
	if result != "2025-03-15" {
		t.Errorf("ParseDueString(2025-03-15) = %q, want 2025-03-15", result)
	}
}

func TestValidateTime(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"13:30", true},
		{"00:00", true},
		{"23:59", true},
		{"24:00", false},
		{"12:60", false},
		{"abc", false},
	}
	for _, tt := range tests {
		result := ValidateTime(tt.input)
		if result != tt.valid {
			t.Errorf("ValidateTime(%q) = %v, want %v", tt.input, result, tt.valid)
		}
	}
}

func TestNewDate(t *testing.T) {
	result := NewDate(0)
	expected := time.Now().Format("2006-01-02")
	if result != expected {
		t.Errorf("NewDate(0) = %q, want %q", result, expected)
	}
}

func TestNewDateFormatted(t *testing.T) {
	result := NewDateFormatted(0)
	// Should contain current day of week
	today := time.Now().Format("Monday")
	if !strings.Contains(result, today) {
		t.Errorf("NewDateFormatted(0) = %q, should contain %q", result, today)
	}
}

func TestResolveRescheduleDate(t *testing.T) {
	tests := []struct {
		input string
		check func(string) bool
	}{
		{"2025-03-15", func(s string) bool { return s == "2025-03-15" }},
		{"2025-03-15T14:00", func(s string) bool { return s == "2025-03-15T14:00" }},
		{"7d", func(s string) bool { return s == NewDate(7) }},
		{"7d13:30", func(s string) bool { return s == NewDate(7)+"T13:30" }},
		{"0", func(s string) bool { return s == NewDate(0) }},
	}
	for _, tt := range tests {
		result := ResolveRescheduleDate(tt.input)
		if !tt.check(result) {
			t.Errorf("ResolveRescheduleDate(%q) = %q", tt.input, result)
		}
	}
}

func TestBuildRescheduleMenu(t *testing.T) {
	// Empty input returns 4 preset options
	items := BuildRescheduleMenu("", "test task", "en")
	if len(items) != 4 {
		t.Errorf("BuildRescheduleMenu(\"\") returned %d items, want 4", len(items))
	}

	// Valid input returns 1 option
	items = BuildRescheduleMenu("7", "test task", "en")
	if len(items) != 1 {
		t.Errorf("BuildRescheduleMenu(\"7\") returned %d items, want 1", len(items))
	}

	// NLP input should resolve to a date
	items = BuildRescheduleMenu("tomorrow", "test task", "en")
	if len(items) != 1 {
		t.Errorf("BuildRescheduleMenu(\"tomorrow\") returned %d items, want 1", len(items))
	}
	if strings.Contains(items[0].Title, "Invalid") {
		t.Errorf("expected NLP date for 'tomorrow', got invalid: %q", items[0].Title)
	}
	// Arg should be a resolved date (YYYY-MM-DD), not nlp: prefix
	expected := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	if !strings.HasPrefix(items[0].Arg, expected) {
		t.Errorf("expected arg starting with %q, got %q", expected, items[0].Arg)
	}

	// Truly invalid input
	items = BuildRescheduleMenu("xyzzy", "test task", "en")
	if len(items) != 1 {
		t.Errorf("BuildRescheduleMenu(\"xyzzy\") returned %d items, want 1", len(items))
	}
	if !strings.Contains(items[0].Title, "Invalid") {
		t.Errorf("expected invalid format message, got %q", items[0].Title)
	}
}

func TestParseNaturalDate(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"tomorrow", true},
		{"next friday", true},
		{"in 3 days", true},
		{"", false},
		{"xyzzy", false},
	}
	for _, tt := range tests {
		_, ok := ParseNaturalDate(tt.input, "en")
		if ok != tt.valid {
			t.Errorf("ParseNaturalDate(%q, \"en\") = _, %v, want %v", tt.input, ok, tt.valid)
		}
	}
}

func TestParseNaturalDateLocale(t *testing.T) {
	// Italian
	_, ok := ParseNaturalDate("domani", "it")
	if !ok {
		t.Error("ParseNaturalDate(\"domani\", \"it\") should resolve")
	}

	_, ok = ParseNaturalDate("oggi", "it")
	if !ok {
		t.Error("ParseNaturalDate(\"oggi\", \"it\") should resolve")
	}

	_, ok = ParseNaturalDate("venerdì", "it")
	if !ok {
		t.Error("ParseNaturalDate(\"venerdì\", \"it\") should resolve")
	}

	// German
	_, ok = ParseNaturalDate("morgen", "de")
	if !ok {
		t.Error("ParseNaturalDate(\"morgen\", \"de\") should resolve")
	}

	_, ok = ParseNaturalDate("freitag", "de")
	if !ok {
		t.Error("ParseNaturalDate(\"freitag\", \"de\") should resolve")
	}

	// French
	_, ok = ParseNaturalDate("demain", "fr")
	if !ok {
		t.Error("ParseNaturalDate(\"demain\", \"fr\") should resolve")
	}

	// German multi-word: "nächsten freitag" / "nachsten freitag"
	_, ok = ParseNaturalDate("nachsten freitag", "de")
	if !ok {
		t.Error("ParseNaturalDate(\"nachsten freitag\", \"de\") should resolve")
	}

	_, ok = ParseNaturalDate("nächsten Freitag", "de")
	if !ok {
		t.Error("ParseNaturalDate(\"nächsten Freitag\", \"de\") should resolve")
	}

	// Italian multi-word: "prossimo venerdì"
	_, ok = ParseNaturalDate("prossimo venerdì", "it")
	if !ok {
		t.Error("ParseNaturalDate(\"prossimo venerdì\", \"it\") should resolve")
	}

	// Unknown locale keyword
	_, ok = ParseNaturalDate("xyzzy", "it")
	if ok {
		t.Error("ParseNaturalDate(\"xyzzy\", \"it\") should not resolve")
	}
}

func TestParseDueStringNLP(t *testing.T) {
	// NLP date should resolve instead of showing menu
	result, _, needsMenu := ParseDueString("tomorrow", "test due:tomorrow", "en")
	if needsMenu {
		t.Error("expected NLP resolution for 'tomorrow', got menu")
	}
	if result == "" {
		t.Error("expected resolved date for 'tomorrow', got empty")
	}
	// Should start with tomorrow's date (may include time component)
	expected := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	if !strings.HasPrefix(result, expected) {
		t.Errorf("ParseDueString(tomorrow) = %q, want prefix %q", result, expected)
	}
}

func TestParseDueStringLocale(t *testing.T) {
	// Italian NLP should resolve locally
	result, _, needsMenu := ParseDueString("domani", "test due:domani", "it")
	if needsMenu {
		t.Error("expected locale NLP resolution for 'domani', got menu")
	}
	if result == "" {
		t.Error("expected resolved date for 'domani', got empty")
	}
	expected := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	if result != expected {
		t.Errorf("ParseDueString(domani, it) = %q, want %q", result, expected)
	}
}

func TestFormatResolvedDate(t *testing.T) {
	// Date only
	dt := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)
	if got := FormatResolvedDate(dt); got != "2025-06-15" {
		t.Errorf("FormatResolvedDate(date) = %q, want 2025-06-15", got)
	}
	// Date with time
	dt = time.Date(2025, 6, 15, 14, 30, 0, 0, time.Local)
	if got := FormatResolvedDate(dt); got != "2025-06-15T14:30" {
		t.Errorf("FormatResolvedDate(datetime) = %q, want 2025-06-15T14:30", got)
	}
}

func TestIsNaturalLanguageDate(t *testing.T) {
	if IsNaturalLanguageDate("7d", "en") {
		t.Error("7d should not be NLP")
	}
	if IsNaturalLanguageDate("2025-03-15", "en") {
		t.Error("ISO date should not be NLP")
	}
	if !IsNaturalLanguageDate("tomorrow", "en") {
		t.Error("tomorrow should be NLP")
	}
	if !IsNaturalLanguageDate("domani", "it") {
		t.Error("domani should be NLP with lang=it")
	}
}
