package setup

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnableDisablePreservesZshrc(t *testing.T) {
	dir := t.TempDir()
	zshrc := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(zshrc, []byte("export ORIGINAL=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "twig's bin", "twig")
	plugin := filepath.Join(dir, "twig's bin", "twig.zsh")
	if err := os.Mkdir(filepath.Dir(binary), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plugin, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := Enable(zshrc, binary, plugin); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "export ORIGINAL=1\n") {
		t.Fatalf("existing config changed: %q", content)
	}
	if strings.Count(content, "# >>> twig command >>>") != 1 || strings.Count(content, "# >>> twig completion >>>") != 1 {
		t.Fatalf("enable is not idempotent: %q", content)
	}
	if !strings.Contains(content, "twig() {") || !strings.Contains(content, "source '") {
		t.Fatalf("missing command or plugin: %q", content)
	}
	output, err := exec.Command("/bin/zsh", "-fc", "source "+shellQuote(zshrc)+"; twig hello").CombinedOutput()
	if err != nil || !strings.Contains(string(output), "hello") {
		t.Fatalf("twig function failed: %v\n%s", err, output)
	}
	if err := Disable(zshrc); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(zshrc)
	if err != nil {
		t.Fatal(err)
	}
	content = string(data)
	if strings.Contains(content, "# >>> twig completion >>>") || !strings.Contains(content, "# >>> twig command >>>") {
		t.Fatalf("disable removed the command or kept the plugin: %q", content)
	}
	if err := Disable(zshrc); err != nil {
		t.Fatal(err)
	}
}

func TestDisableWithoutConfig(t *testing.T) {
	if err := Disable(filepath.Join(t.TempDir(), ".zshrc")); err != nil {
		t.Fatal(err)
	}
}

func TestEnableRejectsIncompleteManagedBlock(t *testing.T) {
	zshrc := filepath.Join(t.TempDir(), ".zshrc")
	original := "keep this\n# >>> twig completion >>>\n"
	if err := os.WriteFile(zshrc, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Enable(zshrc, "/tmp/twig", "/tmp/twig.zsh"); err == nil {
		t.Fatal("expected malformed block error")
	}
	data, err := os.ReadFile(zshrc)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != original {
		t.Fatalf("malformed config was changed: %q", data)
	}
}

func TestEnablePreservesZshrcSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real-zshrc")
	zshrc := filepath.Join(dir, ".zshrc")
	if err := os.WriteFile(target, []byte("export ORIGINAL=1\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, zshrc); err != nil {
		t.Fatal(err)
	}
	if err := Enable(zshrc, "/tmp/twig", "/tmp/twig.zsh"); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Lstat(zshrc); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("zshrc symlink was replaced: %v, %v", info, err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("zshrc permissions changed: %v", info.Mode().Perm())
	}
}
