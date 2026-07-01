package capability_test

import (
	"errors"
	"testing"

	"github.com/heurema/debate/internal/debate/capability"
)

func lookup(found ...string) capability.LookPath {
	set := make(map[string]bool, len(found))
	for _, f := range found {
		set[f] = true
	}
	return func(name string) (string, error) {
		if set[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}

func TestDetect_Precedence(t *testing.T) {
	cases := []struct {
		name   string
		found  []string
		want   capability.Family
		wantOK bool
	}{
		{"none", nil, capability.Family{}, false},
		{"claude only", []string{"claude"}, capability.Claude, true},
		{"codex only", []string{"codex"}, capability.Codex, true},
		{"agy only", []string{"agy"}, capability.Gemini, true},
		{"gemini only", []string{"gemini"}, capability.Gemini, true},
		{"claude and codex prefers claude", []string{"claude", "codex"}, capability.Claude, true},
		{"codex and agy prefers codex", []string{"codex", "agy"}, capability.Codex, true},
		{"all prefers claude", []string{"claude", "codex", "agy", "gemini"}, capability.Claude, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := capability.Detect(lookup(tc.found...))
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("Family = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestFamilyForBackend(t *testing.T) {
	cases := []struct {
		backend string
		want    capability.Family
		wantOK  bool
	}{
		{"claude-agent-acp", capability.Claude, true},
		{"codex-acp", capability.Codex, true},
		{"agy", capability.Gemini, true},
		{"echo", capability.Family{}, false},
		{"", capability.Family{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.backend, func(t *testing.T) {
			got, ok := capability.FamilyForBackend(tc.backend)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("Family = %+v, want %+v", got, tc.want)
			}
		})
	}
}
