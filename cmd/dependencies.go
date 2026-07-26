package cmd

import (
	"fmt"
	"os"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/spf13/cobra"
)

var dependenciesCmd = &cobra.Command{
	Use:     "dependencies",
	Aliases: []string{"deps"},
	Short:   "List available Spring Boot dependencies grouped by category",
	Long: `Fetch and display all available Spring Boot dependencies from Spring Initializr.

Dependencies are fetched live from start.spring.io and displayed in their
original category groups. This is a read-only command; it does not modify
any project files.

Use the interactive dependency picker inside 'springx new' to select
dependencies for a project.`,
	Example: `  springx dependencies
  springx deps`,
	RunE: func(cmd *cobra.Command, args []string) error {
		meta, err := metadata.Fetch()
		if err != nil {
			return fmt.Errorf("failed to fetch dependency metadata: %w\n\nCheck your internet connection and that start.spring.io is reachable", err)
		}
		meta.PrintDependencyGroups(os.Stdout)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dependenciesCmd)
}
