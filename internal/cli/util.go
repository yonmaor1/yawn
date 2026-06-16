package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/yonmaor/yawn/internal/config"
	"github.com/yonmaor/yawn/internal/repo"
)

func today() string     { return time.Now().Format("2006-01-02") }
func timestamp() string { return time.Now().Format("2006-01-02 15:04") }

// currentTask returns the active task: $YAWN_TASK if set (inside a switch
// subshell), else the current branch when it is not the base branch.
func currentTask(r *repo.Repo) string {
	if t := os.Getenv("YAWN_TASK"); t != "" {
		return t
	}
	b, err := r.CurrentBranch()
	if err != nil || b == repo.BaseBranch {
		return ""
	}
	return b
}

// configPath returns the on-disk path to a task's config.yaml (valid when the
// task's branch, or a descendant, is checked out).
func configPath(r *repo.Repo, name string) string {
	return filepath.Join(r.Dir, name, "config.yaml")
}

func readmePath(r *repo.Repo, name string) string {
	return filepath.Join(r.Dir, name, "README.md")
}

// loadTask reads a task's config from its own branch without checking it out.
func loadTask(r *repo.Repo, name string) (*config.Task, error) {
	out, err := r.Show(name, name+"/config.yaml")
	if err != nil {
		return nil, err
	}
	return config.Parse([]byte(out))
}

// loadTasks returns every task (branch != base) mapped to its config.
func loadTasks(r *repo.Repo) (map[string]*config.Task, error) {
	branches, err := r.ListBranches()
	if err != nil {
		return nil, err
	}
	tasks := map[string]*config.Task{}
	for _, b := range branches {
		if b == repo.BaseBranch {
			continue
		}
		t, err := loadTask(r, b)
		if err != nil {
			continue // branch without a yawn config; skip
		}
		tasks[b] = t
	}
	return tasks, nil
}

// openEditor opens a file in $EDITOR (default vi), inheriting the terminal.
func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %q", editor, path))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
