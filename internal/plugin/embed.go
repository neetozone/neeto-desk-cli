package plugin

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

//go:embed skill.md
var skillContent string

const (
	// PluginName matches `.claude-plugin/plugin.json#name` and the entry
	// under `marketplace.json#plugins[].name`.
	PluginName = "neetodesk"
	// MarketplaceName matches `.claude-plugin/marketplace.json#name`.
	MarketplaceName = "neetodesk"
)

// SkillBody returns the SKILL.md content with YAML frontmatter stripped.
func SkillBody() string {
	content := skillContent
	if strings.HasPrefix(content, "---") {
		if idx := strings.Index(content[3:], "---"); idx != -1 {
			content = strings.TrimLeft(content[3+idx+3:], "\n")
		}
	}
	return content
}

// ExtractClaudePlugin writes a Claude Code plugin (and a single-plugin
// marketplace pointing at it) to dest. The layout follows
// https://code.claude.com/docs/en/plugins-reference:
//
//	dest/
//	├── .claude-plugin/
//	│   ├── plugin.json
//	│   └── marketplace.json
//	├── skills/<name>/SKILL.md
//	├── commands/<cmd>.md
//	└── hooks/{hooks.json,*.sh}
func ExtractClaudePlugin(dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	manifest := map[string]interface{}{
		"name":        PluginName,
		"description": "NeetoDesk CLI",
		"author": map[string]string{
			"name":  "BigBinary",
			"email": "support@bigbinary.com",
		},
		"homepage":   "https://github.com/neetozone/neeto-desk-cli",
		"repository": "https://github.com/neetozone/neeto-desk-cli",
	}
	manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
	if err := writeFile(filepath.Join(dest, ".claude-plugin", "plugin.json"), manifestJSON, 0o644); err != nil {
		return err
	}

	marketplace := map[string]interface{}{
		"name": MarketplaceName,
		"owner": map[string]string{
			"name":  "BigBinary",
			"email": "support@bigbinary.com",
		},
		"plugins": []map[string]string{
			{
				"name":        PluginName,
				"source":      "./",
				"description": "NeetoDesk CLI",
			},
		},
	}
	marketplaceJSON, _ := json.MarshalIndent(marketplace, "", "  ")
	if err := writeFile(filepath.Join(dest, ".claude-plugin", "marketplace.json"), marketplaceJSON, 0o644); err != nil {
		return err
	}

	hooks := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []interface{}{
				map[string]interface{}{
					"hooks": []interface{}{
						map[string]interface{}{
							"type":    "command",
							"command": "${CLAUDE_PLUGIN_ROOT}/hooks/session-start.sh",
							"timeout": 5,
						},
					},
				},
			},
		},
	}
	hooksJSON, _ := json.MarshalIndent(hooks, "", "  ")
	if err := writeFile(filepath.Join(dest, "hooks", "hooks.json"), hooksJSON, 0o644); err != nil {
		return err
	}

	sessionStartSh := "#!/bin/sh\n" +
		"# NeetoDesk CLI — session-start hook for Claude Code\n" +
		"# Lightweight auth liveness check. Always exits 0 (informational).\n" +
		"\n" +
		"if ! command -v neetodesk >/dev/null 2>&1; then\n" +
		"  echo \"NeetoDesk CLI is not installed or not on PATH.\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"\n" +
		"if neetodesk whoami >/dev/null 2>&1; then\n" +
		"  echo \"NeetoDesk plugin active.\"\n" +
		"else\n" +
		"  echo \"NeetoDesk CLI installed but not authenticated. Run 'neetodesk login' to authenticate.\"\n" +
		"fi\n" +
		"\n" +
		"exit 0\n"
	if err := writeFile(filepath.Join(dest, "hooks", "session-start.sh"), []byte(sessionStartSh), 0o755); err != nil {
		return err
	}

	doctorMd := "---\n" +
		"name: neetodesk-doctor\n" +
		"description: Check NeetoDesk CLI health — auth, API connectivity.\n" +
		"invocable: true\n" +
		"---\n" +
		"\n" +
		"Run `neetodesk doctor` and report the results to the user.\n"
	if err := writeFile(filepath.Join(dest, "commands", "doctor.md"), []byte(doctorMd), 0o644); err != nil {
		return err
	}

	if err := writeFile(filepath.Join(dest, "skills", "neetodesk", "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		return err
	}

	return nil
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}
