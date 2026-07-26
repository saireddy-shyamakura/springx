// Package cmd implements the springx CLI commands.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Verbose and Debug are set by the --verbose / --debug persistent flags.
// Subcommands can read them to gate detailed output.
var (
	Verbose bool
	Debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "springx",
	Short: "A fast, interactive Spring Boot project generator",
	Long: `springx — Spring Boot project generator

springx creates production-ready Spring Boot projects via Spring Initializr.
It provides an interactive terminal UI for selecting dependencies, supports
opinionated project templates, and runs post-generation automation hooks.

Examples:
  springx new                         # interactive project wizard
  springx new --template rest-api     # start from a preset
  springx new --template jpa --no-hooks
  springx template list               # show all presets
  springx config show                 # view your defaults
  springx version                     # show build information

Documentation: https://github.com/saireddy-shyamakura/springx`,
}

// Execute is called by main.main(). It runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false,
		"Show more output (HTTP requests, hook details)")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false,
		"Show debug-level output (full stack traces, raw responses)")

	// Silence the default "completion" command that cobra generates.
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
