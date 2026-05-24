// Package cmd wires up the cobra command tree for the todo CLI.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	todo "github.com/h3jfc/todo/internal"
	"github.com/spf13/cobra"
)

// newStore is a package-level hook that tests can replace.
var newStore = func() *todo.Store {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot determine home directory:", err)
		os.Exit(1)
	}
	dir := filepath.Join(home, "ToDo")
	s := todo.NewStore(dir)
	if err := s.EnsureDir(); err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot create ToDo directory:", err)
		os.Exit(1)
	}
	return s
}

// completeTodoNames returns the existing todo titles for shell completion.
func completeTodoNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	names, err := newStore().List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeTodoNamesFirstArg completes the first positional arg with todo names
// and disables completion for any subsequent args.
func completeTodoNamesFirstArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeTodoNames(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

// New returns the root cobra.Command for the todo CLI.
func New() *cobra.Command {
	root := &cobra.Command{
		Use:   "todo",
		Short: "A dead-simple file-based todo list",
		Long: `todo manages a directory of Markdown files in ~/ToDo.

Each todo is a plain .md file – create one, delete it when done.
No database, no sync, no friction.`,
		// When invoked with no sub-command and no args: list todos.
		// When invoked with positional args: create a todo.
		Args: cobra.ArbitraryArgs,
		RunE: runRoot,
	}

	root.AddCommand(
		newInitCmd(),
		newSubCmd(),
		newEditCmd(),
		newFinishCmd(),
		newRmCmd(),
	)

	// Cobra attaches `completion` and `help` automatically.
	return root
}

// runRoot handles `todo` (list) and `todo <title> [desc]` (create).
func runRoot(cmd *cobra.Command, args []string) error {
	s := newStore()

	out := cmd.OutOrStdout()
	if len(args) == 0 {
		names, err := s.List()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			fmt.Fprintln(out, "No todos yet. Run `todo <title>` to create one.")
			return nil
		}
		tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "TITLE\tDESCRIPTION")
		for _, n := range names {
			desc, _ := s.Description(n)
			if len(desc) > 100 {
				desc = desc[:100]
			}
			fmt.Fprintf(tw, "%s\t%s\n", n, desc)
		}
		return tw.Flush()
	}

	title := args[0]
	desc := ""
	if len(args) > 1 {
		desc = args[1]
	}
	path, created, err := s.Create(title, desc)
	if err != nil {
		return err
	}
	if created {
		fmt.Fprintf(out, "created %s\n", path)
	} else {
		fmt.Fprintf(out, "already exists: %s\n", path)
	}
	return nil
}
