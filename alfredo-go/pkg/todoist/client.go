package todoist

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Todoist API client
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
	syncURL    string
}

// Task represents a Todoist task
type Task struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Due     *Due   `json:"due"`
}

// Due represents a task's due date
type Due struct {
	Date string `json:"date"`
}

// SyncResponse represents the response from the sync API
type SyncResponse struct {
	Items []Task `json:"items"`
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

// Output represents the Alfred workflow output format
type Output struct {
	Items []OutputItem `json:"items"`
}

// OutputItem represents a single item in Alfred workflow output
type OutputItem struct {
	Title    string              `json:"title"`
	Subtitle string              `json:"subtitle"`
	Arg      string              `json:"arg"`
	Mods     map[string]ModsItem `json:"mods,omitempty"`
}

// ModsItem represents modifier keys in Alfred workflow
type ModsItem struct {
	Arg      string `json:"arg"`
	Subtitle string `json:"subtitle"`
}

// NewClient creates a new Todoist API client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.todoist.com/rest/v1",
		syncURL:    "https://api.todoist.com/sync/v8",
	}
}

// GetTasks fetches tasks from the sync API
func (c *Client) GetTasks() (*SyncResponse, error) {
	apiURL := fmt.Sprintf("%s/sync", c.syncURL)

	data := url.Values{}
	data.Set("sync_token", "*")
	data.Set("resource_types", `["items"]`)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
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
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var syncResp SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, err
	}

	return &syncResp, nil
}

// GetStats fetches completion statistics
func (c *Client) GetStats() (*StatsResponse, error) {
	apiURL := fmt.Sprintf("%s/completed/get_stats", c.syncURL)

	req, err := http.NewRequest("GET", apiURL, nil)
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
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var statsResp StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&statsResp); err != nil {
		return nil, err
	}

	return &statsResp, nil
}

// CompleteTask marks a task as completed
func (c *Client) CompleteTask(taskID string) error {
	apiURL := fmt.Sprintf("%s/tasks/%s/close", c.baseURL, taskID)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to complete task: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
