// Package cli implements the `2m team` subcommands for managing team configs.
//
// Subcommands:
//   2m team list           — List all available teams
//   2m team show <name>    — Show team configuration details
//   2m history <team>      — Show last session's team channel log
//   2m config set <key>    — Set global config values
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/2mcode/2mcode/internal/bus"
	"github.com/2mcode/2mcode/internal/team"
)

// teamCmd is the parent command for team management.
var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Manage agent team configurations",
	Long: `View and manage your agent team configurations.

Subcommands:
  list  — List all available teams
  show  — Show detailed configuration for a specific team`,
}

// teamListCmd lists all available teams.
var teamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available teams",
	Long:  "Show all team configurations found in project-local, global, and bundled locations.",
	RunE:  listTeams,
}

// teamShowCmd shows details for a specific team.
var teamShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show team configuration details",
	Long:  "Display the full configuration for a specific team, including all agents and workflow settings.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  showTeam,
}

// historyCmd shows the last session's log.
var historyCmd = &cobra.Command{
	Use:   "history <team>",
	Short: "Show last session's team channel log",
	Long:  "Display the conversation history from the most recent session with a team.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  showHistory,
}

// configCmd manages global configuration.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global 2M Code configuration",
}

// configSetCmd sets a global config value.
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a global configuration value",
	Long: `Set a global configuration value.

Available keys:
  default_team     — Team to use when none is specified
  default_provider — Provider to use when creating new teams
  verbose          — Enable verbose output (true/false)`,
	Args: cobra.ExactArgs(2),
	RunE: setConfig,
}

func init() {
	// Register subcommands
	teamCmd.AddCommand(teamListCmd)
	teamCmd.AddCommand(teamShowCmd)
	rootCmd.AddCommand(teamCmd)
	rootCmd.AddCommand(historyCmd)

	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}

// listTeams handles `2m team list`.
func listTeams(cmd *cobra.Command, args []string) error {
	teams, err := team.ListTeams()
	if err != nil {
		return fmt.Errorf("cannot list teams: %w", err)
	}

	if len(teams) == 0 {
		fmt.Println("No teams found. Run '2m new-team' to create one.")
		return nil
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))

	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	fmt.Println(headerStyle.Render("Available teams:"))
	fmt.Println()

	for name, path := range teams {
		// Load the team to show description
		t, err := team.LoadTeam(name)
		desc := ""
		if err == nil && t.Description != "" {
			desc = " — " + t.Description
		}

		fmt.Printf("  %s%s\n", headerStyle.Render(name), desc)
		fmt.Printf("  %s\n\n", pathStyle.Render(path))
	}

	return nil
}

// showTeam handles `2m team show <name>`.
// The team name may contain spaces so all positional args are joined.
func showTeam(cmd *cobra.Command, args []string) error {
	teamName := strings.Join(args, " ")
	renderer := NewRenderer()

	t, err := team.LoadTeam(teamName)
	if err != nil {
		renderer.PrintError(err.Error())
		return err
	}

	renderer.PrintTeamInfo(t)

	// Show agent details
	for _, agent := range t.Agents {
		detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

		fmt.Printf("  %s\n", detailStyle.Render("System prompt (first 200 chars):"))
		prompt := agent.SystemPrompt
		if len(prompt) > 200 {
			prompt = prompt[:200] + "..."
		}
		fmt.Printf("  %s\n\n", detailStyle.Render(prompt))
	}

	// Show API key status
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).Render("API Key Status:"))
	seen := make(map[string]bool)
	for _, agent := range t.Agents {
		if seen[agent.Provider] {
			continue
		}
		seen[agent.Provider] = true

		_, err := team.GetProviderAPIKey(agent.Provider)
		if err != nil {
			fmt.Printf("  %s: %s\n", agent.Provider, lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✗ not set"))
		} else {
			fmt.Printf("  %s: %s\n", agent.Provider, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓ configured"))
		}
	}
	fmt.Println()

	return nil
}

// showHistory handles `2m history <team>`.
// The team name may contain spaces so all positional args are joined.
func showHistory(cmd *cobra.Command, args []string) error {
	teamName := strings.Join(args, " ")
	renderer := NewRenderer()

	// Load team to get agent roles/colors for rendering
	t, err := team.LoadTeam(teamName)
	if err != nil {
		// Non-fatal — we can still show history without agent metadata
		t = &team.Team{Name: teamName}
	}

	// Build agent lookup (name → agent)
	agentLookup := make(map[string]team.Agent)
	for _, a := range t.Agents {
		agentLookup[a.Name] = a
	}

	// Get the sessions directory for this team
	sessDir, err := team.SessionsPath(teamName)
	if err != nil {
		return fmt.Errorf("cannot determine sessions path: %w", err)
	}

	// Check if the sessions directory exists
	if _, err := os.Stat(sessDir); os.IsNotExist(err) {
		renderer.PrintInfo(fmt.Sprintf("No sessions found for team '%s'", teamName))
		renderer.PrintInfo("Run a task first with: 2m run " + teamName + " \"<task>\"")
		return nil
	}

	// Find the latest .db file by modification time
	entries, err := os.ReadDir(sessDir)
	if err != nil {
		return fmt.Errorf("cannot read sessions directory %s: %w", sessDir, err)
	}

	var latestDB string
	var latestTime time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".db") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestDB = filepath.Join(sessDir, e.Name())
		}
	}

	if latestDB == "" {
		renderer.PrintInfo(fmt.Sprintf("No session databases found for team '%s'", teamName))
		renderer.PrintInfo("Run a task first with: 2m run " + teamName + " \"<task>\"")
		return nil
	}

	// Open the database
	db, err := bus.InitDB(latestDB)
	if err != nil {
		return fmt.Errorf("cannot open session database: %w", err)
	}
	defer db.Close()

	eventBus := bus.New(db)

	// Get the latest session ID from the sessions table
	sessionID, err := eventBus.GetLatestSessionID(teamName)
	if err != nil {
		return fmt.Errorf("cannot get latest session: %w", err)
	}
	if sessionID == "" {
		renderer.PrintInfo(fmt.Sprintf("No sessions found for team '%s'", teamName))
		return nil
	}

	// Get all messages
	messages, err := eventBus.GetAllMessages(sessionID)
	if err != nil {
		return fmt.Errorf("cannot read messages: %w", err)
	}

	if len(messages) == 0 {
		renderer.PrintInfo("Session is empty")
		return nil
	}

	// Display session header
	renderer.PrintInfo(fmt.Sprintf("History for team '%s': %d messages", teamName, len(messages)))
	renderer.PrintInfo(fmt.Sprintf("Session: %s | %s", sessionID, latestTime.Format("Jan 2, 2006 15:04:05")))
	fmt.Println()

	// Display each message
	for _, msg := range messages {
		timeStr := msg.CreatedAt.Format("15:04:05")
		if msg.Role == "user" {
			renderer.PrintInfo(fmt.Sprintf("[%s] You:", timeStr))
			fmt.Printf("  %s\n\n", msg.Content)
		} else {
			// Use agent's assigned color if available
			a, known := agentLookup[msg.AgentName]
			if known {
				clr := colorMap["cyan"]
				if c, ok := colorMap[a.Color]; ok {
					clr = c
				}
				badge := lipgloss.NewStyle().
					Background(clr).
					Foreground(lipgloss.Color("0")).
					Padding(0, 1).
					Render(fmt.Sprintf(" %s · %s ", msg.AgentName, a.Role))
				fmt.Printf("  %s  [%s]\n", badge, timeStr)
			} else {
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Padding(0, 1)
				fmt.Printf("  %s  [%s]\n", style.Render(msg.AgentName), timeStr)
			}

			// Print content line by line
			for _, line := range strings.Split(msg.Content, "\n") {
				fmt.Printf("  │ %s\n", line)
			}

			if msg.ToolCalls != "" {
				fmt.Printf("  │ \x1b[90m[Tool calls: %s]\x1b[0m\n", msg.ToolCalls)
			}
			fmt.Println()
		}
	}

	return nil
}

// setConfig handles `2m config set <key> <value>`.
func setConfig(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := team.LoadConfig()
	if err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}

	switch key {
	case "default_team":
		cfg.DefaultTeam = value
	case "default_provider":
		cfg.DefaultProvider = value
	case "verbose":
		cfg.Verbose = value == "true"
	default:
		return fmt.Errorf("unknown config key '%s' — use: default_team, default_provider, verbose", key)
	}

	if err := team.SaveConfig(cfg); err != nil {
		return fmt.Errorf("cannot save config: %w", err)
	}

	fmt.Printf("Set %s = %s\n", key, value)
	return nil
}
