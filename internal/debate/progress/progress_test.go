package progress

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEmitterSerializesConcurrentEventsAsValidPrefixedJSON(t *testing.T) {
	var out bytes.Buffer
	emitter := NewEmitter(&out)
	emitter.RunStarted()

	const (
		goroutines = 32
		eventsEach = 25
	)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(worker int) {
			defer wg.Done()
			<-start
			for j := 0; j < eventsEach; j++ {
				silence := time.Duration(j+1) * time.Millisecond
				if (worker+j)%2 == 0 {
					emitter.Heartbeat(1, "alice", silence)
				} else {
					emitter.TurnCompleted(1, "alice", silence)
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	wantLines := 1 + goroutines*eventsEach
	if len(lines) != wantLines {
		t.Fatalf("progress lines = %d, want %d", len(lines), wantLines)
	}
	for _, line := range lines {
		if !strings.HasPrefix(line, Prefix) {
			t.Fatalf("line missing prefix %q: %q", Prefix, line)
		}
		var ev Event
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, Prefix)), &ev); err != nil {
			t.Fatalf("invalid progress JSON %q: %v", line, err)
		}
		if ev.Version != 1 || ev.Type == "" || ev.Stage == "" {
			t.Fatalf("invalid progress event: %+v", ev)
		}
	}
}
