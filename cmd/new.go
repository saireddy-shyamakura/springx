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

springx fetches live metadata from start.spring.io, presents an interactive
terminal UI for selecting dependencies, then downloads and extracts the
generated project into the current directory.

Post-generation hooks run automatically after extraction to set up git,
Docker, VS Code settings, and more. Use --hook to run specific hooks or
--no-hooks to skip them entirely.`,
	Example: `  # Interactive wizard — no flags needed
  springx new

  # Start from an opinionated template
  springx new --template rest-api
  springx new --template jpa
  springx new --template kafka

  # Template with specific hooks only
  springx new --template jpa --hook git --hook docker

  # Skip all post-generation hooks
  springx new --no-hooks

  # See all available templates
  springx template list`,

	RunE: func(cmd *cobra.Command, args []string) error {
		templateName, _ := cmd.Flags().GetString("template")
		hookNames, _    := cmd.Flags().GetStringArray("hook")
		noHooks, _      := cmd.Flags().GetBool("no-hooks")

		// ── 1. Load plugins ────────────────────────────────────────────────
		plugins.LoadDisabledIntoRegistry()

		// ── 2. Fetch metadata ──────────────────────────────────────────────
		if Verbose {
			fmt.Fprintln(os.Stderr, "  → Fetching metadata from start.spring.io…")
		}
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

		// ── 5. Build generation pipeline ───────────────────────────────────
		//
		// Each step is a pure function → tea.Msg. The UI never directly
		// performs I/O; it calls these as tea.Cmd values and receives their
		// results as messages. State is shared via closure captures.

		var zipFile string // set by download, read by extract + cleanup

		downloadStep := func() tea.Msg {
			zf, dlErr := initializr.Download(cfg)
			if dlErr != nil {
				return ui.StepFailedMsg{Err: fmt.Errorf("download failed: %w", dlErr)}
			}
			if Verbose {
				fmt.Fprintf(os.Stderr, "  → Downloaded: %s\n", zf)
			}
			zipFile = zf
			return ui.StepDoneMsg{Detail: zipFile}
		}

		extractStep := func() tea.Msg {
			if err := extract.Unzip(zipFile, cfg.ProjectName); err != nil {
				return ui.StepFailedMsg{Err: fmt.Errorf("extraction failed: %w", err)}
			}
			return ui.StepDoneMsg{}
		}

		cleanupStep := func() tea.Msg {
			if zipFile != "" {
				os.Remove(zipFile) //nolint:errcheck
			}
			return ui.StepDoneMsg{}
		}

		labels := []string{
			"Downloading from Spring Initializr",
			"Extracting project",
			"Cleaning up",
		}
		steps := []ui.StepFunc{downloadStep, extractStep, cleanupStep}

		if !noHooks {
			hooks, resolveErr := postgen.ResolveHooks(hookNames)
			if resolveErr != nil {
				return resolveErr
			}

			hooksStep := func() tea.Msg {
				if Verbose {
					fmt.Fprintf(os.Stderr, "  → Running %d post-generation hook(s)…\n", len(hooks))
				}
				_, hookErr := postgen.RunHooks(postgen.RunOptions{
					ProjectPath: cfg.ProjectName,
					Config:      cfg,
					Hooks:       hooks,
					Out:         os.Stderr,
				})
				if hookErr != nil {
					if Debug {
						fmt.Fprintf(os.Stderr, "  [debug] hook errors: %v\n", hookErr)
					}
					return ui.StepDoneMsg{Detail: "some hooks reported errors"}
				}
				return ui.StepDoneMsg{}
			}

			labels = append(labels, "Running post-generation hooks")
			steps  = append(steps, hooksStep)
		}

		// Determine "next steps" hint from the selected build tool.
		buildCmd := "./mvnw spring-boot:run"
		for _, v := range meta.Type.Values {
			if v.ID == cfg.BuildTool {
				if v.ID == "gradle-project" || v.ID == "gradle-project-kotlin" {
					buildCmd = "./gradlew bootRun"
				}
				break
			}
		}

		// ── 6. Run the progress TUI ────────────────────────────────────────
		pcfg := ui.ProgressConfig{
			Labels:      labels,
			Steps:       steps,
			ProjectName: cfg.ProjectName,
			NextSteps:   []string{"cd " + cfg.ProjectName, buildCmd},
		}

		if genErr := ui.RunProgressProgram(pcfg); genErr != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", genErr)
			return fmt.Errorf("project generation failed")
		}

		return nil
	},
}

func init() {
	newCmd.Flags().StringP("template", "t", "",
		"Bootstrap from a built-in template (e.g. rest-api, jpa, kafka) — see 'springx template list'")
	newCmd.Flags().StringArray("hook", nil,
		"Run only a specific post-generation hook (repeatable, e.g. --hook git --hook docker)")
	newCmd.Flags().Bool("no-hooks", false,
		"Skip all post-generation automation")
	rootCmd.AddCommand(newCmd)
}
