package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/saireddy-shyamakura/springx/internal/plugins"
	"github.com/spf13/cobra"
)

// ── Root plugin command ───────────────────────────────────────────────────────

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage springx plugins",
	Long: `Manage springx plugins.

Plugins extend springx with additional templates, post-generation hooks,
and dependency groups. They are compiled into the binary via blank imports
and discovered at runtime from ~/.config/springx/plugins/.

Use the subcommands to inspect and toggle plugins:

  springx plugin list
  springx plugin info <name>
  springx plugin enable <name>
  springx plugin disable <name>`,
}

// ── plugin list ───────────────────────────────────────────────────────────────

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered plugins",
	Long:  `Display all plugins compiled into this springx binary, together with their version, status, and contributions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all := plugins.Registered()
		if len(all) == 0 {
			fmt.Println("No plugins registered.")
			fmt.Println("See the plugin authoring guide: https://github.com/saireddy-shyamakura/springx#plugins")
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

// ── plugin info ───────────────────────────────────────────────────────────────

var pluginInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		p, ok := plugins.Lookup(name)
		if !ok {
			return fmt.Errorf("plugin %q is not registered — run 'springx plugin list' to see available plugins", name)
		}

		m := p.Manifest()
		s := plugins.Summary(p)

		status := "enabled"
		if !plugins.IsEnabled(m.Name) {
			status = "disabled"
		}

		fmt.Printf("\n")
		fmt.Printf("  Plugin:       %s\n", m.Name)
		fmt.Printf("  Version:      %s\n", m.Version)
		fmt.Printf("  Author:       %s\n", m.Author)
		fmt.Printf("  Status:       %s\n", status)
		fmt.Printf("  Description:  %s\n", m.Description)
		if m.Homepage != "" {
			fmt.Printf("  Homepage:     %s\n", m.Homepage)
		}
		if len(m.Tags) > 0 {
			fmt.Printf("  Tags:         %s\n", strings.Join(m.Tags, ", "))
		}

		fmt.Printf("\n  Contributions:\n")
		if s.Templates > 0 {
			fmt.Printf("    Templates:         %d\n", s.Templates)
			if tp, ok := p.(plugins.TemplatePlugin); ok {
				for _, t := range tp.Templates() {
					fmt.Printf("      • %s — %s\n", t.Name, t.Description)
				}
			}
		}
		if s.Hooks > 0 {
			fmt.Printf("    Hooks:             %d\n", s.Hooks)
			if hp, ok := p.(plugins.HookPlugin); ok {
				for _, h := range hp.Hooks() {
					fmt.Printf("      • %s\n", h.Name())
				}
			}
		}
		if s.DependencyGroups > 0 {
			fmt.Printf("    Dependency groups: %d (%d total dependencies)\n",
				s.DependencyGroups, s.Dependencies)
			if dp, ok := p.(plugins.DependencyProvider); ok {
				for _, g := range dp.DependencyGroups() {
					fmt.Printf("      • %s (%d deps)\n", g.Name, len(g.Values))
				}
			}
		}
		if s.Templates == 0 && s.Hooks == 0 && s.DependencyGroups == 0 {
			fmt.Printf("    (none — this plugin only satisfies the base Plugin interface)\n")
		}
		fmt.Printf("\n")
		return nil
	},
}

// ── plugin enable ─────────────────────────────────────────────────────────────

var pluginEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a disabled plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if _, ok := plugins.Lookup(name); !ok {
			return fmt.Errorf("plugin %q is not registered", name)
		}
		if err := plugins.SetEnabled(name, true); err != nil {
			return err
		}
		if err := plugins.PersistEnabled(name); err != nil {
			return fmt.Errorf("plugin enabled in memory but could not save to disk: %w", err)
		}
		fmt.Printf("✔ Plugin %q enabled.\n", name)
		return nil
	},
}

// ── plugin disable ────────────────────────────────────────────────────────────

var pluginDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a plugin without removing it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if _, ok := plugins.Lookup(name); !ok {
			return fmt.Errorf("plugin %q is not registered", name)
		}
		if err := plugins.SetEnabled(name, false); err != nil {
			return err
		}
		if err := plugins.PersistDisabled(name); err != nil {
			return fmt.Errorf("plugin disabled in memory but could not save to disk: %w", err)
		}
		fmt.Printf("✔ Plugin %q disabled.\n", name)
		return nil
	},
}

// ── registration ──────────────────────────────────────────────────────────────

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	rootCmd.AddCommand(pluginCmd)
}
