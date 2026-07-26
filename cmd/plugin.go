package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/saireddy-shyamakura/springx/internal/plugins"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage springx plugins",
	Long: `Manage springx plugins.

Plugins extend springx with additional templates, post-generation hooks,
and dependency groups. They are compiled into the binary via blank imports
and discovered at runtime from:

  Linux / macOS : ~/.config/springx/plugins/<name>/plugin.json
  Windows       : %APPDATA%\springx\plugins\<name>\plugin.json

Disabling a plugin prevents its templates, hooks, and dependency groups
from being active without requiring a rebuild. The enable/disable state
is persisted between runs.`,
	Example: `  springx plugin list
  springx plugin info aws
  springx plugin enable aws
  springx plugin disable aws`,
}

var pluginListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all registered plugins",
	Long:    `Display all plugins compiled into this binary with their version, status, and contribution summary.`,
	Example: `  springx plugin list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all := plugins.Registered()
		if len(all) == 0 {
			fmt.Println("No plugins are registered in this build.")
			fmt.Println()
			fmt.Println("To add a plugin, blank-import its package in main.go and rebuild.")
			fmt.Println("See: https://github.com/saireddy-shyamakura/springx#plugin-system")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tSTATUS\tDESCRIPTION")
		fmt.Fprintln(w, "────\t───────\t──────\t───────────")
		for _, p := range all {
			m := p.Manifest()
			status := "enabled"
			if !plugins.IsEnabled(m.Name) {
				status = "disabled"
			}
			desc := m.Description
			if len(desc) > 60 {
				desc = desc[:57] + "…"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Name, m.Version, status, desc)
		}
		return w.Flush()
	},
}

var pluginInfoCmd = &cobra.Command{
	Use:     "info <name>",
	Short:   "Show detailed information about a plugin",
	Long:    `Display full details for a plugin: manifest metadata, status, and every template, hook, and dependency group it contributes.`,
	Example: `  springx plugin info aws`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		p, ok := plugins.Lookup(name)
		if !ok {
			return fmt.Errorf("plugin %q is not registered\n\nRun 'springx plugin list' to see available plugins", name)
		}

		m := p.Manifest()
		s := plugins.Summary(p)

		status := "enabled"
		if !plugins.IsEnabled(m.Name) {
			status = "disabled"
		}

		fmt.Printf("\n  Plugin       : %s\n", m.Name)
		fmt.Printf("  Version      : %s\n", m.Version)
		fmt.Printf("  Author       : %s\n", m.Author)
		fmt.Printf("  Status       : %s\n", status)
		fmt.Printf("  Description  : %s\n", m.Description)
		if m.Homepage != "" {
			fmt.Printf("  Homepage     : %s\n", m.Homepage)
		}
		if len(m.Tags) > 0 {
			fmt.Printf("  Tags         : %s\n", strings.Join(m.Tags, ", "))
		}

		fmt.Printf("\n  Contributions:\n")
		if s.Templates > 0 {
			fmt.Printf("    Templates (%d):\n", s.Templates)
			if tp, ok := p.(plugins.TemplatePlugin); ok {
				for _, t := range tp.Templates() {
					fmt.Printf("      • %-20s %s\n", t.Name, t.Description)
				}
			}
		}
		if s.Hooks > 0 {
			fmt.Printf("    Hooks (%d):\n", s.Hooks)
			if hp, ok := p.(plugins.HookPlugin); ok {
				for _, h := range hp.Hooks() {
					fmt.Printf("      • %s\n", h.Name())
				}
			}
		}
		if s.DependencyGroups > 0 {
			fmt.Printf("    Dependency groups (%d, %d total deps):\n",
				s.DependencyGroups, s.Dependencies)
			if dp, ok := p.(plugins.DependencyProvider); ok {
				for _, g := range dp.DependencyGroups() {
					fmt.Printf("      • %s (%d deps)\n", g.Name, len(g.Values))
				}
			}
		}
		if s.Templates == 0 && s.Hooks == 0 && s.DependencyGroups == 0 {
			fmt.Println("    (no contributions beyond the base Plugin interface)")
		}
		fmt.Println()
		return nil
	},
}

var pluginEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a disabled plugin",
	Long: `Re-enable a plugin that was previously disabled.

The plugin's templates, hooks, and dependency groups become active
immediately and the enabled state is saved for future runs.`,
	Example: `  springx plugin enable aws`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if _, ok := plugins.Lookup(name); !ok {
			return fmt.Errorf("plugin %q is not registered — run 'springx plugin list'", name)
		}
		if err := plugins.SetEnabled(name, true); err != nil {
			return err
		}
		if err := plugins.PersistEnabled(name); err != nil {
			return fmt.Errorf("plugin enabled in memory but state could not be saved: %w", err)
		}
		fmt.Printf("✔  Plugin %q enabled.\n", name)
		return nil
	},
}

var pluginDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a plugin without removing it",
	Long: `Disable a plugin so its contributions are no longer active.

The plugin remains compiled into the binary; disabling only prevents
its templates, hooks, and dependency groups from being visible or
executed. Run 'springx plugin enable <name>' to re-activate it.`,
	Example: `  springx plugin disable aws`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if _, ok := plugins.Lookup(name); !ok {
			return fmt.Errorf("plugin %q is not registered — run 'springx plugin list'", name)
		}
		if err := plugins.SetEnabled(name, false); err != nil {
			return err
		}
		if err := plugins.PersistDisabled(name); err != nil {
			return fmt.Errorf("plugin disabled in memory but state could not be saved: %w", err)
		}
		fmt.Printf("✔  Plugin %q disabled.\n", name)
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	rootCmd.AddCommand(pluginCmd)
}
