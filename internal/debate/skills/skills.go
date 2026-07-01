// Package skills installs and repairs the bundled debate Agent Skill into
// detected local agent client skill directories. Installation happens only
// during "debate init"; it never runs during ordinary debate runs.
package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/heurema/debate/internal/debate/capability"
)

// MetadataFileName is the managed metadata file written alongside an
// installed skill's content, under the skill's own directory.
const MetadataFileName = ".debate-skill.json"

const metadataSchema = "debate.skill.v1alpha1"

// Metadata records enough information to distinguish a debate-managed skill
// copy from unrelated user content and to support idempotent repair (AC5, AC6).
type Metadata struct {
	Schema        string `json:"schema"`
	BinaryVersion string `json:"debate_binary_version"`
	Checksum      string `json:"checksum"`
	Target        string `json:"target"`
	Source        string `json:"source"`
}

// Options configures InstallOrRepair. Home, LookPath, and Bundled are
// injectable so tests never touch the real user home directory or PATH.
type Options struct {
	// Home is the user's home directory. Required; a target is skipped with a
	// warning when empty.
	Home string
	// LookPath resolves an executable's path, mirroring os/exec.LookPath.
	LookPath capability.LookPath
	// Bundled is the skill content bundled with the running debate binary,
	// rooted at the skill directory (e.g. bundled.Skill()).
	Bundled fs.FS
	// BinaryVersion is the running debate binary's version, recorded in
	// managed metadata.
	BinaryVersion string
}

// Result reports the outcome of installing or repairing one target.
type Result struct {
	// Path is the target skill directory, e.g. ~/.agents/skills/debate.
	Path string
	// Action is one of "created", "updated", "current", or "skipped".
	Action string
	// Warning is non-empty when Action is "skipped" and explains why.
	Warning string
}

// InstallOrRepair installs or repairs the bundled debate skill into every
// detected local agent client target. It never returns an error: unwritable
// or unsafe targets, missing HOME, or no detected client are reported as
// skipped Results with an explanatory Warning rather than failing init.
func InstallOrRepair(opts Options) []Result {
	if strings.TrimSpace(opts.Home) == "" {
		return []Result{{
			Action:  "skipped",
			Warning: "HOME is not set; global debate skill not installed",
		}}
	}

	standard := detectStandard(opts.Home, opts.LookPath)
	claude := detectClaude(opts.Home, opts.LookPath)
	if !standard && !claude {
		return []Result{{
			Action: "skipped",
			Warning: "no supported local agent client detected (looked for codex, gemini, " +
				"or claude executables on PATH, and ~/.agents or ~/.claude directories); " +
				"global debate skill not installed. Re-run `debate init` after installing one.",
		}}
	}

	var results []Result
	if standard {
		results = append(results, installTarget(filepath.Join(opts.Home, ".agents", "skills", "debate"), "agents", opts))
	}
	if claude {
		results = append(results, installTarget(filepath.Join(opts.Home, ".claude", "skills", "debate"), "claude", opts))
	}
	return results
}

func detectStandard(home string, lookup capability.LookPath) bool {
	if lookup != nil {
		if _, err := lookup("codex"); err == nil {
			return true
		}
		if _, err := lookup("gemini"); err == nil {
			return true
		}
	}
	return dirExists(filepath.Join(home, ".agents"))
}

func detectClaude(home string, lookup capability.LookPath) bool {
	if lookup != nil {
		if _, err := lookup("claude"); err == nil {
			return true
		}
	}
	return dirExists(filepath.Join(home, ".claude"))
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// installTarget installs or repairs the skill at targetDir, following the
// two-step checksum comparison in AC5.
func installTarget(targetDir, name string, opts Options) Result {
	if !safeTarget(opts.Home, targetDir) {
		return Result{Path: targetDir, Action: "skipped",
			Warning: fmt.Sprintf("%s: unsafe target (symlink detected in path); preserving existing content", targetDir)}
	}

	bundledSum, err := hashFS(opts.Bundled)
	if err != nil {
		return Result{Path: targetDir, Action: "skipped", Warning: fmt.Sprintf("%s: reading bundled skill content: %v", targetDir, err)}
	}

	info, err := os.Stat(targetDir)
	switch {
	case os.IsNotExist(err):
		if err := writeFresh(targetDir, name, bundledSum, opts); err != nil {
			return Result{Path: targetDir, Action: "skipped", Warning: fmt.Sprintf("%s: %v", targetDir, err)}
		}
		return Result{Path: targetDir, Action: "created"}
	case err != nil:
		return Result{Path: targetDir, Action: "skipped", Warning: fmt.Sprintf("%s: %v", targetDir, err)}
	case !info.IsDir():
		return Result{Path: targetDir, Action: "skipped", Warning: fmt.Sprintf("%s: exists and is not a directory; preserving", targetDir)}
	}

	meta, ok := readMetadata(targetDir)
	if !ok {
		return Result{Path: targetDir, Action: "skipped",
			Warning: fmt.Sprintf("%s: exists without recognizable debate-managed metadata; preserving local content", targetDir)}
	}

	installedSum, err := hashDir(targetDir)
	if err != nil {
		return Result{Path: targetDir, Action: "skipped", Warning: fmt.Sprintf("%s: %v", targetDir, err)}
	}
	if installedSum != meta.Checksum {
		return Result{Path: targetDir, Action: "skipped",
			Warning: fmt.Sprintf("%s: locally modified since it was last installed; preserving (remove it and re-run `debate init` to reinstall)", targetDir)}
	}

	if bundledSum == meta.Checksum {
		return Result{Path: targetDir, Action: "current"}
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return Result{Path: targetDir, Action: "skipped", Warning: fmt.Sprintf("%s: %v", targetDir, err)}
	}
	if err := writeFresh(targetDir, name, bundledSum, opts); err != nil {
		return Result{Path: targetDir, Action: "skipped", Warning: fmt.Sprintf("%s: %v", targetDir, err)}
	}
	return Result{Path: targetDir, Action: "updated"}
}

func writeFresh(targetDir, name, checksum string, opts Options) error {
	if err := writeBundled(targetDir, opts.Bundled); err != nil {
		return err
	}
	meta := Metadata{
		Schema:        metadataSchema,
		BinaryVersion: opts.BinaryVersion,
		Checksum:      checksum,
		Target:        name,
		Source:        "debate-bundled-skill",
	}
	return writeMetadata(targetDir, meta)
}

// safeTarget reports whether targetDir is a descendant of home with no
// symlinked component along the way, following AC7's path-safety rules.
func safeTarget(home, targetDir string) bool {
	rel, err := filepath.Rel(home, targetDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return false
	}
	cur := home
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		cur = filepath.Join(cur, part)
		info, err := os.Lstat(cur)
		if err != nil {
			if os.IsNotExist(err) {
				return true // remaining components do not exist yet; safe to create
			}
			return false
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return false
		}
	}
	return true
}

func readMetadata(targetDir string) (Metadata, bool) {
	data, err := os.ReadFile(filepath.Join(targetDir, MetadataFileName))
	if err != nil {
		return Metadata{}, false
	}
	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return Metadata{}, false
	}
	if meta.Schema != metadataSchema || meta.Checksum == "" {
		return Metadata{}, false
	}
	return meta, true
}

func writeMetadata(targetDir string, meta Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(targetDir, MetadataFileName), data, 0644)
}

func writeBundled(targetDir string, bundled fs.FS) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	return fs.WalkDir(bundled, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		dest := filepath.Join(targetDir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := fs.ReadFile(bundled, path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
}

// hashFiles is the AC6 checksum algorithm: SHA-256 over each file's relative
// path and content, sorted by relative path and concatenated (path length,
// path, content length, content, per file), hashed once.
func hashFiles(files map[string][]byte) string {
	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		fmt.Fprintf(h, "%d:%s:%d:", len(k), k, len(files[k]))
		h.Write(files[k])
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashFS(bundled fs.FS) (string, error) {
	files := make(map[string][]byte)
	err := fs.WalkDir(bundled, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path == MetadataFileName {
			return nil
		}
		data, err := fs.ReadFile(bundled, path)
		if err != nil {
			return err
		}
		files[path] = data
		return nil
	})
	if err != nil {
		return "", err
	}
	return hashFiles(files), nil
}

func hashDir(root string) (string, error) {
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == MetadataFileName {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = data
		return nil
	})
	if err != nil {
		return "", err
	}
	return hashFiles(files), nil
}
