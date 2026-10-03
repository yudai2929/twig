package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestInteractiveCompletion(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("this PTY test uses macOS script")
	}
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "twig")
	build := exec.Command("go", "build", "-o", binary, "./cmd/twig")
	build.Dir = root
	build.Env = append(os.Environ(), "GOCACHE="+filepath.Join(t.TempDir(), "go-cache"))
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "script", "-q", "/dev/null", "/bin/zsh", "-f", "-i")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { stdin.Close(); _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	var output bytes.Buffer
	var mu sync.Mutex
	go func() {
		var chunk [4096]byte
		for {
			n, err := stdout.Read(chunk[:])
			if n > 0 {
				mu.Lock()
				output.Write(chunk[:n])
				mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	getOutput := func() string {
		mu.Lock()
		defer mu.Unlock()
		return output.String()
	}
	setup := fmt.Sprintf("export TWIG_BIN=%s; source %s; print -r -- TWIG_READY\r", shellQuote(binary), shellQuote(filepath.Join(root, "zsh", "twig.zsh")))
	if _, err := io.WriteString(stdin, setup); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "TWIG_READY")
	if _, err := io.WriteString(stdin, "g"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "git")
	if _, err := io.WriteString(stdin, "i"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "› git")
	start := len(getOutput())
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› add")
	if _, err := io.WriteString(stdin, "c"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "commit  record changes")
	if _, err := io.WriteString(stdin, strings.Repeat("\x18j", 5)); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "› commit")
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "git commit")
	if _, err := io.WriteString(stdin, "\x1b"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(600 * time.Millisecond)
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "fatal: not a git repository")

	start = len(getOutput())
	if _, err := io.WriteString(stdin, "git rem"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› remote")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› add")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x03"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")

	start = len(getOutput())
	if _, err := io.WriteString(stdin, "g"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "git")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x7f"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if strings.Contains(getOutput()[start:], "› ") {
		t.Fatal("completion menu remained after deleting all input")
	}
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "git\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "usage: git")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "print -r -- TWIG_HISTORY\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "TWIG_HISTORY\r\n")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "g"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "git")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x1b[A"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "print -r -- TWIG_HISTORY")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x03"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")

	fileDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(fileDir, "hello world.txt"), []byte("TWIG_FILE_CONTENT\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(stdin, "cd "+shellQuote(fileDir)+"; print -r -- TWIG_CD_READY\r"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "TWIG_CD_READY")
	if _, err := io.WriteString(stdin, "cat hel"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "hello world.txt")
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "cat hello\\ world.txt")
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "TWIG_FILE_CONTENT")

	if _, err := io.WriteString(stdin, "git cherry-p\t"); err != nil {
		t.Fatal(err)
	}
	waitOutput(t, ctx, getOutput, "git cherry-pick")
	time.Sleep(250 * time.Millisecond)
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x03"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")

	start = len(getOutput())
	if _, err := io.WriteString(stdin, "git c"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› checkout")
	if _, err := io.WriteString(stdin, strings.Repeat("\x18j", 5)); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› commit")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "git commit")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, " "); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(stdin, "-"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "--all")

	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x03"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "git c"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› checkout")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "git checkout")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x7f"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› checkout")

	for _, scenario := range []struct {
		command string
		want    string
		binary  string
	}{
		{command: "gh pr", want: "› pr  Manage pull requests", binary: "gh"},
		{command: "mise u", want: "  use  ", binary: "mise"},
		{command: "npm ru", want: "› run", binary: "npm"},
		{command: "go b", want: "build  compile packages and dependencies", binary: "go"},
		{command: "go mod t", want: "tidy  add missing and remove unused modules", binary: "go"},
		{command: "cargo b", want: "build  Compile", binary: "cargo"},
		{command: "cargo build --r", want: "--release", binary: "cargo"},
	} {
		if _, err := exec.LookPath(scenario.binary); err != nil {
			continue
		}
		start = len(getOutput())
		if _, err := io.WriteString(stdin, "\x03"); err != nil {
			t.Fatal(err)
		}
		waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")
		start = len(getOutput())
		if _, err := io.WriteString(stdin, scenario.command); err != nil {
			t.Fatal(err)
		}
		waitOutputAfter(t, ctx, getOutput, start, scenario.want)
	}

	if _, err := exec.LookPath("docker"); err == nil {
		start = len(getOutput())
		if _, err := io.WriteString(stdin, "\x03"); err != nil {
			t.Fatal(err)
		}
		waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")
		start = len(getOutput())
		if _, err := io.WriteString(stdin, "docker ru"); err != nil {
			t.Fatal(err)
		}
		waitOutputAfter(t, ctx, getOutput, start, "Create and run a new container")
		start = len(getOutput())
		if _, err := io.WriteString(stdin, "\r"); err != nil {
			t.Fatal(err)
		}
		waitOutputAfter(t, ctx, getOutput, start, "docker run")
	}

	// A command without a zsh completion definition still gets nested help candidates.
	cliDir := t.TempDir()
	cli := "#!/bin/sh\ncase \"$*\" in\n  '--help') printf 'Commands:\\n  build  Build it\\n' ;;\n  'build --help') printf 'Commands:\\n  release  Release it\\n' ;;\nesac\n"
	if err := os.WriteFile(filepath.Join(cliDir, "twigfixture"), []byte(cli), 0o700); err != nil {
		t.Fatal(err)
	}
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\x03"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "export PATH="+shellQuote(cliDir)+":$PATH; print -r -- TWIG_FIXTURE_READY\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "TWIG_FIXTURE_READY\r\n")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "twigfixture b"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› build  Build it")
	start = len(getOutput())
	if _, err := io.WriteString(stdin, "\r"); err != nil {
		t.Fatal(err)
	}
	waitOutputAfter(t, ctx, getOutput, start, "› release  Release it")

	if _, err := exec.LookPath("codex"); err == nil {
		start = len(getOutput())
		if _, err := io.WriteString(stdin, "\x03"); err != nil {
			t.Fatal(err)
		}
		waitOutputAfter(t, ctx, getOutput, start, "\x1b[?2004h")
		start = len(getOutput())
		if _, err := io.WriteString(stdin, "codex ex"); err != nil {
			t.Fatal(err)
		}
		waitOutputAfter(t, ctx, getOutput, start, "› exec")
	}
}

func waitOutput(t *testing.T, ctx context.Context, output func() string, want string) {
	waitOutputAfter(t, ctx, output, 0, want)
}

func waitOutputAfter(t *testing.T, ctx context.Context, output func() string, start int, want string) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()
	for {
		current := output()
		if len(current) >= start && strings.Contains(current[start:], want) {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("waiting for %q: %v\noutput:\n%s", want, ctx.Err(), output())
		case <-deadline.C:
			current := output()
			if start < len(current) {
				current = current[start:]
			}
			t.Fatalf("waiting for %q timed out\nrecent output:\n%s", want, current)
		case <-ticker.C:
		}
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
