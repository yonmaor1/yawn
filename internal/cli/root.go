// Package cli wires up the yawn command-line interface.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yonmaor/yawn/internal/repo"
)

var rootCmd = &cobra.Command{
	Use:           "yawn",
	Short:         "a lazy task management CLI",
	Long:          "yawn is a bare-bones task management system that lives in your terminal, backed by git.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "yawn:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		initCmd,
		newCmd,
		switchCmd,
		openCmd,
		prioritizeCmd,
		listCmd,
		updateCmd,
		doneCmd,
		archiveCmd,
		carryOverCmd,
	)
}

// requireRepo resolves the repo and errors if yawn has not been initialized.
func requireRepo() (*repo.Repo, error) {
	r, err := repo.Resolve()
	if err != nil {
		return nil, err
	}
	if !r.Exists() {
		return nil, fmt.Errorf("not initialized; run `yawn init` first (looked in %s)", r.Dir)
	}
	return r, nil
}
