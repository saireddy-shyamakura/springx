// Package ui — progress pipeline TUI.
//
// Architecture
// ────────────
// Generation work is executed via tea.Cmd functions. Each step returns a
// tea.Msg when it completes. The model advances the pipeline by issuing the
// next tea.Cmd only after the previous message arrives. This eliminates all
// goroutine-to-channel races and the "send on closed channel" panic.
//
// Terminal lifecycle
// ──────────────────
// tea.WithAltScreen() is used so Bubble Tea owns alternate-screen entry and
// exit. When the program exits (normally, via Ctrl+C, or due to a panic) the
// runtime deferred cleanup inside RunProgressProgram restores the terminal.
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Step status ───────────────────────────────────────────────────────────────

// StepStatus represents the execution state of a pipeline step.
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepDone
	StepFailed
	StepSkipped
)

// Step is one item in the generation pipeline.
type Step struct {
	Label  string
	Detail string // shown inline (e.g. filename, error text)
	Status StepStatus
}

// ── Pipeline messages (public — callers build them) ───────────────────────────

// StepDoneMsg marks the current running step as done and carries an optional
// detail string (e.g. the downloaded filename).
type StepDoneMsg struct{ Detail string }

// StepFailedMsg marks the current running step as failed.
type StepFailedMsg struct{ Err error }

// ── Internal messages ─────────────────────────────────────────────────────────

// pipelineStartMsg kicks off the first generation step.
type pipelineStartMsg struct{}

// ── Generation function type ──────────────────────────────────────────────────

// StepFunc is a function that performs one generation step synchronously.
// It must return either a StepDoneMsg or a StepFailedMsg.
// It must never panic — recover() is wrapped around it in the runner.
type StepFunc func() tea.Msg

// ── ProgressConfig ────────────────────────────────────────────────────────────

// ProgressConfig describes the full generation pipeline to run.
type ProgressConfig struct {
	// Steps is the ordered list of (label, work function) pairs.
	// len(Steps) must equal len(Labels).
	Labels []string
	Steps  []StepFunc

	// ProjectName is shown on the success screen.
	ProjectName string

	// NextSteps are the shell commands shown on the success screen.
	NextSteps []string
}

// ── Model ─────────────────────────────────────────────────────────────────────

// ProgressModel is the Bubble Tea model for the generation pipeline view.
type ProgressModel struct {
	cfg      ProgressConfig
	steps    []Step
	finalErr error // non-nil when a step failed fatally
	spinner  spinner.Model
	zipPath  string // preserved on extraction failure
	width    int
	height   int
	done     bool // all steps finished
	aborted  bool // Ctrl+C
}

// NewProgressModel builds the model from a ProgressConfig.
// To inspect steps in tests use m.StepList().
func NewProgressModel(cfg ProgressConfig) ProgressModel {
	steps := make([]Step, len(cfg.Labels))
	for i, l := range cfg.Labels {
		steps[i] = Step{Label: l, Status: StepPending}
	}
	if len(steps) > 0 {
		steps[0].Status = StepRunning
	}

	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = SpinnerStyle

	return ProgressModel{
		cfg:     cfg,
		steps:   steps,
		spinner: sp,
		width:   80,
		height:  24,
	}
}

// NewProgressModelLabels is a convenience constructor used in tests.
// It accepts a plain slice of step labels with no StepFuncs.
func NewProgressModelLabels(labels []string) ProgressModel {
	return NewProgressModel(ProgressConfig{Labels: labels})
}

// StepList returns the current step list. Used by tests that need to inspect
// step status without accessing unexported fields directly.
func (m ProgressModel) StepList() []Step {
	return m.steps
}

// ── Bubble Tea interface ──────────────────────────────────────────────────────

func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		// Kick off first step immediately after the first frame renders.
		func() tea.Msg { return pipelineStartMsg{} },
	)
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.done || m.aborted {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.aborted = true
			return m, tea.Quit
		case "enter":
			// Enter on the final success/error screen exits.
			if m.done {
				return m, tea.Quit
			}
		}
		return m, nil

	case pipelineStartMsg:
		// Run the first step.
		return m, m.runCurrentStep()

	case StepDoneMsg:
		idx := m.currentIdx()
		if idx >= 0 {
			m.steps[idx].Status = StepDone
			if msg.Detail != "" {
				m.steps[idx].Detail = msg.Detail
				// Track zip path for error recovery display.
				if strings.HasSuffix(msg.Detail, ".zip") {
					m.zipPath = msg.Detail
				}
			}
		}
		// Advance to next step.
		next := m.advanceToNext()
		m.steps = next.steps
		if m.allFinished() {
			m.done = true
			return m, nil
		}
		return m, m.runCurrentStep()

	case StepFailedMsg:
		idx := m.currentIdx()
		if idx >= 0 {
			m.steps[idx].Status = StepFailed
			if msg.Err != nil {
				m.steps[idx].Detail = msg.Err.Error()
			}
		}
		m.finalErr = msg.Err
		// Mark all remaining pending steps as skipped.
		for i := range m.steps {
			if m.steps[i].Status == StepPending {
				m.steps[i].Status = StepSkipped
			}
		}
		m.done = true
		return m, nil
	}

	return m, nil
}

// runCurrentStep wraps the current step's StepFunc in a tea.Cmd so it runs
// off the main goroutine and returns its result as a tea.Msg. A deferred
// recover() catches any panic inside the work function and converts it to a
// StepFailedMsg so the terminal is always restored cleanly.
func (m ProgressModel) runCurrentStep() tea.Cmd {
	idx := m.currentIdx()
	if idx < 0 || idx >= len(m.cfg.Steps) {
		return nil
	}
	fn := m.cfg.Steps[idx]
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				msg = StepFailedMsg{
					Err: fmt.Errorf("panic: %v", r),
				}
			}
		}()
		return fn()
	}
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m ProgressModel) View() string {
	if m.done && m.finalErr == nil {
		return m.viewSuccess()
	}
	if m.done && m.finalErr != nil {
		return m.viewError()
	}
	return m.viewProgress()
}

// viewProgress renders the running pipeline.
func (m ProgressModel) viewProgress() string {
	// ── dialog content ────────────────────────────────────────────────────
	var lines []string
	lines = append(lines,
		ProgressTitleStyle.Render("Generating Spring Boot project"),
		"",
	)

	for _, step := range m.steps {
		lines = append(lines, m.renderStep(step))
	}

	lines = append(lines, "", AppSubtitleStyle.Render("  Ctrl+C to abort"))

	return m.centreDialog(lines, clrAccent)
}

// viewSuccess renders the final success screen.
func (m ProgressModel) viewSuccess() string {
	loc := m.cfg.ProjectName
	if loc == "" {
		loc = "."
	}
	// Make path prettier when relative.
	if !filepath.IsAbs(loc) {
		if abs, err := filepath.Abs(loc); err == nil {
			loc = abs
		}
	}

	var lines []string
	lines = append(lines,
		SuccessStyle.Render("  ✔  Project generated successfully!"),
		"",
		ConfirmLabelStyle.Render("  Location  ")+ConfirmValueStyle.Render(loc),
	)

	if len(m.cfg.NextSteps) > 0 {
		lines = append(lines, "", ConfirmLabelStyle.Render("  Next steps"))
		for _, s := range m.cfg.NextSteps {
			lines = append(lines, "    "+DepDescStyle.Render(s))
		}
	}

	lines = append(lines, "", AppSubtitleStyle.Render("  Press Enter to exit"))

	return m.centreDialog(lines, clrGreen)
}

// viewError renders the final error screen.
func (m ProgressModel) viewError() string {
	var lines []string
	lines = append(lines, ErrorTitleStyle.Render("  ✗  Generation failed"), "")

	// Find the failed step for context.
	for _, s := range m.steps {
		if s.Status == StepFailed && s.Detail != "" {
			lines = append(lines,
				ConfirmLabelStyle.Render("  Step    ")+ConfirmValueStyle.Render(s.Label),
				ConfirmLabelStyle.Render("  Reason  ")+ErrorReasonStyle.Render(wrapText(s.Detail, m.dialogWidth()-24)),
			)
			break
		}
	}

	// If extraction failed and we have a zip, tell the user where it is.
	if m.zipPath != "" {
		lines = append(lines, "", AppSubtitleStyle.Render(
			"  The downloaded ZIP has been preserved at: "+m.zipPath))
	}

	lines = append(lines, "", AppSubtitleStyle.Render("  Press Enter to exit"))

	return m.centreDialog(lines, clrRed)
}

// centreDialog wraps lines in a box with the given border color and centers
// it on screen. The box width is capped to the terminal width minus margins.
func (m ProgressModel) centreDialog(lines []string, borderClr lipgloss.Color) string {
	dw := m.dialogWidth()

	// Clamp each line to dialog width to prevent overflow.
	clamped := make([]string, len(lines))
	for i, l := range lines {
		// Strip ANSI for width measurement, then re-render if too wide.
		vis := lipgloss.Width(l)
		if vis > dw {
			// Hard-truncate the visible portion — last resort safety.
			clamped[i] = l[:len(l)-(vis-dw)]
		} else {
			clamped[i] = l
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderClr).
		Padding(1, 3).
		Width(dw).
		Render(strings.Join(clamped, "\n"))

	bw := lipgloss.Width(box)
	bh := lipgloss.Height(box)

	padH := (m.width - bw) / 2
	padV := (m.height - bh) / 2
	if padH < 0 {
		padH = 0
	}
	if padV < 0 {
		padV = 0
	}

	// Build vertically centered output. Use a full-height background so the
	// alt-screen has no leftover content from previous renders.
	var sb strings.Builder
	sb.WriteString(strings.Repeat("\n", padV))

	prefix := strings.Repeat(" ", padH)
	for i, line := range strings.Split(box, "\n") {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(prefix)
		sb.WriteString(line)
	}

	return sb.String()
}

// dialogWidth returns the inner content width for the progress dialog,
// capped to avoid overflow on narrow terminals.
func (m ProgressModel) dialogWidth() int {
	// Target 60% of terminal width, between 52 and 80 cols.
	w := (m.width * 60) / 100
	if w < 52 {
		w = 52
	}
	if w > 80 {
		w = 80
	}
	// Never wider than the terminal minus 4 cols of margin.
	if w > m.width-4 {
		w = m.width - 4
	}
	if w < 20 {
		w = 20
	}
	return w
}

// renderStep formats one step row with the appropriate icon and style.
func (m ProgressModel) renderStep(step Step) string {
	var icon, label string

	switch step.Status {
	case StepDone:
		icon = ProgressDoneStyle.Render("  ✓")
		label = ProgressDoneStyle.Render("  " + step.Label)
	case StepFailed:
		icon = ProgressErrorStyle.Render("  ✗")
		label = ProgressErrorStyle.Render("  " + step.Label)
	case StepRunning:
		icon = "  " + m.spinner.View()
		label = ProgressCurrentStyle.Render("  " + step.Label)
	case StepSkipped:
		icon = ProgressPendingStyle.Render("  –")
		label = ProgressPendingStyle.Render("  " + step.Label)
	default: // StepPending
		icon = ProgressPendingStyle.Render("  ○")
		label = ProgressPendingStyle.Render("  " + step.Label)
	}

	line := icon + label
	if step.Detail != "" && step.Status != StepFailed {
		// Show short inline detail (filename, etc) — truncate so it never wraps.
		line += DepDescStyle.Render("  " + truncate(step.Detail, 36))
	}
	return line
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (m ProgressModel) currentIdx() int {
	for i, s := range m.steps {
		if s.Status == StepRunning {
			return i
		}
	}
	return -1
}

func (m ProgressModel) advanceToNext() ProgressModel {
	for i, s := range m.steps {
		if s.Status == StepPending {
			m.steps[i].Status = StepRunning
			return m
		}
	}
	return m
}

func (m ProgressModel) allFinished() bool {
	for _, s := range m.steps {
		if s.Status == StepPending || s.Status == StepRunning {
			return false
		}
	}
	return true
}

// wrapText hard-wraps s at maxCols. Used for error detail display.
func wrapText(s string, maxCols int) string {
	if maxCols <= 0 || len(s) <= maxCols {
		return s
	}
	return s[:maxCols] + "…"
}

// ── Error helpers (used by cmd/new.go) ───────────────────────────────────────

// RenderError formats err for display outside the TUI with contextual suggestions.
func RenderError(title string, err error, suggestions []string) string {
	var rows []string
	rows = append(rows,
		ErrorTitleStyle.Render("✗  "+title),
		"",
		ErrorReasonStyle.Render("Reason:"),
		ErrorReasonStyle.Render("  "+err.Error()),
	)

	if len(suggestions) > 0 {
		rows = append(rows, "", ErrorReasonStyle.Render("Suggestions:"))
		for _, s := range suggestions {
			rows = append(rows, ErrorSuggestionStyle.Render("  • "+s))
		}
	}

	return ErrorBoxStyle.Render(strings.Join(rows, "\n"))
}

// MetadataErrorSuggestions returns standard suggestions for metadata failures.
func MetadataErrorSuggestions() []string {
	return []string{
		"Check your internet connection.",
		"Verify that start.spring.io is reachable.",
		"Try again in a few seconds.",
	}
}

// DownloadErrorSuggestions returns standard suggestions for download failures.
func DownloadErrorSuggestions() []string {
	return []string{
		"Check your internet connection.",
		"Verify that start.spring.io is reachable.",
		"Confirm that all selected dependency IDs are valid.",
	}
}

// ── Public entry point ────────────────────────────────────────────────────────

// RunProgressProgram runs the generation pipeline as a full-screen Bubble Tea
// TUI. It blocks until the user presses Enter on the final screen (or Ctrl+C).
//
// The returned error is non-nil if any generation step failed. The terminal is
// always fully restored before this function returns, even on panic.
func RunProgressProgram(cfg ProgressConfig) (finalErr error) {
	// Panic safety — if something in the TUI itself panics, restore the
	// terminal before re-panicking so the shell is not left broken.
	defer func() {
		if r := recover(); r != nil {
			// Attempt a best-effort terminal restore via the OS.
			os.Stdout.WriteString("\033[?1049l\033[?25h\033[0m") //nolint:errcheck
			finalErr = fmt.Errorf("internal panic: %v", r)
		}
	}()

	m := NewProgressModel(cfg)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("progress TUI error: %w", err)
	}

	fm, ok := final.(ProgressModel)
	if !ok {
		return fmt.Errorf("unexpected model type returned from progress TUI")
	}

	if fm.aborted {
		return fmt.Errorf("generation aborted by user")
	}

	return fm.finalErr
}

// ── Compatibility shim ────────────────────────────────────────────────────────
// The old RunProgressProgram(labels []string) channel-based API is removed.
// Callers now use RunProgressProgram(ProgressConfig{...}).

// Ensure os is used (for the panic-recovery terminal escape sequence).
var _ = os.Stdout
