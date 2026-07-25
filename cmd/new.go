package cmd

import (
	"fmt"
	"os"

	"github.com/saireddy-shyamakura/springx/internal/extract"
	"github.com/saireddy-shyamakura/springx/internal/initializr"
	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"

	// Side-effect: registers all built-in post-generation hooks.
	_ "github.com/saireddy-shyamakura/springx/internal/postgen/hooks"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Spring Boot project",
	Long: `Create a new Spring Boot project via Spring Initializr.

--template   Bootstrap from a preset (run 'springx template list' to see all).
--hook       Run a specific post-generation hook (repeatable). Omit to run all.
--no-hooks   Skip all post-generation automation.

Examples:
  springx new
  springx new --template rest-api
  springx new --template jpa --hook git --hook docker --hook compose`,

	RunE: func(cmd *cobra.Command, args []string) error {
		templateName, _ := cmd.Flags().GetString("template")
		hookNames, _ := cmd.Flags().GetStringArray("hook")
		noHooks, _ := cmd.Flags().GetBool("no-hooks")

		// ── 1. Interactive configuration ──────────────────────────────────
		var (
			cfg *prompt.ProjectConfig
			err error
		)

		if templateName != "" {
			cfg, err = prompt.PromptForConfigWithTemplate(templateName)
		} else {
			cfg, err = prompt.PromptForConfig()
		}
		if err != nil {
			return err
		}

		// ── 2. Download ZIP from Spring Initializr ────────────────────────
		zipFile, err := initializr.Download(cfg)
		if err != nil {
			return err
		}
		fmt.Printf("  ✔ %-20s\n", "Generated project")

		// ── 3. Extract ZIP ────────────────────────────────────────────────
		if err := extract.Unzip(zipFile, cfg.ProjectName); err != nil {
			return fmt.Errorf("failed to extract project: %w", err)
		}
		if err := os.Remove(zipFile); err != nil {
			return fmt.Errorf("failed to remove zip file: %w", err)
		}

		// ── 4. Post-generation hooks ──────────────────────────────────────
		if !noHooks {
			hooks, err := postgen.ResolveHooks(hookNames)
			if err != nil {
				return err
			}

			fmt.Println("\nRunning post-generation hooks:")

			results, hookErr := postgen.RunHooks(postgen.RunOptions{
				ProjectPath: cfg.ProjectName,
				Config:      cfg,
				Hooks:       hooks,
				Out:         os.Stdout,
			})

			// Print a per-hook summary regardless of errors.
			_ = results

			if hookErr != nil {
				// Surface hook failures as warnings, not fatal errors, so
				// the user always ends up with a usable project.
				fmt.Fprintf(os.Stderr, "\nWarning: some hooks reported errors:\n%v\n", hookErr)
			}
		}

		fmt.Printf("\n  ✔ %-20s\n", "Completed")
		fmt.Printf("\nYour project is ready at: ./%s\n", cfg.ProjectName)

		return nil
	},
}

func init() {
	newCmd.Flags().StringP("template", "t", "", "Bootstrap from a built-in project template (e.g. rest-api, jpa, kafka)")
	newCmd.Flags().StringArray("hook", nil, "Run a specific post-generation hook (repeatable, e.g. --hook git --hook docker)")
	newCmd.Flags().Bool("no-hooks", false, "Skip all post-generation automation")
	rootCmd.AddCommand(newCmd)
}
