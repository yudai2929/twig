package collector

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/yudai2929/twig/internal/completion"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
var helpHint = regexp.MustCompile(`['"]([A-Za-z0-9_.-]+) help ([A-Za-z0-9_ -]+)['"]`)

// helpCandidates is the fallback for installed CLIs without usable zsh metadata.
// The command path is restricted to plain words before invoking a help entrypoint.
func helpCandidates(ctx context.Context, buffer, dir string) []completion.Candidate {
	fields := strings.Fields(buffer)
	if len(fields) == 0 || len(fields) > 5 {
		return nil
	}
	tool := fields[0]
	if _, err := exec.LookPath(tool); err != nil {
		return nil
	}
	path := fields[1:]
	if !strings.HasSuffix(buffer, " ") && !strings.HasSuffix(buffer, "\t") {
		path = path[:len(path)-1]
	}
	for _, part := range path {
		if !plainCommandWord(part) {
			return nil
		}
	}
	prefix := currentPrefix(buffer)
	output := runHelp(ctx, tool, append(append([]string{}, path...), "--help"), dir)
	if candidates := parseHelp(output, tool, prefix); len(candidates) > 0 {
		return candidates
	}

	// Some CLIs direct users to a help subcommand for the requested topic.
	// Follow only topics printed by the CLI itself, with a small bounded fanout.
	seen := make(map[string]bool)
	var topics [][]string
	for _, match := range helpHint.FindAllStringSubmatch(output, 6) {
		if match[1] != tool {
			continue
		}
		words := strings.Fields(match[2])
		if len(words) > 3 {
			continue
		}
		valid := true
		for _, word := range words {
			valid = valid && plainCommandWord(word)
		}
		if valid {
			topics = append(topics, words)
		}
	}
	if len(topics) == 0 && len(path) > 0 {
		topics = append(topics, path)
	}
	for _, topic := range topics {
		key := strings.Join(topic, " ")
		if seen[key] {
			continue
		}
		seen[key] = true
		args := append([]string{"help"}, topic...)
		if candidates := parseHelp(runHelp(ctx, tool, args, dir), tool, prefix); len(candidates) > 0 {
			return candidates
		}
	}
	return nil
}

func plainCommandWord(word string) bool {
	return word != "" && !strings.HasPrefix(word, "-") && strings.IndexFunc(word, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_')
	}) < 0
}

func runHelp(ctx context.Context, tool string, args []string, dir string) string {
	commandCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(commandCtx, tool, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "NPM_CONFIG_UPDATE_NOTIFIER=false")
	output, _ := cmd.CombinedOutput()
	if len(output) > 1<<20 {
		return ""
	}
	return stripOverstrike(ansiEscape.ReplaceAllString(string(output), ""))
}

func parseHelp(output, tool, prefix string) []completion.Candidate {
	output = stripOverstrike(output)
	section := ""
	var candidates []completion.Candidate
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if line == trimmed {
			heading := strings.ToLower(strings.TrimSuffix(trimmed, ":"))
			if strings.HasSuffix(trimmed, ":") || strings.ToUpper(trimmed) == trimmed {
				switch {
				case strings.Contains(heading, "command") && !strings.Contains(heading, "usage"):
					section = "commands"
				case strings.Contains(heading, "option") || strings.Contains(heading, "flag"):
					section = "flags"
				default:
					section = ""
				}
			}
			continue
		}
		if strings.HasPrefix(trimmed, "-") {
			if !strings.HasPrefix(prefix, "-") && prefix != "" {
				continue
			}
			for _, field := range strings.Fields(trimmed) {
				if !strings.HasPrefix(field, "-") {
					break
				}
				name := strings.TrimRight(field, ",.")
				if index := strings.IndexAny(name, "=<["); index >= 0 {
					name = name[:index]
				}
				if len(name) < 2 || !strings.HasPrefix(name, prefix) {
					continue
				}
				candidates = append(candidates, completion.Candidate{Group: tool + " flags", Value: name, Description: helpDescription(trimmed)})
			}
			continue
		}
		if section != "commands" || strings.HasPrefix(prefix, "-") {
			continue
		}
		if strings.Contains(trimmed, ",") && helpDescription(trimmed) == "" {
			for _, word := range strings.Fields(strings.ReplaceAll(trimmed, ",", " ")) {
				if plainCommandWord(word) && strings.HasPrefix(word, prefix) {
					candidates = append(candidates, completion.Candidate{Group: tool + " commands", Value: word, Suffix: " "})
				}
			}
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimRight(parts[0], ",:")
		if !plainCommandWord(name) || !strings.HasPrefix(name, prefix) {
			continue
		}
		description := helpDescription(trimmed)
		if description == "" {
			continue
		}
		candidates = append(candidates, completion.Candidate{Group: tool + " commands", Value: name, Description: description, Suffix: " "})
	}
	return completion.Filter(candidates, prefix)
}

func stripOverstrike(input string) string {
	output := make([]rune, 0, len(input))
	for _, char := range input {
		if char == '\b' {
			if len(output) > 0 {
				output = output[:len(output)-1]
			}
			continue
		}
		output = append(output, char)
	}
	return string(output)
}

func helpDescription(line string) string {
	for i := 1; i < len(line)-1; i++ {
		if line[i] == ' ' && line[i+1] == ' ' {
			return strings.TrimSpace(line[i+2:])
		}
	}
	return ""
}
