package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnableDisableInNewZsh(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("this zsh integration test targets macOS")
	}
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	bundle := t.TempDir()
	if err := os.Mkdir(filepath.Join(bundle, "bin"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(bundle, "zsh"), 0o700); err != nil {
		t.Fatal(err)
	}
	plugin, err := os.ReadFile(filepath.Join(root, "zsh", "twig.zsh"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "zsh", "twig.zsh"), plugin, 0o600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(bundle, "bin", "twig")
	build := exec.Command("go", "build", "-o", binary, "./cmd/twig")
	build.Dir = root
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "go-cache"))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	resolvedBinary, err := filepath.EvalSymlinks(binary)
	if err != nil {
		t.Fatal(err)
	}
	zdotdir := t.TempDir()
	dataDir := t.TempDir()
	env := append(os.Environ(), "ZDOTDIR="+zdotdir, "XDG_DATA_HOME="+dataDir)
	run := func(command string) {
		t.Helper()
		cmd := exec.Command(binary, command)
		cmd.Env = env
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("twig %s: %v\n%s", command, err, output)
		}
	}
	checkZsh := func(want string) {
		t.Helper()
		cmd := exec.Command("/bin/zsh", "-ic", "whence -w twig; print -r -- TWIG_BIN=$TWIG_BIN WIDGET=${+functions[_twig_accept]}")
		cmd.Env = env
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("zsh: %v\n%s", err, output)
		}
		if !strings.Contains(string(output), want) {
			t.Fatalf("zsh output missing %q: %s", want, output)
		}
	}
	run("enable")
	checkZsh("TWIG_BIN=" + resolvedBinary)
	checkZsh("WIDGET=1")
	bindings := exec.Command("/bin/zsh", "-ic", "bindkey '^[[A'; bindkey '^Xj'; bindkey '^Xk'")
	bindings.Env = env
	if output, err := bindings.CombinedOutput(); err != nil ||
		!strings.Contains(string(output), "up-line-or-history") ||
		!strings.Contains(string(output), "_twig_down") ||
		!strings.Contains(string(output), "_twig_up") {
		t.Fatalf("history and completion bindings conflict: %v\n%s", err, output)
	}
	reload := exec.Command("/bin/zsh", "-ic", "twig disable; twig enable; bindkey '^M'")
	reload.Env = env
	if output, err := reload.CombinedOutput(); err != nil || !strings.Contains(string(output), "_twig_accept") {
		t.Fatalf("same-shell reload failed: %v\n%s", err, output)
	}
	for _, command := range []string{"twig disable", "twig enable", "twig disable"} {
		cmd := exec.Command("/bin/zsh", "-ic", command+"; bindkey '^M'")
		cmd.Env = env
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v\n%s", command, err, output)
		}
		if command == "twig enable" {
			if !strings.Contains(string(output), "_twig_accept") {
				t.Fatalf("completion did not start in current shell: %s", output)
			}
			checkZsh("WIDGET=1")
		} else {
			if strings.Contains(string(output), "_twig_accept") {
				t.Fatalf("completion did not stop in current shell: %s", output)
			}
			checkZsh("WIDGET=0")
		}
	}
	checkZsh("WIDGET=0")
	data, err := os.ReadFile(filepath.Join(zdotdir, ".zshrc"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "# >>> twig completion >>>") {
		t.Fatalf("completion still enabled: %s", data)
	}
	firstShellDir := t.TempDir()
	firstShell := exec.Command("/bin/zsh", "-ic", "source "+shellQuote(filepath.Join(bundle, "zsh", "twig.zsh"))+"; eval \"$("+shellQuote(binary)+" enable --shell)\"; bindkey '^M'; whence -w twig")
	firstShell.Env = append(os.Environ(), "ZDOTDIR="+firstShellDir, "XDG_DATA_HOME="+dataDir)
	if output, err := firstShell.CombinedOutput(); err != nil || !strings.Contains(string(output), "_twig_accept") || !strings.Contains(string(output), "twig: function") {
		t.Fatalf("first-shell enable failed: %v\n%s", err, output)
	}
}

func TestStandaloneBinaryEnable(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("this zsh integration test targets macOS")
	}
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	dataDir := filepath.Join(t.TempDir(), "data with spaces")
	zdotdir := t.TempDir()
	binDir := filepath.Join(t.TempDir(), "bin")
	env := append(os.Environ(), "XDG_DATA_HOME="+dataDir, "ZDOTDIR="+zdotdir, "GOBIN="+binDir, "GOCACHE="+filepath.Join(t.TempDir(), "go-cache"))
	install := exec.Command("go", "install", "./cmd/twig")
	install.Dir = root
	install.Env = env
	if output, err := install.CombinedOutput(); err != nil {
		t.Fatalf("go install: %v\n%s", err, output)
	}
	binary := filepath.Join(binDir, "twig")
	for range 2 {
		enable := exec.Command(binary, "enable")
		enable.Env = env
		if output, err := enable.CombinedOutput(); err != nil {
			t.Fatalf("standalone enable: %v\n%s", err, output)
		}
	}
	plugin := filepath.Join(dataDir, "twig", "twig.zsh")
	for _, path := range []string{binary, plugin} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("installed file %s: %v", path, err)
		}
	}
	config, err := os.ReadFile(filepath.Join(zdotdir, ".zshrc"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(config), "# >>> twig completion >>>") != 1 {
		t.Fatalf("installer duplicated zsh configuration: %s", config)
	}
	check := exec.Command("/bin/zsh", "-ic", "whence -w twig; print -r -- WIDGET=${+functions[_twig_accept]}; twig complete --buffer 'git c' --json >/dev/null")
	check.Env = env
	if output, err := check.CombinedOutput(); err != nil || !strings.Contains(string(output), "twig: function") || !strings.Contains(string(output), "WIDGET=1") {
		t.Fatalf("installed Twig unavailable in new zsh: %v\n%s", err, output)
	}
}
