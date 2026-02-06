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
	ID          string   `json:"id"`
	Content     string   `json:"content"`
	Due         *Due     `json:"due"`
	Labels      []string `json:"labels"`
	Priority    int      `json:"priority"`
	ProjectID   string   `json:"project_id"`
	SectionID   string   `json:"section_id"`
	IsRecurring bool     `json:"is_recurring"`
}

// Due represents a task's due date
type Due struct {
	Date string `json:"date"`
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

// SyncResponse represents the full sync API response (used for cache rebuild)
type SyncResponse struct {
	Items    []Task    `json:"items"`
	Projects []Project `json:"projects"`
	Sections []Section `json:"sections"`
	Labels   []Label   `json:"labels"`
	Stats    *StatsResponse `json:"stats"`
	User     *UserInfo      `json:"user"`
}

// paginatedResponse is the generic wrapper for paginated API v1 responses
type paginatedResponse struct {
	Results    json.RawMessage `json:"results"`
	NextCursor *string         `json:"next_cursor"`
}

// NewClient creates a new Todoist API client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.todoist.com",
	}
}

// getPaginated fetches all pages from a paginated endpoint
func (c *Client) getPaginated(endpoint string) ([]json.RawMessage, error) {
	var allPages []json.RawMessage
	cursor := ""

	for {
		url := c.baseURL + endpoint
		if cursor != "" {
			url += "?cursor=" + cursor
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("%s: API request failed with status %d: %s", endpoint, resp.StatusCode, string(body))
		}

		var page paginatedResponse
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			return nil, fmt.Errorf("%s: failed to decode response: %w", endpoint, err)
		}

		allPages = append(allPages, page.Results)

		if page.NextCursor == nil || *page.NextCursor == "" {
			break
		}
		cursor = *page.NextCursor
	}

	return allPages, nil
}

// GetTasks fetches all active tasks from the API v1
func (c *Client) GetTasks() ([]Task, error) {
	pages, err := c.getPaginated("/api/v1/tasks")
	if err != nil {
		return nil, err
	}

	var all []Task
	for _, page := range pages {
		var tasks []Task
		if err := json.Unmarshal(page, &tasks); err != nil {
			return nil, fmt.Errorf("tasks: failed to unmarshal: %w", err)
		}
		all = append(all, tasks...)
	}
	return all, nil
}

// GetProjects fetches all projects
func (c *Client) GetProjects() ([]Project, error) {
	pages, err := c.getPaginated("/api/v1/projects")
	if err != nil {
		return nil, err
	}

	var all []Project
	for _, page := range pages {
		var projects []Project
		if err := json.Unmarshal(page, &projects); err != nil {
			return nil, fmt.Errorf("projects: failed to unmarshal: %w", err)
		}
		all = append(all, projects...)
	}
	return all, nil
}

// GetSections fetches all sections
func (c *Client) GetSections() ([]Section, error) {
	pages, err := c.getPaginated("/api/v1/sections")
	if err != nil {
		return nil, err
	}

	var all []Section
	for _, page := range pages {
		var sections []Section
		if err := json.Unmarshal(page, &sections); err != nil {
			return nil, fmt.Errorf("sections: failed to unmarshal: %w", err)
		}
		all = append(all, sections...)
	}
	return all, nil
}

// GetLabels fetches all labels
func (c *Client) GetLabels() ([]Label, error) {
	pages, err := c.getPaginated("/api/v1/labels")
	if err != nil {
		return nil, err
	}

	var all []Label
	for _, page := range pages {
		var labels []Label
		if err := json.Unmarshal(page, &labels); err != nil {
			return nil, fmt.Errorf("labels: failed to unmarshal: %w", err)
		}
		all = append(all, labels...)
	}
	return all, nil
}

// GetStats fetches completion statistics and user info via the Sync API
func (c *Client) GetStats() (*StatsResponse, *UserInfo, error) {
	form := "sync_token=*&resource_types=" + `["user","stats"]`
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/sync", strings.NewReader(form))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("/api/v1/sync (stats): API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var syncResp struct {
		Stats *StatsResponse `json:"stats"`
		User  *UserInfo      `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, nil, err
	}
	return syncResp.Stats, syncResp.User, nil
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

// BuildDueObject converts a date string (YYYY-MM-DD or YYYY-MM-DDTHH:MM) into
// the API v1 due object format: {"date": "...", "time": "..."}
func BuildDueObject(dateStr string) map[string]string {
	due := map[string]string{}
	if strings.Contains(dateStr, "T") {
		parts := strings.SplitN(dateStr, "T", 2)
		due["date"] = parts[0]
		due["time"] = parts[1]
	} else {
		due["date"] = dateStr
	}
	return due
}

// CreateTask creates a new task
func (c *Client) CreateTask(content string, labels []string, projectID, sectionID, dueDate string, priority int) error {
	payload := map[string]any{
		"content":  content,
		"priority": priority,
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
	if dueDate != "" {
		if strings.Contains(dueDate, "T") {
			payload["due_datetime"] = dueDate + ":00"
		} else {
			payload["due_date"] = dueDate
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

// CreateLabel creates a new label
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
