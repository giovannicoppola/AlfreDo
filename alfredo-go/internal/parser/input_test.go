package parser

import (
	"strings"
	"testing"
	"time"
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
		Lang:          "en",
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

func TestParseNewTaskInputDeadline(t *testing.T) {
	ctx := &InputContext{
		AllLabels:     []string{},
		AllProjects:   []string{"#Work"},
		LabelCounts:   map[string]int{},
		ProjectCounts: map[string]int{"Work": 5},
		PartialMatch:  true,
		Lang:          "en",
	}

	// Test {deadline} with ISO date
	parsed, _, needsExit := ParseNewTaskInput("buy groceries {2025-06-30} #Work", ctx)
	if needsExit {
		t.Fatal("expected no exit for deadline input")
	}
	if parsed.Content != "buy groceries" {
		t.Errorf("Content = %q, want %q", parsed.Content, "buy groceries")
	}
	if parsed.Deadline != "2025-06-30" {
		t.Errorf("Deadline = %q, want 2025-06-30", parsed.Deadline)
	}
	if parsed.DeadlineRaw != "2025-06-30" {
		t.Errorf("DeadlineRaw = %q, want 2025-06-30", parsed.DeadlineRaw)
	}

	// Test {deadline} with relative days
	parsed, _, needsExit = ParseNewTaskInput("report {7d}", ctx)
	if needsExit {
		t.Fatal("expected no exit for deadline input")
	}
	if parsed.Content != "report" {
		t.Errorf("Content = %q, want %q", parsed.Content, "report")
	}
	if parsed.Deadline == "" {
		t.Error("Deadline should not be empty for {7d}")
	}

	// Test {deadline} with English NLP
	parsed, _, needsExit = ParseNewTaskInput("report {tomorrow}", ctx)
	if needsExit {
		t.Fatal("expected no exit for NLP deadline input")
	}
	if parsed.Content != "report" {
		t.Errorf("Content = %q, want %q", parsed.Content, "report")
	}
	if parsed.Deadline == "" {
		t.Error("Deadline should not be empty for {tomorrow}")
	}
	if parsed.DeadlineRaw != "tomorrow" {
		t.Errorf("DeadlineRaw = %q, want 'tomorrow'", parsed.DeadlineRaw)
	}

	// Test {deadline} with unrecognized string (no lang set) — silently ignored
	parsed, _, needsExit = ParseNewTaskInput("report {domani}", ctx)
	if needsExit {
		t.Fatal("expected no exit for unresolved deadline input")
	}
	if parsed.Content != "report" {
		t.Errorf("Content = %q, want %q", parsed.Content, "report")
	}
	if parsed.Deadline != "" {
		t.Errorf("Deadline should be empty for unresolved NLP (lang=en), got %q", parsed.Deadline)
	}
	if parsed.DeadlineRaw != "domani" {
		t.Errorf("DeadlineRaw = %q, want 'domani'", parsed.DeadlineRaw)
	}

	// Test {deadline} with Italian locale — now resolves locally
	ctxIT := &InputContext{
		AllLabels:     []string{},
		AllProjects:   []string{"#Work"},
		LabelCounts:   map[string]int{},
		ProjectCounts: map[string]int{"Work": 5},
		PartialMatch:  true,
		Lang:          "it",
	}
	parsed, _, needsExit = ParseNewTaskInput("report {domani}", ctxIT)
	if needsExit {
		t.Fatal("expected no exit for Italian deadline input")
	}
	if parsed.Deadline == "" {
		t.Error("Deadline should not be empty for {domani} with lang=it")
	}
	expected := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	if parsed.Deadline != expected {
		t.Errorf("Deadline = %q, want %q", parsed.Deadline, expected)
	}
}

func TestParseNewTaskInputNLPDue(t *testing.T) {
	ctx := &InputContext{
		AllLabels:     []string{},
		AllProjects:   []string{},
		LabelCounts:   map[string]int{},
		ProjectCounts: map[string]int{},
		PartialMatch:  true,
		Lang:          "en",
	}

	// Test due:tomorrow (NLP single word)
	parsed, _, needsExit := ParseNewTaskInput("buy milk due:tomorrow", ctx)
	if needsExit {
		t.Fatal("expected no exit for due:tomorrow")
	}
	if parsed.DueDate == "" {
		t.Error("DueDate should not be empty for due:tomorrow")
	}
	if parsed.Content != "buy milk" {
		t.Errorf("Content = %q, want 'buy milk'", parsed.Content)
	}

	// Test due:next friday (NLP multi-word)
	parsed, _, needsExit = ParseNewTaskInput("buy milk due:next friday", ctx)
	if needsExit {
		t.Fatal("expected no exit for due:next friday")
	}
	if parsed.DueDate == "" {
		t.Error("DueDate should not be empty for due:next friday")
	}
	if parsed.Content != "buy milk" {
		t.Errorf("Content = %q, want 'buy milk'", parsed.Content)
	}

	// Test due:domani with Italian locale — resolves locally
	ctxIT := &InputContext{
		AllLabels:     []string{},
		AllProjects:   []string{},
		LabelCounts:   map[string]int{},
		ProjectCounts: map[string]int{},
		PartialMatch:  true,
		Lang:          "it",
	}
	parsed, _, needsExit = ParseNewTaskInput("buy milk due:domani", ctxIT)
	if needsExit {
		t.Fatal("expected no exit for due:domani with lang=it")
	}
	if parsed.DueDate == "" {
		t.Error("DueDate should not be empty for due:domani with lang=it")
	}
	expected := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	if !strings.HasPrefix(parsed.DueDate, expected) {
		t.Errorf("DueDate = %q, want prefix %q", parsed.DueDate, expected)
	}
	if parsed.Content != "buy milk" {
		t.Errorf("Content = %q, want 'buy milk'", parsed.Content)
	}

	// Test due:domani without Italian locale — shows menu (not resolved)
	ctxEN := &InputContext{
		AllLabels:     []string{},
		AllProjects:   []string{},
		LabelCounts:   map[string]int{},
		ProjectCounts: map[string]int{},
		PartialMatch:  true,
		Lang:          "en",
	}
	_, _, needsExit = ParseNewTaskInput("buy milk due:domani", ctxEN)
	if !needsExit {
		t.Error("expected menu for due:domani with lang=en (unresolved)")
	}
}

func TestParseNewTaskInputInlineNLP(t *testing.T) {
	ctx := &InputContext{
		AllLabels:     []string{},
		AllProjects:   []string{},
		LabelCounts:   map[string]int{},
		ProjectCounts: map[string]int{},
		PartialMatch:  true,
		Lang:          "en",
	}

	// Inline: "new task tomorrow" -> content="new task", due resolved
	parsed, _, needsExit := ParseNewTaskInput("new task tomorrow", ctx)
	if needsExit {
		t.Fatal("expected no exit for inline NLP")
	}
	if parsed.Content != "new task" {
		t.Errorf("Content = %q, want 'new task'", parsed.Content)
	}
	if parsed.DueDate == "" {
		t.Error("DueDate should not be empty for inline NLP 'tomorrow'")
	}

	// Inline: "meeting next friday at 3pm" -> content="meeting", due resolved
	parsed, _, needsExit = ParseNewTaskInput("meeting next friday at 3pm", ctx)
	if needsExit {
		t.Fatal("expected no exit for inline NLP")
	}
	if parsed.Content != "meeting" {
		t.Errorf("Content = %q, want 'meeting'", parsed.Content)
	}
	if parsed.DueDate == "" {
		t.Error("DueDate should not be empty for inline NLP")
	}

	// No false positive: "buy milk" -> no date
	parsed, _, needsExit = ParseNewTaskInput("buy milk", ctx)
	if needsExit {
		t.Fatal("expected no exit")
	}
	if parsed.DueDate != "" {
		t.Errorf("DueDate should be empty for 'buy milk', got %q", parsed.DueDate)
	}
}

func TestExtractDuration(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedMinutes int
		expectedContent string
	}{
		{
			name:            "30 minutes",
			content:         "Do something for 30 minutes",
			expectedMinutes: 30,
			expectedContent: "Do something",
		},
		{
			name:            "30m",
			content:         "Do something for 30m",
			expectedMinutes: 30,
			expectedContent: "Do something",
		},
		{
			name:            "2 hours",
			content:         "Do something for 2 hours",
			expectedMinutes: 120,
			expectedContent: "Do something",
		},
		{
			name:            "2h",
			content:         "Do something for 2h",
			expectedMinutes: 120,
			expectedContent: "Do something",
		},
		{
			name:            "2h15m",
			content:         "Do something for 2h15m",
			expectedMinutes: 135,
			expectedContent: "Do something",
		},
		{
			name:            "2h 15m",
			content:         "Do something for 2h 15m",
			expectedMinutes: 135,
			expectedContent: "Do something",
		},
		{
			name:            "3 hours",
			content:         "Do something for 3 hours",
			expectedMinutes: 180,
			expectedContent: "Do something",
		},
		{
			name:            "no duration",
			content:         "Do something",
			expectedMinutes: 0,
			expectedContent: "Do something",
		},
		{
			name:            "1 hour 30 minutes",
			content:         "Meeting for 1 hour 30 minutes",
			expectedMinutes: 90,
			expectedContent: "Meeting",
		},
		{
			name:            "45 mins",
			content:         "Call for 45 mins",
			expectedMinutes: 45,
			expectedContent: "Call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minutes, content := extractDuration(tt.content)
			if minutes != tt.expectedMinutes {
				t.Errorf("extractDuration(%q) minutes = %d, want %d", tt.content, minutes, tt.expectedMinutes)
			}
			if content != tt.expectedContent {
				t.Errorf("extractDuration(%q) content = %q, want %q", tt.content, content, tt.expectedContent)
			}
		})
	}
}

func TestIsDueDateWithTime(t *testing.T) {
	tests := []struct {
		name     string
		dueDate  string
		expected bool
	}{
		{
			name:     "ISO 8601 with time",
			dueDate:  "2025-05-06T17:00:00",
			expected: true,
		},
		{
			name:     "date with colon time",
			dueDate:  "2025-05-06 17:00",
			expected: true,
		},
		{
			name:     "date only",
			dueDate:  "2025-05-06",
			expected: false,
		},
		{
			name:     "empty string",
			dueDate:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDueDateWithTime(tt.dueDate)
			if result != tt.expected {
				t.Errorf("isDueDateWithTime(%q) = %v, want %v", tt.dueDate, result, tt.expected)
			}
		})
	}
}

func TestParseNewTaskInputWithDuration(t *testing.T) {
	ctx := &InputContext{
		AllLabels:     []string{},
		AllProjects:   []string{},
		LabelCounts:   map[string]int{},
		ProjectCounts: map[string]int{},
		PartialMatch:  true,
		Lang:          "en",
	}

	tests := []struct {
		name            string
		input           string
		expectedContent string
		expectedMinutes int
		hasDueDateTime  bool
	}{
		{
			name:            "tomorrow at 5pm for 30 minutes",
			input:           "Do something tomorrow at 5pm for 30 minutes",
			expectedContent: "Do something",
			expectedMinutes: 30,
			hasDueDateTime:  true,
		},
		{
			name:            "tomorrow at 5pm for 30m",
			input:           "Do something tomorrow at 5pm for 30m",
			expectedContent: "Do something",
			expectedMinutes: 30,
			hasDueDateTime:  true,
		},
		{
			name:            "at 5pm for 2h tomorrow",
			input:           "Do something at 5pm for 2h tomorrow",
			expectedContent: "Do something",
			expectedMinutes: 120,
			hasDueDateTime:  true,
		},
		{
			name:            "at 5pm for 2h15m tomorrow",
			input:           "Do something at 5pm for 2h15m tomorrow",
			expectedContent: "Do something",
			expectedMinutes: 135,
			hasDueDateTime:  true,
		},
		{
			name:            "for 3 hours tomorrow at 5pm",
			input:           "Do something for 3 hours tomorrow at 5pm",
			expectedContent: "Do something",
			expectedMinutes: 180,
			hasDueDateTime:  true,
		},
		{
			name:            "tomorrow for 30m (no time, no duration extraction)",
			input:           "Do something tomorrow for 30m",
			expectedContent: "Do something for 30m",
			expectedMinutes: 0,
			hasDueDateTime:  false,
		},
		{
			name:            "at 5pm for 1h (today)",
			input:           "Do something at 5pm for 1h",
			expectedContent: "Do something",
			expectedMinutes: 60,
			hasDueDateTime:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, _, needsExit := ParseNewTaskInput(tt.input, ctx)
			if needsExit {
				t.Fatal("expected no exit")
			}

			if parsed.Content != tt.expectedContent {
				t.Errorf("Content = %q, want %q", parsed.Content, tt.expectedContent)
			}

			if tt.hasDueDateTime {
				if parsed.DueDate == "" {
					t.Error("DueDate should not be empty")
				}
				if !isDueDateWithTime(parsed.DueDate) {
					t.Errorf("DueDate %q should contain time component", parsed.DueDate)
				}
			}

			if parsed.DurationMinutes != tt.expectedMinutes {
				t.Errorf("DurationMinutes = %d, want %d", parsed.DurationMinutes, tt.expectedMinutes)
			}
		})
	}
}
