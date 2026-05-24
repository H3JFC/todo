package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <title> [description]",
		Short: "Create a new todo",
		Long: `Create ~/ToDo/<title>.md with a Markdown template.

If the file already exists it is left untouched.`,
		Args:    cobra.RangeArgs(1, 2),
		RunE:    runInit,
		Example: "  todo init buy-milk\n  todo init buy-milk \"get oat milk from the co-op\"",
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	title := args[0]
	desc := ""
	if len(args) > 1 {
		desc = args[1]
	}
	s := newStore()
	path, created, err := s.Create(title, desc)
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if created {
		fmt.Fprintf(out, "created %s\n", path)
	} else {
		fmt.Fprintf(out, "already exists: %s\n", path)
	}
	return nil
}
