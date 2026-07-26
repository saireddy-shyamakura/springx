package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build-time variables injected via -ldflags.
// See Makefile and .goreleaser.yaml for the injection commands.
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show springx version and build information",
	Long: `Display the current springx version together with build metadata.

The version, commit hash, and build date are injected at compile time via
-ldflags and are embedded in every official release binary.`,
	Example: `  springx version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("springx %s\n", Version)
		fmt.Printf("  commit     : %s\n", Commit)
		fmt.Printf("  built      : %s\n", BuildDate)
		fmt.Printf("  go version : %s\n", runtime.Version())
		fmt.Printf("  platform   : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
