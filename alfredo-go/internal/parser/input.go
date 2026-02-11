package parser

import (
	"alfredo-go/pkg/utils"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var inputPattern = regexp.MustCompile(`\s*([@#]\([^)]+\)|\S+)\s*`)
var deadlinePattern = regexp.MustCompile(`\{([^}]+)\}`)

// ParseInput tokenizes user input, keeping together elements with spaces if they are
// in parentheses and preceded by # or @
func ParseInput(input string) []string {
	matches := inputPattern.FindAllStringSubmatch(input, -1)
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		result = append(result, m[1])
	}
	return result
}

// ParsedTask holds the result of parsing a new-task input string
type ParsedTask struct {
	Content      string
	Labels       []string
	ProjectName  string // includes # prefix
	ProjectID    string
	SectionName  string
	SectionID    string
	DueDate      string
	DueString    string // raw natural language text (e.g., "tomorrow")
	DueLang      string // language for Todoist NLP (e.g., "en")
	Deadline     string // resolved deadline date (YYYY-MM-DD)
	DeadlineRaw  string // raw deadline text for NLP (e.g., "friday")
	Priority     int    // Todoist API priority (4=highest, 1=lowest)
	PrioString   string
	RawInput     string
}

// InputContext holds contextual data needed during parsing
type InputContext struct {
	AllLabels     []string        // prefixed with @
	AllProjects   []string        // prefixed with #
	LabelCounts   map[string]int  // label name (no prefix) -> count
	ProjectCounts map[string]int  // project name (no prefix) -> count
	PartialMatch  bool
	Lang          string          // system language code (e.g., "it", "de", "en")
}

// ParseNewTaskInput parses raw input for new task creation.
// Returns (parsedTask, autocompleteItems, needsExit).
// If autocompleteItems is non-nil, caller should display them and exit.
func ParseNewTaskInput(input string, ctx *InputContext) (*ParsedTask, []AutocompleteItem, bool) {
	lang := ctx.Lang

	// Extract {deadline} before tokenizing
	var deadlineRaw string
	cleanedInput := input
	if m := deadlinePattern.FindStringSubmatch(input); m != nil {
		deadlineRaw = strings.TrimSpace(m[1])
		cleanedInput = deadlinePattern.ReplaceAllString(input, "")
		cleanedInput = strings.TrimSpace(cleanedInput)
		// collapse double spaces
		for strings.Contains(cleanedInput, "  ") {
			cleanedInput = strings.ReplaceAll(cleanedInput, "  ", " ")
		}
	}

	elements := ParseInput(cleanedInput)
	utils.Log("input elements: %v", elements)

	parsed := &ParsedTask{
		Priority: 1,
		RawInput: input,
	}

	// Resolve deadline
	if deadlineRaw != "" {
		parsed.DeadlineRaw = deadlineRaw
		parsed.Deadline = resolveDeadline(deadlineRaw, lang)
	}

	var taskElements []string

	for i := 0; i < len(elements); i++ {
		item := NormalizeUnicode(elements[i])

		if strings.HasPrefix(item, "@") {
			// Handle label
			item = unwrapParens(item, "@")

			if containsStr(ctx.AllLabels, item) {
				parsed.Labels = append(parsed.Labels, item[1:])
			} else {
				// Autocomplete for labels
				subset := filterMatch(ctx.AllLabels, item, "@", ctx.PartialMatch)
				remaining := removeElement(elements, elements[i])
				remainingStr := strings.Join(remaining, " ")

				if len(subset) > 0 {
					items := make([]AutocompleteItem, 0, len(subset))
					for _, label := range subset {
						labelStr := formatWithParens(label, "@")
						var arg string
						if remainingStr != "" {
							arg = remainingStr + " " + labelStr + " "
						} else {
							arg = labelStr + " "
						}
						name := label[1:] // strip @
						count := ctx.LabelCounts[name]
						items = append(items, AutocompleteItem{
							Title:    label + " (" + itoa(count) + ")",
							Subtitle: arg,
							Arg:      arg,
							Icon:     "icons/label.png",
						})
					}
					return nil, items, true
				}
				// No matches — offer to create new label
				items := []AutocompleteItem{{
					Title:    "no labels matching, create a new label named '" + item[1:] + "'?",
					Subtitle: "press Enter to create a new label",
					Arg:      input + " ",
					Icon:     "icons/newLabel.png",
					Variables: map[string]any{
						"mySource":   "createLabel",
						"myNewLabel": item[1:],
					},
				}}
				return nil, items, true
			}

		} else if strings.HasPrefix(item, "#") {
			// Handle project
			item = unwrapParens(item, "#")

			if containsStr(ctx.AllProjects, item) {
				parsed.ProjectName = item
				// ProjectID resolution is handled by the caller
			} else {
				// Autocomplete for projects
				subset := filterMatch(ctx.AllProjects, item, "#", ctx.PartialMatch)
				remaining := removeElement(elements, elements[i])
				remainingStr := strings.Join(remaining, " ")

				if len(subset) > 0 {
					items := make([]AutocompleteItem, 0, len(subset))
					for _, proj := range subset {
						projStr := formatWithParens(proj, "#")
						var arg string
						if remainingStr != "" {
							arg = remainingStr + " " + projStr + " "
						} else {
							arg = projStr + " "
						}
						name := proj[1:] // strip #
						count := ctx.ProjectCounts[name]
						items = append(items, AutocompleteItem{
							Title:    proj + " (" + itoa(count) + ")",
							Subtitle: arg,
							Arg:      arg,
							Icon:     "icons/project.png",
						})
					}
					return nil, items, true
				}
				// No matches
				items := []AutocompleteItem{{
					Title:    "no projects matching",
					Subtitle: "try another query?",
					Arg:      "",
					Icon:     "icons/Warning.png",
				}}
				return nil, items, true
			}

		} else if strings.EqualFold(item, "p1") || strings.EqualFold(item, "p2") ||
			strings.EqualFold(item, "p3") || strings.EqualFold(item, "p4") {
			switch strings.ToLower(item) {
			case "p1":
				parsed.Priority = 4
				parsed.PrioString = "p1"
			case "p2":
				parsed.Priority = 3
				parsed.PrioString = "p2"
			case "p3":
				parsed.Priority = 2
				parsed.PrioString = "p3"
			case "p4":
				parsed.Priority = 1
				parsed.PrioString = "p4"
			}

		} else if strings.HasPrefix(item, "due:") {
			dueStr := item[4:]
			// Try single-token resolution first (coded formats + NLP with locale)
			resolved, menuItems, needsMenu := ParseDueString(dueStr, cleanedInput, lang)
			if needsMenu {
				// Before showing menu, try consuming following tokens for multi-word NLP
				// e.g., "due:next friday" is tokenized as "due:next" + "friday"
				combined := dueStr
				consumed := 0
				for j := i + 1; j < len(elements); j++ {
					candidate := combined + " " + elements[j]
					if t, ok := ParseNaturalDate(candidate, lang); ok {
						combined = candidate
						consumed = j - i
						resolved = FormatResolvedDate(t)
					} else {
						break
					}
				}
				if consumed > 0 {
					// Skip consumed tokens
					i += consumed
					parsed.DueDate = resolved
				} else {
					return nil, menuItems, true
				}
			} else {
				parsed.DueDate = resolved
			}

		} else {
			taskElements = append(taskElements, item)
		}
	}

	parsed.Content = strings.Join(taskElements, " ")

	// Inline date detection: if no explicit due: was set, try NLP on the content
	if parsed.DueDate == "" && parsed.Content != "" {
		if nlp := ParseNaturalDateInText(parsed.Content, lang); nlp != nil {
			parsed.DueDate = FormatResolvedDate(nlp.Time)
			// Strip matched date text from content
			cleaned := parsed.Content[:nlp.Start] + parsed.Content[nlp.End:]
			cleaned = strings.TrimSpace(cleaned)
			for strings.Contains(cleaned, "  ") {
				cleaned = strings.ReplaceAll(cleaned, "  ", " ")
			}
			parsed.Content = cleaned
		}
	}

	return parsed, nil, false
}

// AutocompleteItem represents an autocomplete suggestion
type AutocompleteItem struct {
	Title     string
	Subtitle  string
	Arg       string
	Icon      string
	Variables map[string]any
}

func unwrapParens(item, prefix string) string {
	if strings.HasPrefix(item, prefix+"(") && strings.HasSuffix(item, ")") && strings.Contains(item, " ") {
		item = strings.Replace(item, "(", "", 1)
		item = strings.TrimSuffix(item, ")")
		item = strings.TrimSpace(item)
	}
	return item
}

func formatWithParens(item, prefix string) string {
	name := item[len(prefix):]
	if strings.Contains(name, " ") {
		return prefix + "(" + name + ")"
	}
	return item
}

func containsStr(slice []string, s string) bool {
	normalized := NormalizeUnicode(s)
	for _, v := range slice {
		if NormalizeUnicode(v) == normalized {
			return true
		}
	}
	return false
}

func filterMatch(all []string, fragment, prefix string, partialMatch bool) []string {
	var result []string
	search := strings.ToLower(fragment)
	searchNoPrefix := strings.ToLower(fragment[len(prefix):])
	for _, item := range all {
		lower := strings.ToLower(item)
		if partialMatch {
			if strings.Contains(lower, searchNoPrefix) {
				result = append(result, item)
			}
		} else {
			if strings.Contains(lower, search) {
				result = append(result, item)
			}
		}
	}
	return result
}

func removeElement(slice []string, elem string) []string {
	result := make([]string, 0, len(slice))
	removed := false
	for _, s := range slice {
		if s == elem && !removed {
			removed = true
			continue
		}
		result = append(result, s)
	}
	return result
}

// NormalizeUnicode applies NFC normalization
func NormalizeUnicode(text string) string {
	return norm.NFC.String(strings.TrimSpace(text))
}

// resolveDeadline resolves a deadline string to YYYY-MM-DD.
// Tries coded formats first (YYYY-MM-DD, Nd, Nw, Nm), then NLP with locale.
func resolveDeadline(raw, lang string) string {
	if absDatePattern.MatchString(raw) {
		return raw
	}
	if m := relDaysPattern.FindStringSubmatch(raw); m != nil {
		days, _ := strconv.Atoi(m[1])
		return NewDate(days)
	}
	if m := relWeeksPattern.FindStringSubmatch(raw); m != nil {
		weeks, _ := strconv.Atoi(m[1])
		return NewDate(weeks * 7)
	}
	if m := relMonthsPattern.FindStringSubmatch(raw); m != nil {
		months, _ := strconv.Atoi(m[1])
		return NewDate(months * 30)
	}
	// Try NLP with locale
	if t, ok := ParseNaturalDate(raw, lang); ok {
		return t.Format("2006-01-02")
	}
	// Unrecognized — return empty so it's silently ignored
	return ""
}

// containsLetter returns true if the string contains at least one Unicode letter.
func containsLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
