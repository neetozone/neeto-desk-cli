package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neetozone/neeto-desk-cli/internal/plugin"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up NeetoDesk for AI coding assistants",
	Long: "Set up NeetoDesk for AI coding assistants.\n\n" +
		"Every subcommand except \"claude\" writes into the current project directory,\n" +
		"so run it from the root of the project you want the assistant to use NeetoDesk in.",
}

// --- Claude Code ---

var setupClaudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Register NeetoDesk plugin with Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("Could not determine home directory: %w", err)
		}

		if _, err := os.Stat(filepath.Join(home, ".claude")); os.IsNotExist(err) {
			return fmt.Errorf("Claude Code not found (~/.claude/ does not exist).")
		}

		dest := filepath.Join(home, ".config", "neetodesk", "claude-plugin")
		if err := os.RemoveAll(dest); err != nil {
			return fmt.Errorf("Could not clean destination: %w", err)
		}

		if err := plugin.ExtractClaudePlugin(dest); err != nil {
			return fmt.Errorf("Could not extract plugin: %w", err)
		}

		// Make hooks executable
		hooksDir := filepath.Join(dest, "hooks")
		entries, _ := os.ReadDir(hooksDir)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".sh") {
				_ = os.Chmod(filepath.Join(hooksDir, e.Name()), 0o755)
			}
		}

		fmt.Printf("NeetoDesk plugin extracted to %s\n", dest)
		fmt.Println()
		fmt.Println("To finish installation, open Claude Code and run these slash commands:")
		fmt.Printf("  /plugin marketplace add %s\n", dest)
		fmt.Printf("  /plugin install %s@%s\n", plugin.PluginName, plugin.MarketplaceName)
		fmt.Println()
		fmt.Println("(Claude Code installs plugins via interactive slash commands — there is no shell equivalent today.)")
		return nil
	},
}

// --- Cursor ---

var cursorRulesPath = filepath.Join(".cursor", "rules", "neetodesk.mdc")

var setupCursorCmd = &cobra.Command{
	Use:   "cursor",
	Short: "Write NeetoDesk rules for Cursor IDE",
	Long:  ruleFileHelp(cursorRulesPath),
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeCreateMode(cmd.OutOrStdout(), cursorRulesPath, cursorContent())
	},
}

func cursorContent() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("description: \"NeetoDesk CLI\"\n")
	b.WriteString("alwaysApply: true\n")
	b.WriteString("---\n\n")
	b.WriteString(plugin.SkillBody())
	return b.String()
}

// --- Windsurf ---

var windsurfRulesPath = filepath.Join(".windsurf", "rules", "neetodesk.md")

var setupWindsurfCmd = &cobra.Command{
	Use:   "windsurf",
	Short: "Write NeetoDesk rules for Windsurf IDE",
	Long:  ruleFileHelp(windsurfRulesPath),
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeCreateMode(cmd.OutOrStdout(), windsurfRulesPath, windsurfContent())
	},
}

func windsurfContent() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("trigger: always_on\n")
	b.WriteString("description: \"NeetoDesk CLI\"\n")
	b.WriteString("---\n\n")
	b.WriteString(plugin.SkillBody())
	return b.String()
}

// --- Copilot ---

var copilotInstructionsPath = filepath.Join(".github", "copilot-instructions.md")

var setupCopilotCmd = &cobra.Command{
	Use:   "copilot",
	Short: "Add NeetoDesk instructions for GitHub Copilot",
	Long:  sectionHelp(copilotInstructionsPath),
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeSection(cmd.OutOrStdout(), copilotInstructionsPath, plugin.SkillBody())
	},
}

// --- Gemini ---

var geminiInstructionsPath = "GEMINI.md"

var setupGeminiCmd = &cobra.Command{
	Use:   "gemini",
	Short: "Add NeetoDesk instructions for Gemini CLI",
	Long:  sectionHelp(geminiInstructionsPath),
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeSection(cmd.OutOrStdout(), geminiInstructionsPath, plugin.SkillBody())
	},
}

// --- Codex ---

var codexInstructionsPath = "AGENTS.md"

var setupCodexCmd = &cobra.Command{
	Use:   "codex",
	Short: "Add NeetoDesk instructions for OpenAI Codex",
	Long:  sectionHelp(codexInstructionsPath),
	RunE: func(cmd *cobra.Command, args []string) error {
		return writeSection(cmd.OutOrStdout(), codexInstructionsPath, plugin.SkillBody())
	},
}

func sectionHelp(target string) string {
	return fmt.Sprintf(
		"Add a NeetoDesk section to %s in the current project directory.\n\n"+
			"Existing content in the file is kept. Re-running replaces the NeetoDesk section\n"+
			"instead of adding a duplicate, so run it again after every upgrade.",
		target,
	)
}

func ruleFileHelp(target string) string {
	return fmt.Sprintf(
		"Write the NeetoDesk rule file to %s in the current project directory.\n\n"+
			"An existing file is left untouched. Delete it and re-run to regenerate it.",
		target,
	)
}

// --- Helpers ---

// writeCreateMode writes a file, creating parent dirs. Skips if already present.
func writeCreateMode(w io.Writer, target, content string) error {
	if _, err := os.Stat(target); err == nil {
		fmt.Fprintf(w, "Already installed: %s\n", target)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return err
	}

	fmt.Fprintf(w, "Wrote %s\n", target)
	return nil
}

func sectionMarkers() (string, string) {
	name := rootCmd.Name()
	return "<!-- " + name + ":start -->", "<!-- " + name + ":end -->"
}

func writeSection(w io.Writer, target, body string) error {
	_, statErr := os.Stat(target)
	start, end := sectionMarkers()
	if _, err := upsertBlock(target, start, end, "## NeetoDesk CLI\n\n"+body); err != nil {
		return err
	}
	if statErr == nil {
		fmt.Fprintf(w, "Updated %s\n", target)
	} else {
		fmt.Fprintf(w, "Wrote %s\n", target)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(setupClaudeCmd)
	setupCmd.AddCommand(setupCursorCmd)
	setupCmd.AddCommand(setupWindsurfCmd)
	setupCmd.AddCommand(setupCopilotCmd)
	setupCmd.AddCommand(setupGeminiCmd)
	setupCmd.AddCommand(setupCodexCmd)
}
