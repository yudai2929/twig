package collector

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/yudai2929/twig/internal/completion"
)

//go:embed collector.zsh
var collectorScript []byte

type Request struct {
	Buffer string
	Cursor int
	Dir    string
}

func Collect(ctx context.Context, request Request) ([]completion.Candidate, error) {
	cursorByte, err := byteOffset(request.Buffer, request.Cursor)
	if err != nil {
		return nil, err
	}
	if request.Dir == "" {
		request.Dir = "."
	}
	beforeCursor := request.Buffer[:cursorByte]
	if isCommandName(beforeCursor) {
		if _, err := exec.LookPath(beforeCursor); err == nil {
			return nil, nil
		}
		return commandNames(beforeCursor), nil
	}
	dataDir, err := os.MkdirTemp("", "twig-collector-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dataDir)
	scriptPath := filepath.Join(dataDir, "collector.zsh")
	bufferPath := filepath.Join(dataDir, "buffer")
	resultPath := filepath.Join(dataDir, "result")
	donePath := filepath.Join(dataDir, "done")
	if err := os.WriteFile(scriptPath, collectorScript, 0o600); err != nil {
		return nil, err
	}
	if err := os.WriteFile(bufferPath, []byte(request.Buffer), 0o600); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "script", "-q", "/dev/null", "/bin/zsh", "-f", "-i")
	cmd.Dir = request.Dir
	cmd.Env = append(os.Environ(),
		"TWIG_BUFFER_FILE="+bufferPath,
		"TWIG_RESULT_FILE="+resultPath,
		"TWIG_DONE_FILE="+donePath,
		fmt.Sprintf("TWIG_CURSOR=%d", request.Cursor),
		"TERM=xterm-256color",
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer func() {
		stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()
	var output bytes.Buffer
	var outputMu sync.Mutex
	go func() {
		var chunk [4096]byte
		for {
			n, err := stdout.Read(chunk[:])
			if n > 0 {
				outputMu.Lock()
				output.Write(chunk[:n])
				outputMu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	getOutput := func() string {
		outputMu.Lock()
		defer outputMu.Unlock()
		return output.String()
	}
	if _, err := io.WriteString(stdin, "source "+scriptPath+"\r"); err != nil {
		return nil, err
	}
	if err := waitFor(ctx, 2*time.Second, func() bool { return strings.Contains(getOutput(), "TWIG_READY") }); err != nil {
		return nil, fmt.Errorf("collector did not initialize: %w; output: %s", err, getOutput())
	}
	if _, err := stdin.Write([]byte{0x18, 'g'}); err != nil {
		return nil, err
	}
	if err := waitFor(ctx, 30*time.Second, func() bool { _, err := os.Stat(donePath); return err == nil }); err != nil {
		return nil, fmt.Errorf("collector did not finish: %w; output: %s", err, getOutput())
	}
	data, err := os.ReadFile(resultPath)
	if os.IsNotExist(err) {
		return classifyCandidates(helpCandidates(ctx, beforeCursor, request.Dir), beforeCursor, request.Dir), nil
	}
	if err != nil {
		return nil, err
	}
	candidates, err := completion.DecodeRecords(data)
	if err != nil {
		return nil, err
	}
	candidates = completion.Filter(candidates, currentPrefix(beforeCursor), currentWord(beforeCursor))
	for i := range candidates {
		if candidates[i].Suffix == "" && !strings.HasPrefix(candidates[i].Value, "-") &&
			(strings.Contains(candidates[i].Description, " -- ") || (candidates[i].Group == "" && candidates[i].Description == candidates[i].Value)) {
			candidates[i].Suffix = " "
		}
	}
	if len(candidates) == 0 {
		candidates = helpCandidates(ctx, beforeCursor, request.Dir)
	} else if pathOnlyCandidates(candidates) {
		if help := helpCandidates(ctx, beforeCursor, request.Dir); len(help) > 0 {
			var commands []completion.Candidate
			for _, candidate := range help {
				if strings.HasSuffix(candidate.Group, " commands") {
					commands = append(commands, candidate)
				}
			}
			if len(commands) > 0 {
				candidates = commands
			}
		}
	}
	return classifyCandidates(candidates, beforeCursor, request.Dir), nil
}

func classifyCandidates(candidates []completion.Candidate, beforeCursor, dir string) []completion.Candidate {
	wordStart := strings.LastIndexAny(beforeCursor, " \t\n") + 1
	word := beforeCursor[wordStart:]
	pathPrefix := ""
	if slash := strings.LastIndexByte(word, '/'); slash >= 0 {
		pathPrefix = word[:slash+1]
	}
	for i := range candidates {
		candidates[i] = classifyCandidate(candidates[i], dir, pathPrefix)
	}
	return candidates
}

func classifyCandidate(candidate completion.Candidate, dir, pathPrefix string) completion.Candidate {
	if strings.HasPrefix(candidate.Value, "-") || strings.HasSuffix(candidate.Group, " flags") {
		candidate.Kind = "flag"
		return candidate
	}
	if candidate.Group == "commands" || strings.HasSuffix(candidate.Group, " commands") {
		candidate.Kind = "command"
		return candidate
	}
	lowerGroup := strings.ToLower(candidate.Group)
	pathCandidate := candidate.Description == "" || candidate.Description == candidate.Value || strings.Contains(lowerGroup, "file") || strings.Contains(lowerGroup, "dir")
	if pathCandidate {
		path := filepath.Join(dir, pathPrefix, candidate.Value)
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				candidate.Kind = "directory"
			} else {
				candidate.Kind = "file"
			}
			return candidate
		}
	}
	if candidate.Description != "" && candidate.Suffix == " " {
		candidate.Kind = "command"
	} else {
		candidate.Kind = "value"
	}
	return candidate
}

func pathOnlyCandidates(candidates []completion.Candidate) bool {
	for _, candidate := range candidates {
		if (candidate.Group != "" && candidate.Group != "-default-") || candidate.Description != "" {
			return false
		}
	}
	return len(candidates) > 0
}

func isCommandName(buffer string) bool {
	return buffer != "" && !strings.ContainsAny(buffer, " \t\n/|;&<>")
}

func commandNames(prefix string) []completion.Candidate {
	seen := make(map[string]bool)
	var names []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			if seen[name] || !strings.HasPrefix(name, prefix) {
				continue
			}
			info, err := os.Stat(filepath.Join(dir, name))
			if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	fpath := os.Getenv("FPATH")
	if fpath == "" {
		if output, err := exec.Command("/bin/zsh", "-fc", "print -rl -- $fpath").Output(); err == nil {
			fpath = strings.Join(strings.Fields(string(output)), string(os.PathListSeparator))
		}
	}
	completionDirs := filepath.SplitList(fpath)
	hasCompletion := make(map[string]bool, len(names))
	for _, name := range names {
		for _, dir := range completionDirs {
			if info, err := os.Stat(filepath.Join(dir, "_"+name)); err == nil && info.Mode().IsRegular() {
				hasCompletion[name] = true
				break
			}
		}
	}
	sort.Slice(names, func(i, j int) bool {
		if hasCompletion[names[i]] != hasCompletion[names[j]] {
			return hasCompletion[names[i]]
		}
		if len(names[i]) != len(names[j]) {
			return len(names[i]) < len(names[j])
		}
		return names[i] < names[j]
	})
	result := make([]completion.Candidate, 0, len(names))
	for _, name := range names {
		result = append(result, completion.Candidate{Group: "commands", Value: name, Suffix: " ", Kind: "command"})
	}
	return result
}

func byteOffset(buffer string, cursor int) (int, error) {
	if cursor < 0 || cursor > utf8.RuneCountInString(buffer) {
		return 0, fmt.Errorf("cursor %d is outside the buffer", cursor)
	}
	for index := range buffer {
		if cursor == 0 {
			return index, nil
		}
		cursor--
	}
	return len(buffer), nil
}

func waitFor(ctx context.Context, maxWait time.Duration, ready func() bool) error {
	deadline := time.NewTimer(maxWait)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if ready() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("timed out")
		case <-ticker.C:
		}
	}
}

func currentPrefix(buffer string) string {
	end := len(buffer)
	for end > 0 && buffer[end-1] != ' ' && buffer[end-1] != '\t' && buffer[end-1] != '/' {
		end--
	}
	return buffer[end:]
}

func currentWord(buffer string) string {
	return buffer[strings.LastIndexAny(buffer, " \t\n")+1:]
}
