package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "open the current task's readme",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := requireRepo()
		if err != nil {
			return err
		}
		name := currentTask(r)
		if name == "" {
			return fmt.Errorf("not in a task; `yawn switch <name>` first")
		}
		// Make sure the task's branch is checked out so its files are present.
		if cur, _ := r.CurrentBranch(); cur != name {
			if err := r.Switch(name); err != nil {
				return err
			}
		}

		// Append a timestamped heading, then open the editor.
		f, err := os.OpenFile(readmePath(r, name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(f, "\n## %s\n", timestamp()); err != nil {
			f.Close()
			return err
		}
		f.Close()

		if err := openEditor(readmePath(r, name)); err != nil {
			return err
		}
		return r.CommitAll("open: " + name + " " + timestamp())
	},
}
