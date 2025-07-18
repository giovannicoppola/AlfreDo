package service

import (
	"alfredo-go/pkg/todoist"
	"fmt"
	"sort"
	"time"
)

// TaskService handles task-related operations
type TaskService struct {
	client *todoist.Client
}

// NewTaskService creates a new TaskService
func NewTaskService(client *todoist.Client) *TaskService {
	return &TaskService{
		client: client,
	}
}

// GetTasksOutput generates Alfred workflow output for tasks
func (s *TaskService) GetTasksOutput(mode string) (*todoist.Output, error) {
	// Get tasks and stats in parallel
	tasksChan := make(chan *todoist.SyncResponse, 1)
	tasksErrChan := make(chan error, 1)
	statsChan := make(chan *todoist.StatsResponse, 1)
	statsErrChan := make(chan error, 1)

	go func() {
		tasks, err := s.client.GetTasks()
		if err != nil {
			tasksErrChan <- err
			return
		}
		tasksChan <- tasks
	}()

	go func() {
		stats, err := s.client.GetStats()
		if err != nil {
			statsErrChan <- err
			return
		}
		statsChan <- stats
	}()

	// Wait for both operations to complete
	var tasks *todoist.SyncResponse
	var stats *todoist.StatsResponse

	select {
	case tasks = <-tasksChan:
	case err := <-tasksErrChan:
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	select {
	case stats = <-statsChan:
	case err := <-statsErrChan:
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return s.buildOutput(tasks, stats, mode)
}

// CompleteTask completes a task by ID
func (s *TaskService) CompleteTask(taskID string) error {
	return s.client.CompleteTask(taskID)
}

// GetStats returns completion statistics
func (s *TaskService) GetStats() (*todoist.StatsResponse, error) {
	return s.client.GetStats()
}

func (s *TaskService) buildOutput(tasks *todoist.SyncResponse, stats *todoist.StatsResponse, mode string) (*todoist.Output, error) {
	today := time.Now().Format("2006-01-02")

	// Calculate stats
	var soFarCompleted, dailyGoal, weeklyGoal, totalWeekCompleted int

	// Find today's completed tasks
	for _, dayItem := range stats.DaysItems {
		if dayItem.Date == today {
			soFarCompleted = dayItem.TotalCompleted
			break
		}
	}

	dailyGoal = stats.Goals.DailyGoal
	weeklyGoal = stats.Goals.WeeklyGoal

	if len(stats.WeekItems) > 0 {
		totalWeekCompleted = stats.WeekItems[0].TotalCompleted
	}

	// Status indicators
	statusDay := "❌"
	if soFarCompleted >= dailyGoal {
		statusDay = "✅"
	}

	statusWeek := "❌"
	if totalWeekCompleted >= weeklyGoal {
		statusWeek = "✅"
	}

	// Filter tasks with due dates
	var dueDateTasks []todoist.Task
	for _, task := range tasks.Items {
		if task.Due != nil && task.Due.Date <= today {
			dueDateTasks = append(dueDateTasks, task)
		}
	}

	// Filter by mode
	if mode == "today" {
		var todayTasks []todoist.Task
		for _, task := range dueDateTasks {
			if task.Due.Date == today {
				todayTasks = append(todayTasks, task)
			}
		}
		dueDateTasks = todayTasks
	}

	// Sort by due date
	sort.Slice(dueDateTasks, func(i, j int) bool {
		return dueDateTasks[i].Due.Date < dueDateTasks[j].Due.Date
	})

	output := &todoist.Output{Items: []todoist.OutputItem{}}

	if len(dueDateTasks) == 0 {
		// No tasks left
		output.Items = append(output.Items, todoist.OutputItem{
			Title: "no tasks left to do today 🙌",
			Subtitle: fmt.Sprintf("Daily: %d/%d%s Weekly: %d/%d%s",
				soFarCompleted, dailyGoal, statusDay,
				totalWeekCompleted, weeklyGoal, statusWeek),
			Arg: "",
			Mods: map[string]todoist.ModsItem{
				"shift": {
					Arg:      "",
					Subtitle: "nothing to see here",
				},
			},
		})
	} else {
		// Build task items
		totalMatchCount := len(dueDateTasks)
		dueToday := len(dueDateTasks)

		for i, task := range dueDateTasks {
			countR := i + 1
			subtitle := fmt.Sprintf("%s-%d/%d-%d due today. Daily: %d/%d%s Weekly: %d/%d%s",
				task.Due.Date, countR, totalMatchCount, dueToday,
				soFarCompleted, dailyGoal, statusDay,
				totalWeekCompleted, weeklyGoal, statusWeek)

			output.Items = append(output.Items, todoist.OutputItem{
				Title:    task.Content,
				Subtitle: subtitle,
				Arg:      fmt.Sprintf("%s;;%d", task.ID, dueToday),
			})
		}
	}

	return output, nil
}
