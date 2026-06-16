// Package state manages cross-task state that does not belong to any single
// branch — currently the per-day todo lists. It is stored as an untracked file
// under the repo's .git directory so it is branch-independent and never appears
// as untracked working-tree content.
package state

import (
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// State is the on-disk cross-task state.
type State struct {
	// Todos maps a date (YYYY-MM-DD) to the task names prioritized that day.
	Todos map[string][]string `yaml:"todos"`
}

func path(repoDir string) string {
	return filepath.Join(repoDir, ".git", "yawn", "state.yaml")
}

// Load reads state from the repo, returning an empty state if none exists.
func Load(repoDir string) (*State, error) {
	b, err := os.ReadFile(path(repoDir))
	if os.IsNotExist(err) {
		return &State{Todos: map[string][]string{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var s State
	if err := yaml.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	if s.Todos == nil {
		s.Todos = map[string][]string{}
	}
	return &s, nil
}

// Save writes state back to the repo.
func (s *State) Save(repoDir string) error {
	if err := os.MkdirAll(filepath.Dir(path(repoDir)), 0o755); err != nil {
		return err
	}
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path(repoDir), b, 0o644)
}

// Add appends a task name to a date's list, ignoring duplicates.
func (s *State) Add(date, name string) {
	for _, n := range s.Todos[date] {
		if n == name {
			return
		}
	}
	s.Todos[date] = append(s.Todos[date], name)
}

// Get returns the task names for a date.
func (s *State) Get(date string) []string {
	return s.Todos[date]
}

// MostRecentBefore returns the latest date strictly before `date` that has a
// non-empty todo list. Dates are YYYY-MM-DD so lexical order is chronological.
func (s *State) MostRecentBefore(date string) (string, bool) {
	var dates []string
	for d, names := range s.Todos {
		if d < date && len(names) > 0 {
			dates = append(dates, d)
		}
	}
	if len(dates) == 0 {
		return "", false
	}
	sort.Strings(dates)
	return dates[len(dates)-1], true
}
