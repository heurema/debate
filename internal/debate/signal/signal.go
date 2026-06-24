// Package signal parses structured convergence signals from turn content.
package signal

import (
	"encoding/json"
	"strings"
)

// Signal carries the structured convergence state a speaker embeds in its turn.
type Signal struct {
	Position   string   `json:"position"`
	Objections []string `json:"objections"`
	Done       bool     `json:"done"`
}

// Parse extracts the last fenced ```signal block from content and decodes it.
// Returns (Signal, true) on success; (zero Signal, false) otherwise.
// Invariant: if Done is true but Objections is non-empty, Done is set to false.
func Parse(content string) (Signal, bool) {
	body, ok := lastSignalBlock(content)
	if !ok {
		return Signal{}, false
	}
	trimmed := strings.TrimSpace(body)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return Signal{}, false
	}
	var s Signal
	if err := json.Unmarshal([]byte(trimmed), &s); err != nil {
		return Signal{}, false
	}
	if s.Objections == nil {
		s.Objections = []string{}
	}
	if s.Done && len(s.Objections) > 0 {
		s.Done = false
	}
	return s, true
}

// lastSignalBlock returns the body of the last ```signal ... ``` block found in content.
func lastSignalBlock(content string) (string, bool) {
	lines := strings.Split(content, "\n")
	var last string
	found := false

	i := 0
	for i < len(lines) {
		line := strings.TrimRight(lines[i], "\r")
		if line == "```signal" {
			i++
			var body strings.Builder
			closed := false
			for i < len(lines) {
				bl := strings.TrimRight(lines[i], "\r")
				if bl == "```" {
					closed = true
					i++
					break
				}
				if body.Len() > 0 {
					body.WriteByte('\n')
				}
				body.WriteString(bl)
				i++
			}
			if closed {
				last = body.String()
				found = true
			}
		} else {
			i++
		}
	}
	return last, found
}
