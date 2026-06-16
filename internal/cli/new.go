package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/config"
	"github.com/yonmaor/yawn/internal/repo"
)

var newIn string

var newCmd = &cobra.Command{
	Use:   "new [--in <parent>] <name>",
	Short: "create a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := requireRepo()
		if err != nil {
			return err
		}
		name := args[0]
		if err := validName(name); err != nil {
			return err
		}

		parent := newIn
		if parent == "" {
			parent = repo.BaseBranch
		}
		if r.BranchExists(name) {
			return fmt.Errorf("task %q already exists", name)
		}
		if !r.BranchExists(parent) {
			return fmt.Errorf("parent task %q does not exist", parent)
		}

		prev, err := r.CurrentBranch()
		if err != nil {
			return err
		}
		if err := r.CreateBranch(name, parent); err != nil {
			return err
		}

		// Scaffold <name>/README.md and <name>/config.yaml.
		dir := r.Dir + string(os.PathSeparator) + name
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(readmePath(r, name), []byte("# "+name+"\n"), 0o644); err != nil {
			return err
		}
		if err := config.Default(parent).Save(configPath(r, name)); err != nil {
			return err
		}

		// Let the user fill in directory/rc/cleanup.
		fmt.Printf("opening %s — fill in directory, rc, and cleanup\n", configPath(r, name))
		if err := openEditor(configPath(r, name)); err != nil {
			fmt.Fprintf(os.Stderr, "yawn: could not open editor (%v); edit %s manually\n", err, configPath(r, name))
		}

		if err := r.CommitAll("new task: " + name); err != nil {
			return err
		}
		// Return to where we were (the base branch when at rest).
		if err := r.Switch(prev); err != nil {
			return err
		}
		fmt.Printf("created task %q off %q\n", name, parent)
		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newIn, "in", "", "parent task to branch off (default: base branch)")
}

func validName(name string) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if name == repo.BaseBranch {
		return fmt.Errorf("task name %q is reserved", repo.BaseBranch)
	}
	if strings.ContainsAny(name, " \t/\\:~^?*[") {
		return fmt.Errorf("invalid task name %q: avoid spaces and git-unsafe characters", name)
	}
	return nil
}
