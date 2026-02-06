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
			result, items, needsMenu := ParseDueString(tt.dueStr, "test input")
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
	result, _, _ := ParseDueString("7d", "")
	expected := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	if result != expected {
		t.Errorf("ParseDueString(7d) = %q, want %q", result, expected)
	}
}

func TestParseDueStringDaysWithTime(t *testing.T) {
	result, _, _ := ParseDueString("7d13:30", "")
	expected := time.Now().AddDate(0, 0, 7).Format("2006-01-02") + "T13:30"
	if result != expected {
		t.Errorf("ParseDueString(7d13:30) = %q, want %q", result, expected)
	}
}

func TestParseDueStringWeeks(t *testing.T) {
	result, _, _ := ParseDueString("2w", "")
	expected := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	if result != expected {
		t.Errorf("ParseDueString(2w) = %q, want %q", result, expected)
	}
}

func TestParseDueStringAbsolute(t *testing.T) {
	result, _, _ := ParseDueString("2025-03-15", "")
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
	items := BuildRescheduleMenu("", "test task")
	if len(items) != 4 {
		t.Errorf("BuildRescheduleMenu(\"\") returned %d items, want 4", len(items))
	}

	// Valid input returns 1 option
	items = BuildRescheduleMenu("7", "test task")
	if len(items) != 1 {
		t.Errorf("BuildRescheduleMenu(\"7\") returned %d items, want 1", len(items))
	}

	// Invalid input
	items = BuildRescheduleMenu("abc", "test task")
	if len(items) != 1 {
		t.Errorf("BuildRescheduleMenu(\"abc\") returned %d items, want 1", len(items))
	}
	if !strings.Contains(items[0].Title, "Invalid") {
		t.Errorf("expected invalid format message, got %q", items[0].Title)
	}
}
