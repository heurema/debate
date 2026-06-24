package transport_test

import (
	"fmt"
	"testing"

	"github.com/heurema/debate/internal/engine/transport"
)

type classifyCase struct {
	err       error
	wantKind  string
	wantRetry bool
}

var sentinelCases = []classifyCase{
	{transport.ErrRateLimit, "rate_limit", true},
	{transport.ErrIdleTimeout, "idle_timeout", true},
	{transport.ErrTransportDrop, "transport_drop", true},
	{transport.ErrServerError, "server_error", true},
	{transport.ErrDeadline, "deadline", true},
	{transport.ErrAuth, "auth", false},
	{transport.ErrClientError, "client_error", false},
	{transport.ErrCanceled, "canceled", false},
}

func TestClassifyBaresentinel(t *testing.T) {
	for _, tc := range sentinelCases {
		got := transport.Classify(tc.err)
		if got.Kind != tc.wantKind {
			t.Errorf("Classify(%v).Kind = %q, want %q", tc.err, got.Kind, tc.wantKind)
		}
		if got.Retryable != tc.wantRetry {
			t.Errorf("Classify(%v).Retryable = %v, want %v", tc.err, got.Retryable, tc.wantRetry)
		}
	}
}

func TestClassifyWrappedSentinel(t *testing.T) {
	for _, tc := range sentinelCases {
		wrapped := fmt.Errorf("outer: %w", tc.err)
		got := transport.Classify(wrapped)
		if got.Kind != tc.wantKind {
			t.Errorf("Classify(wrapped %v).Kind = %q, want %q", tc.err, got.Kind, tc.wantKind)
		}
		if got.Retryable != tc.wantRetry {
			t.Errorf("Classify(wrapped %v).Retryable = %v, want %v", tc.err, got.Retryable, tc.wantRetry)
		}
	}
}

func TestClassifyNil(t *testing.T) {
	got := transport.Classify(nil)
	if got.Kind != "none" {
		t.Errorf("Classify(nil).Kind = %q, want none", got.Kind)
	}
	if got.Retryable {
		t.Error("Classify(nil).Retryable = true, want false")
	}
}

func TestClassifyUnknown(t *testing.T) {
	got := transport.Classify(fmt.Errorf("some random error"))
	if got.Kind != "unknown" {
		t.Errorf("Classify(unknown).Kind = %q, want unknown", got.Kind)
	}
	if got.Retryable {
		t.Error("Classify(unknown).Retryable = true, want false")
	}
}
