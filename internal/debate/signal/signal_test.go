package signal_test

import (
	"testing"

	"github.com/heurema/debate/internal/debate/signal"
)

func TestParse_WellFormed(t *testing.T) {
	content := "Some reply.\n\n```signal\n{\"position\": \"agreed\", \"objections\": [], \"done\": true}\n```"
	sig, ok := signal.Parse(content)
	if !ok {
		t.Fatal("Parse returned ok=false, want true")
	}
	if sig.Position != "agreed" {
		t.Errorf("Position = %q, want %q", sig.Position, "agreed")
	}
	if !sig.Done {
		t.Error("Done = false, want true")
	}
	if len(sig.Objections) != 0 {
		t.Errorf("Objections = %v, want empty", sig.Objections)
	}
}

func TestParse_TrailingProse(t *testing.T) {
	// Text after the last signal block is permitted and does not prevent a successful parse.
	content := "Reply.\n\n```signal\n{\"position\": \"ok\", \"done\": true}\n```\n\nMore text after the signal."
	sig, ok := signal.Parse(content)
	if !ok {
		t.Fatal("Parse returned ok=false, want true")
	}
	if !sig.Done {
		t.Error("Done = false, want true")
	}
}

func TestParse_LastBlockUsed(t *testing.T) {
	// When multiple signal blocks exist, the last one is used.
	content := "```signal\n{\"done\": false}\n```\nMiddle text.\n```signal\n{\"done\": true}\n```"
	sig, ok := signal.Parse(content)
	if !ok {
		t.Fatal("Parse returned ok=false, want true")
	}
	if !sig.Done {
		t.Error("Done = false; should use last signal block where done=true")
	}
}

func TestParse_NoBlock(t *testing.T) {
	_, ok := signal.Parse("No fenced blocks here at all.")
	if ok {
		t.Error("Parse returned ok=true, want false")
	}
}

func TestParse_GarbledJSON(t *testing.T) {
	_, ok := signal.Parse("```signal\nnot-valid-json\n```")
	if ok {
		t.Error("Parse returned ok=true for invalid JSON, want false")
	}
}

func TestParse_NonObjectJSON(t *testing.T) {
	// A JSON array is not a valid signal body.
	_, ok := signal.Parse("```signal\n[\"item\"]\n```")
	if ok {
		t.Error("Parse returned ok=true for JSON array, want false")
	}
}

func TestParse_NullJSON(t *testing.T) {
	// null is not a JSON object.
	_, ok := signal.Parse("```signal\nnull\n```")
	if ok {
		t.Error("Parse returned ok=true for null JSON, want false")
	}
}

func TestParse_EmptyObject(t *testing.T) {
	// {} is a valid signal with all zero fields.
	sig, ok := signal.Parse("```signal\n{}\n```")
	if !ok {
		t.Fatal("Parse returned ok=false for {}, want true")
	}
	if sig.Done {
		t.Error("Done = true, want false for empty object")
	}
	if sig.Objections == nil {
		t.Error("Objections should be non-nil empty slice, got nil")
	}
}

func TestParse_DoneWithObjections_Invariant(t *testing.T) {
	// Done=true with non-empty Objections must be returned with Done=false.
	content := "```signal\n{\"done\": true, \"objections\": [\"blocking issue\"]}\n```"
	sig, ok := signal.Parse(content)
	if !ok {
		t.Fatal("Parse returned ok=false, want true")
	}
	if sig.Done {
		t.Error("Done should be false when Objections is non-empty (invariant)")
	}
	if len(sig.Objections) != 1 || sig.Objections[0] != "blocking issue" {
		t.Errorf("Objections = %v, want [blocking issue]", sig.Objections)
	}
}

func TestParse_NonSignalFencedBlock(t *testing.T) {
	// A ```json block is not a signal block.
	_, ok := signal.Parse("```json\n{\"done\": true}\n```")
	if ok {
		t.Error("Parse returned ok=true for non-signal fenced block, want false")
	}
}

func TestParse_MultiLineJSON(t *testing.T) {
	// Multi-line JSON inside the signal block should parse correctly.
	content := "```signal\n{\n  \"position\": \"ok\",\n  \"done\": true\n}\n```"
	sig, ok := signal.Parse(content)
	if !ok {
		t.Fatal("Parse returned ok=false for multi-line JSON, want true")
	}
	if sig.Position != "ok" {
		t.Errorf("Position = %q, want %q", sig.Position, "ok")
	}
}
