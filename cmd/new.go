package cmd

import (
	"fmt"
	"os"

	"github.com/saireddy-shyamakura/springx/internal/extract"
	"github.com/saireddy-shyamakura/springx/internal/initializr"
	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/plugins"
	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
	"github.com/saireddy-shyamakura/springx/internal/ui"

	// Side-effect: registers all built-in post-generation hooks.
	_ "github.com/saireddy-shyamakura/springx/internal/postgen/hooks"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Spring Boot project",
	Long: `Create a new Spring Boot project via Spring Initializr.

Options:
  --template   Bootstrap from a preset (run 'springx template list' to see all).
  --hook       Run a specific post-generation hook (repeatable). Omit to run all.
  --no-hooks   Skip all post-generation automation.

Examples:
  springx new
  springx new --template rest-api
  springx new --template jpa --hook git --hook docker`,

	RunE: func(cmd *cobra.Command, args []string) error {
		templateName, _ := cmd.Flags().GetString("template")
		hookNames, _    := cmd.Flags().GetStringArray("hook")
		noHooks, _      := cmd.Flags().GetBool("no-hooks")

		// ── 1. Load plugins ────────────────────────────────────────────────
		plugins.LoadDisabledIntoRegistry()

		// ── 2. Fetch metadata ──────────────────────────────────────────────
		meta, err := metadata.Fetch()
		if err != nil {
			fmt.Fprintln(os.Stderr, ui.RenderError(
				"Unable to fetch Spring Initializr metadata.",
				err,
				ui.MetadataErrorSuggestions(),
			))
			return fmt.Errorf("metadata fetch failed")
		}

		// ── 3. Apply plugins ───────────────────────────────────────────────
		plugins.Apply(meta)

		// ── 4. Interactive configuration (dep picker TUI) ──────────────────
		var cfg *prompt.ProjectConfig
		if templateName != "" {
			cfg, err = prompt.PromptForConfigWithTemplate(templateName)
		} else {
			cfg, err = prompt.PromptForConfig()
		}
		if err != nil {
			return err
		}

		// ── 5. Build generation pipeline ───────────────────────────────────
		//
		// Each step is a pure function → tea.Msg. The UI never directly
		// performs I/O; it calls these functions as tea.Cmd values and
		// receives their results as messages.
		//
		// State shared between steps is captured in local variables that are
		// closed over by each StepFunc. Steps communicate via those variables
		// rather than channels.
		var (
			zipFile string // set by the download step, read by extract+cleanup
		)

		// step: download
		downloadStep := func() tea.Msg {
			zf, dlErr := initializr.Download(cfg)
			if dlErr != nil {
				return ui.StepFailedMsg{Err: fmt.Errorf("download failed: %w", dlErr)}
			}
			zipFile = zf
			// Pass the full relative path as detail — the progress model
			// truncates it for display and tracks it for error recovery.
			return ui.StepDoneMsg{Detail: zipFile}
		}

		// step: extract
		extractStep := func() tea.Msg {
			if err := extract.Unzip(zipFile, cfg.ProjectName); err != nil {
				return ui.StepFailedMsg{Err: fmt.Errorf("extraction failed: %w", err)}
			}
			return ui.StepDoneMsg{}
		}

		// step: delete zip (after successful extraction)
		cleanupStep := func() tea.Msg {
			if zipFile != "" {
				os.Remove(zipFile) //nolint:errcheck
			}
			return ui.StepDoneMsg{}
		}

		// Assemble label + step-func slices in lock-step.
		labels := []string{
			"Downloading from Spring Initializr",
			"Extracting project",
			"Cleaning up",
		}
		steps := []ui.StepFunc{
			downloadStep,
			extractStep,
			cleanupStep,
		}

		if !noHooks {
			// Resolve hook list before the TUI starts so any "unknown hook"
			// error surfaces before the user sees the progress screen.
			hooks, resolveErr := postgen.ResolveHooks(hookNames)
			if resolveErr != nil {
				return resolveErr
			}

			hooksStep := func() tea.Msg {
				_, hookErr := postgen.RunHooks(postgen.RunOptions{
					ProjectPath: cfg.ProjectName,
					Config:      cfg,
					Hooks:       hooks,
					Out:         os.Stderr,
				})
				if hookErr != nil {
					// Hook errors are non-fatal: report as done with a detail note.
					return ui.StepDoneMsg{Detail: "some hooks reported errors"}
				}
				return ui.StepDoneMsg{}
			}

			labels = append(labels, "Running post-generation hooks")
			steps  = append(steps, hooksStep)
		}

		// Determine "next steps" shown on the success screen.
		buildCmd := "./mvnw spring-boot:run"
		for _, v := range meta.Type.Values {
			if v.ID == cfg.BuildTool {
				if v.ID == "gradle-project" || v.ID == "gradle-project-kotlin" {
					buildCmd = "./gradlew bootRun"
				}
				break
			}
		}
		nextSteps := []string{
			"cd " + cfg.ProjectName,
			buildCmd,
		}

		// ── 6. Run the progress TUI ────────────────────────────────────────
		pcfg := ui.ProgressConfig{
			Labels:      labels,
			Steps:       steps,
			ProjectName: cfg.ProjectName,
			NextSteps:   nextSteps,
		}

		if genErr := ui.RunProgressProgram(pcfg); genErr != nil {
			// The error screen was already shown inside the TUI.
			// Emit a minimal message to stderr for CI/log consumers.
			fmt.Fprintf(os.Stderr, "\nError: %v\n", genErr)
			return fmt.Errorf("project generation failed")
		}

		return nil
	},
}

func init() {
	newCmd.Flags().StringP("template", "t", "",
		"Bootstrap from a built-in project template (e.g. rest-api, jpa, kafka)")
	newCmd.Flags().StringArray("hook", nil,
		"Run a specific post-generation hook (repeatable, e.g. --hook git --hook docker)")
	newCmd.Flags().Bool("no-hooks", false,
		"Skip all post-generation automation")
	rootCmd.AddCommand(newCmd)
}
