// Package cli implements the `2m skill` subcommands for managing skills.
//
// Subcommands:
//
//	2m skill list   — List all available skills
//	2m skill show   — Show full content of a specific skill
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// skillCmd is the parent command for skill management.
var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "List and view built-in skills",
	Long: `List and view built-in skills that can be injected into agent prompts.

Skills are structured instruction sets stored in the Skills/ directory.
When a skill name is detected in your message, its instructions are
automatically injected into the agent's system prompt.

Subcommands:
  list  — List all available skills
  show  — Show full content of a specific skill`,
}

// skillListCmd lists all discovered skills.
var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available skills",
	Long:  "Show all skills found in the Skills/ directory with their name, description, and license.",
	RunE:  listSkills,
}

// skillShowCmd shows full content of a specific skill.
var skillShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show full content of a specific skill",
	Long:  "Display the complete content of a skill by name, including its YAML frontmatter and body.",
	Args:  cobra.ExactArgs(1),
	RunE:  showSkill,
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillShowCmd)
}

func listSkills(cmd *cobra.Command, args []string) error {
	engineURL := "http://127.0.0.1:8765/skills"
	resp, err := http.Get(engineURL)
	if err != nil {
		return fmt.Errorf("connecting to agent engine: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Skills []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			License     string `json:"license"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	badge := lipgloss.NewStyle().
		Background(lipgloss.Color("#555555")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	if len(result.Skills) == 0 {
		fmt.Println(badge.Render("No skills found"))
		fmt.Println()
		fmt.Println("Place SKILL.md files in the Skills/ directory:")
		fmt.Println("  Skills/<skill-name>/SKILL.md")
		return nil
	}

	fmt.Println(title.Render("Available Skills"))
	fmt.Println()
	for _, s := range result.Skills {
		fmt.Printf("  %s\n", value.Render(s.Name))
		if s.Description != "" {
			// Truncate long descriptions for list view
			desc := s.Description
			if len(desc) > 120 {
				desc = desc[:117] + "..."
			}
			fmt.Printf("    %s %s\n", label.Render("description:"), desc)
		}
		if s.License != "" {
			fmt.Printf("    %s %s\n", label.Render("license:"), s.License)
		}
		fmt.Println()
	}

	fmt.Printf("Use %s to view full content.\n", value.Render("2m skill show <name>"))
	return nil
}

func showSkill(cmd *cobra.Command, args []string) error {
	name := args[0]

	engineURL := fmt.Sprintf("http://127.0.0.1:8765/skills/%s", name)
	resp, err := http.Get(engineURL)
	if err != nil {
		return fmt.Errorf("connecting to agent engine: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 404 {
		return fmt.Errorf("skill '%s' not found", name)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("engine returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Skill struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			License     string `json:"license"`
			Content     string `json:"content"`
			Path        string `json:"path"`
		} `json:"skill"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	fmt.Printf("%s %s\n", title.Render("Skill:"), value.Render(result.Skill.Name))
	if result.Skill.Description != "" {
		fmt.Printf("%s %s\n", label.Render("Description:"), result.Skill.Description)
	}
	if result.Skill.License != "" {
		fmt.Printf("%s %s\n", label.Render("License:"), result.Skill.License)
	}
	if result.Skill.Path != "" {
		fmt.Printf("%s %s\n", label.Render("Path:"), result.Skill.Path)
	}
	fmt.Println()
	fmt.Println(title.Render("Content"))
	fmt.Println()
	// Render content with line breaks preserved
	lines := strings.Split(result.Skill.Content, "\n")
	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}
