// Package transport defines how the engine sends and receives messages.
package transport

import (
	"context"
	"errors"
)

// Spec describes how to open a session for one participant.
type Spec struct {
	ID       string
	Model    string
	Effort   string
	System   string
	ReadOnly bool
	Command  []string // optional; backend-specific invocation
}

// Usage holds token counters from a model response.
type Usage struct {
	InputTokens  int
	OutputTokens int
	CacheRead    int
	CacheWrite   int
}

// Result is the response from a single Send call.
type Result struct {
	Content string
	Usage   Usage
}

// Session represents a persistent connection to one participant for one run.
type Session interface {
	Send(ctx context.Context, prompt string) (Result, error)
	Close() error
}

// Transport opens sessions.
type Transport interface {
	Open(ctx context.Context, spec Spec) (Session, error)
}

// Sentinel errors — retryable.
var (
	ErrRateLimit     = errors.New("transport: rate limit")
	ErrIdleTimeout   = errors.New("transport: idle timeout")
	ErrTransportDrop = errors.New("transport: transport drop")
	ErrServerError   = errors.New("transport: server error")
	ErrDeadline      = errors.New("transport: deadline")
)

// Sentinel errors — non-retryable.
var (
	ErrAuth        = errors.New("transport: auth")
	ErrClientError = errors.New("transport: client error")
	ErrCanceled    = errors.New("transport: canceled")
)

// ErrorClass is the result of Classify.
type ErrorClass struct {
	Retryable bool
	Kind      string
}

var errorTable = []struct {
	sentinel  error
	kind      string
	retryable bool
}{
	{ErrRateLimit, "rate_limit", true},
	{ErrIdleTimeout, "idle_timeout", true},
	{ErrTransportDrop, "transport_drop", true},
	{ErrServerError, "server_error", true},
	{ErrDeadline, "deadline", true},
	{ErrAuth, "auth", false},
	{ErrClientError, "client_error", false},
	{ErrCanceled, "canceled", false},
}

// Classify maps an error to its ErrorClass.
// nil -> Kind "none", Retryable false.
// Unrecognized errors -> Kind "unknown", Retryable false.
func Classify(err error) ErrorClass {
	if err == nil {
		return ErrorClass{Kind: "none"}
	}
	for _, e := range errorTable {
		if errors.Is(err, e.sentinel) {
			return ErrorClass{Retryable: e.retryable, Kind: e.kind}
		}
	}
	return ErrorClass{Kind: "unknown"}
}
