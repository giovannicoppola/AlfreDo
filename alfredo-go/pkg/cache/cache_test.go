package cache

import (
	"alfredo-go/pkg/config"
	"alfredo-go/pkg/todoist"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNeedsRefresh_MissingFile(t *testing.T) {
	cfg := &config.Config{
		DataFolder:  t.TempDir(),
		RefreshRate: 1,
	}
	c := NewCache(nil, cfg)
	if !c.NeedsRefresh() {
		t.Error("NeedsRefresh should return true when database file is missing")
	}
}

func TestNeedsRefresh_FreshFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		DataFolder:  dir,
		RefreshRate: 1,
	}

	// Create a fresh file
	path := filepath.Join(dir, "allData.json")
	os.WriteFile(path, []byte("{}"), 0644)

	c := NewCache(nil, cfg)
	if c.NeedsRefresh() {
		t.Error("NeedsRefresh should return false when database file is fresh")
	}
}

func TestNeedsRefresh_StaleFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		DataFolder:  dir,
		RefreshRate: 0, // 0 days = always refresh
	}

	// Create a file and backdate it
	path := filepath.Join(dir, "allData.json")
	os.WriteFile(path, []byte("{}"), 0644)
	oldTime := time.Now().Add(-48 * time.Hour)
	os.Chtimes(path, oldTime, oldTime)

	c := NewCache(nil, cfg)
	if !c.NeedsRefresh() {
		t.Error("NeedsRefresh should return true when database file is stale")
	}
}

func TestNeedsRefresh_EmptyDataFolder(t *testing.T) {
	cfg := &config.Config{
		DataFolder:  "",
		RefreshRate: 1,
	}
	c := NewCache(nil, cfg)
	if !c.NeedsRefresh() {
		t.Error("NeedsRefresh should return true when DataFolder is empty")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		DataFolder:  dir,
		RefreshRate: 1,
	}

	c := NewCache(nil, cfg)
	c.data = &CachedData{
		Tasks: []todoist.Task{
			{ID: "1", Content: "Test task"},
		},
		Projects: []todoist.Project{
			{ID: "p1", Name: "Inbox"},
		},
		FetchedAt: time.Now(),
	}

	if err := c.save(); err != nil {
		t.Fatalf("save() error: %v", err)
	}

	// Load it back
	c2 := NewCache(nil, cfg)
	if err := c2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if len(c2.data.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(c2.data.Tasks))
	}
	if c2.data.Tasks[0].Content != "Test task" {
		t.Errorf("task content = %q, want %q", c2.data.Tasks[0].Content, "Test task")
	}
}

func TestComputeLabelCounts(t *testing.T) {
	tasks := []todoist.Task{
		{Labels: []string{"work", "urgent"}},
		{Labels: []string{"work"}},
		{Labels: []string{"personal"}},
	}
	labels := []todoist.Label{
		{Name: "work"},
		{Name: "urgent"},
		{Name: "personal"},
		{Name: "unused"},
	}

	counts := ComputeLabelCounts(tasks, labels)
	if counts["work"] != 2 {
		t.Errorf("work count = %d, want 2", counts["work"])
	}
	if counts["urgent"] != 1 {
		t.Errorf("urgent count = %d, want 1", counts["urgent"])
	}
	if counts["unused"] != 0 {
		t.Errorf("unused count = %d, want 0", counts["unused"])
	}
}

func TestComputeProjectCounts(t *testing.T) {
	tasks := []todoist.Task{
		{ProjectID: "p1"},
		{ProjectID: "p1"},
		{ProjectID: "p2", SectionID: "s1"},
	}
	projects := []todoist.Project{
		{ID: "p1", Name: "Inbox"},
		{ID: "p2", Name: "Work"},
	}
	sections := []todoist.Section{
		{ID: "s1", Name: "Urgent", ProjectID: "p2"},
	}

	counts := ComputeProjectCounts(tasks, projects, sections)
	if counts["Inbox"] != 2 {
		t.Errorf("Inbox count = %d, want 2", counts["Inbox"])
	}
	if counts["Work/Urgent"] != 1 {
		t.Errorf("Work/Urgent count = %d, want 1", counts["Work/Urgent"])
	}
}
