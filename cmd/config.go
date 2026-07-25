package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/saireddy-shyamakura/springx/internal/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage springx CLI configuration settings",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Configuration file already exists at %s\n", path)
			return nil
		}

		def := config.DefaultConfig()
		if err := config.Save(&def); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		fmt.Printf("✔ Initialized configuration at %s\n", path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display active configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := config.GetConfigPath()
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Printf("Config File : %s\n", path)
		fmt.Println("----------------------------------")
		fmt.Printf("Group ID       : %s\n", cfg.GroupID)
		fmt.Printf("Artifact Prefix: %s\n", cfg.ArtifactPrefix)
		fmt.Printf("Package Prefix : %s\n", cfg.PackagePrefix)
		fmt.Printf("Java Version   : %s\n", cfg.JavaVersion)
		fmt.Printf("Build Tool     : %s\n", cfg.BuildTool)
		fmt.Printf("Packaging      : %s\n", cfg.Packaging)
		fmt.Printf("Language       : %s\n", cfg.Language)
		fmt.Println("----------------------------------")
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration file in default editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		// Create file if it doesn't exist yet
		if _, err := os.Stat(path); os.IsNotExist(err) {
			def := config.DefaultConfig()
			if err := config.Save(&def); err != nil {
				return err
			}
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "nano"
			}
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
	Short: "Delete configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		if err := config.Reset(); err != nil {
			return err
		}

		fmt.Printf("✔ Reset configuration (deleted %s)\n", path)
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
