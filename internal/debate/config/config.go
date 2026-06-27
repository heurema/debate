// Package config handles .heurema/debate workspace discovery and loading.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/heurema/debate/internal/debate/persona"
)

const (
	heuremaDirName   = ".heurema"
	debateDirName    = "debate"
	personasDirName  = "personas"
	tablesDirName    = "tables"
	defaultTableName = "default"

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

type table struct {
	Panel []string
	Synth string
}

type tableYAML struct {
	Version int      `yaml:"version"`
	Panel   []string `yaml:"panel"`
	Synth   string   `yaml:"synth"`
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

// ValidSegment reports whether s is a path-safe workspace identifier segment.
func ValidSegment(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// Load discovers the .heurema/debate workspace rooted at or above startDir,
// then loads personas and tables, and resolves the debater panel and synthesizer.
// withList, if non-empty, overrides the selected table panel. synthOverride, if
// non-empty, names the synthesizer.
func Load(startDir, tableName string, withList []string, synthOverride string) (Workspace, error) {
	debDir, err := Discover(startDir)
	if err != nil {
		return Workspace{}, err
	}

	personas, err := loadPersonas(filepath.Join(debDir, personasDirName))
	if err != nil {
		return Workspace{}, err
	}
	byID := make(map[string]persona.Persona, len(personas))
	for _, p := range personas {
		if _, exists := byID[p.ID]; exists {
			return Workspace{}, fmt.Errorf("persona %s: duplicate persona ID", p.ID)
		}
		byID[p.ID] = p
	}

	tables, err := loadTables(filepath.Join(debDir, tablesDirName))
	if err != nil {
		if !errors.Is(err, errTablesMissing) || len(withList) == 0 || tableName != "" {
			return Workspace{}, err
		}
		tables = map[string]table{}
	}

	selectedTable, haveTable, err := selectTable(tables, tableName, len(withList) > 0)
	if err != nil {
		return Workspace{}, err
	}

	panelSelectors := selectedTable.Panel
	if len(withList) > 0 {
		panelSelectors = withList
	}
	panel, err := resolvePanel(byID, panelSelectors)
	if err != nil {
		return Workspace{}, err
	}

	synthSelector := ""
	if haveTable {
		synthSelector = selectedTable.Synth
	}
	synth, err := resolveSynthesizer(byID, synthOverride, synthSelector)
	if err != nil {
		return Workspace{}, err
	}

	return Workspace{
		Dir:         debDir,
		Panel:       panel,
		Synthesizer: synth,
	}, nil
}

func loadPersonas(personasPath string) ([]persona.Persona, error) {
	entries, err := os.ReadDir(personasPath)
	if err != nil {
		return nil, fmt.Errorf("personas: %w", err)
	}

	var personas []persona.Persona
	for _, entry := range entries {
		name := entry.Name()
		if isHidden(name) {
			continue
		}
		full := filepath.Join(personasPath, name)
		switch {
		case entry.IsDir():
			if !ValidSegment(name) {
				return nil, fmt.Errorf("personas: invalid namespace %q (must contain only letters, digits, hyphens, and underscores)", name)
			}
			nsPersonas, err := loadNamespacePersonas(full, name)
			if err != nil {
				return nil, err
			}
			personas = append(personas, nsPersonas...)
		case filepath.Ext(name) == ".md":
			id := strings.TrimSuffix(name, ".md")
			if !ValidSegment(id) {
				return nil, fmt.Errorf("persona %s: invalid name %q (must contain only letters, digits, hyphens, and underscores)", id, id)
			}
			p, err := persona.ParseFileWithID(full, id)
			if err != nil {
				return nil, err
			}
			personas = append(personas, p)
		}
	}
	sort.Slice(personas, func(i, j int) bool { return personas[i].ID < personas[j].ID })
	return personas, nil
}

func loadNamespacePersonas(namespacePath, namespace string) ([]persona.Persona, error) {
	entries, err := os.ReadDir(namespacePath)
	if err != nil {
		return nil, fmt.Errorf("personas/%s: %w", namespace, err)
	}

	var personas []persona.Persona
	for _, entry := range entries {
		name := entry.Name()
		if isHidden(name) {
			continue
		}
		full := filepath.Join(namespacePath, name)
		if entry.IsDir() {
			hasMD, err := containsMarkdown(full)
			if err != nil {
				return nil, fmt.Errorf("personas/%s/%s: %w", namespace, name, err)
			}
			if hasMD {
				return nil, fmt.Errorf("personas/%s/%s: nested persona files are not supported in v1", namespace, name)
			}
			continue
		}
		if filepath.Ext(name) != ".md" {
			continue
		}
		shortName := strings.TrimSuffix(name, ".md")
		if !ValidSegment(shortName) {
			return nil, fmt.Errorf("persona %s/%s: invalid name %q (must contain only letters, digits, hyphens, and underscores)", namespace, shortName, shortName)
		}
		id := namespace + "/" + shortName
		p, err := persona.ParseFileWithID(full, id)
		if err != nil {
			return nil, err
		}
		personas = append(personas, p)
	}
	return personas, nil
}

func containsMarkdown(root string) (bool, error) {
	found := false
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && isHidden(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isHidden(d.Name()) && filepath.Ext(d.Name()) == ".md" {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found, err
}

var errTablesMissing = errors.New("tables directory missing")

func loadTables(tablesPath string) (map[string]table, error) {
	entries, err := os.ReadDir(tablesPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("tables: %w", errTablesMissing)
		}
		return nil, fmt.Errorf("tables: %w", err)
	}

	tables := make(map[string]table)
	for _, entry := range entries {
		name := entry.Name()
		if isHidden(name) || entry.IsDir() || filepath.Ext(name) != ".yml" {
			continue
		}
		tableName := strings.TrimSuffix(name, ".yml")
		if !ValidSegment(tableName) {
			return nil, fmt.Errorf("table %q: invalid name (must contain only letters, digits, hyphens, and underscores)", tableName)
		}
		tbl, err := parseTableFile(filepath.Join(tablesPath, name), tableName)
		if err != nil {
			return nil, err
		}
		tables[tableName] = tbl
	}
	return tables, nil
}

func parseTableFile(path, name string) (table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return table{}, fmt.Errorf("table %s: %w", name, err)
	}
	var fields tableYAML
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&fields); err != nil {
		return table{}, fmt.Errorf("table %s: %w", name, err)
	}
	if fields.Version == 0 {
		return table{}, fmt.Errorf("table %s: missing required field: version", name)
	}
	if fields.Version != 1 {
		return table{}, fmt.Errorf("table %s: unsupported version %d (must be 1)", name, fields.Version)
	}
	if len(fields.Panel) == 0 {
		return table{}, fmt.Errorf("table %s: panel must not be empty", name)
	}
	return table{Panel: fields.Panel, Synth: fields.Synth}, nil
}

func selectTable(tables map[string]table, tableName string, hasWithOverride bool) (table, bool, error) {
	if tableName != "" {
		if !ValidSegment(tableName) {
			return table{}, false, fmt.Errorf("table %q: invalid name (must contain only letters, digits, hyphens, and underscores)", tableName)
		}
		tbl, ok := tables[tableName]
		if !ok {
			return table{}, false, fmt.Errorf("table %q not found", tableName)
		}
		return tbl, true, nil
	}

	if tbl, ok := tables[defaultTableName]; ok {
		return tbl, true, nil
	}
	if hasWithOverride {
		return table{}, false, nil
	}
	if len(tables) == 0 {
		return table{}, false, fmt.Errorf("tables: no table files found")
	}
	return table{}, false, fmt.Errorf("table %q not found", defaultTableName)
}

func resolvePanel(byID map[string]persona.Persona, selectors []string) ([]persona.Persona, error) {
	if len(selectors) == 0 {
		return nil, fmt.Errorf("panel: resolved panel is empty")
	}
	panel := make([]persona.Persona, 0, len(selectors))
	seen := make(map[string]struct{}, len(selectors))
	for _, selector := range selectors {
		p, err := resolveSelector(byID, selector)
		if err != nil {
			return nil, fmt.Errorf("panel: %w", err)
		}
		if p.Role == "synthesizer" {
			return nil, fmt.Errorf("panel: persona %q has role synthesizer and cannot be in the debater panel", p.ID)
		}
		if _, ok := seen[p.ID]; ok {
			return nil, fmt.Errorf("panel: duplicate persona %q", p.ID)
		}
		seen[p.ID] = struct{}{}
		panel = append(panel, p)
	}
	return panel, nil
}

func resolveSynthesizer(byID map[string]persona.Persona, override, tableSynth string) (persona.Persona, error) {
	if override != "" {
		return resolveSynthSelector(byID, override, "synthesizer")
	}
	if tableSynth != "" {
		return resolveSynthSelector(byID, tableSynth, "synthesizer")
	}

	p, err := resolveSelector(byID, "synthesizer")
	if err == nil {
		if p.Role != "synthesizer" {
			return persona.Persona{}, fmt.Errorf("synthesizer: persona %q has role %s and cannot be used as synthesizer", p.ID, p.Role)
		}
		return p, nil
	}
	if isNotFound(err) {
		return buildDefaultSynthesizer()
	}
	return persona.Persona{}, fmt.Errorf("synthesizer: %w", err)
}

func resolveSynthSelector(byID map[string]persona.Persona, selector, label string) (persona.Persona, error) {
	p, err := resolveSelector(byID, selector)
	if err != nil {
		return persona.Persona{}, fmt.Errorf("%s: %w", label, err)
	}
	if p.Role != "synthesizer" {
		return persona.Persona{}, fmt.Errorf("%s: persona %q has role %s and cannot be used as synthesizer", label, p.ID, p.Role)
	}
	return p, nil
}

type selectorNotFoundError struct {
	selector string
}

func (e selectorNotFoundError) Error() string {
	return fmt.Sprintf("selector %q did not match any persona", e.selector)
}

func isNotFound(err error) bool {
	var notFound selectorNotFoundError
	return errors.As(err, &notFound)
}

func resolveSelector(byID map[string]persona.Persona, selector string) (persona.Persona, error) {
	if strings.Contains(selector, "/") {
		p, ok := byID[selector]
		if !ok {
			return persona.Persona{}, selectorNotFoundError{selector: selector}
		}
		return p, nil
	}
	if p, ok := byID[selector]; ok {
		return p, nil
	}

	var matches []persona.Persona
	for _, p := range byID {
		if shortName(p.ID) == selector {
			matches = append(matches, p)
		}
	}
	switch len(matches) {
	case 0:
		return persona.Persona{}, selectorNotFoundError{selector: selector}
	case 1:
		return matches[0], nil
	default:
		sort.Slice(matches, func(i, j int) bool { return matches[i].ID < matches[j].ID })
		ids := make([]string, len(matches))
		for i, p := range matches {
			ids[i] = p.ID
		}
		return persona.Persona{}, fmt.Errorf("selector %q is ambiguous; use a qualified persona ID: %s", selector, strings.Join(ids, ", "))
	}
}

func shortName(id string) string {
	if idx := strings.LastIndex(id, "/"); idx >= 0 {
		return id[idx+1:]
	}
	return id
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

func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}
