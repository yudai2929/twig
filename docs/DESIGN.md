# Twig design

Twig runs in interactive zsh on macOS. ZLE handles input and display; the Go executable collects suggestions.

## Input and display

`zsh/twig.zsh` watches the editing buffer and cursor position and displays suggestions inside the terminal. Suggestion collection runs in a separate process so typing remains responsive while a request is in progress. A generation number discards results for older input. Twig avoids redrawing unchanged suggestions and filters existing results when it can.

Enter inserts the selected suggestion or runs the command when no suggestions are open. Tab invokes zsh's standard completion. The up and down arrows move through suggestions while they are open and browse command history otherwise. `Ctrl-X j` and `Ctrl-X k` also move through suggestions. An icon before each candidate identifies commands, flags, files, directories, and other values.

## Collecting suggestions

`internal/collector/` searches in this order:

1. For the first word, find executables on the current `PATH` and rank commands with zsh completion definitions first.
2. For arguments, run `list-choices` in an isolated interactive zsh and record candidates passed to `compadd`.
3. If a CLI has no zsh completion definition but its help describes a completion generator, load its generated zsh completion.
4. If no candidates are found, parse commands and options from the CLI's `--help` output. Follow a referenced `help` page when the CLI points to one.

Twig keeps the candidate order from existing zsh completions when possible. It has no fixed lists for individual commands.

## Limits

The collector starts zsh with `zsh -f`. It can use definitions on the completion path but cannot share functions or variables loaded only in the user's current shell. zsh has no API for retrieving all candidates as structured data, so Twig observes calls to `compadd`. Descriptions, suffixes, quoting, or order may differ from standard completion for some definitions. Tab always uses the current shell's standard completion.

Twig does not send the command line to an external service. However, existing zsh completions and CLI completion generators may run external commands or make network requests during collection. Automatic suggestions can be disabled per command.
