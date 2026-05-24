package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFinishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "finish <title>",
		Short: "Archive a completed todo",
		Long: `Move ~/ToDo/<title>.md to ~/ToDo/Finished/YYYY.MM.DD<title>.md.

Use this to keep a dated record of completed work without cluttering
the active list.`,
		Args:              cobra.ExactArgs(1),
		RunE:              runFinish,
		ValidArgsFunction: completeTodoNamesFirstArg,
		Example:           "  todo finish buy-milk",
	}
}

func runFinish(cmd *cobra.Command, args []string) error {
	s := newStore()
	title := args[0]
	if err := s.Finish(title); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "finished %s\n", title)
	return nil
}
