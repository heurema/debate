// Package bundled embeds the debate Agent Skill shipped with the debate binary.
package bundled

import (
	"embed"
	"io/fs"
)

//go:embed all:debate
var files embed.FS

// Skill returns the bundled debate Agent Skill content, rooted at the skill
// directory itself (paths like "SKILL.md" and "references/cli-reference.md"
// rather than "debate/SKILL.md").
func Skill() fs.FS {
	sub, err := fs.Sub(files, "debate")
	if err != nil {
		// files is embedded at build time from a directory literally named
		// "debate"; this cannot fail at runtime.
		panic(err)
	}
	return sub
}
