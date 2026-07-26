package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"os"

	"github.com/saireddy-shyamakura/springx/internal/templates"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:     "template",
	Aliases: []string{"tmpl"},
	Short:   "View and inspect opinionated project templates",
	Long: `List and inspect opinionated project templates.

Templates are pre-configured dependency sets that let you bootstrap a
well-structured project with a single flag instead of selecting every
dependency by hand. They are applied before the interactive dependency
picker opens, so you can still add or remove dependencies afterward.

Use 'springx new --template <name>' to create a project from a template.`,
	Example: `  springx template list
  springx template info rest-api
  springx new --template jpa`,
}

var templateListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available project templates",
	Long: `Display all built-in project templates with their descriptions.

Templates are opinionated presets maintained by the springx team. Plugins
may register additional templates; run 'springx plugin list' to see any
third-party templates that are currently enabled.`,
	Example: `  springx template list
  springx tmpl ls`,
	Run: func(cmd *cobra.Command, args []string) {
		all := templates.List()
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION")
		fmt.Fprintln(w, "────\t───────────")
		for _, t := range all {
			fmt.Fprintf(w, "%s\t%s\n", t.Name, t.Description)
		}
		w.Flush() //nolint:errcheck
		fmt.Println()
		fmt.Println("Use 'springx template info <name>' for full details.")
		fmt.Println("Use 'springx new --template <name>' to scaffold a project.")
	},
}

var templateInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a template",
	Long: `Show all details for a named template: description, dependency list,
and default settings (Java version, build tool, packaging).`,
	Example: `  springx template info rest-api
  springx template info jpa
  springx tmpl info kafka`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		tmpl, err := templates.Get(name)
		if err != nil {
			return err
		}

		fmt.Printf("\n  Template     : %s\n", tmpl.Name)
		fmt.Printf("  Description  : %s\n", tmpl.Description)
		fmt.Printf("  Dependencies : %s\n", strings.Join(tmpl.Dependencies, ", "))
		fmt.Println("  Defaults:")
		if tmpl.Defaults.JavaVersion != "" {
			fmt.Printf("    Java Version : %s\n", tmpl.Defaults.JavaVersion)
		}
		if tmpl.Defaults.BuildTool != "" {
			fmt.Printf("    Build Tool   : %s\n", tmpl.Defaults.BuildTool)
		}
		if tmpl.Defaults.Packaging != "" {
			fmt.Printf("    Packaging    : %s\n", tmpl.Defaults.Packaging)
		}
		fmt.Printf("\n  To use: springx new --template %s\n\n", tmpl.Name)
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateInfoCmd)
	rootCmd.AddCommand(templateCmd)
}
