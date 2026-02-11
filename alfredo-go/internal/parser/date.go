package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

var (
	// Matches: Nd, Nw, Nm, or Nd with time (e.g., 7d13:30)
	relDaysPattern    = regexp.MustCompile(`^(\d+)d$`)
	relDaysTimePattern = regexp.MustCompile(`^(\d+)d(\d{2}:\d{2})$`)
	relWeeksPattern   = regexp.MustCompile(`^(\d+)w$`)
	relMonthsPattern  = regexp.MustCompile(`^(\d+)m$`)
	absDatePattern    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	absDateTimePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$`)
	// Menu pattern: number with optional w/m, optional time
	menuPattern = regexp.MustCompile(`^(\d+)([wm]?)(?:(\d{2}):(\d{2}))?$`)
)

// ParseDueString resolves a due string from user input.
// lang is the system language code (e.g., "it", "de") for locale-aware NLP.
// Returns (resolvedDate, menuItems, needsMenu).
// If needsMenu is true, the caller should display menuItems and exit.
func ParseDueString(dueStr, fullInput, lang string) (string, []AutocompleteItem, bool) {
	if m := relDaysPattern.FindStringSubmatch(dueStr); m != nil {
		days, _ := strconv.Atoi(m[1])
		return NewDate(days), nil, false
	}
	if m := relDaysTimePattern.FindStringSubmatch(dueStr); m != nil {
		days, _ := strconv.Atoi(m[1])
		timeStr := m[2]
		if ValidateTime(timeStr) {
			return NewDate(days) + "T" + timeStr, nil, false
		}
	}
	if m := relWeeksPattern.FindStringSubmatch(dueStr); m != nil {
		weeks, _ := strconv.Atoi(m[1])
		return NewDate(weeks * 7), nil, false
	}
	if m := relMonthsPattern.FindStringSubmatch(dueStr); m != nil {
		months, _ := strconv.Atoi(m[1])
		return NewDate(months * 30), nil, false
	}
	if absDatePattern.MatchString(dueStr) {
		return dueStr, nil, false
	}
	if absDateTimePattern.MatchString(dueStr) {
		return dueStr, nil, false
	}

	// Try natural language parsing before showing menu
	if t, ok := ParseNaturalDate(dueStr, lang); ok {
		return FormatResolvedDate(t), nil, false
	}

	// Could not resolve — show due date menu
	items := BuildDueMenu(dueStr, fullInput)
	return "", items, true
}

// BuildDueMenu builds a date picker menu for the given input
func BuildDueMenu(customDays, inputThrough string) []AutocompleteItem {
	// Remove the due: value from the passthrough text
	duePat := regexp.MustCompile(`(?:due:)\d*[wm]?(?:(\d{2}):(\d{2}))?`)
	inputThroughF := duePat.ReplaceAllString(inputThrough, "")
	inputThroughF = strings.TrimSpace(inputThroughF)
	if inputThroughF != "" {
		inputThroughF += " "
	}

	var items []AutocompleteItem

	if customDays == "" {
		items = []AutocompleteItem{
			{Title: fmt.Sprintf("Due today 🗓️ %s 🔥", NewDateFormatted(0)), Arg: inputThroughF + "due:0d ", Icon: "icons/today.png"},
			{Title: fmt.Sprintf("Due tomorrow 🗓️ %s 🧨", NewDateFormatted(1)), Arg: inputThroughF + "due:1d ", Icon: "icons/today.png"},
			{Title: fmt.Sprintf("Due in a week 🗓️ %s 🍹", NewDateFormatted(7)), Arg: inputThroughF + "due:7d ", Icon: "icons/today.png"},
			{Title: fmt.Sprintf("Due in a month 🗓️ %s 🏖️", NewDateFormatted(30)), Arg: inputThroughF + "due:30d ", Icon: "icons/today.png"},
		}
		return items
	}

	if m := menuPattern.FindStringSubmatch(customDays); m != nil {
		numStr, letter := m[1], m[2]
		hours, minutes := m[3], m[4]

		num, _ := strconv.Atoi(numStr)
		var timeString, timeStringSF string
		if hours != "" && minutes != "" {
			timeString = fmt.Sprintf(", %s:%s", hours, minutes)
			timeStringSF = fmt.Sprintf("%s:%s", hours, minutes)
		}

		if letter == "w" {
			num *= 7
		} else if letter == "m" {
			num *= 30
		}

		dayString := "days"
		if num == 1 {
			dayString = "day"
		}

		items = append(items, AutocompleteItem{
			Title: fmt.Sprintf("Due in %d %s 🗓️ %s%s", num, dayString, NewDateFormatted(num), timeString),
			Arg:   fmt.Sprintf("%sdue:%dd%s ", inputThroughF, num, timeStringSF),
			Icon:  "icons/today.png",
		})
		return items
	}

	// Invalid format
	items = append(items, AutocompleteItem{
		Title:    "Invalid format!",
		Subtitle: "enter an integer (days) or add 'w' (weeks) or 'm' (months). Optional: time in 24h format",
		Arg:      "",
		Icon:     "icons/warning.png",
	})
	return items
}

// BuildRescheduleMenu builds a reschedule date picker menu.
// lang is the system language code for locale-aware NLP.
func BuildRescheduleMenu(customDays, taskContent, lang string) []AutocompleteItem {
	var items []AutocompleteItem

	if customDays == "" {
		items = []AutocompleteItem{
			{Title: fmt.Sprintf("Reschedule to today 🗓️ %s 🔥", NewDateFormatted(0)), Subtitle: taskContent, Arg: "0", Icon: "icons/today.png"},
			{Title: fmt.Sprintf("Reschedule to tomorrow 🗓️ %s 🧨", NewDateFormatted(1)), Subtitle: taskContent, Arg: "1", Icon: "icons/today.png"},
			{Title: fmt.Sprintf("Reschedule in a week 🗓️ %s 🍹", NewDateFormatted(7)), Subtitle: taskContent, Arg: "7", Icon: "icons/today.png"},
			{Title: fmt.Sprintf("Reschedule in a month 🗓️ %s 🏖️", NewDateFormatted(30)), Subtitle: taskContent, Arg: "30", Icon: "icons/today.png"},
		}
		return items
	}

	if m := menuPattern.FindStringSubmatch(customDays); m != nil {
		numStr, letter := m[1], m[2]
		hours, minutes := m[3], m[4]

		num, _ := strconv.Atoi(numStr)
		var timeString, timeStringSF string
		if hours != "" && minutes != "" {
			timeString = fmt.Sprintf(", %s:%s", hours, minutes)
			timeStringSF = fmt.Sprintf("%s:%s", hours, minutes)
		}

		if letter == "w" {
			num *= 7
		} else if letter == "m" {
			num *= 30
		}

		dayString := "days"
		if num == 1 {
			dayString = "day"
		}

		items = append(items, AutocompleteItem{
			Title:    fmt.Sprintf("Reschedule in %d %s 🗓️ %s%s", num, dayString, NewDateFormatted(num), timeString),
			Subtitle: taskContent,
			Arg:      fmt.Sprintf("%dd%s", num, timeStringSF),
			Icon:     "icons/today.png",
		})
		return items
	}

	if absDatePattern.MatchString(customDays) {
		daysTo, dateF := HandleINTDate(customDays)
		dayString := "days"
		if daysTo == 1 || daysTo == -1 {
			dayString = "day"
		}
		items = append(items, AutocompleteItem{
			Title:    fmt.Sprintf("Reschedule in %d %s 🗓️ %s", daysTo, dayString, dateF),
			Subtitle: taskContent,
			Arg:      customDays,
			Icon:     "icons/today.png",
		})
		return items
	}

	if absDateTimePattern.MatchString(customDays) {
		daysTo, dateF := HandleINTDateHour(customDays)
		dayString := "days"
		if daysTo == 1 || daysTo == -1 {
			dayString = "day"
		}
		items = append(items, AutocompleteItem{
			Title:    fmt.Sprintf("Reschedule in %d %s 🗓️ %s", daysTo, dayString, dateF),
			Subtitle: taskContent,
			Arg:      customDays,
			Icon:     "icons/today.png",
		})
		return items
	}

	// Try natural language parsing before showing invalid
	if t, ok := ParseNaturalDate(customDays, lang); ok {
		daysTo := daysFromToday(t)
		dayString := "days"
		if daysTo == 1 || daysTo == -1 {
			dayString = "day"
		}
		dateF := t.Format("Monday, January 02, 2006")
		if t.Hour() != 0 || t.Minute() != 0 {
			dateF = t.Format("Monday, January 02, 2006, 15:04")
		}
		resolved := FormatResolvedDate(t)
		items = append(items, AutocompleteItem{
			Title:    fmt.Sprintf("Reschedule in %d %s 🗓️ %s", daysTo, dayString, dateF),
			Subtitle: taskContent,
			Arg:      resolved,
			Icon:     "icons/today.png",
		})
		return items
	}

	// Invalid format
	items = append(items, AutocompleteItem{
		Title:    "Invalid format!",
		Subtitle: "enter an integer (days) or add 'w' (weeks) or 'm' (months). Optional: time in 24h format",
		Arg:      "",
		Icon:     "icons/warning.png",
	})
	return items
}

// ResolveRescheduleDate converts a reschedule input (days string from menu) into an API date
func ResolveRescheduleDate(input string) string {
	if strings.Contains(input, "-") && !strings.Contains(input, "d") {
		// Full ISO date or datetime
		return input
	}
	if strings.Contains(input, "d") {
		parts := strings.SplitN(input, "d", 2)
		days, _ := strconv.Atoi(parts[0])
		newDate := NewDate(days)
		if len(parts) > 1 && parts[1] != "" {
			newDate += "T" + parts[1]
		}
		return newDate
	}
	// Plain number of days
	days, err := strconv.Atoi(input)
	if err != nil {
		return input
	}
	return NewDate(days)
}

// NewDate returns an ISO date string N days from today
func NewDate(days int) string {
	return time.Now().AddDate(0, 0, days).Format("2006-01-02")
}

// NewDateFormatted returns a human-readable date string N days from today
func NewDateFormatted(days int) string {
	return time.Now().AddDate(0, 0, days).Format("Monday, January 02, 2006")
}

// HandleINTDate parses an ISO date and returns days until that date and a formatted string
func HandleINTDate(dateString string) (int, string) {
	d, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return 0, dateString
	}
	daysTo := daysFromToday(d)
	formatted := d.Format("Monday, January 02, 2006")
	return daysTo, formatted
}

// HandleINTDateHour parses an ISO datetime and returns days until that date and a formatted string
func HandleINTDateHour(dateString string) (int, string) {
	d, err := time.Parse("2006-01-02T15:04", dateString)
	if err != nil {
		return 0, dateString
	}
	daysTo := daysFromToday(d)
	formatted := d.Format("Monday, January 02, 2006, 15:04")
	return daysTo, formatted
}

// daysFromToday calculates the number of days between today (local) and the given time.
// Uses calendar date difference to avoid timezone/truncation issues.
func daysFromToday(t time.Time) int {
	now := time.Now()
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	targetDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return int(targetDate.Sub(todayDate).Hours() / 24)
}

// newWhenParser creates a configured when parser for English NLP dates
func newWhenParser() *when.Parser {
	w := when.New(nil)
	w.Add(en.All...)
	w.Add(common.All...)
	return w
}

// NLPResult holds the result of natural language date parsing within text
type NLPResult struct {
	Time  time.Time
	Text  string // the matched portion of input (e.g., "tomorrow", "next friday at 3pm")
	Start int    // start index in input
	End   int    // end index in input
}

// ParseNaturalDate tries to parse a natural language date string.
// Tries English NLP (olebedev/when) first, then locale-specific keywords.
// lang is the system language code (e.g., "it", "de", "en").
func ParseNaturalDate(input, lang string) (time.Time, bool) {
	if input == "" {
		return time.Time{}, false
	}
	// Try English NLP first (works for all locales since English is widely understood)
	w := newWhenParser()
	r, err := w.Parse(input, time.Now())
	if err == nil && r != nil {
		t := r.Time
		// If the input has no explicit time indicator, strip hours/minutes
		// (e.g., "today" returns current time, but we want date-only)
		if !hasTimeIndicator(input) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		return t, true
	}
	// Try locale-specific keywords
	return ParseLocaleDateKeyword(input, lang)
}

// ParseNaturalDateInText finds a natural language date within a larger text.
// Returns the result with match position info, or nil if no date found.
// lang is the system language code for locale-aware parsing.
func ParseNaturalDateInText(input, lang string) *NLPResult {
	if input == "" {
		return nil
	}
	// Try English NLP first
	w := newWhenParser()
	r, err := w.Parse(input, time.Now())
	if err == nil && r != nil {
		t := r.Time
		if !hasTimeIndicator(r.Text) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		}
		return &NLPResult{
			Time:  t,
			Text:  strings.TrimSpace(r.Text),
			Start: r.Index,
			End:   r.Index + len(r.Text),
		}
	}
	// Try locale keywords on individual words
	if lang != "" && lang != "en" {
		words := strings.Fields(input)
		for _, word := range words {
			if t, ok := ParseLocaleDateKeyword(word, lang); ok {
				idx := strings.Index(input, word)
				return &NLPResult{
					Time:  t,
					Text:  word,
					Start: idx,
					End:   idx + len(word),
				}
			}
		}
	}
	return nil
}

// hasTimeIndicator checks if the input text contains explicit time markers
// (e.g., "at 3pm", "at 14:00", "3pm", "15:00").
var timeIndicatorPattern = regexp.MustCompile(`(?i)\b(at\s+)?\d{1,2}(:\d{2})?\s*(am|pm)\b|\b\d{1,2}:\d{2}\b`)

func hasTimeIndicator(input string) bool {
	return timeIndicatorPattern.MatchString(input)
}

// FormatNaturalDate formats a time for human-readable display
func FormatNaturalDate(t time.Time) string {
	if t.Hour() == 0 && t.Minute() == 0 {
		return t.Format("Monday, January 02, 2006")
	}
	return t.Format("Monday, January 02, 2006, 15:04")
}

// FormatResolvedDate formats a time as ISO date or datetime for the API
func FormatResolvedDate(t time.Time) string {
	if t.Hour() == 0 && t.Minute() == 0 {
		return t.Format("2006-01-02")
	}
	return t.Format("2006-01-02T15:04")
}

// IsNaturalLanguageDate checks if the input is a natural language date (not coded format).
// lang is the system language code for locale-aware parsing.
func IsNaturalLanguageDate(input, lang string) bool {
	if relDaysPattern.MatchString(input) || relDaysTimePattern.MatchString(input) ||
		relWeeksPattern.MatchString(input) || relMonthsPattern.MatchString(input) ||
		absDatePattern.MatchString(input) || absDateTimePattern.MatchString(input) {
		return false
	}
	_, ok := ParseNaturalDate(input, lang)
	return ok
}

// ValidateTime checks if a time string is valid HH:MM in 24h format
func ValidateTime(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return h >= 0 && h < 24 && m >= 0 && m < 60
}
