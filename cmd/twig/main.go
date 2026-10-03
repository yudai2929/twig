package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/yudai2929/twig"
	"github.com/yudai2929/twig/internal/collector"
	"github.com/yudai2929/twig/internal/completion"
	"github.com/yudai2929/twig/internal/setup"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "enable":
		if len(os.Args) != 2 && !(len(os.Args) == 3 && os.Args[2] == "--shell") {
			usage()
			os.Exit(2)
		}
		if err := enable(len(os.Args) == 3); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	case "disable":
		if len(os.Args) != 2 {
			usage()
			os.Exit(2)
		}
		if err := disable(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	case "complete":
	default:
		usage()
		os.Exit(2)
	}
	complete()
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: twig enable [--shell] | disable | complete --buffer TEXT --cursor N [--cwd DIR] [--json]")
}

func enable(emitShell bool) error {
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	binary, err = filepath.EvalSymlinks(binary)
	if err != nil {
		return err
	}
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	plugin := filepath.Join(dataDir, "twig", "twig.zsh")
	if err := os.MkdirAll(filepath.Dir(plugin), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(plugin, twig.PluginScript(), 0o600); err != nil {
		return err
	}
	zshrc, err := setup.ZshrcPath()
	if err != nil {
		return err
	}
	if err := setup.Enable(zshrc, binary, plugin); err != nil {
		return err
	}
	if emitShell {
		fmt.Print(setup.ShellCode(binary, plugin))
		return nil
	}
	fmt.Printf("Twig enabled in %s. New zsh sessions will load completion.\n", zshrc)
	return nil
}

func disable() error {
	zshrc, err := setup.ZshrcPath()
	if err != nil {
		return err
	}
	if err := setup.Disable(zshrc); err != nil {
		return err
	}
	fmt.Printf("Twig disabled in %s. New zsh sessions will skip completion.\n", zshrc)
	return nil
}

func complete() {
	flags := flag.NewFlagSet("complete", flag.ExitOnError)
	buffer := flags.String("buffer", "", "zsh editing buffer")
	cursor := flags.Int("cursor", -1, "cursor byte offset")
	dir := flags.String("cwd", ".", "working directory")
	jsonOutput := flags.Bool("json", false, "emit JSON instead of NUL records")
	if err := flags.Parse(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *cursor == -1 {
		*cursor = utf8.RuneCountInString(*buffer)
	}
	if isDisabled(*buffer) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	candidates, err := collector.Collect(ctx, collector.Request{Buffer: *buffer, Cursor: *cursor, Dir: *dir})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *jsonOutput {
		if err := json.NewEncoder(os.Stdout).Encode(candidates); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if _, err := os.Stdout.Write(completion.EncodeRecords(candidates)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isDisabled(buffer string) bool {
	command := strings.Fields(buffer)
	if len(command) == 0 {
		return false
	}
	for _, disabled := range strings.Split(os.Getenv("TWIG_DISABLED_COMMANDS"), ",") {
		if strings.TrimSpace(disabled) == command[0] {
			return true
		}
	}
	return false
}
