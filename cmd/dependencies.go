package cmd

import (
	"fmt"
	"os"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/spf13/cobra"
)

// dependenciesCmd represents the dependencies command
var dependenciesCmd = &cobra.Command{
	Use:     "dependencies",
	Aliases: []string{"deps"},
	Short:   "List available dependencies from Spring Initializr grouped by category",
	RunE: func(cmd *cobra.Command, args []string) error {
		meta, err := metadata.Fetch()
		if err != nil {
			return fmt.Errorf("failed to fetch metadata: %w", err)
		}

		meta.PrintDependencyGroups(os.Stdout)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dependenciesCmd)
}
