package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <title>",
		Short: "Delete a todo without archiving",
		Long: `Permanently delete ~/ToDo/<title>.md.

Unlike `+"`todo finish`"+`, no archive copy is kept. Use this for todos that were
created by mistake or are no longer relevant.`,
		Args:              cobra.ExactArgs(1),
		RunE:              runRm,
		ValidArgsFunction: completeTodoNamesFirstArg,
		Example:           "  todo rm buy-milk",
	}
}

func runRm(cmd *cobra.Command, args []string) error {
	s := newStore()
	title := args[0]
	if err := s.Remove(title); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", title)
	return nil
}
