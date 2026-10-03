# Development and testing

## Development environment

The root `mise.toml` specifies the Go version. Twig has no external Go libraries.

```sh
mise install
mise exec -- go build -o ./bin/twig ./cmd/twig
```

User installation instructions are in the [README](../README.md). The release binary embeds the zsh plugin; `twig enable` extracts it to `${XDG_DATA_HOME:-$HOME/.local/share}/twig/twig.zsh`. To try a development build in the current interactive zsh, run `eval "$(./bin/twig enable --shell)"` after building.

## Tests

```sh
mise exec -- go test -p 1 ./...
mise exec -- go test -race -p 1 ./...
mise exec -- go vet ./...
```

The `e2e` tests start interactive zsh through a macOS PTY. They check suggestion display, selection, execution, Tab, Space, file-name quoting, and suggestions after accepting a candidate.

To inspect only the collector output:

```sh
./bin/twig complete --buffer 'git c' --json
```

## Code map

- `zsh/twig.zsh`: Input monitoring, suggestion display, and key bindings
- `internal/collector/`: Candidate collection, observing zsh completion, and parsing help text
- `internal/completion/`: Encoding and filtering candidates
- `internal/setup/`: Enabling and disabling Twig in zsh configuration
- `e2e/`: Integration tests with interactive zsh
