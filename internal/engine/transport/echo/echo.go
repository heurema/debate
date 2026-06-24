// Package echo provides a deterministic offline transport backend for testing and demos.
package echo

import (
	"context"

	"github.com/heurema/debate/internal/engine/transport"
)

// cannedReply is the fixed response returned for every Send call.
// It includes a valid signal block with done=true so debates converge.
const cannedReply = "I agree with the current approach.\n\n" +
	"```signal\n" +
	"{\"position\": \"agree\", \"objections\": [], \"done\": true}\n" +
	"```"

// New returns a transport.Transport that opens echo sessions.
// Echo sessions return cannedReply for every Send call — no network or subprocess.
func New() transport.Transport { return echoTransport{} }

type echoTransport struct{}

func (echoTransport) Open(_ context.Context, _ transport.Spec) (transport.Session, error) {
	return &echoSession{}, nil
}

type echoSession struct{}

func (*echoSession) Send(_ context.Context, _ string) (transport.Result, error) {
	return transport.Result{Content: cannedReply}, nil
}

func (*echoSession) Close() error { return nil }
