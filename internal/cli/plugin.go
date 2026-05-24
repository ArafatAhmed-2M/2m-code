// Package cli implements the `2m plugin` subcommands for managing plugins.
//
// Subcommands:
//
//	2m plugin list   — List all discovered plugins
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/2mcode/2mcode/internal/bridge"
)

var (
	pluginDirFlag string
)

// pluginCmd is the parent command for plugin management.
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage agent engine plugins",
	Long: `List and manage plugins that extend the agent engine.

Plugins are Python files placed in:
  ~/.2mcode/plugins/   (global)
  .2mcode/plugins/      (project-local)

Subcommands:
  list  — List all discovered plugins and their hooks`,
}

// pluginListCmd lists all discovered plugins.
var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered plugins",
	Long:  "Show all plugins found in global and project-local directories, including their lifecycle hooks.",
	RunE:  listPlugins,
}

func init() {
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.AddCommand(pluginListCmd)

	pluginListCmd.Flags().StringVarP(&pluginDirFlag, "dir", "d", "", "Scan a specific directory instead of default locations")
}

func listPlugins(cmd *cobra.Command, args []string) error {
	// Scan filesystem for plugin files
	dirs := []string{
		filepath.Join(os.Getenv("HOME"), ".2mcode", "plugins"),
		".2mcode" + string(filepath.Separator) + "plugins",
	}
	if pluginDirFlag != "" {
		dirs = []string{pluginDirFlag}
	}

	type pluginFile struct {
		Path string
		Dir  string
	}
	var localPlugins []pluginFile

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				continue
			}
			return fmt.Errorf("reading plugin directory %s: %w", dir, err)
		}
		dirLabel := "project"
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" && strings.HasPrefix(dir, homeDir) {
			dirLabel = "global"
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".py") || strings.HasPrefix(e.Name(), "_") {
				continue
			}
			localPlugins = append(localPlugins, pluginFile{
				Path: filepath.Join(dir, e.Name()),
				Dir:  dirLabel,
			})
		}
	}

	// Fetch loaded plugin info from the running engine
	var loadedPlugins []map[string]any
	engineURL := "http://127.0.0.1:8765/plugins"
	resp, err := http.Get(engineURL)
	if err == nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var result struct {
			Plugins []map[string]any `json:"plugins"`
		}
		if err := json.Unmarshal(body, &result); err == nil {
			loadedPlugins = result.Plugins
		}
	}

	// Render output
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	value := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	badge := lipgloss.NewStyle().
		Background(lipgloss.Color("#555555")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	if len(localPlugins) == 0 && len(loadedPlugins) == 0 {
		fmt.Println(badge.Render("No plugins found"))
		fmt.Println()
		fmt.Println("To create a plugin, place a .py file in:")
		fmt.Println("  " + bridge.StylePath("~/.2mcode/plugins/"))
		fmt.Println("  " + bridge.StylePath(".2mcode/plugins/"))
		fmt.Println()
		fmt.Println("See agent_engine/plugin_base.py for the Plugin interface.")
		return nil
	}

	// Filesystem plugins
	if len(localPlugins) > 0 {
		fmt.Println(title.Render("Plugin Files"))
		for _, p := range localPlugins {
			dirLabel := badge.Render(p.Dir)
			fmt.Printf("  %s  %s\n", dirLabel, value.Render(p.Path))
		}
		fmt.Println()
	}

	// Loaded engine plugins (with hook info)
	if len(loadedPlugins) > 0 {
		fmt.Println(title.Render("Loaded Plugins"))
		for _, p := range loadedPlugins {
			name, _ := p["name"].(string)
			fmt.Printf("  %s\n", value.Render(name))

			hooksRaw, _ := p["hooks"].([]any)
			if len(hooksRaw) > 0 {
				var hooks []string
				for _, h := range hooksRaw {
					hooks = append(hooks, fmt.Sprintf("%v", h))
				}
				fmt.Printf("    %s %s\n", label.Render("hooks:"), strings.Join(hooks, ", "))
			}
		}
	}

	return nil
}
