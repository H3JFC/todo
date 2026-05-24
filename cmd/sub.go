package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSubCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sub <title> <task>",
		Short: "Append a sub-task to a todo",
		Long: `Append a GFM checkbox item ("- [ ] <task>") to ~/ToDo/<title>.md.

The todo file is created with an empty description if it does not exist yet.`,
		Args:              cobra.ExactArgs(2),
		RunE:              runSub,
		ValidArgsFunction: completeTodoNamesFirstArg,
		Example:           "  todo sub buy-milk \"check expiry dates\"",
	}
}

func runSub(cmd *cobra.Command, args []string) error {
	title, task := args[0], args[1]
	s := newStore()
	if err := s.AddSubTask(title, task); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "added sub-task to %s\n", title)
	return nil
}
