// Package persona parses and validates debate participant persona files.
package persona

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Persona describes a single debate participant.
type Persona struct {
	ID      string
	Role    string
	Model   string
	Effort  string
	Backend string
	Tags    []string
	System  string
}

type frontmatterFields struct {
	Role    string   `yaml:"role"`
	Model   string   `yaml:"model"`
	Effort  string   `yaml:"effort"`
	Backend string   `yaml:"backend"`
	Tags    []string `yaml:"tags"`
}

// ParseFile reads and parses a persona from path.
// The basename without .md becomes the persona ID.
func ParseFile(path string) (Persona, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Persona{}, fmt.Errorf("persona %s: %w", filepath.Base(path), err)
	}
	id := strings.TrimSuffix(filepath.Base(path), ".md")
	return parse(id, data)
}

func parse(id string, content []byte) (Persona, error) {
	fm, body, err := splitFrontmatter(content)
	if err != nil {
		return Persona{}, fmt.Errorf("persona %s: %w", id, err)
	}

	var fields frontmatterFields
	if len(bytes.TrimSpace(fm)) > 0 {
		dec := yaml.NewDecoder(bytes.NewReader(fm))
		dec.KnownFields(true)
		if err := dec.Decode(&fields); err != nil {
			return Persona{}, fmt.Errorf("persona %s: %w", id, err)
		}
	}

	role := fields.Role
	if role == "" {
		role = "debater"
	}
	if role != "debater" && role != "synthesizer" {
		return Persona{}, fmt.Errorf("persona %s: invalid role %q (must be debater or synthesizer)", id, role)
	}

	if fields.Model == "" {
		return Persona{}, fmt.Errorf("persona %s: missing required field: model", id)
	}
	if fields.Effort == "" {
		return Persona{}, fmt.Errorf("persona %s: missing required field: effort", id)
	}

	system := strings.TrimSpace(body)
	if system == "" {
		return Persona{}, fmt.Errorf("persona %s: system prompt body is empty", id)
	}

	backend := fields.Backend
	if backend == "" {
		if backend, err = InferBackend(fields.Model); err != nil {
			return Persona{}, fmt.Errorf("persona %s: %w", id, err)
		}
	}

	tags := fields.Tags
	if tags == nil {
		tags = []string{}
	}

	return Persona{
		ID:      id,
		Role:    role,
		Model:   fields.Model,
		Effort:  fields.Effort,
		Backend: backend,
		Tags:    tags,
		System:  system,
	}, nil
}

// InferBackend infers the backend identifier from a model name.
// Returns an error if no backend can be inferred.
func InferBackend(model string) (string, error) {
	lower := strings.ToLower(model)
	switch {
	case strings.HasPrefix(lower, "claude-"),
		lower == "opus", lower == "sonnet", lower == "haiku", lower == "fable":
		return "claude-agent-acp", nil
	case strings.HasPrefix(lower, "gpt-"),
		lower == "codex",
		len(lower) >= 2 && lower[0] == 'o' && lower[1] >= '0' && lower[1] <= '9':
		return "codex-acp", nil
	case strings.HasPrefix(lower, "gemini-"):
		return "agy", nil
	default:
		return "", fmt.Errorf("cannot infer backend for model %q: set backend explicitly", model)
	}
}

// splitFrontmatter separates YAML frontmatter from the body using line scanning.
// Returns (nil, fullContent, nil) when the file does not start with "---".
func splitFrontmatter(content []byte) (fm []byte, body string, err error) {
	s := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return nil, s, nil
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return []byte(strings.Join(lines[1:i], "\n")),
				strings.Join(lines[i+1:], "\n"),
				nil
		}
	}
	return nil, "", errors.New("unclosed YAML frontmatter")
}
