/*
Copyright © 2026 springx contributors
*/
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/saireddy-shyamakura/springx/cmd"
)

func main() {
	// Restore terminal on SIGTERM / SIGHUP in addition to the normal path.
	// SIGINT (Ctrl+C) is handled by Bubble Tea's own signal handler inside
	// each TUI program, but SIGTERM can arrive from process managers and
	// would otherwise leave the terminal in raw/alt-screen mode.
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
