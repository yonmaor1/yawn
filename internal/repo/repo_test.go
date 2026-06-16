package repo

import (
	"os"
	"path/filepath"
	"testing"
)

// newTestRepo returns an initialized repo in a temp dir with a deterministic
// git identity so commits succeed regardless of the host's global config.
func newTestRepo(t *testing.T) *Repo {
	t.Helper()
	t.Setenv("GIT_AUTHOR_NAME", "yawn test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "yawn test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	r := &Repo{Dir: filepath.Join(t.TempDir(), "yawn")}
	if err := r.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	return r
}

func TestInitAndBranchLifecycle(t *testing.T) {
	r := newTestRepo(t)

	if !r.Exists() {
		t.Fatal("repo should exist after init")
	}
	if b, _ := r.CurrentBranch(); b != BaseBranch {
		t.Fatalf("current branch = %q, want %q", b, BaseBranch)
	}

	// Create a task branch with a config file and commit it.
	if err := r.CreateBranch("foo", BaseBranch); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(r.Dir, "foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.Dir, "foo", "config.yaml"), []byte("status: pending\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := r.CommitAll("new task: foo"); err != nil {
		t.Fatal(err)
	}

	if !r.BranchExists("foo") {
		t.Fatal("foo branch should exist")
	}

	// Read the config off the branch without checking it out.
	if err := r.Switch(BaseBranch); err != nil {
		t.Fatal(err)
	}
	out, err := r.Show("foo", "foo/config.yaml")
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if out != "status: pending" {
		t.Fatalf("show returned %q", out)
	}

	// CommitAll with no changes is a no-op (must not error).
	if err := r.CommitAll("noop"); err != nil {
		t.Fatalf("empty commit should be a no-op: %v", err)
	}

	// Merge foo into done.
	if err := r.Merge("foo"); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if _, err := os.Stat(filepath.Join(r.Dir, "foo", "config.yaml")); err != nil {
		t.Fatalf("merged file missing on base branch: %v", err)
	}

	branches, err := r.ListBranches()
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %v", branches)
	}
}
