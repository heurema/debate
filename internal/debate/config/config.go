// Package config handles .heurema/debate workspace discovery and loading.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"github.com/heurema/debate/internal/debate/persona"
)

const (
	heuremaDirName  = ".heurema"
	debateDirName   = "debate"
	configFileName  = "config.yml"
	personasDirName = "personas"

	defaultSynthModel  = "claude-haiku-4-5"
	defaultSynthEffort = "low"
	defaultSynthSystem = "You are a neutral synthesizer. Review the discussion and produce a concise synthesis: areas of agreement, unresolved objections, and a proposed resolution."
)

// Workspace holds all loaded and resolved debate workspace data.
type Workspace struct {
	Dir         string
	Panel       []persona.Persona
	Synthesizer persona.Persona
}

// Discover walks up from startDir to find the first .heurema/debate directory.
// Mirrors git's .git discovery: walks to filesystem root and returns an error if not found.
func Discover(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("discover: %w", err)
	}
	for {
		candidate := filepath.Join(dir, heuremaDirName, debateDirName)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .heurema/debate directory found (searched from %s)", startDir)
		}
		dir = parent
	}
}

type configYAML struct {
	Table []string `yaml:"table"`
}

// Load discovers the .heurema/debate workspace rooted at or above startDir,
// then loads config and personas, and resolves the debater panel and synthesizer.
// withList, if non-empty, overrides the panel selector. synthOverride, if non-empty, names the synthesizer.
func Load(startDir string, withList []string, synthOverride string) (Workspace, error) {
	// 1. Discover
	debDir, err := Discover(startDir)
	if err != nil {
		return Workspace{}, err
	}

	// 2. Parse config.yml
	var table []string
	cfgPath := filepath.Join(debDir, configFileName)
	if data, err := os.ReadFile(cfgPath); err == nil {
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		var cfg configYAML
		if err := dec.Decode(&cfg); err != nil {
			return Workspace{}, fmt.Errorf("config.yml: %w", err)
		}
		table = cfg.Table
	} else if !errors.Is(err, os.ErrNotExist) {
		return Workspace{}, fmt.Errorf("config.yml: %w", err)
	}

	// 3. Parse personas in lexicographic filename order
	personasPath := filepath.Join(debDir, personasDirName)
	entries, err := filepath.Glob(filepath.Join(personasPath, "*.md"))
	if err != nil {
		return Workspace{}, fmt.Errorf("personas: %w", err)
	}
	sort.Strings(entries)

	personas := make([]persona.Persona, 0, len(entries))
	for _, entry := range entries {
		p, err := persona.ParseFile(entry)
		if err != nil {
			return Workspace{}, err
		}
		personas = append(personas, p)
	}

	byID := make(map[string]persona.Persona, len(personas))
	for _, p := range personas {
		byID[p.ID] = p
	}

	// 5. Resolve panel
	panel, err := resolvePanel(personas, byID, withList, table)
	if err != nil {
		return Workspace{}, err
	}

	// 6. Resolve synthesizer
	synth, err := resolveSynthesizer(byID, synthOverride)
	if err != nil {
		return Workspace{}, err
	}

	return Workspace{
		Dir:         debDir,
		Panel:       panel,
		Synthesizer: synth,
	}, nil
}

// resolvePanel resolves the debater panel using withList, table, or all debater personas.
// Synthesizer-role personas are never included; naming one explicitly is a fail-fast error.
func resolvePanel(personas []persona.Persona, byID map[string]persona.Persona, withList, table []string) ([]persona.Persona, error) {
	selectors := withList
	if len(selectors) == 0 {
		selectors = table
	}

	if len(selectors) > 0 {
		panel := make([]persona.Persona, 0, len(selectors))
		for _, name := range selectors {
			p, ok := byID[name]
			if !ok {
				return nil, fmt.Errorf("panel: persona %q not found", name)
			}
			if p.Role == "synthesizer" {
				return nil, fmt.Errorf("panel: persona %q has role synthesizer and cannot be in the debater panel", name)
			}
			panel = append(panel, p)
		}
		if len(panel) == 0 {
			return nil, fmt.Errorf("panel: resolved panel is empty")
		}
		return panel, nil
	}

	// Default: all debater personas in lexicographic order by persona ID.
	var panel []persona.Persona
	for _, p := range personas {
		if p.Role == "debater" {
			panel = append(panel, p)
		}
	}
	if len(panel) == 0 {
		return nil, fmt.Errorf("panel: no debater personas found")
	}
	sort.Slice(panel, func(i, j int) bool { return panel[i].ID < panel[j].ID })
	return panel, nil
}

// resolveSynthesizer returns the synthesizer: explicit override, then personas["synthesizer"], then built-in default.
func resolveSynthesizer(byID map[string]persona.Persona, override string) (persona.Persona, error) {
	if override != "" {
		p, ok := byID[override]
		if !ok {
			return persona.Persona{}, fmt.Errorf("synthesizer: persona %q not found", override)
		}
		return p, nil
	}
	if p, ok := byID["synthesizer"]; ok {
		return p, nil
	}
	return buildDefaultSynthesizer()
}

func buildDefaultSynthesizer() (persona.Persona, error) {
	backend, err := persona.InferBackend(defaultSynthModel)
	if err != nil {
		return persona.Persona{}, fmt.Errorf("default synthesizer: %w", err)
	}
	return persona.Persona{
		ID:      "synthesizer",
		Role:    "synthesizer",
		Model:   defaultSynthModel,
		Effort:  defaultSynthEffort,
		Backend: backend,
		Tags:    []string{},
		System:  defaultSynthSystem,
	}, nil
}
