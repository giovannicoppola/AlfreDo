package todoist

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents a Todoist API client
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// Task represents a Todoist task
type Task struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Due         *Due      `json:"due"`
	Deadline    *Deadline `json:"deadline"`
	Labels      []string  `json:"labels"`
	Priority    int       `json:"priority"`
	ProjectID   string    `json:"project_id"`
	SectionID   string    `json:"section_id"`
	IsRecurring bool      `json:"is_recurring"`
}

// Due represents a task's due date
type Due struct {
	Date string `json:"date"`
}

// Deadline represents a task's deadline
type Deadline struct {
	Date string `json:"date"`
	Lang string `json:"lang,omitempty"`
}

// Project represents a Todoist project
type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsDeleted  bool   `json:"is_deleted"`
	IsArchived bool   `json:"is_archived"`
}

// Section represents a Todoist section
type Section struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
}

// Label represents a Todoist label
type Label struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDeleted bool   `json:"is_deleted"`
}

// StatsResponse represents the response from the stats API
type StatsResponse struct {
	DaysItems []DayItem  `json:"days_items"`
	Goals     Goals      `json:"goals"`
	WeekItems []WeekItem `json:"week_items"`
}

// DayItem represents completed tasks for a day
type DayItem struct {
	Date           string `json:"date"`
	TotalCompleted int    `json:"total_completed"`
}

// WeekItem represents completed tasks for a week
type WeekItem struct {
	TotalCompleted int `json:"total_completed"`
}

// Goals represents daily and weekly goals
type Goals struct {
	DailyGoal  int `json:"daily_goal"`
	WeeklyGoal int `json:"weekly_goal"`
}

// UserInfo holds daily/weekly goal info from sync API
type UserInfo struct {
	DailyGoal  int `json:"daily_goal"`
	WeeklyGoal int `json:"weekly_goal"`
}

// SyncAllResponse represents the full sync API response
type SyncAllResponse struct {
	Items    []Task         `json:"items"`
	Projects []Project      `json:"projects"`
	Sections []Section      `json:"sections"`
	Labels   []Label        `json:"labels"`
	Stats    *StatsResponse `json:"stats"`
	User     *UserInfo      `json:"user"`
}

// NewClient creates a new Todoist API client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.todoist.com",
	}
}

// SyncAll fetches all data in a single API call via the Sync endpoint
func (c *Client) SyncAll() (*SyncAllResponse, error) {
	form := `sync_token=*&resource_types=["all"]`
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/sync", strings.NewReader(form))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("sync failed with status %d: %s", resp.StatusCode, string(body))
	}

	var syncResp SyncAllResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, fmt.Errorf("failed to decode sync response: %w", err)
	}
	return &syncResp, nil
}

// CompleteTask marks a task as completed
func (c *Client) CompleteTask(taskID string) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/tasks/%s/close", c.baseURL, taskID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to complete task: status %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}

// CreateTask creates a new task via the REST API
func (c *Client) CreateTask(content string, labels []string, projectID, sectionID, dueDate, dueString, dueLang string, priority int, deadline *Deadline, description string) error {
	payload := map[string]any{
		"content":  content,
		"priority": priority,
	}
	if description != "" {
		payload["description"] = description
	}
	if len(labels) > 0 {
		payload["labels"] = labels
	}
	if projectID != "" {
		payload["project_id"] = projectID
	}
	if sectionID != "" {
		payload["section_id"] = sectionID
	}
	if dueString != "" {
		// Use Todoist's NLP: send due_string + due_lang
		payload["due_string"] = dueString
		if dueLang != "" {
			payload["due_lang"] = dueLang
		}
	} else if dueDate != "" {
		if strings.Contains(dueDate, "T") {
			payload["due_datetime"] = dueDate + ":00"
		} else {
			payload["due_date"] = dueDate
		}
	}
	if deadline != nil {
		payload["deadline"] = map[string]string{
			"date": deadline.Date,
			"lang": deadline.Lang,
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/tasks", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create task: status %d, body: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// UpdateTask updates a task via the Sync API (item_update command)
func (c *Client) UpdateTask(taskID string, updates map[string]any) error {
	args := map[string]any{"id": taskID}
	for k, v := range updates {
		args[k] = v
	}

	commands := []map[string]any{
		{
			"type": "item_update",
			"uuid": fmt.Sprintf("%d", time.Now().UnixNano()),
			"args": args,
		},
	}

	cmdJSON, err := json.Marshal(commands)
	if err != nil {
		return err
	}

	form := "commands=" + string(cmdJSON)
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/sync", strings.NewReader(form))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update task: status %d, body: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// CreateLabel creates a new label via the REST API
func (c *Client) CreateLabel(name string) error {
	payload := map[string]string{"name": name}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/labels", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create label: status %d, body: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
