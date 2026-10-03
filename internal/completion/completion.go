package completion

import (
	"bytes"
	"fmt"
	"strings"
)

// Candidate is a single suggestion captured from a zsh completion call.
type Candidate struct {
	Group       string `json:"group"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Suffix      string `json:"suffix,omitempty"`
	Kind        string `json:"kind,omitempty"`
}

func DecodeRecords(data []byte) ([]Candidate, error) {
	if len(data) == 0 {
		return nil, nil
	}
	if data[len(data)-1] != 0 {
		return nil, fmt.Errorf("completion protocol is not NUL terminated")
	}
	fields := bytes.Split(data[:len(data)-1], []byte{0})
	if len(fields)%5 != 0 {
		return nil, fmt.Errorf("completion protocol has %d fields; want a multiple of five", len(fields))
	}
	out := make([]Candidate, 0, len(fields)/5)
	for i := 0; i < len(fields); i += 5 {
		out = append(out, Candidate{
			Group:       string(fields[i]),
			Value:       string(fields[i+1]),
			Description: string(fields[i+2]),
			Suffix:      string(fields[i+3]),
			Kind:        string(fields[i+4]),
		})
	}
	return out, nil
}

func EncodeRecords(candidates []Candidate) []byte {
	var buf bytes.Buffer
	for _, candidate := range candidates {
		for _, field := range [...]string{candidate.Group, candidate.Value, candidate.Description, candidate.Suffix, candidate.Kind} {
			buf.WriteString(field)
			buf.WriteByte(0)
		}
	}
	return buf.Bytes()
}

func Filter(candidates []Candidate, prefix string) []Candidate {
	out := make([]Candidate, 0, len(candidates))
	seen := make(map[string]int)
	for _, candidate := range candidates {
		if !strings.HasPrefix(candidate.Value, prefix) {
			continue
		}
		if index, exists := seen[candidate.Value]; exists {
			if out[index].Description == "" && candidate.Description != "" {
				out[index].Description = candidate.Description
			}
			if out[index].Kind == "" && candidate.Kind != "" {
				out[index].Kind = candidate.Kind
			}
			continue
		}
		seen[candidate.Value] = len(out)
		out = append(out, candidate)
	}
	return out
}
