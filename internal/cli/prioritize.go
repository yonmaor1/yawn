package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/state"
)

var prioritizeCmd = &cobra.Command{
	Use:   "prioritize <name>",
	Short: "add a task to today's todo",
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
		s, err := state.Load(r.Dir)
		if err != nil {
			return err
		}
		s.Add(today(), name)
		if err := s.Save(r.Dir); err != nil {
			return err
		}
		fmt.Printf("prioritized %q for %s\n", name, today())
		return nil
	},
}
