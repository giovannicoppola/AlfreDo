package cache

import (
	"alfredo-go/pkg/config"
	"alfredo-go/pkg/todoist"
	"alfredo-go/pkg/utils"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
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

// Refresh fetches all data from the API in parallel and saves to disk
func (c *Cache) Refresh() error {
	utils.Log("refreshing cache...")

	var (
		tasks    []todoist.Task
		projects []todoist.Project
		sections []todoist.Section
		labels   []todoist.Label
		stats    *todoist.StatsResponse
		user     *todoist.UserInfo
		wg       sync.WaitGroup
		mu       sync.Mutex
		errs     []error
	)

	collect := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}

	collect(func() error {
		t, err := c.client.GetTasks()
		if err != nil {
			return err
		}
		mu.Lock()
		tasks = t
		mu.Unlock()
		return nil
	})

	collect(func() error {
		p, err := c.client.GetProjects()
		if err != nil {
			return err
		}
		mu.Lock()
		projects = p
		mu.Unlock()
		return nil
	})

	collect(func() error {
		s, err := c.client.GetSections()
		if err != nil {
			return err
		}
		mu.Lock()
		sections = s
		mu.Unlock()
		return nil
	})

	collect(func() error {
		l, err := c.client.GetLabels()
		if err != nil {
			return err
		}
		mu.Lock()
		labels = l
		mu.Unlock()
		return nil
	})

	// Stats fetch is non-fatal — if it fails, we still have tasks/projects/labels
	collect(func() error {
		st, u, err := c.client.GetStats()
		if err != nil {
			utils.Log("Warning: could not fetch stats: %s", err.Error())
			return nil
		}
		mu.Lock()
		stats = st
		user = u
		mu.Unlock()
		return nil
	})

	wg.Wait()

	if len(errs) > 0 {
		return errs[0]
	}

	c.data = &CachedData{
		Tasks:     tasks,
		Projects:  projects,
		Sections:  sections,
		Labels:    labels,
		Stats:     stats,
		User:      &todoist.UserInfo{},
		FetchedAt: time.Now(),
	}

	// Use user info from sync response if available
	if user != nil {
		c.data.User = user
	} else if stats != nil {
		c.data.User = &todoist.UserInfo{
			DailyGoal:  stats.Goals.DailyGoal,
			WeeklyGoal: stats.Goals.WeeklyGoal,
		}
	}

	if err := c.save(); err != nil {
		return err
	}

	// Save label and project counts
	labelCounts := ComputeLabelCounts(tasks, labels)
	if err := saveJSON(c.labelCountsPath(), labelCounts); err != nil {
		utils.Log("warning: failed to save label counts: %v", err)
	}

	projectCounts := ComputeProjectCounts(tasks, projects, sections)
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
