package bundled_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/heurema/debate/internal/debate/skills/bundled"
)

func TestSkill_ContainsValidSkillMD(t *testing.T) {
	skill := bundled.Skill()
	data, err := fs.ReadFile(skill, "SKILL.md")
	if err != nil {
		t.Fatalf("ReadFile(SKILL.md): %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf("SKILL.md does not start with YAML frontmatter")
	}
	if !strings.Contains(content, "name: debate") {
		t.Errorf("SKILL.md frontmatter missing name: debate")
	}
	if !strings.Contains(content, "description:") {
		t.Errorf("SKILL.md frontmatter missing description")
	}
}

func TestSkill_ReferencesPresent(t *testing.T) {
	skill := bundled.Skill()
	for _, name := range []string{
		"references/cli-reference.md",
		"references/workspace-format.md",
		"references/progress-stream.md",
		"references/panel-guidance.md",
	} {
		if _, err := fs.ReadFile(skill, name); err != nil {
			t.Errorf("ReadFile(%s): %v", name, err)
		}
	}
}
