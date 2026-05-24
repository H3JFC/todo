package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// editorFunc is replaceable in tests.
var editorFunc = openEditor

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <title> [description]",
		Short: "Open a todo in $EDITOR",
		Long: `Open ~/ToDo/<title>.md in $EDITOR (falls back to vi).

The file is created from the standard template first if it does not exist.
An optional description is only used during that initial creation.`,
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runEdit,
		ValidArgsFunction: completeTodoNamesFirstArg,
		Example:           "  todo edit buy-milk\n  todo edit buy-milk \"get oat milk\"",
	}
}

func runEdit(cmd *cobra.Command, args []string) error {
	title := args[0]
	desc := ""
	if len(args) > 1 {
		desc = args[1]
	}
	s := newStore()
	path, _, err := s.Create(title, desc)
	if err != nil {
		return err
	}
	return editorFunc(path)
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}
	return nil
}
