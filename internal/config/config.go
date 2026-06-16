// Package config handles a task's config.yaml.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Status values yawn understands. Status is stored free-form, but these are the
// canonical ones; "done" triggers a merge into the parent.
const (
	StatusPending  = "pending"
	StatusDone     = "done"
	StatusArchived = "archived"
)

// Task is the metadata stored in <name>/config.yaml on the task's branch.
type Task struct {
	Directory string   `yaml:"directory"`
	Status    string   `yaml:"status"`
	Parent    string   `yaml:"parent"`
	RC        []string `yaml:"rc"`
	Cleanup   []string `yaml:"cleanup"`
}

// Default returns a fresh task config for a task created off the given parent.
func Default(parent string) *Task {
	return &Task{Status: StatusPending, Parent: parent}
}

// Parse decodes a Task from YAML bytes (e.g. the output of `git show`).
func Parse(b []byte) (*Task, error) {
	var t Task
	if err := yaml.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Bytes encodes the task as YAML.
func (t *Task) Bytes() ([]byte, error) {
	return yaml.Marshal(t)
}

// Load reads a task config from disk.
func Load(path string) (*Task, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Save writes the task config to disk.
func (t *Task) Save(path string) error {
	b, err := t.Bytes()
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
