//go:build exec_integration

package exec

import (
	"context"
	"os"
	"testing"

	"github.com/heurema/debate/internal/engine/transport"
)

// TestIntegration_Agy exercises the real agy CLI.
// Requires: build tag exec_integration AND env var DEBATE_EXEC_INTEGRATION set (non-empty).
// Without DEBATE_EXEC_INTEGRATION the test compiles and skips; no subprocess is invoked.
func TestIntegration_Agy(t *testing.T) {
	if os.Getenv("DEBATE_EXEC_INTEGRATION") == "" {
		t.Skip("set DEBATE_EXEC_INTEGRATION=1 to run exec integration tests")
	}

	model := os.Getenv("AGY_MODEL")
	if model == "" {
		t.Skip("set AGY_MODEL to run exec integration tests")
	}

	tr, err := New(BackendAgy, os.Getenv, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	sess, err := tr.Open(context.Background(), transport.Spec{
		ID:    "agy-test",
		Model: model,
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
