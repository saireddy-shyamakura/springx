package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
)

// ── Step state ────────────────────────────────────────────────────────────────

// StepStatus represents the state of one progress step.
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepDone
	StepFailed
	StepSkipped
)

// Step is a single item in the generation pipeline.
type Step struct {
	Label  string
	Status StepStatus
	Detail string // shown inline when non-empty (e.g. the downloaded filename)
}

// ── Messages callers send ─────────────────────────────────────────────────────

// StepDoneMsg advances the current running step to StepDone.
type StepDoneMsg struct{ Detail string }

// StepFailedMsg marks the current running step as StepFailed.
type StepFailedMsg struct{ Err error }

// ProgressDoneMsg signals all steps are finished.
type ProgressDoneMsg struct{}

// ── Model ─────────────────────────────────────────────────────────────────────

// ProgressModel is the Bubble Tea model for the linear progress pipeline view.
type ProgressModel struct {
	Steps   []Step
	spinner spinner.Model
	width   int
	height  int
	done    bool
}

// NewProgressModel builds a model from step labels. The first step is set to
// StepRunning automatically.
func NewProgressModel(labels []string) ProgressModel {
	steps := make([]Step, len(labels))
	for i, l := range labels {
		steps[i] = Step{Label: l, Status: StepPending}
	}
	if len(steps) > 0 {
		steps[0].Status = StepRunning
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	return ProgressModel{
		Steps:   steps,
		spinner: sp,
		width:   80,
		height:  24,
	}
}

// ── Bubble Tea interface ──────────────────────────────────────────────────────

func (m ProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case StepDoneMsg:
		m = m.markCurrent(StepDone, msg.Detail)
		m = m.advanceToNext()
		if m.allFinished() {
			m.done = true
			return m, func() tea.Msg { return ProgressDoneMsg{} }
		}

	case StepFailedMsg:
		detail := ""
		if msg.Err != nil {
			detail = msg.Err.Error()
		}
		m = m.markCurrent(StepFailed, detail)
		m = m.advanceToNext()
		if m.allFinished() {
			m.done = true
			return m, func() tea.Msg { return ProgressDoneMsg{} }
		}

	case ProgressDoneMsg:
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ProgressModel) View() string {
	var rows []string
	rows = append(rows, ProgressTitleStyle.Render("Generating Spring Boot project"))
	rows = append(rows, "")

	for _, step := range m.Steps {
		rows = append(rows, m.renderStep(step))
	}

	rows = append(rows, "")
	if m.done {
		rows = append(rows, SuccessStyle.Render("  ✔  All done!"))
	}

	box := ProgressBoxStyle.Render(strings.Join(rows, "\n"))

	bh  := gloss.Height(box)
	bw  := gloss.Width(box)
	padV := (m.height - bh) / 2
	padH := (m.width - bw) / 2
	if padV < 0 {
		padV = 0
	}
	if padH < 0 {
		padH = 0
	}

	return strings.Repeat("\n", padV) + strings.Repeat(" ", padH) + box + "\n"
}

func (m ProgressModel) renderStep(step Step) string {
	var line string

	switch step.Status {
	case StepDone:
		line = ProgressDoneStyle.Render("  ✔  " + step.Label)
	case StepFailed:
		line = ProgressErrorStyle.Render("  ✗  " + step.Label)
	case StepRunning:
		line = ProgressCurrentStyle.Render(m.spinner.View() + "  " + step.Label)
	case StepSkipped:
		line = ProgressPendingStyle.Render("  –  " + step.Label)
	default:
		line = ProgressPendingStyle.Render("  ○  " + step.Label)
	}

	if step.Detail != "" {
		line += DepDescStyle.Render("  " + truncate(step.Detail, 48))
	}
	return line
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (m ProgressModel) currentIdx() int {
	for i, s := range m.Steps {
		if s.Status == StepRunning {
			return i
		}
	}
	return -1
}

func (m ProgressModel) markCurrent(status StepStatus, detail string) ProgressModel {
	idx := m.currentIdx()
	if idx >= 0 {
		m.Steps[idx].Status = status
		if detail != "" {
			m.Steps[idx].Detail = detail
		}
	}
	return m
}

func (m ProgressModel) advanceToNext() ProgressModel {
	for i, s := range m.Steps {
		if s.Status == StepPending {
			m.Steps[i].Status = StepRunning
			return m
		}
	}
	return m
}

func (m ProgressModel) allFinished() bool {
	for _, s := range m.Steps {
		if s.Status == StepPending || s.Status == StepRunning {
			return false
		}
	}
	return true
}

// ── Friendly error renderer ───────────────────────────────────────────────────

// RenderError formats err for display outside the TUI with contextual suggestions.
func RenderError(title string, err error, suggestions []string) string {
	var rows []string
	rows = append(rows, ErrorTitleStyle.Render("❌  "+title))
	rows = append(rows, "")
	rows = append(rows, ErrorReasonStyle.Render("Reason:"))
	rows = append(rows, ErrorReasonStyle.Render("  "+err.Error()))

	if len(suggestions) > 0 {
		rows = append(rows, "")
		rows = append(rows, ErrorReasonStyle.Render("Suggestions:"))
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

// RunProgressProgram runs the progress model as a full-screen TUI and returns
// a send channel and a wait function the caller uses to drive step updates.
//
//	ch, wait := ui.RunProgressProgram(labels)
//	ch <- ui.StepDoneMsg{}
//	ch <- ui.StepDoneMsg{Detail: "demo.zip"}
//	wait()
func RunProgressProgram(labels []string) (chan<- tea.Msg, func()) {
	m  := NewProgressModel(labels)
	p  := tea.NewProgram(m, tea.WithAltScreen())
	ch := make(chan tea.Msg, 8)

	go func() {
		for msg := range ch {
			p.Send(msg)
		}
	}()

	done := make(chan struct{})
	go func() {
		defer close(done)
		p.Run() //nolint:errcheck
	}()

	// Small pause so the program goroutine starts before the first Send.
	time.Sleep(40 * time.Millisecond)

	wait := func() {
		close(ch)
		<-done
	}

	return ch, wait
}

// Ensure fmt is used.
var _ = fmt.Sprintf
