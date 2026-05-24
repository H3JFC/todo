# todo

A dead-simple, file-based todo list. Each task is a Markdown file in `~/ToDo/`.
Create it, delete it when done. No database, no sync, no friction.

```
todo buy-milk          # create ~/ToDo/buy-milk.md
todo sub buy-milk "check dates"   # append a sub-task
todo                   # list active todos
todo finish buy-milk   # archive to ~/ToDo/Finished/
todo rm buy-milk       # delete without archiving
todo edit buy-milk     # open in $EDITOR
```

---

## Installation

```bash
git clone https://github.com/h3jfc/todo
cd todo
go install .
```

Requires Go 1.21+.

---

## Commands

| Command                    | Description                                       |
| -------------------------- | ------------------------------------------------- |
| `todo`                     | List all active todos                             |
| `todo <title> [desc]`      | Create a new todo (shorthand for `init`)          |
| `todo init <title> [desc]` | Create a new todo explicitly                      |
| `todo sub <title> <task>`  | Append `- [ ] <task>` to a todo                   |
| `todo edit <title> [desc]` | Open in `$EDITOR`; create first if missing        |
| `todo finish <title>`      | Archive to `~/ToDo/Finished/YYYY.MM.DD<title>.md` |
| `todo rm <title>`          | Delete without archiving                          |
| `todo help`                | Show help                                         |
| `todo completion <shell>`  | Generate shell completion script                  |

Titles may omit the `.md` extension – `todo rm buy-milk` and `todo rm buy-milk.md` are identical.

---

## Shell Auto-completion

Cobra generates completion scripts automatically via `todo completion <shell>`.

### Bash

```bash
# One-shot (current session)
source <(todo completion bash)

# Permanent
todo completion bash > /etc/bash_completion.d/todo
# or, for user-local install:
todo completion bash > ~/.local/share/bash-completion/completions/todo
```

### Zsh

```zsh
# Ensure completion is initialised in your ~/.zshrc:
autoload -U compinit && compinit

# One-shot
source <(todo completion zsh)

# Permanent (pick a directory already on your $fpath)
todo completion zsh > "${fpath[1]}/_todo"
```

### Fish

```fish
todo completion fish | source

# Permanent
todo completion fish > ~/.config/fish/completions/todo.fish
```

### PowerShell

```powershell
# One-shot
todo completion powershell | Out-String | Invoke-Expression

# Permanent – add to your $PROFILE:
todo completion powershell >> $PROFILE
```

### Completing todo titles

Sub-commands like `finish`, `rm`, `edit`, and `sub` all take a todo title as
their first argument. To teach your shell to complete _existing_ todo titles
you can wrap the generated completion with a small helper.

**Bash example** – add after sourcing the completion script:

```bash
_todo_titles() {
  local titles
  titles=$(ls "$HOME/ToDo" 2>/dev/null | sed 's/\.md$//')
  COMPREPLY=($(compgen -W "$titles" -- "${COMP_WORDS[COMP_CWORD]}"))
}
# Override the positional-arg completion for commands that take a title.
for _cmd in finish rm edit sub; do
  complete -F _todo_titles todo "$_cmd"
done
```

**Zsh example** – add to `~/.zshrc` after `compinit`:

```zsh
_todo_title_complete() {
  local titles
  titles=(${(f)"$(ls $HOME/ToDo 2>/dev/null | sed 's/\.md$//')"})
  _describe 'todo title' titles
}
compdef _todo_title_complete 'todo finish' 'todo rm' 'todo edit' 'todo sub'
```

**Fish example** – add to `~/.config/fish/completions/todo.fish`:

```fish
function __todo_titles
    ls $HOME/ToDo 2>/dev/null | string replace -r '\.md$' ''
end

for cmd in finish rm edit sub
    complete -c todo -n "__fish_seen_subcommand_from $cmd" \
             -a "(__todo_titles)" --no-files
end
```

---

## File layout

```
~/ToDo/
├── buy-milk.md
├── fix-laptop.md
└── Finished/
    ├── 2024.06.14read-book.md
    └── 2024.06.15buy-milk.md
```

Each todo file follows this template:

```markdown
# buy-milk

## Description

get oat milk from the co-op

## Sub-Tasks

- [ ] check expiry dates
- [ ] grab two cartons
```

---

## Development

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Build
go build -o todo .
```

The core file-system logic lives in `internal/todo` and is tested against an
in-memory fake (`FS` interface). The `cmd` package wires up cobra commands and
injects the real store at runtime; the same injection point is used in
`cmd_test.go` to keep tests fast and hermetic.
