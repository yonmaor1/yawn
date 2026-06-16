package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/config"
	"github.com/yonmaor/yawn/internal/repo"
)

var updateCmd = &cobra.Command{
	Use:   "update <status>",
	Short: "update the current task's status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(args[0])
	},
}

var doneCmd = &cobra.Command{
	Use:   "done",
	Short: "mark the current task done and merge it into its parent",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(config.StatusDone)
	},
}

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "archive the current task",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(config.StatusArchived)
	},
}

// runUpdate sets the current task's status and, when the status is "done",
// merges the task branch into its parent.
func runUpdate(status string) error {
	r, err := requireRepo()
	if err != nil {
		return err
	}
	name := currentTask(r)
	if name == "" {
		return fmt.Errorf("not in a task; `yawn switch <name>` first")
	}

	// The task's branch must be checked out to edit its config.
	if cur, _ := r.CurrentBranch(); cur != name {
		if err := r.Switch(name); err != nil {
			return err
		}
	}

	t, err := config.Load(configPath(r, name))
	if err != nil {
		return err
	}
	t.Status = status
	if err := t.Save(configPath(r, name)); err != nil {
		return err
	}
	if err := r.CommitAll(fmt.Sprintf("status: %s -> %s", name, status)); err != nil {
		return err
	}
	fmt.Printf("%s is now %q\n", name, status)

	if status == config.StatusDone {
		parent := t.Parent
		if parent == "" {
			parent = repo.BaseBranch
		}
		if err := r.Switch(parent); err != nil {
			return err
		}
		if err := r.Merge(name); err != nil {
			return err
		}
		fmt.Printf("merged %q into %q\n", name, parent)
		if currentTask(r) == name {
			fmt.Println("exit this shell to leave the task")
		}
	}
	return nil
}
