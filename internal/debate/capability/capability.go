// Package capability detects which local agent runtime is usable and resolves
// the v1 default (model, backend) pair for it. Detection is signal-based only:
// it checks for an executable on PATH and does not verify authentication,
// network reachability, or model entitlement.
package capability

import "os/exec"

// LookPath resolves an executable's path, mirroring os/exec.LookPath. Callers
// inject a fake in tests to simulate different local tooling without touching
// the real PATH.
type LookPath func(file string) (string, error)

// DefaultLookup is the production LookPath backed by os/exec.
var DefaultLookup LookPath = exec.LookPath

// Family is one of the three v1 runtime backend defaults.
type Family struct {
	Model   string
	Backend string
}

// The three v1 runtime default capability families.
var (
	Claude = Family{Model: "claude-haiku-4-5", Backend: "claude-agent-acp"}
	Codex  = Family{Model: "codex", Backend: "codex-acp"}
	Gemini = Family{Model: "gemini-pro", Backend: "agy"}
)

// Detect resolves the v1 runtime backend default from executables on PATH,
// checked in precedence order: claude, then codex, then agy or gemini.
// Reports ok=false when none of the supported executables are found.
func Detect(lookup LookPath) (Family, bool) {
	if _, err := lookup("claude"); err == nil {
		return Claude, true
	}
	if _, err := lookup("codex"); err == nil {
		return Codex, true
	}
	if _, err := lookup("agy"); err == nil {
		return Gemini, true
	}
	if _, err := lookup("gemini"); err == nil {
		return Gemini, true
	}
	return Family{}, false
}

// FamilyForBackend maps a resolved persona backend id to its matching v1
// family, when it is one of the three supported runtime families.
func FamilyForBackend(backend string) (Family, bool) {
	switch backend {
	case Claude.Backend:
		return Claude, true
	case Codex.Backend:
		return Codex, true
	case Gemini.Backend:
		return Gemini, true
	default:
		return Family{}, false
	}
}
