package mock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/heurema/debate/internal/engine/transport"
	"github.com/heurema/debate/internal/engine/transport/mock"
)

func TestSessionScriptedResults(t *testing.T) {
	script := []mock.ScriptedResult{
		{Result: transport.Result{Content: "hello"}},
		{Result: transport.Result{Content: "world"}},
	}
	s := mock.NewSession(script)

	r1, err := s.Send(context.Background(), "p1")
	if err != nil || r1.Content != "hello" {
		t.Fatalf("Send 1: got (%v, %v), want (hello, nil)", r1.Content, err)
	}
	r2, err := s.Send(context.Background(), "p2")
	if err != nil || r2.Content != "world" {
		t.Fatalf("Send 2: got (%v, %v), want (world, nil)", r2.Content, err)
	}
}

func TestSessionScriptedError(t *testing.T) {
	sentErr := errors.New("scripted error")
	script := []mock.ScriptedResult{
		{Err: sentErr},
	}
	s := mock.NewSession(script)
	_, err := s.Send(context.Background(), "prompt")
	if !errors.Is(err, sentErr) {
		t.Fatalf("err = %v, want sentErr", err)
	}
}

func TestSessionRecordsPrompts(t *testing.T) {
	s := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "a"}},
		{Result: transport.Result{Content: "b"}},
	})
	s.Send(context.Background(), "first")
	s.Send(context.Background(), "second")
	prompts := s.Prompts()
	if len(prompts) != 2 || prompts[0] != "first" || prompts[1] != "second" {
		t.Errorf("prompts = %v, want [first second]", prompts)
	}
}

func TestSessionExhausted(t *testing.T) {
	s := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "only"}},
	})
	s.Send(context.Background(), "ok")
	_, err := s.Send(context.Background(), "extra")
	if err == nil {
		t.Error("expected error when script exhausted, got nil")
	}
}

func TestSessionCloseIdempotent(t *testing.T) {
	s := mock.NewSession(nil)
	if err := s.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if !s.Closed() {
		t.Error("Closed() should be true after Close()")
	}
	if err := s.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestTransportOpen(t *testing.T) {
	sess := mock.NewSession([]mock.ScriptedResult{
		{Result: transport.Result{Content: "hi"}},
	})
	tr := mock.NewTransport(map[string]*mock.Session{"alice": sess})

	got, err := tr.Open(context.Background(), transport.Spec{ID: "alice"})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if got == nil {
		t.Fatal("Open returned nil session")
	}
}

func TestTransportOpenUnknown(t *testing.T) {
	tr := mock.NewTransport(map[string]*mock.Session{})
	_, err := tr.Open(context.Background(), transport.Spec{ID: "nobody"})
	if err == nil {
		t.Error("expected error for unknown id, got nil")
	}
}
