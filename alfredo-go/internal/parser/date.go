package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
// Returns (resolvedDate, menuItems, needsMenu).
// If needsMenu is true, the caller should display menuItems and exit.
func ParseDueString(dueStr, fullInput string) (string, []AutocompleteItem, bool) {
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

// BuildRescheduleMenu builds a reschedule date picker menu
func BuildRescheduleMenu(customDays, taskContent string) []AutocompleteItem {
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
	today := time.Now().Truncate(24 * time.Hour)
	daysTo := int(d.Sub(today).Hours() / 24)
	formatted := d.Format("Monday, January 02, 2006")
	return daysTo, formatted
}

// HandleINTDateHour parses an ISO datetime and returns days until that date and a formatted string
func HandleINTDateHour(dateString string) (int, string) {
	d, err := time.Parse("2006-01-02T15:04", dateString)
	if err != nil {
		return 0, dateString
	}
	today := time.Now().Truncate(24 * time.Hour)
	daysTo := int(d.Sub(today).Hours() / 24)
	formatted := d.Format("Monday, January 02, 2006, 15:04")
	return daysTo, formatted
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
