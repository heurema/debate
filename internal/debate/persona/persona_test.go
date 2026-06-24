package persona_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/persona"
)

func writePersona(t *testing.T, dir, name, content string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseFile_ValidDebater(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "alice.md", `---
model: claude-sonnet-4-6
effort: high
tags:
  - logic
  - ethics
---
You are Alice, a careful logical reasoner.
`)
	p, err := persona.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "alice" {
		t.Errorf("ID = %q, want alice", p.ID)
	}
	if p.Role != "debater" {
		t.Errorf("Role = %q, want debater", p.Role)
	}
	if p.Model != "claude-sonnet-4-6" {
		t.Errorf("Model = %q, want claude-sonnet-4-6", p.Model)
	}
	if p.Effort != "high" {
		t.Errorf("Effort = %q, want high", p.Effort)
	}
	if p.Backend != "claude-agent-acp" {
		t.Errorf("Backend = %q, want claude-agent-acp", p.Backend)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "logic" || p.Tags[1] != "ethics" {
		t.Errorf("Tags = %v, want [logic ethics]", p.Tags)
	}
	if !strings.Contains(p.System, "Alice") {
		t.Errorf("System missing expected content: %q", p.System)
	}
}

func TestParseFile_ValidSynthesizer(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "synthesizer.md", `---
role: synthesizer
model: claude-haiku-4-5
effort: low
---
You are the synthesizer. Summarize agreement and disagreement.
`)
	p, err := persona.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Role != "synthesizer" {
		t.Errorf("Role = %q, want synthesizer", p.Role)
	}
	if p.ID != "synthesizer" {
		t.Errorf("ID = %q, want synthesizer", p.ID)
	}
	if p.Backend != "claude-agent-acp" {
		t.Errorf("Backend = %q, want claude-agent-acp", p.Backend)
	}
}

func TestParseFile_ExplicitBackendOverridesInference(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bob.md", `---
model: claude-opus-4-8
effort: medium
backend: some-custom-backend
---
Bob's system prompt.
`)
	p, err := persona.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Backend != "some-custom-backend" {
		t.Errorf("Backend = %q, want some-custom-backend", p.Backend)
	}
}

func TestParseFile_RoleDefaultsToDebater(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "carol.md", `---
model: gpt-4o
effort: medium
---
Carol's system prompt.
`)
	p, err := persona.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Role != "debater" {
		t.Errorf("Role = %q, want debater", p.Role)
	}
	if p.Backend != "codex-acp" {
		t.Errorf("Backend = %q, want codex-acp", p.Backend)
	}
}

func TestParseFile_GeminiBackend(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "dave.md", `---
model: gemini-pro
effort: medium
---
Dave's system prompt.
`)
	p, err := persona.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Backend != "agy" {
		t.Errorf("Backend = %q, want agy", p.Backend)
	}
}

func TestParseFile_UnknownFrontmatterKey(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bad.md", `---
model: claude-sonnet-4-6
effort: high
unknown_key: should_fail
---
Body text.
`)
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for unknown frontmatter key, got nil")
	}
}

func TestParseFile_InvalidRole(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bad.md", `---
role: moderator
model: claude-sonnet-4-6
effort: high
---
Body text.
`)
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for invalid role, got nil")
	}
	if !strings.Contains(err.Error(), "invalid role") {
		t.Errorf("error should mention invalid role: %v", err)
	}
}

func TestParseFile_MissingModel(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bad.md", `---
effort: high
---
Body text.
`)
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for missing model, got nil")
	}
	if !strings.Contains(err.Error(), "model") {
		t.Errorf("error should mention model: %v", err)
	}
}

func TestParseFile_MissingEffort(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bad.md", `---
model: claude-sonnet-4-6
---
Body text.
`)
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for missing effort, got nil")
	}
	if !strings.Contains(err.Error(), "effort") {
		t.Errorf("error should mention effort: %v", err)
	}
}

func TestParseFile_EmptyBody(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bad.md", `---
model: claude-sonnet-4-6
effort: high
---
`)
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should mention empty: %v", err)
	}
}

func TestParseFile_WhitespaceOnlyBody(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bad.md", "---\nmodel: claude-sonnet-4-6\neffort: high\n---\n   \n\t\n")
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for whitespace-only body, got nil")
	}
}

func TestParseFile_UninfernableModel(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "bad.md", `---
model: some-unknown-llm
effort: high
---
Body text.
`)
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error for uninferrable model backend, got nil")
	}
	if !strings.Contains(err.Error(), "cannot infer backend") {
		t.Errorf("error should mention cannot infer backend: %v", err)
	}
}

func TestInferBackend(t *testing.T) {
	cases := []struct {
		model   string
		backend string
		wantErr bool
	}{
		{"claude-sonnet-4-6", "claude-agent-acp", false},
		{"claude-opus-4-8", "claude-agent-acp", false},
		{"claude-haiku-4-5", "claude-agent-acp", false},
		{"opus", "claude-agent-acp", false},
		{"sonnet", "claude-agent-acp", false},
		{"haiku", "claude-agent-acp", false},
		{"fable", "claude-agent-acp", false},
		{"gpt-4o", "codex-acp", false},
		{"gpt-3.5-turbo", "codex-acp", false},
		{"codex", "codex-acp", false},
		{"o1", "codex-acp", false},
		{"o3-mini", "codex-acp", false},
		{"gemini-pro", "agy", false},
		{"gemini-1.5-flash", "agy", false},
		{"unknown-model", "", true},
		{"llama3", "", true},
	}
	for _, tc := range cases {
		got, err := persona.InferBackend(tc.model)
		if tc.wantErr {
			if err == nil {
				t.Errorf("InferBackend(%q): expected error, got %q", tc.model, got)
			}
		} else {
			if err != nil {
				t.Errorf("InferBackend(%q): unexpected error: %v", tc.model, err)
			}
			if got != tc.backend {
				t.Errorf("InferBackend(%q) = %q, want %q", tc.model, got, tc.backend)
			}
		}
	}
}

func TestParseFile_NoFrontmatterFails(t *testing.T) {
	dir := t.TempDir()
	path := writePersona(t, dir, "nofm.md", "Just a body with no frontmatter at all.\n")
	_, err := persona.ParseFile(path)
	if err == nil {
		t.Fatal("expected error (missing model), got nil")
	}
}
