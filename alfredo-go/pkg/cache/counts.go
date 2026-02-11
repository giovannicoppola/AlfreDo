package cache

import (
	"alfredo-go/pkg/todoist"
	"sort"
)

// ComputeLabelCounts computes label usage counts from tasks, including zero-count active labels
func ComputeLabelCounts(tasks []todoist.Task, labels []todoist.Label) map[string]int {
	counts := make(map[string]int)
	for _, task := range tasks {
		for _, label := range task.Labels {
			counts[label]++
		}
	}
	// Include active labels with 0 count
	for _, label := range labels {
		if !label.IsDeleted {
			if _, exists := counts[label.Name]; !exists {
				counts[label.Name] = 0
			}
		}
	}
	return counts
}

// ComputeProjectCounts computes project (and project/section) usage counts from tasks
func ComputeProjectCounts(tasks []todoist.Task, projects []todoist.Project, sections []todoist.Section) map[string]int {
	counts := make(map[string]int)
	projMap := projectNameMap(projects)
	sectMap := sectionNameMap(sections)

	for _, task := range tasks {
		name := projMap[task.ProjectID]
		if task.SectionID != "" {
			if sectName, ok := sectMap[task.SectionID]; ok {
				name = name + "/" + sectName
			}
		}
		counts[name]++
	}

	// Include active projects with 0 count
	for _, p := range projects {
		if !p.IsDeleted && !p.IsArchived {
			if _, exists := counts[p.Name]; !exists {
				counts[p.Name] = 0
			}
		}
	}
	// Include sections with 0 count
	for _, s := range sections {
		pName := projMap[s.ProjectID]
		fullName := pName + "/" + s.Name
		if _, exists := counts[fullName]; !exists {
			counts[fullName] = 0
		}
	}

	return counts
}

// FetchLabelsFromSubset computes label counts from a task subset, returns counts and sorted label list (prefixed with @)
func FetchLabelsFromSubset(tasks []todoist.Task) (map[string]int, []string) {
	counts := make(map[string]int)
	for _, task := range tasks {
		for _, label := range task.Labels {
			counts[label]++
		}
	}

	labels := make([]string, 0, len(counts))
	for k := range counts {
		labels = append(labels, k)
	}
	sort.Slice(labels, func(i, j int) bool {
		return counts[labels[i]] > counts[labels[j]]
	})

	prefixed := make([]string, len(labels))
	for i, l := range labels {
		prefixed[i] = "@" + l
	}
	return counts, prefixed
}

// FetchProjectsFromSubset computes project counts from a task subset
func FetchProjectsFromSubset(tasks []todoist.Task, projects []todoist.Project, sections []todoist.Section) (map[string]int, []string) {
	counts := make(map[string]int)
	projMap := projectNameMap(projects)
	sectMap := sectionNameMap(sections)

	for _, task := range tasks {
		name := projMap[task.ProjectID]
		if task.SectionID != "" {
			if sectName, ok := sectMap[task.SectionID]; ok {
				name = name + "/" + sectName
			}
		}
		counts[name]++
	}

	names := make([]string, 0, len(counts))
	for k := range counts {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool {
		return counts[names[i]] > counts[names[j]]
	})

	prefixed := make([]string, len(names))
	for i, n := range names {
		prefixed[i] = "#" + n
	}
	return counts, prefixed
}

// FetchSectionsFromSubset computes section counts and parent project mapping
func FetchSectionsFromSubset(tasks []todoist.Task, sections []todoist.Section, projects []todoist.Project) (map[string]int, []string, map[string]string) {
	counts := make(map[string]int)
	parentProjects := make(map[string]string)
	projMap := projectNameMap(projects)
	sectMap := sectionNameMap(sections)
	sectProjMap := make(map[string]string) // sectionID -> projectID
	for _, s := range sections {
		sectProjMap[s.ID] = s.ProjectID
	}

	for _, task := range tasks {
		if task.SectionID != "" {
			sectName := sectMap[task.SectionID]
			parentProjects[sectName] = projMap[sectProjMap[task.SectionID]]
			counts[sectName]++
		}
	}

	names := make([]string, 0, len(counts))
	for k := range counts {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool {
		return counts[names[i]] > counts[names[j]]
	})

	prefixed := make([]string, len(names))
	for i, n := range names {
		prefixed[i] = "^" + n
	}
	return counts, prefixed, parentProjects
}

func projectNameMap(projects []todoist.Project) map[string]string {
	m := make(map[string]string, len(projects))
	for _, p := range projects {
		m[p.ID] = p.Name
	}
	return m
}

func sectionNameMap(sections []todoist.Section) map[string]string {
	m := make(map[string]string, len(sections))
	for _, s := range sections {
		m[s.ID] = s.Name
	}
	return m
}
