package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	commandStart    = "# >>> twig command >>>"
	commandEnd      = "# <<< twig command <<<"
	completionStart = "# >>> twig completion >>>"
	completionEnd   = "# <<< twig completion <<<"
)

func ZshrcPath() (string, error) {
	dir := os.Getenv("ZDOTDIR")
	if dir == "" {
		var err error
		dir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(dir, ".zshrc"), nil
}

func Enable(zshrc, binary, plugin string) error {
	path, content, mode, err := readZshrc(zshrc)
	if err != nil {
		return err
	}
	commandBlock := commandStart + "\n" + commandFunction(binary, plugin) + commandEnd + "\n"
	completionBlock := completionStart + "\nTWIG_BIN=" + shellQuote(binary) + "\nsource " + shellQuote(plugin) + "\n" + completionEnd + "\n"
	content, err = replaceBlock(content, commandStart, commandEnd, commandBlock)
	if err != nil {
		return err
	}
	content, err = replaceBlock(content, completionStart, completionEnd, completionBlock)
	if err != nil {
		return err
	}
	return writeZshrc(path, content, mode)
}

func ShellCode(binary, plugin string) string {
	return "if (( $+functions[_twig_unload] )); then _twig_unload; fi\n" +
		"TWIG_BIN=" + shellQuote(binary) + "\n" + commandFunction(binary, plugin) + "source " + shellQuote(plugin) + "\n"
}

func commandFunction(binary, plugin string) string {
	return "twig() {\n  local action=\"$1\"\n  " + shellQuote(binary) + " \"$@\" || return\n  case \"$action\" in\n    enable) source " + shellQuote(plugin) + " || return ;;\n    disable) if (( $+functions[_twig_unload] )); then _twig_unload; fi ;;\n  esac\n}\n"
}

func Disable(zshrc string) error {
	path, content, mode, err := readZshrc(zshrc)
	if err != nil {
		return err
	}
	updated, err := replaceBlock(content, completionStart, completionEnd, "")
	if err != nil {
		return err
	}
	if updated == content {
		return nil
	}
	return writeZshrc(path, updated, mode)
}

func readZshrc(path string) (string, string, os.FileMode, error) {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	} else if !os.IsNotExist(err) {
		return "", "", 0, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return path, "", 0o600, nil
	}
	if err != nil {
		return "", "", 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", "", 0, err
	}
	return path, string(data), info.Mode().Perm(), nil
}

func replaceBlock(content, start, end, replacement string) (string, error) {
	startCount := strings.Count(content, start)
	endCount := strings.Count(content, end)
	if startCount > 1 || endCount > 1 || startCount != endCount {
		return "", fmt.Errorf("incomplete or duplicate Twig block in .zshrc")
	}
	if startCount == 0 {
		if replacement == "" {
			return content, nil
		}
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return content + replacement, nil
	}
	startIndex := strings.Index(content, start)
	endIndex := strings.Index(content, end)
	if endIndex < startIndex {
		return "", fmt.Errorf("misordered Twig block in .zshrc")
	}
	endIndex += len(end)
	if endIndex < len(content) && content[endIndex] == '\n' {
		endIndex++
	}
	return content[:startIndex] + replacement + content[endIndex:], nil
}

func writeZshrc(path, content string, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".twig-zshrc-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if err := tmp.Chmod(mode); err != nil {
		return err
	}
	if _, err := tmp.WriteString(content); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
