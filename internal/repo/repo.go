// Package repo is a thin wrapper around the git CLI for the yawn metadata repo.
// All operations target the repo explicitly via `git -C <dir>`, so they work
// regardless of the caller's current working directory (important inside a
// switch subshell, where the cwd is the task's work directory).
package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BaseBranch is the root branch of the yawn repo; tasks branch off it by default
// and "done" tasks are merged back into their parent (ultimately this branch).
const BaseBranch = "done"

type Repo struct {
	Dir string
}

// Resolve locates the yawn repo: $YAWN_DIR if set, else ~/yawn.
func Resolve() (*Repo, error) {
	dir := os.Getenv("YAWN_DIR")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(home, "yawn")
	}
	return &Repo{Dir: dir}, nil
}

// Exists reports whether the repo has been initialized.
func (r *Repo) Exists() bool {
	info, err := os.Stat(filepath.Join(r.Dir, ".git"))
	return err == nil && info.IsDir()
}

// git runs a git command in the repo and returns trimmed stdout.
func (r *Repo) git(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", r.Dir}, args...)...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(out.String()),
			fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(errb.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// gitOK runs a git command and reports whether it exited zero.
func (r *Repo) gitOK(args ...string) bool {
	cmd := exec.Command("git", append([]string{"-C", r.Dir}, args...)...)
	return cmd.Run() == nil
}

// Init creates the repo directory, runs `git init -b done`, and makes the
// initial commit so branches have a common base.
func (r *Repo) Init() error {
	if err := os.MkdirAll(r.Dir, 0o755); err != nil {
		return err
	}
	if _, err := r.git("init", "-b", BaseBranch); err != nil {
		return err
	}
	readme := filepath.Join(r.Dir, "README.md")
	if err := os.WriteFile(readme, []byte("# yawn tasks\n\nThis repository is managed by yawn. Each task is a branch.\n"), 0o644); err != nil {
		return err
	}
	if err := r.CommitAll("yawn init"); err != nil {
		return err
	}
	return nil
}

// CurrentBranch returns the branch the repo is currently on.
func (r *Repo) CurrentBranch() (string, error) {
	return r.git("rev-parse", "--abbrev-ref", "HEAD")
}

// BranchExists reports whether a local branch with the given name exists.
func (r *Repo) BranchExists(name string) bool {
	return r.gitOK("rev-parse", "--verify", "--quiet", "refs/heads/"+name)
}

// CreateBranch creates and checks out a new branch off `from`.
func (r *Repo) CreateBranch(name, from string) error {
	_, err := r.git("switch", "-c", name, from)
	return err
}

// Switch checks out an existing branch.
func (r *Repo) Switch(name string) error {
	_, err := r.git("switch", name)
	return err
}

// Show returns the contents of a file on a given branch without checking it out.
func (r *Repo) Show(branch, path string) (string, error) {
	return r.git("show", branch+":"+path)
}

// ListBranches returns all local branch names.
func (r *Repo) ListBranches() ([]string, error) {
	out, err := r.git("for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// CommitAll stages everything and commits. It is a no-op when there is nothing
// to commit.
func (r *Repo) CommitAll(msg string) error {
	if _, err := r.git("add", "-A"); err != nil {
		return err
	}
	// `diff --cached --quiet` exits 0 when there are no staged changes.
	if r.gitOK("diff", "--cached", "--quiet") {
		return nil
	}
	_, err := r.git("commit", "-m", msg)
	return err
}

// Merge merges `branch` into the currently checked-out branch. On conflict it
// aborts the merge and returns an error so the repo is left clean.
func (r *Repo) Merge(branch string) error {
	if _, err := r.git("merge", "--no-edit", branch); err != nil {
		r.gitOK("merge", "--abort")
		return fmt.Errorf("merge of %q failed (aborted): %w", branch, err)
	}
	return nil
}
