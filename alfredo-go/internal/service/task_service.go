package service

import (
	"alfredo-go/internal/parser"
	"alfredo-go/pkg/alfred"
	"alfredo-go/pkg/cache"
	"alfredo-go/pkg/config"
	"alfredo-go/pkg/todoist"
	"alfredo-go/pkg/utils"
	"fmt"
	"sort"
	"strings"
	"time"
)

// TaskService handles task-related operations
type TaskService struct {
	client *todoist.Client
	cache  *cache.Cache
	cfg    *config.Config
}

// NewTaskService creates a new TaskService
func NewTaskService(client *todoist.Client, c *cache.Cache, cfg *config.Config) *TaskService {
	return &TaskService{
		client: client,
		cache:  c,
		cfg:    cfg,
	}
}

// QueryTasks is the main query function, porting alfredo-query.py logic
func (s *TaskService) QueryTasks(mode, input string) (*alfred.Output, error) {
	if err := s.cache.EnsureFresh(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	data := s.cache.Data()
	now := time.Now()
	today := now.Format("2006-01-02")
	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Build goals string
	goalsString := ""
	if s.cfg.ShowGoals && data.Stats != nil && data.User != nil {
		// Goals come from the user object, not stats
		dailyGoal := data.User.DailyGoal
		weeklyGoal := data.User.WeeklyGoal

		var soFarCompleted int
		for _, d := range data.Stats.DaysItems {
			if d.Date == today {
				soFarCompleted = d.TotalCompleted
				break
			}
		}

		var totalWeekCompleted int
		if len(data.Stats.WeekItems) > 0 {
			totalWeekCompleted = data.Stats.WeekItems[0].TotalCompleted
		}

		statusDay := "❌"
		if soFarCompleted >= dailyGoal {
			statusDay = "✅"
		}
		statusWeek := "❌"
		if totalWeekCompleted >= weeklyGoal {
			statusWeek = "✅"
		}
		goalsString = fmt.Sprintf(" Daily: %d/%d %s Weekly: %d/%d %s ",
			soFarCompleted, dailyGoal, statusDay, totalWeekCompleted, weeklyGoal, statusWeek)
	}

	// Subset tasks based on mode
	var toShow []todoist.Task
	var icon string

	switch mode {
	case "today":
		seen := map[string]bool{}
		for _, t := range data.Tasks {
			if t.Due != nil && strings.Split(t.Due.Date, "T")[0] == today {
				toShow = append(toShow, t)
				seen[t.ID] = true
			}
		}
		for _, t := range data.Tasks {
			if !seen[t.ID] && t.Deadline != nil && t.Deadline.Date == today {
				toShow = append(toShow, t)
			}
		}
		sort.Slice(toShow, func(i, j int) bool {
			di := ""
			if toShow[i].Due != nil {
				di = toShow[i].Due.Date
			}
			dj := ""
			if toShow[j].Due != nil {
				dj = toShow[j].Due.Date
			}
			return di < dj
		})
		icon = "icons/today.png"

	case "due":
		seen := map[string]bool{}
		for _, t := range data.Tasks {
			if t.Due != nil && t.Due.Date < today {
				toShow = append(toShow, t)
				seen[t.ID] = true
			}
		}
		for _, t := range data.Tasks {
			if !seen[t.ID] && t.Deadline != nil && t.Deadline.Date < today {
				toShow = append(toShow, t)
			}
		}
		sort.Slice(toShow, func(i, j int) bool {
			di := ""
			if toShow[i].Due != nil {
				di = toShow[i].Due.Date
			}
			dj := ""
			if toShow[j].Due != nil {
				dj = toShow[j].Due.Date
			}
			return di < dj
		})
		icon = "icons/overdue.png"

	case "all":
		toShow = make([]todoist.Task, len(data.Tasks))
		copy(toShow, data.Tasks)
		sort.Slice(toShow, func(i, j int) bool {
			di := "9999-12-31"
			dj := "9999-12-31"
			if toShow[i].Due != nil {
				di = toShow[i].Due.Date
			}
			if toShow[j].Due != nil {
				dj = toShow[j].Due.Date
			}
			return di < dj
		})
		icon = "icons/bullet.png"

	case "deadline":
		for _, t := range data.Tasks {
			if t.Deadline != nil && t.Deadline.Date != "" {
				toShow = append(toShow, t)
			}
		}
		sort.Slice(toShow, func(i, j int) bool {
			return toShow[i].Deadline.Date < toShow[j].Deadline.Date
		})
		icon = "icons/deadline.png"
	}

	// Get counts from subset
	_, labelsAll := cache.FetchLabelsFromSubset(toShow)
	_, projectsAll := cache.FetchProjectsFromSubset(toShow, data.Projects, data.Sections)

	// Parse input
	inputItems := parser.ParseInput(input)
	var filterLabels, filterProjects, filterSections, searchStrings []string
	finalInput := make([]string, len(inputItems))
	copy(finalInput, inputItems)

	labelFlag := false
	projectFlag := false
	var tagFrag, projFrag string

	output := &alfred.Output{Items: []alfred.OutputItem{}}

	for _, item := range inputItems {
		item = parser.NormalizeUnicode(item)

		if strings.HasPrefix(item, "@") {
			// Unwrap parentheses
			cleaned := unwrapParens(item, "@")

			if containsStr(labelsAll, cleaned) {
				filterLabels = append(filterLabels, cleaned[1:])
			} else {
				labelFlag = true
				tagFrag = cleaned
				finalInput = removeElement(finalInput, item)
			}

		} else if strings.HasPrefix(item, "#") {
			cleaned := unwrapParens(item, "#")

			if containsStr(projectsAll, cleaned) {
				if strings.Contains(cleaned, "/") {
					parts := strings.SplitN(cleaned, "/", 2)
					projID := getProjectID(data.Projects, parts[0][1:])
					sectID := getSectionID(data.Projects, data.Sections, cleaned)
					filterProjects = append(filterProjects, projID)
					filterSections = append(filterSections, sectID)
				} else {
					projID := getProjectID(data.Projects, cleaned[1:])
					filterProjects = append(filterProjects, projID)
				}
			} else {
				projectFlag = true
				projFrag = cleaned
				finalInput = removeElement(finalInput, item)
			}

		} else {
			searchStrings = append(searchStrings, item)
		}
	}

	myInput := strings.Join(finalInput, " ")

	// Apply filters
	toShow = filterTasks(toShow, filterLabels, filterProjects, filterSections, searchStrings)

	// Label autocomplete
	if labelFlag {
		labelCounts, labels := cache.FetchLabelsFromSubset(toShow)
		var subset []string
		if s.cfg.PartialMatch {
			for _, l := range labels {
				if strings.Contains(strings.ToLower(l), strings.ToLower(tagFrag[1:])) {
					subset = append(subset, l)
				}
			}
		} else {
			for _, l := range labels {
				if strings.Contains(strings.ToLower(l), strings.ToLower(tagFrag)) {
					subset = append(subset, l)
				}
			}
		}

		if len(subset) > 0 {
			for _, label := range subset {
				labelStr := formatWithParens(label, "@")
				var arg string
				if myInput != "" {
					arg = myInput + " " + labelStr + " "
				} else {
					arg = labelStr + " "
				}
				output.Items = append(output.Items, alfred.OutputItem{
					Title:    fmt.Sprintf("%s (%d)", label, labelCounts[label[1:]]),
					Subtitle: myInput,
					Arg:      "",
					Variables: map[string]any{
						"myIter": true,
						"myArg":  arg,
						"myMode": mode,
					},
					Icon: &alfred.Icon{Path: "icons/label.png"},
				})
			}
		} else {
			output.Items = append(output.Items, alfred.OutputItem{
				Title:    "no labels matching",
				Subtitle: "try another query?",
				Arg:      "",
				Variables: map[string]any{
					"myIter": true,
					"myArg":  myInput + " ",
					"myMode": mode,
				},
				Icon: &alfred.Icon{Path: "icons/Warning.png"},
			})
		}
		return output, nil
	}

	// Project autocomplete
	if projectFlag {
		projectCounts, projectList := cache.FetchProjectsFromSubset(toShow, data.Projects, data.Sections)
		var subset []string
		if s.cfg.PartialMatch {
			for _, p := range projectList {
				if strings.Contains(strings.ToLower(p), strings.ToLower(projFrag[1:])) {
					subset = append(subset, p)
				}
			}
		} else {
			for _, p := range projectList {
				if strings.Contains(strings.ToLower(p), strings.ToLower(projFrag)) {
					subset = append(subset, p)
				}
			}
		}

		if len(subset) > 0 {
			for _, proj := range subset {
				projStr := formatWithParens(proj, "#")
				var arg string
				if myInput != "" {
					arg = myInput + " " + projStr + " "
				} else {
					arg = projStr + " "
				}
				output.Items = append(output.Items, alfred.OutputItem{
					Title:    fmt.Sprintf("%s (%d)", proj, projectCounts[proj[1:]]),
					Subtitle: myInput,
					Arg:      "",
					Variables: map[string]any{
						"myIter": true,
						"myArg":  arg,
						"myMode": mode,
					},
					Icon: &alfred.Icon{Path: "icons/project.png"},
				})
			}
		} else {
			output.Items = append(output.Items, alfred.OutputItem{
				Title:    "no projects matching",
				Subtitle: "try another query?",
				Arg:      "",
				Variables: map[string]any{
					"myIter": true,
					"myArg":  myInput + " ",
					"myMode": mode,
				},
				Icon: &alfred.Icon{Path: "icons/Warning.png"},
			})
		}
		return output, nil
	}

	// Build task output
	if len(toShow) > 0 {
		matchCount := len(toShow)
		countR := 1

		for _, task := range toShow {
			dueString := ""
			if task.Due != nil {
				dueDate := parseDueDate(task.Due.Date)
				if !dueDate.IsZero() {
					dueDays := int(todayDate.Sub(dueDate).Hours() / 24)
					dayWord := "days"
					if abs(dueDays) == 1 {
						dayWord = "day"
					}
					if dueDays == 0 {
						dueString = "DUE TODAY"
					} else if dueDays < 0 {
						dueString = fmt.Sprintf("due in %d %s ⚠️", abs(dueDays), dayWord)
					} else {
						dueString = fmt.Sprintf("%d %s overdue❗", dueDays, dayWord)
					}
				}
			}

			deadlineString := ""
			if task.Deadline != nil && task.Deadline.Date != "" {
				dlDate := parseDueDate(task.Deadline.Date)
				if !dlDate.IsZero() {
					dlDays := int(todayDate.Sub(dlDate).Hours() / 24)
					dayWord := "days"
					if abs(dlDays) == 1 {
						dayWord = "day"
					}
					if dlDays == 0 {
						deadlineString = "🎯 DEADLINE TODAY"
					} else if dlDays < 0 {
						deadlineString = fmt.Sprintf("🎯 deadline in %d %s", abs(dlDays), dayWord)
					} else {
						deadlineString = fmt.Sprintf("🎯 %d %s past deadline❗", dlDays, dayWord)
					}
				}
			}

			labelsString := ""
			if len(task.Labels) > 0 {
				labelsString = "🏷️ " + strings.Join(task.Labels, ",")
			}

			projectName := getProjectName(data.Projects, task.ProjectID)

			dateInfo := dueString
			if mode == "deadline" {
				dateInfo = deadlineString
			}

			title := fmt.Sprintf("%s (#%s) %s", task.Content, projectName, dateInfo)
			subtitleDeadline := ""
			if mode != "deadline" && deadlineString != "" {
				subtitleDeadline = " " + deadlineString
			}
			subtitle := fmt.Sprintf("%d/%d.%s%s%s", countR, matchCount, goalsString, labelsString, subtitleDeadline)

			// Build reconstructed input string for edit mode
			editArg := reconstructEditInput(task, data.Projects)

			output.Items = append(output.Items, alfred.OutputItem{
				Title:    title,
				Subtitle: subtitle,
				Arg:      "",
				Variables: map[string]any{
					"myIter":        false,
					"myURL":         fmt.Sprintf("https://app.todoist.com/app/task/%s", task.ID),
					"myAppURL":      fmt.Sprintf("todoist://task?id=%s", task.ID),
					"myTaskID":      task.ID,
					"myTaskContent": task.Content,
					"myArg":         input,
					"myMode":        mode,
				},
				Mods: map[string]alfred.ModsItem{
					"alt": {
						Arg:      editArg,
						Subtitle: "Edit this task ✏️",
						Variables: map[string]any{
							"myTaskID": task.ID,
						},
					},
					"cmd+ctrl+alt": {
						Subtitle: "Delete this task 🗑️",
					},
				},
				Icon: &alfred.Icon{Path: icon},
			})
			countR++
		}
	} else if len(searchStrings) > 0 || len(filterLabels) > 0 {
		output.Items = append(output.Items, alfred.OutputItem{
			Title:    "no tasks matching your query 🙁",
			Subtitle: "",
			Arg:      "",
			Variables: map[string]any{
				"myIter": true,
				"myArg":  "",
				"myMode": mode,
			},
			Mods: map[string]alfred.ModsItem{
				"shift": {Arg: "", Subtitle: "nothing to see here"},
				"cmd":   {Arg: "", Subtitle: "nothing to see here"},
				"ctrl":  {Arg: "", Subtitle: "nothing to see here"},
				"alt":   {Arg: "", Subtitle: "nothing to see here"},
			},
		})
	} else {
		emptyTitle := "no tasks left to do today! 🙌"
		switch mode {
		case "due":
			emptyTitle = "no overdue tasks! 🙌"
		case "deadline":
			emptyTitle = "no tasks with a deadline set"
		case "all":
			emptyTitle = "no tasks found"
		}
		output.Items = append(output.Items, alfred.OutputItem{
			Title:    emptyTitle,
			Subtitle: goalsString,
			Arg:      "",
			Mods: map[string]alfred.ModsItem{
				"shift": {Arg: "", Subtitle: "nothing to see here"},
			},
		})
	}

	return output, nil
}

// ParseNewTask handles parse command
func (s *TaskService) ParseNewTask(input string) (*alfred.Output, error) {
	if err := s.cache.EnsureFresh(); err != nil {
		return nil, err
	}

	data := s.cache.Data()

	// Load counts
	labelCounts, err := s.cache.LoadLabelCounts()
	if err != nil {
		labelCounts = make(map[string]int)
	}
	projectCounts, err := s.cache.LoadProjectCounts()
	if err != nil {
		projectCounts = make(map[string]int)
	}

	// Build label list with @ prefix
	allLabels := make([]string, 0, len(labelCounts))
	for name := range labelCounts {
		allLabels = append(allLabels, "@"+name)
	}

	// Build project list with # prefix (NFC-normalized)
	allProjects := make([]string, 0, len(projectCounts))
	for name := range projectCounts {
		allProjects = append(allProjects, "#"+parser.NormalizeUnicode(name))
	}

	ctx := &parser.InputContext{
		AllLabels:     allLabels,
		AllProjects:   allProjects,
		LabelCounts:   labelCounts,
		ProjectCounts: projectCounts,
		PartialMatch:  s.cfg.PartialMatch,
		Lang:          s.cfg.DueLang,
	}

	parsed, autocomplete, needsExit := parser.ParseNewTaskInput(input, ctx)

	output := &alfred.Output{Items: []alfred.OutputItem{}}

	if needsExit {
		for _, ac := range autocomplete {
			item := alfred.OutputItem{
				Title:    ac.Title,
				Subtitle: ac.Subtitle,
				Arg:      ac.Arg,
				Icon:     &alfred.Icon{Path: ac.Icon},
			}
			if ac.Variables != nil {
				item.Variables = ac.Variables
			}
			output.Items = append(output.Items, item)
		}
		return output, nil
	}

	// Resolve project ID
	if parsed.ProjectName != "" {
		projName := parsed.ProjectName
		if strings.Contains(projName, "/") {
			parts := strings.SplitN(projName, "/", 2)
			parsed.ProjectID = getProjectID(data.Projects, parts[0][1:])
			parsed.SectionID = getSectionID(data.Projects, data.Sections, projName)
			parsed.SectionName = parts[1]
		} else {
			parsed.ProjectID = getProjectID(data.Projects, projName[1:])
		}
	} else {
		parsed.ProjectName = "#Inbox"
		parsed.ProjectID = getProjectID(data.Projects, "Inbox")
	}

	// Build preview
	tagString := strings.Join(parsed.Labels, ",,..,,")
	var tagStringF string
	if len(parsed.Labels) > 0 {
		tagStringF = "🏷️" + strings.Join(parsed.Labels, ",")
	}

	var dueStringF string
	if parsed.DueDate != "" {
		dueStringF = "🗓️ due:" + parsed.DueDate
	}

	var sectStringF string
	if parsed.SectionName != "" {
		sectStringF = "🧩 section:" + parsed.SectionName
	}

	var prioStringF string
	switch parsed.PrioString {
	case "p1":
		prioStringF = "p1️⃣"
	case "p2":
		prioStringF = "p2️⃣"
	case "p3":
		prioStringF = "p3️⃣"
	}

	var deadlineStringF string
	if parsed.Deadline != "" {
		deadlineStringF = "🎯 deadline:" + parsed.Deadline
	}

	projStringF := "📋" + parsed.ProjectName

	subtitle := fmt.Sprintf("%s %s %s %s %s %s ⇧↩️ to create",
		projStringF, sectStringF, tagStringF, prioStringF, dueStringF, deadlineStringF)

	output.Items = append(output.Items, alfred.OutputItem{
		Title:    parsed.Content,
		Subtitle: subtitle,
		Arg:      input,
		Variables: map[string]any{
			"myTaskText":    parsed.Content,
			"myTagString":   tagString,
			"myProjectID":   parsed.ProjectID,
			"mySectionID":   parsed.SectionID,
			"myDueDate":     parsed.DueDate,
			"myDueString":   parsed.DueString,
			"myDueLang":     parsed.DueLang,
			"myDeadline":    parsed.Deadline,
			"myDeadlineRaw": parsed.DeadlineRaw,
			"myPriority":    parsed.Priority,
		},
		Icon: &alfred.Icon{Path: "icons/newTask.png"},
	})

	return output, nil
}

// CompleteTask completes a task and refreshes cache
func (s *TaskService) CompleteTask(taskID string) error {
	if err := s.client.CompleteTask(taskID); err != nil {
		return err
	}
	// Refresh cache in background (best effort)
	if err := s.cache.Refresh(); err != nil {
		utils.Log("warning: cache refresh failed: %v", err)
	}
	return nil
}

// DeleteTask deletes a task and refreshes cache
func (s *TaskService) DeleteTask(taskID string) error {
	if err := s.client.DeleteTask(taskID); err != nil {
		return err
	}
	// Refresh cache in background (best effort)
	if err := s.cache.Refresh(); err != nil {
		utils.Log("warning: cache refresh failed: %v", err)
	}
	return nil
}

// CreateTask creates a new task via the API
func (s *TaskService) CreateTask(content, labelsStr, projectID, sectionID, dueDate, dueString, dueLang string, priority int, deadline, deadlineLang string) error {
	var labels []string
	if labelsStr != "" {
		labels = strings.Split(labelsStr, ",,..,,")
	}

	var dl *todoist.Deadline
	if deadline != "" {
		lang := deadlineLang
		if lang == "" {
			lang = s.cfg.DueLang
		}
		dl = &todoist.Deadline{Date: deadline, Lang: lang}
	}

	if dueLang == "" && dueString != "" {
		dueLang = s.cfg.DueLang
	}

	// Build description from TASK_STAMP template
	description := ""
	if s.cfg.TaskStamp != "" {
		description = strings.ReplaceAll(s.cfg.TaskStamp, "{timestamp}",
			time.Now().Format("Monday, January 2, 2006, 3:04:05 pm"))
	}

	if err := s.client.CreateTask(content, labels, projectID, sectionID, dueDate, dueString, dueLang, priority, dl, description); err != nil {
		return err
	}

	// Refresh cache
	if err := s.cache.Refresh(); err != nil {
		utils.Log("warning: cache refresh failed: %v", err)
	}
	return nil
}

// CreateLabel creates a label and updates the counts file
func (s *TaskService) CreateLabel(name string) error {
	// Check if label already exists
	counts, err := s.cache.LoadLabelCounts()
	if err == nil {
		if _, exists := counts[name]; exists {
			return nil // Already exists
		}
	}

	if err := s.client.CreateLabel(name); err != nil {
		return err
	}

	// Update label counts file
	if counts == nil {
		counts = make(map[string]int)
	}
	counts[name] = 0
	return s.cache.SaveLabelCounts(counts)
}

// RescheduleTask reschedules a task to a new date
func (s *TaskService) RescheduleTask(taskID, dateInput string) error {
	newDate := parser.ResolveRescheduleDate(dateInput)
	utils.Log("rescheduling task %s to %s", taskID, newDate)

	updates := map[string]any{
		"due": map[string]string{"date": newDate},
	}

	if err := s.client.UpdateTask(taskID, updates); err != nil {
		return err
	}

	if err := s.cache.Refresh(); err != nil {
		utils.Log("warning: cache refresh failed: %v", err)
	}
	return nil
}

// BuildRescheduleMenu builds the reschedule date menu
func (s *TaskService) BuildRescheduleMenu(customDays, taskContent string) *alfred.Output {
	items := parser.BuildRescheduleMenu(customDays, taskContent, s.cfg.DueLang)
	output := &alfred.Output{Items: make([]alfred.OutputItem, 0, len(items))}
	for _, item := range items {
		output.Items = append(output.Items, alfred.OutputItem{
			Title:    item.Title,
			Subtitle: item.Subtitle,
			Arg:      item.Arg,
			Icon:     &alfred.Icon{Path: item.Icon},
		})
	}
	return output
}

// ForceRebuild forces a cache refresh
func (s *TaskService) ForceRebuild() (*alfred.Output, error) {
	if err := s.cache.Refresh(); err != nil {
		return nil, err
	}
	return &alfred.Output{
		Items: []alfred.OutputItem{{
			Title:    "Done!",
			Subtitle: "ready to use AlfreDo now ✅",
			Arg:      "",
			Icon:     &alfred.Icon{Path: "icons/done.png"},
		}},
	}, nil
}

// GetStats returns completion statistics from the cache
func (s *TaskService) GetStats() (*todoist.StatsResponse, error) {
	if err := s.cache.EnsureFresh(); err != nil {
		return nil, err
	}
	data := s.cache.Data()
	return data.Stats, nil
}

// EditTask updates an existing task via the API
func (s *TaskService) EditTask(taskID, content, labelsStr, projectID, sectionID, dueDate, dueString, dueLang string, priority int, deadline, deadlineLang string) error {
	var labels []string
	if labelsStr != "" {
		labels = strings.Split(labelsStr, ",,..,,")
	}

	updates := map[string]any{
		"content":  content,
		"labels":   labels,
		"priority": priority,
	}

	if projectID != "" {
		updates["project_id"] = projectID
	}
	if sectionID != "" {
		updates["section_id"] = sectionID
	}

	if dueString != "" {
		if dueLang == "" {
			dueLang = s.cfg.DueLang
		}
		updates["due"] = map[string]any{
			"string": dueString,
			"lang":   dueLang,
		}
	} else if dueDate != "" {
		updates["due"] = map[string]string{"date": dueDate}
	} else {
		updates["due"] = nil
	}

	if deadline != "" {
		lang := deadlineLang
		if lang == "" {
			lang = s.cfg.DueLang
		}
		if lang == "" {
			lang = "en"
		}
		updates["deadline"] = map[string]string{"date": deadline, "lang": lang}
	} else {
		updates["deadline"] = nil
	}

	if err := s.client.UpdateTask(taskID, updates); err != nil {
		return err
	}

	if err := s.cache.Refresh(); err != nil {
		utils.Log("warning: cache refresh failed: %v", err)
	}
	return nil
}

// --- helpers ---

// reconstructEditInput builds a string that mirrors what the user would type to create a task,
// used for pre-populating the edit input field
func reconstructEditInput(task todoist.Task, projects []todoist.Project) string {
	parts := []string{task.Content}

	// Labels
	for _, label := range task.Labels {
		if strings.Contains(label, " ") {
			parts = append(parts, "@("+label+")")
		} else {
			parts = append(parts, "@"+label)
		}
	}

	// Project (skip Inbox as it's the default)
	projectName := getProjectName(projects, task.ProjectID)
	if projectName != "" && projectName != "Inbox" {
		if strings.Contains(projectName, " ") {
			parts = append(parts, "#("+projectName+")")
		} else {
			parts = append(parts, "#"+projectName)
		}
	}

	// Priority (reverse-map: API 4→p1, 3→p2, 2→p3, 1→default/skip)
	switch task.Priority {
	case 4:
		parts = append(parts, "p1")
	case 3:
		parts = append(parts, "p2")
	case 2:
		parts = append(parts, "p3")
	}

	// Due date
	if task.Due != nil && task.Due.Date != "" {
		if strings.Contains(task.Due.Date, "T") {
			// Has time component: due:YYYY-MM-DDTHH:MM
			datePart := task.Due.Date
			if len(datePart) > 16 {
				datePart = datePart[:16] // Trim seconds
			}
			parts = append(parts, "due:"+datePart)
		} else {
			parts = append(parts, "due:"+task.Due.Date)
		}
	}

	// Deadline
	if task.Deadline != nil && task.Deadline.Date != "" {
		parts = append(parts, "{"+task.Deadline.Date+"}")
	}

	return strings.Join(parts, " ")
}

func parseDueDate(dateStr string) time.Time {
	if strings.Contains(dateStr, "T") {
		if strings.Contains(dateStr, "Z") {
			t, _ := time.Parse("2006-01-02T15:04:05Z", dateStr)
			return t
		}
		t, _ := time.Parse("2006-01-02T15:04:05", dateStr)
		if t.IsZero() {
			t, _ = time.Parse("2006-01-02T15:04", dateStr)
		}
		return t
	}
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func getProjectName(projects []todoist.Project, id string) string {
	for _, p := range projects {
		if p.ID == id {
			return p.Name
		}
	}
	return ""
}

func getProjectID(projects []todoist.Project, name string) string {
	normalized := parser.NormalizeUnicode(name)
	for _, p := range projects {
		if parser.NormalizeUnicode(p.Name) == normalized {
			return p.ID
		}
	}
	return ""
}

func getSectionID(projects []todoist.Project, sections []todoist.Section, nameS string) string {
	parts := strings.SplitN(nameS, "/", 2)
	if len(parts) < 2 {
		return ""
	}
	projName := parts[0]
	if strings.HasPrefix(projName, "#") {
		projName = projName[1:]
	}
	sectName := parts[1]
	projID := getProjectID(projects, projName)

	for _, s := range sections {
		if s.Name == sectName && s.ProjectID == projID {
			return s.ID
		}
	}
	return ""
}

func filterTasks(tasks []todoist.Task, labels, projects, sections, search []string) []todoist.Task {
	if len(labels) == 0 && len(projects) == 0 && len(sections) == 0 && len(search) == 0 {
		return tasks
	}

	var result []todoist.Task
	for _, t := range tasks {
		if !matchLabels(t, labels) {
			continue
		}
		if !matchProjects(t, projects) {
			continue
		}
		if len(sections) > 0 && !matchSections(t, sections) {
			continue
		}
		if !matchSearch(t, search) {
			continue
		}
		result = append(result, t)
	}
	return result
}

func matchLabels(t todoist.Task, labels []string) bool {
	for _, l := range labels {
		found := false
		for _, tl := range t.Labels {
			if tl == l {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func matchProjects(t todoist.Task, projects []string) bool {
	for _, p := range projects {
		if !strings.Contains(t.ProjectID, p) {
			return false
		}
	}
	return true
}

func matchSections(t todoist.Task, sections []string) bool {
	if t.SectionID == "" {
		return false
	}
	for _, s := range sections {
		if !strings.Contains(t.SectionID, s) {
			return false
		}
	}
	return true
}

func matchSearch(t todoist.Task, search []string) bool {
	contentLower := strings.ToLower(t.Content)
	for _, s := range search {
		if !strings.Contains(contentLower, strings.ToLower(s)) {
			return false
		}
	}
	return true
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
	normalized := parser.NormalizeUnicode(s)
	for _, v := range slice {
		if parser.NormalizeUnicode(v) == normalized {
			return true
		}
	}
	return false
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

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
