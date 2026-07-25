package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/saireddy-shyamakura/springx/internal/extract"
	"github.com/saireddy-shyamakura/springx/internal/initializr"
	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/plugins"
	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
	"github.com/saireddy-shyamakura/springx/internal/ui"

	// Side-effect: registers all built-in post-generation hooks.
	_ "github.com/saireddy-shyamakura/springx/internal/postgen/hooks"

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
  springx new --template aws-lambda
  springx new --template jpa --hook git --hook docker --hook compose`,

	RunE: func(cmd *cobra.Command, args []string) error {
		templateName, _ := cmd.Flags().GetString("template")
		hookNames, _    := cmd.Flags().GetStringArray("hook")
		noHooks, _      := cmd.Flags().GetBool("no-hooks")

		// ── 1. Plugin state ────────────────────────────────────────────────
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

		// ── 4. Interactive configuration ───────────────────────────────────
		var cfg *prompt.ProjectConfig

		if templateName != "" {
			cfg, err = prompt.PromptForConfigWithTemplate(templateName)
		} else {
			cfg, err = prompt.PromptForConfig()
		}
		if err != nil {
			return err
		}

		// ── 5. Generation pipeline (progress TUI) ─────────────────────────
		stepLabels := []string{
			"Downloading from Spring Initializr",
			"Extracting project",
		}
		if !noHooks {
			stepLabels = append(stepLabels, "Running post-generation hooks")
		}
		stepLabels = append(stepLabels, "Done")

		ch, wait := ui.RunProgressProgram(stepLabels)

		var genErr error

		go func() {
			// Step 1 — download.
			zipFile, dlErr := initializr.Download(cfg)
			if dlErr != nil {
				ch <- ui.StepFailedMsg{Err: dlErr}
				// Skip remaining steps.
				ch <- ui.StepFailedMsg{Err: fmt.Errorf("skipped")}
				if !noHooks {
					ch <- ui.StepFailedMsg{Err: fmt.Errorf("skipped")}
				}
				ch <- ui.StepFailedMsg{Err: fmt.Errorf("skipped")}
				genErr = fmt.Errorf("download failed: %w", dlErr)
				return
			}
			ch <- ui.StepDoneMsg{Detail: zipFile}

			// Step 2 — extract.
			if exErr := extract.Unzip(zipFile, cfg.ProjectName); exErr != nil {
				ch <- ui.StepFailedMsg{Err: exErr}
				if !noHooks {
					ch <- ui.StepFailedMsg{Err: fmt.Errorf("skipped")}
				}
				ch <- ui.StepFailedMsg{Err: fmt.Errorf("skipped")}
				genErr = fmt.Errorf("extraction failed: %w", exErr)
				return
			}
			os.Remove(zipFile) //nolint:errcheck
			ch <- ui.StepDoneMsg{}

			// Step 3 — hooks (optional).
			if !noHooks {
				hooks, resolveErr := postgen.ResolveHooks(hookNames)
				if resolveErr != nil {
					ch <- tea.Msg(ui.StepFailedMsg{Err: resolveErr})
					ch <- tea.Msg(ui.StepFailedMsg{Err: fmt.Errorf("skipped")})
					genErr = resolveErr
					return
				}
				_, hookErr := postgen.RunHooks(postgen.RunOptions{
					ProjectPath: cfg.ProjectName,
					Config:      cfg,
					Hooks:       hooks,
					Out:         os.Stderr, // write to stderr so TUI is not polluted
				})
				if hookErr != nil {
					ch <- tea.Msg(ui.StepDoneMsg{Detail: "some hooks reported errors (see above)"})
				} else {
					ch <- tea.Msg(ui.StepDoneMsg{})
				}
			}

			// Final step — done.
			ch <- tea.Msg(ui.StepDoneMsg{})
		}()

		wait()

		if genErr != nil {
			// The progress UI already showed the failure inline.
			// Emit a friendly error box to stderr for the shell.
			switch {
			case isDownloadErr(genErr):
				fmt.Fprintln(os.Stderr, ui.RenderError(
					"Unable to download the project from Spring Initializr.",
					genErr,
					ui.DownloadErrorSuggestions(),
				))
			default:
				fmt.Fprintf(os.Stderr, "Error: %v\n", genErr)
			}
			return fmt.Errorf("project generation failed")
		}

		fmt.Printf("\nYour project is ready at: ./%s\n", cfg.ProjectName)
		return nil
	},
}

func isDownloadErr(err error) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) > 8 && err.Error()[:8] == "download"
}

func init() {
	newCmd.Flags().StringP("template", "t", "", "Bootstrap from a built-in project template (e.g. rest-api, jpa, kafka)")
	newCmd.Flags().StringArray("hook", nil, "Run a specific post-generation hook (repeatable, e.g. --hook git --hook docker)")
	newCmd.Flags().Bool("no-hooks", false, "Skip all post-generation automation")
	rootCmd.AddCommand(newCmd)
}
