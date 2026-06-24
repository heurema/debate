//go:build acp_integration

package acp

import (
	"context"
	"os"
	"testing"

	"github.com/heurema/debate/internal/engine/transport"
)

// TestIntegration_ClaudeAgentACP exercises the real claude-agent-acp adapter subprocess.
// Requires: build tag acp_integration AND env var DEBATE_ACP_INTEGRATION=1.
// Without DEBATE_ACP_INTEGRATION the test compiles and skips, no network or subprocess is invoked.
func TestIntegration_ClaudeAgentACP(t *testing.T) {
	if os.Getenv("DEBATE_ACP_INTEGRATION") != "1" {
		t.Skip("set DEBATE_ACP_INTEGRATION=1 to run ACP integration tests")
	}
	tr, err := New(BackendClaude, os.Getenv, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sess, err := tr.Open(context.Background(), transport.Spec{
		ID:     "claude-test",
		Model:  os.Getenv("ANTHROPIC_MODEL"),
		Effort: "low",
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer sess.Close()

	result, err := sess.Send(context.Background(), "Reply with exactly: hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty response")
	}
	t.Logf("response: %s", result.Content)
}

// TestIntegration_CodexACP exercises the real codex-acp adapter subprocess.
func TestIntegration_CodexACP(t *testing.T) {
	if os.Getenv("DEBATE_ACP_INTEGRATION") != "1" {
		t.Skip("set DEBATE_ACP_INTEGRATION=1 to run ACP integration tests")
	}
	tr, err := New(BackendCodex, os.Getenv, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	sess, err := tr.Open(context.Background(), transport.Spec{
		ID:    "codex-test",
		Model: os.Getenv("OPENAI_MODEL"),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer sess.Close()

	result, err := sess.Send(context.Background(), "Reply with exactly: hello")
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty response")
	}
	t.Logf("response: %s", result.Content)
}
