// Package mock provides a scripted transport backend for tests.
package mock

import (
	"context"
	"fmt"

	"github.com/heurema/debate/internal/engine/transport"
)

// ScriptedResult is one pre-configured response for a Send call.
type ScriptedResult struct {
	Result transport.Result
	Err    error
}

// Session is a mock transport.Session that returns pre-scripted responses.
// Close is idempotent.
type Session struct {
	script  []ScriptedResult
	idx     int
	prompts []string
	closed  bool
}

// NewSession creates a mock session with the given scripted responses.
func NewSession(script []ScriptedResult) *Session {
	return &Session{script: script}
}

// Send returns the next scripted result and records the prompt.
func (s *Session) Send(_ context.Context, prompt string) (transport.Result, error) {
	s.prompts = append(s.prompts, prompt)
	if s.idx >= len(s.script) {
		return transport.Result{}, fmt.Errorf("mock: no scripted result for Send #%d (script length %d)", s.idx+1, len(s.script))
	}
	r := s.script[s.idx]
	s.idx++
	return r.Result, r.Err
}

// Close marks the session closed. Safe to call multiple times.
func (s *Session) Close() error {
	s.closed = true
	return nil
}

// Prompts returns a copy of the prompts received by this session in order.
func (s *Session) Prompts() []string {
	out := make([]string, len(s.prompts))
	copy(out, s.prompts)
	return out
}

// Closed reports whether Close has been called.
func (s *Session) Closed() bool { return s.closed }

// Transport is a mock transport.Transport that hands out pre-configured sessions by Spec.ID.
type Transport struct {
	sessions map[string]*Session
}

// NewTransport creates a mock transport whose Open returns sessions keyed by Spec.ID.
func NewTransport(sessions map[string]*Session) *Transport {
	return &Transport{sessions: sessions}
}

// Open returns the pre-configured session for spec.ID, or an error if none is registered.
func (t *Transport) Open(_ context.Context, spec transport.Spec) (transport.Session, error) {
	s, ok := t.sessions[spec.ID]
	if !ok {
		return nil, fmt.Errorf("mock: no session configured for id %q", spec.ID)
	}
	return s, nil
}
