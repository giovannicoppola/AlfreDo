package cache

import (
	"alfredo-go/pkg/config"
	"alfredo-go/pkg/todoist"
	"alfredo-go/pkg/utils"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// CachedData holds all cached Todoist data
type CachedData struct {
	Tasks     []todoist.Task       `json:"tasks"`
	Projects  []todoist.Project    `json:"projects"`
	Sections  []todoist.Section    `json:"sections"`
	Labels    []todoist.Label      `json:"labels"`
	Stats     *todoist.StatsResponse `json:"stats"`
	User      *todoist.UserInfo    `json:"user"`
	FetchedAt time.Time            `json:"fetched_at"`
}

// Cache manages local caching of Todoist data
type Cache struct {
	client *todoist.Client
	cfg    *config.Config
	data   *CachedData
}

// NewCache creates a new Cache
func NewCache(client *todoist.Client, cfg *config.Config) *Cache {
	return &Cache{
		client: client,
		cfg:    cfg,
	}
}

func (c *Cache) dbPath() string {
	return filepath.Join(c.cfg.DataFolder, "allData.json")
}

func (c *Cache) labelCountsPath() string {
	return filepath.Join(c.cfg.DataFolder, "labelCounts.json")
}

func (c *Cache) projectCountsPath() string {
	return filepath.Join(c.cfg.DataFolder, "projectCounts.json")
}

// NeedsRefresh returns true if the cache is stale or missing
func (c *Cache) NeedsRefresh() bool {
	if c.cfg.DataFolder == "" {
		return true
	}
	info, err := os.Stat(c.dbPath())
	if err != nil {
		return true
	}
	elapsed := time.Since(info.ModTime())
	return elapsed.Hours() >= float64(c.cfg.RefreshRate*24)
}

// Refresh fetches all data from the API in a single sync call and saves to disk
func (c *Cache) Refresh() error {
	utils.Log("refreshing cache...")

	syncResp, err := c.client.SyncAll()
	if err != nil {
		return err
	}

	c.data = &CachedData{
		Tasks:     syncResp.Items,
		Projects:  syncResp.Projects,
		Sections:  syncResp.Sections,
		Labels:    syncResp.Labels,
		Stats:     syncResp.Stats,
		User:      syncResp.User,
		FetchedAt: time.Now(),
	}

	if c.data.User == nil {
		c.data.User = &todoist.UserInfo{}
	}

	if err := c.save(); err != nil {
		return err
	}

	// Save label and project counts
	labelCounts := ComputeLabelCounts(c.data.Tasks, c.data.Labels)
	if err := saveJSON(c.labelCountsPath(), labelCounts); err != nil {
		utils.Log("warning: failed to save label counts: %v", err)
	}

	projectCounts := ComputeProjectCounts(c.data.Tasks, c.data.Projects, c.data.Sections)
	if err := saveJSON(c.projectCountsPath(), projectCounts); err != nil {
		utils.Log("warning: failed to save project counts: %v", err)
	}

	utils.Log("cache refreshed")
	return nil
}

// Load reads cached data from disk
func (c *Cache) Load() error {
	f, err := os.Open(c.dbPath())
	if err != nil {
		return err
	}
	defer f.Close()

	c.data = &CachedData{}
	return json.NewDecoder(f).Decode(c.data)
}

// EnsureFresh loads from cache if fresh, otherwise refreshes
func (c *Cache) EnsureFresh() error {
	if c.cfg.DataFolder == "" {
		return c.Refresh()
	}
	if c.NeedsRefresh() {
		return c.Refresh()
	}
	return c.Load()
}

// Data returns the cached data
func (c *Cache) Data() *CachedData {
	return c.data
}

// LoadLabelCounts reads label counts from disk
func (c *Cache) LoadLabelCounts() (map[string]int, error) {
	return loadJSONMap(c.labelCountsPath())
}

// LoadProjectCounts reads project counts from disk
func (c *Cache) LoadProjectCounts() (map[string]int, error) {
	return loadJSONMap(c.projectCountsPath())
}

// SaveLabelCounts writes label counts to disk
func (c *Cache) SaveLabelCounts(counts map[string]int) error {
	return saveJSON(c.labelCountsPath(), counts)
}

func (c *Cache) save() error {
	if c.cfg.DataFolder == "" {
		return nil
	}
	return saveJSON(c.dbPath(), c.data)
}

func saveJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	return enc.Encode(v)
}

func loadJSONMap(path string) (map[string]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m map[string]int
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}
