package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/repo"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize yawn",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := repo.Resolve()
		if err != nil {
			return err
		}
		if r.Exists() {
			fmt.Printf("yawn already initialized at %s\n", r.Dir)
			return nil
		}
		if err := r.Init(); err != nil {
			return err
		}
		fmt.Printf("initialized yawn at %s (base branch %q)\n", r.Dir, repo.BaseBranch)
		return nil
	},
}
