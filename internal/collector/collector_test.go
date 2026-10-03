package collector

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCollectGitSubcommands(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "git c", Cursor: len("git c"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, candidate := range got {
		if candidate.Value == "commit" {
			found = true
			if !strings.Contains(candidate.Description, "record") {
				t.Fatalf("missing zsh description: %+v", candidate)
			}
		}
	}
	if !found {
		t.Fatalf("commit missing from %d candidates: %+v", len(got), got)
	}
}

func TestCollectCommandNames(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "g", Cursor: 1, Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for index, candidate := range got {
		if candidate.Value == "git" {
			if index >= 8 {
				t.Fatalf("git is outside the visible menu: index %d", index)
			}
			return
		}
	}
	t.Fatalf("git missing from command candidates: %+v", got)
}

func TestCollectExactCommandCanRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "git", Cursor: len("git"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("exact command should be executable without menu: %+v", got)
	}
}

func TestCollectGitRemoteCanContinue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "git rem", Cursor: len("git rem"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "remote" {
			if candidate.Suffix != " " {
				t.Fatalf("git remote should insert a space before deeper completion: %+v", candidate)
			}
			return
		}
	}
	t.Fatalf("git remote missing: %+v", got)
}

func TestCollectGoModSubcommands(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not on PATH")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "go mod t", Cursor: len("go mod t"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "tidy" {
			return
		}
	}
	t.Fatalf("go mod tidy missing: %+v", got)
}

func TestCollectCargoBuildFlags(t *testing.T) {
	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "cargo build --r", Cursor: len("cargo build --r"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "--release" {
			return
		}
	}
	t.Fatalf("cargo build --release missing: %+v", got)
}

func TestCollectGoTestFlags(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not on PATH")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "go test -v", Cursor: len("go test -v"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "-v" {
			return
		}
	}
	t.Fatalf("go test -v missing: %+v", got)
}

func TestCollectGoSubcommands(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not on PATH")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "go b", Cursor: len("go b"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "build" {
			return
		}
	}
	t.Fatalf("go build missing from candidates: %+v", got)
}

func TestCollectCargoSubcommands(t *testing.T) {
	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "cargo b", Cursor: len("cargo b"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "build" {
			return
		}
	}
	t.Fatalf("cargo build missing from candidates: %+v", got)
}

func TestCollectDockerSubcommands(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "docker ru", Cursor: len("docker ru"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "run" {
			return
		}
	}
	t.Fatalf("docker run missing from candidates: %+v", got)
}

func TestCollectKubectlSubcommands(t *testing.T) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		t.Skip("kubectl is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "kubectl ge", Cursor: len("kubectl ge"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "get" {
			return
		}
	}
	t.Fatalf("kubectl get missing from candidates: %+v", got)
}

func TestCollectGHSubcommands(t *testing.T) {
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "gh pr", Cursor: len("gh pr"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "pr" {
			return
		}
	}
	t.Fatalf("gh pr missing from candidates: %+v", got)
}

func TestCollectMiseSubcommands(t *testing.T) {
	if _, err := exec.LookPath("mise"); err != nil {
		t.Skip("mise is not installed")
	}
	if _, err := exec.LookPath("usage"); err != nil {
		t.Skip("mise completion requires usage")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "mise u", Cursor: len("mise u"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "use" {
			return
		}
	}
	t.Fatalf("mise use missing from candidates: %+v", got)
}

func TestCollectNPMSubcommands(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "npm r", Cursor: len("npm r"), Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "run" {
			return
		}
	}
	t.Fatalf("npm run missing from candidates: %+v", got)
}

func TestCollectFromCustomFpath(t *testing.T) {
	dir := t.TempDir()
	completion := "#compdef twigtest\nlocal -a values\nvalues=(alpha beta)\ncompadd -a values\n"
	if err := os.WriteFile(filepath.Join(dir, "_twigtest"), []byte(completion), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "twigtest"), []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	base, err := exec.Command("zsh", "-fc", "print -r -- ${(j.:.)fpath}").Output()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FPATH", dir+":"+strings.TrimSpace(string(base)))
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "twigtest al", Cursor: len("twigtest al"), Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "alpha" {
			return
		}
	}
	t.Fatalf("custom completion missing: %+v", got)
}

func TestCollectPathWithSpaces(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello world.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "cat hel", Cursor: len("cat hel"), Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "hello world.txt" {
			return
		}
	}
	t.Fatalf("file missing from candidates: %+v", got)
}

func TestCursorUsesCharacters(t *testing.T) {
	buffer := "echo あ file"
	got, err := byteOffset(buffer, len([]rune("echo あ")))
	if err != nil {
		t.Fatal(err)
	}
	if got != len("echo あ") {
		t.Fatalf("got byte offset %d, want %d", got, len("echo あ"))
	}
}

func TestCollectGenericHelpSubcommandsAndFlags(t *testing.T) {
	dir := t.TempDir()
	cli := `#!/bin/sh
case "$*" in
  "--help") cat <<'EOF'
Usage: sample [COMMAND]
Commands:
  build       Build the project
  deploy      Deploy it
Options:
  -v, --verbose  Print more details
EOF
  ;;
  "build --help") cat <<'EOF'
Usage: sample build [COMMAND]
Commands:
  release     Make a release
Options:
  -r, --release  Optimize output
EOF
  ;;
  *) exit 2 ;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "sample"), []byte(cli), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	for _, tc := range []struct{ buffer, want, description string }{
		{"sample bu", "build", "Build the project"},
		{"sample build ", "release", "Make a release"},
		{"sample build re", "release", "Make a release"},
		{"sample build --r", "--release", "Optimize output"},
	} {
		t.Run(tc.buffer, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			got, err := Collect(ctx, Request{Buffer: tc.buffer, Cursor: len(tc.buffer), Dir: dir})
			if err != nil {
				t.Fatal(err)
			}
			for _, candidate := range got {
				if candidate.Value == tc.want && strings.Contains(candidate.Description, tc.description) {
					return
				}
			}
			t.Fatalf("%s missing from candidates: %+v", tc.want, got)
		})
	}
}

func TestCollectCodexSubcommands(t *testing.T) {
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, tc := range []struct{ buffer, want string }{
		{"codex ex", "exec"},
		{"codex exec re", "resume"},
		{"codex exec --m", "--model"},
	} {
		got, err := Collect(ctx, Request{Buffer: tc.buffer, Cursor: len(tc.buffer), Dir: t.TempDir()})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, candidate := range got {
			if candidate.Value == tc.want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s missing from %s: %+v", tc.want, tc.buffer, got)
		}
	}
}

func TestParseHelpFormatsWithoutCommandNames(t *testing.T) {
	for _, tc := range []struct {
		name, output, prefix, want string
	}{
		{"go style", "The commands are:\n\n\tbuild  compile packages\n\tcheck  inspect packages\n\nAdditional help topics:\n\tcache  cache details\n", "bu", "build"},
		{"cargo aliases", "Commands:\n  build, b    Compile the package\n  check, c    Inspect it\n", "bu", "build"},
		{"npm list", "All commands:\n\n  access, build, check,\n  deploy, doctor\n\nSpecify configs in the ini-formatted file:\n", "bu", "build"},
		{"flag sections", "Compilation Options:\n  -r, --release  Optimize output\n", "--r", "--release"},
		{"overstruck man flags", "  -\b--\b-a\bal\bll\bl  Stage changes\n", "--a", "--all"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := parseHelp(tc.output, "sample", tc.prefix)
			for _, candidate := range got {
				if candidate.Value == tc.want {
					return
				}
			}
			t.Fatalf("%s missing: %+v", tc.want, got)
		})
	}
}

func TestHelpHintUsesCommandOutput(t *testing.T) {
	dir := t.TempDir()
	cli := `#!/bin/sh
case "$*" in
  "--help") printf 'The commands are:\n\tmod  work with modules\n' ;;
  "mod --help") printf "Run 'sample help mod' for usage.\n" ;;
  "help mod") printf 'The commands are:\n\ttidy  tidy modules\n' ;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "sample"), []byte(cli), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got := helpCandidates(ctx, "sample mod t", dir)
	for _, candidate := range got {
		if candidate.Value == "tidy" {
			return
		}
	}
	t.Fatalf("hinted subcommand missing: %+v", got)
}

func TestCollectGeneratedCompletionWithoutCommandRegistry(t *testing.T) {
	dir := t.TempDir()
	cli := `#!/bin/sh
case "$*" in
  "--help") printf 'Commands:\n  completion  Generate shell completions\n' ;;
  "completion zsh") cat <<'EOF'
#compdef twiggenerated
_twiggenerated() {
  local -a values descriptions
  values=(alpha beta)
  descriptions=('First command' 'Second command')
  compadd -a values -d descriptions
}

compdef _twiggenerated twiggenerated
EOF
  ;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "twiggenerated"), []byte(cli), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "twiggenerated al", Cursor: len("twiggenerated al"), Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range got {
		if candidate.Value == "alpha" && candidate.Description == "First command" {
			return
		}
	}
	t.Fatalf("generated completion missing: %+v", got)
}

func TestCommandNamesPreferInstalledCompletionDefinition(t *testing.T) {
	dir := t.TempDir()
	fpath := t.TempDir()
	for _, name := range []string{"twigalpha", "twigbeta"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(fpath, "_twigbeta"), []byte("#compdef twigbeta\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("FPATH", fpath)
	got := commandNames("twig")
	if len(got) != 2 || got[0].Value != "twigbeta" || got[1].Value != "twigalpha" {
		t.Fatalf("completion-aware order: %+v", got)
	}
}

func TestDescriptionMarksSubcommandInsertionWithoutToolName(t *testing.T) {
	dir := t.TempDir()
	completionDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "twigword"), []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	completionScript := "#compdef twigword\nlocal -a values descriptions\nvalues=(branch)\ndescriptions=('branch -- Manage branches')\ncompadd -a values -d descriptions\n"
	if err := os.WriteFile(filepath.Join(completionDir, "_twigword"), []byte(completionScript), 0o600); err != nil {
		t.Fatal(err)
	}
	base, err := exec.Command("zsh", "-fc", "print -r -- ${(j.:.)fpath}").Output()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	t.Setenv("FPATH", completionDir+":"+strings.TrimSpace(string(base)))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	got, err := Collect(ctx, Request{Buffer: "twigword br", Cursor: len("twigword br"), Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != "branch" || got[0].Suffix != " " {
		t.Fatalf("subcommand insertion: %+v", got)
	}
}
