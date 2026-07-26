package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/saireddy-shyamakura/springx/cmd"
)

// Version variables are set at build time via -ldflags:
//
//	-X github.com/saireddy-shyamakura/springx/cmd.Version=v1.0.0
//	-X github.com/saireddy-shyamakura/springx/cmd.Commit=abc1234
//	-X github.com/saireddy-shyamakura/springx/cmd.BuildDate=2024-01-01T00:00:00Z
//
// When built with `go build` directly (no ldflags), they default to
// "dev", "none", and "unknown" respectively (see cmd/version.go).
func main() {
	// Restore the terminal on SIGTERM/SIGHUP (process manager kills).
	// SIGINT (Ctrl+C) is handled by each Bubble Tea program directly.
	restoreCh := make(chan os.Signal, 1)
	signal.Notify(restoreCh, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-restoreCh
		// Best-effort: exit alternate screen, show cursor, reset attributes.
		os.Stdout.WriteString("\033[?1049l\033[?25h\033[0m") //nolint:errcheck
		os.Exit(1)
	}()

	cmd.Execute()
}
