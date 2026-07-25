package cmd

import (
	"fmt"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/templates"
	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:     "template",
	Aliases: []string{"tmpl"},
	Short:   "View and inspect opinionated project templates",
}

var templateListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available built-in project templates",
	Run: func(cmd *cobra.Command, args []string) {
		all := templates.List()
		fmt.Println("Available templates:")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, t := range all {
			fmt.Printf("%-15s - %s\n", t.Name, t.Description)
		}
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("Use 'springx template info <name>' to view template details.")
	},
}

var templateInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed information about a template preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		tmpl, err := templates.Get(name)
		if err != nil {
			return err
		}

		fmt.Printf("Template: %s\n", tmpl.Name)
		fmt.Printf("Description: %s\n", tmpl.Description)
		fmt.Printf("Dependencies: %s\n", strings.Join(tmpl.Dependencies, ", "))
		fmt.Println("Defaults:")
		if tmpl.Defaults.JavaVersion != "" {
			fmt.Printf("  Java Version: %s\n", tmpl.Defaults.JavaVersion)
		}
		if tmpl.Defaults.BuildTool != "" {
			fmt.Printf("  Build Tool  : %s\n", tmpl.Defaults.BuildTool)
		}
		if tmpl.Defaults.Packaging != "" {
			fmt.Printf("  Packaging   : %s\n", tmpl.Defaults.Packaging)
		}
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateInfoCmd)

	rootCmd.AddCommand(templateCmd)
}
