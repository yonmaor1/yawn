package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/config"
	"github.com/yonmaor/yawn/internal/shell"
)

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "switch to a task (opens a subshell)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := requireRepo()
		if err != nil {
			return err
		}
		name := args[0]
		if !r.BranchExists(name) {
			return fmt.Errorf("task %q does not exist", name)
		}

		prev, err := r.CurrentBranch()
		if err != nil {
			return err
		}
		if err := r.Switch(name); err != nil {
			return err
		}

		// Read the task config now that its branch is checked out.
		t, err := config.Load(configPath(r, name))
		if err != nil {
			r.Switch(prev)
			return err
		}
		workdir := t.Directory
		if workdir == "" {
			workdir = r.Dir
		} else if info, err := os.Stat(workdir); err != nil || !info.IsDir() {
			r.Switch(prev)
			return fmt.Errorf("task directory %q is not accessible; fix it with `yawn open`/config", workdir)
		}

		fmt.Printf("switched to %q — exit the shell to return\n", name)
		runErr := shell.Run(workdir, name, t.RC, t.Cleanup)

		// Restore the previous branch regardless of how the shell exited.
		if err := r.Switch(prev); err != nil {
			return err
		}
		return runErr
	},
}
