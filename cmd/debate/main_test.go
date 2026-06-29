package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/engine/transport"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestAssembleTask_TaskDashReadsStdin(t *testing.T) {
	task, err := assembleTask("-", nil, strings.NewReader("stdin task\n"))
	if err != nil {
		t.Fatalf("assembleTask: %v", err)
	}
	if task != "stdin task" {
		t.Fatalf("task = %q, want stdin task", task)
	}
	if strings.Contains(task, "-") {
		t.Fatalf("task should not contain literal hyphen sentinel: %q", task)
	}
}

func TestAssembleTask_PipedStdinWithoutTaskFlagStillWorks(t *testing.T) {
	task, err := assembleTask("", nil, strings.NewReader("piped task\n"))
	if err != nil {
		t.Fatalf("assembleTask: %v", err)
	}
	if task != "piped task" {
		t.Fatalf("task = %q, want piped task", task)
	}
}

func TestE2E_ConfigErrorsDoNotResolveBackends(t *testing.T) {
	for _, tc := range []struct {
		name       string
		mutate     func(t *testing.T, workDir string)
		args       []string
		wantErr    string
		wantCalled int
	}{
		{
			name:    "missing synth override",
			args:    []string{"--synth", "missing", "task"},
			wantErr: `selector "missing" did not match any persona`,
		},
		{
			name: "duplicate table panel",
			mutate: func(t *testing.T, workDir string) {
				t.Helper()
				tablePath := filepath.Join(workDir, ".heurema", "debate", "tables", "default.yml")
				err := os.WriteFile(tablePath, []byte("version: 1\npanel:\n  - alice\n  - alice\n"), 0o644)
				if err != nil {
					t.Fatal(err)
				}
			},
			args:    []string{"task"},
			wantErr: "duplicate persona",
		},
		{
			name:    "duplicate with override",
			args:    []string{"--with", "alice", "--with", "alice", "task"},
			wantErr: "duplicate persona",
		},
		{
			name:    "duplicate with comma-separated override",
			args:    []string{"--with", "alice,alice", "task"},
			wantErr: "duplicate persona",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workDir := makeE2EWorkspace(t)
			if tc.mutate != nil {
				tc.mutate(t, workDir)
			}
			var resolverCalls int
			resolver := func(backend string) (transport.Transport, error) {
				resolverCalls++
				return nil, fmt.Errorf("resolver should not be called for backend %q", backend)
			}

			var stdout, stderr bytes.Buffer
			code := parseAndRun(
				tc.args,
				&stdout, &stderr, strings.NewReader(""),
				resolver, workDir,
			)

			if code != 1 {
				t.Fatalf("exit code = %d, want 1; stderr: %s", code, stderr.String())
			}
			if resolverCalls != tc.wantCalled {
				t.Fatalf("resolver calls = %d, want %d", resolverCalls, tc.wantCalled)
			}
			if !strings.Contains(stderr.String(), tc.wantErr) {
				t.Fatalf("stderr = %q, want to contain %q", stderr.String(), tc.wantErr)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestCLI_WithEmptySelectorsFailBeforeResolver(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "empty", args: []string{"--with", "", "task"}},
		{name: "whitespace only", args: []string{"--with", "   ", "task"}},
		{name: "comma only", args: []string{"--with", ",", "task"}},
		{name: "trailing comma", args: []string{"--with", "proposer,", "task"}},
		{name: "leading comma", args: []string{"--with", ",skeptic", "task"}},
		{name: "repeated comma", args: []string{"--with", "proposer,,skeptic", "task"}},
		{name: "whitespace entry", args: []string{"--with", "proposer, ,skeptic", "task"}},
		{name: "equals comma only", args: []string{"--with=,", "task"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			workDir := makeE2EWorkspace(t)
			var resolverCalls int
			resolver := func(backend string) (transport.Transport, error) {
				resolverCalls++
				return nil, fmt.Errorf("resolver should not be called for backend %q", backend)
			}

			var stdout, stderr bytes.Buffer
			code := parseAndRun(
				tc.args,
				&stdout, &stderr, strings.NewReader(""),
				resolver, workDir,
			)

			if code != 1 {
				t.Fatalf("exit code = %d, want 1; stderr: %s", code, stderr.String())
			}
			if resolverCalls != 0 {
				t.Fatalf("resolver calls = %d, want 0", resolverCalls)
			}
			errText := stderr.String()
			if !strings.Contains(errText, "--with") || !strings.Contains(errText, "empty persona selector") {
				t.Fatalf("stderr = %q, want to mention --with and empty persona selector", errText)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}
