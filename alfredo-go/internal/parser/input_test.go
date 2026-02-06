package parser

import (
	"testing"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single word",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "label without spaces",
			input:    "@work",
			expected: []string{"@work"},
		},
		{
			name:     "label with spaces in parens",
			input:    "@(my label)",
			expected: []string{"@(my label)"},
		},
		{
			name:     "project without spaces",
			input:    "#Shopping",
			expected: []string{"#Shopping"},
		},
		{
			name:     "project with spaces in parens",
			input:    "#(My Project)",
			expected: []string{"#(My Project)"},
		},
		{
			name:     "mixed input",
			input:    "buy milk @groceries #Shopping",
			expected: []string{"buy", "milk", "@groceries", "#Shopping"},
		},
		{
			name:     "label and project with spaces",
			input:    "task @(my label) #(My Project)",
			expected: []string{"task", "@(my label)", "#(My Project)"},
		},
		{
			name:     "priority and due",
			input:    "buy milk p1 due:7d",
			expected: []string{"buy", "milk", "p1", "due:7d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInput(tt.input)
			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("ParseInput(%q) = %v (len %d), want %v (len %d)",
					tt.input, result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("ParseInput(%q)[%d] = %q, want %q",
						tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestNormalizeUnicode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"  hello  ", "hello"},
		{"ä", "ä"}, // should normalize combining diacritics
	}
	for _, tt := range tests {
		result := NormalizeUnicode(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeUnicode(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestUnwrapParens(t *testing.T) {
	tests := []struct {
		input    string
		prefix   string
		expected string
	}{
		{"@(my label)", "@", "@my label"},
		{"@work", "@", "@work"},
		{"#(My Project)", "#", "#My Project"},
		{"#Shopping", "#", "#Shopping"},
		{"@(nospace)", "@", "@(nospace)"}, // no space, don't unwrap
	}
	for _, tt := range tests {
		result := unwrapParens(tt.input, tt.prefix)
		if result != tt.expected {
			t.Errorf("unwrapParens(%q, %q) = %q, want %q", tt.input, tt.prefix, result, tt.expected)
		}
	}
}

func TestParseNewTaskInput(t *testing.T) {
	ctx := &InputContext{
		AllLabels:     []string{"@groceries", "@work", "@home"},
		AllProjects:   []string{"#Shopping", "#Inbox", "#Work"},
		LabelCounts:   map[string]int{"groceries": 5, "work": 3, "home": 1},
		ProjectCounts: map[string]int{"Shopping": 2, "Inbox": 10, "Work": 5},
		PartialMatch:  true,
	}

	// Test basic task with known label and project
	parsed, autocomplete, needsExit := ParseNewTaskInput("buy milk @groceries #Shopping due:7d p1", ctx)
	if needsExit {
		t.Fatalf("expected no exit, got autocomplete items: %v", autocomplete)
	}
	if parsed.Content != "buy milk" {
		t.Errorf("Content = %q, want %q", parsed.Content, "buy milk")
	}
	if len(parsed.Labels) != 1 || parsed.Labels[0] != "groceries" {
		t.Errorf("Labels = %v, want [groceries]", parsed.Labels)
	}
	if parsed.ProjectName != "#Shopping" {
		t.Errorf("ProjectName = %q, want #Shopping", parsed.ProjectName)
	}
	if parsed.Priority != 4 {
		t.Errorf("Priority = %d, want 4", parsed.Priority)
	}
	if parsed.DueDate == "" {
		t.Error("DueDate should not be empty")
	}

	// Test autocomplete for partial label
	_, autocomplete, needsExit = ParseNewTaskInput("buy @gro", ctx)
	if !needsExit {
		t.Error("expected exit for partial label match")
	}
	if len(autocomplete) == 0 {
		t.Error("expected autocomplete items for @gro")
	}
}
