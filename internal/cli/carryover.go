package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/config"
	"github.com/yonmaor/yawn/internal/state"
)

var carryOverCmd = &cobra.Command{
	Use:   "carry-over",
	Short: "carry over yesterday's incomplete tasks into today's todo",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := requireRepo()
		if err != nil {
			return err
		}
		s, err := state.Load(r.Dir)
		if err != nil {
			return err
		}
		td := today()
		prev, ok := s.MostRecentBefore(td)
		if !ok {
			fmt.Println("nothing to carry over")
			return nil
		}
		tasks, err := loadTasks(r)
		if err != nil {
			return err
		}

		var carried int
		for _, name := range s.Get(prev) {
			t, exists := tasks[name]
			if !exists {
				continue // task was deleted
			}
			if t.Status == config.StatusDone || t.Status == config.StatusArchived {
				continue // already complete
			}
			before := len(s.Get(td))
			s.Add(td, name)
			if len(s.Get(td)) > before {
				carried++
			}
		}
		if err := s.Save(r.Dir); err != nil {
			return err
		}
		fmt.Printf("carried over %d task(s) from %s into %s\n", carried, prev, td)
		return nil
	},
}
