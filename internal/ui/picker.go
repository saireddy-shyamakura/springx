// Package ui implements the Bubble Tea terminal UI for springx.
// All business logic lives in state.go; this file contains only the Bubble Tea
// model, rendering, and the public RunDependencyPicker entry-point.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/saireddy-shyamakura/springx/internal/metadata"
)

// ── Message types ─────────────────────────────────────────────────────────────

// metadataLoadedMsg is sent when metadata arrives (used when RunDependencyPicker
// is called with a nil *Metadata and loads it asynchronously).
type metadataLoadedMsg struct{ meta *metadata.Metadata }
type metadataErrMsg struct{ err error }

// successDoneMsg fires after the brief success animation completes.
type successDoneMsg struct{}

// ── Focus panels ─────────────────────────────────────────────────────────────

type focusPanel int

const (
	panelDeps focusPanel = iota // dependency list (default)
	panelGroups
)

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	// data
	state       *PickerState
	meta        *metadata.Metadata
	bootVersion string // e.g. "3.4.1"
	javaVersion string // from caller config, shown in status bar
	template    string // active template name, shown in status bar

	// ui state
	searchInput textinput.Model
	spinner     spinner.Model
	focus       focusPanel

	// flags
	loading     bool
	showHelp    bool
	showSuccess bool
	confirmed   bool
	canceled    bool

	// layout
	width  int
	height int

	// success animation
	successStart time.Time
}

const successDisplayDuration = 800 * time.Millisecond

// PickerOptions configures the dependency picker.
type PickerOptions struct {
	// Metadata is the pre-fetched Spring Initializr metadata.
	// If nil the model will fetch it and show a spinner while loading.
	Metadata *metadata.Metadata

	// PreSelected is the list of dependency IDs to pre-check.
	PreSelected []string

	// BootVersion is shown in the status bar (cosmetic only).
	BootVersion string

	// JavaVersion is shown in the status bar (cosmetic only).
	JavaVersion string

	// Template is the active template name shown in the status bar (cosmetic only).
	Template string
}

func newModel(opts PickerOptions) model {
	ti := textinput.New()
	ti.Placeholder = "Type to search…"
	ti.Prompt = ""
	ti.CharLimit = 64

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	m := model{
		searchInput: ti,
		spinner:     sp,
		width:       120,
		height:      36,
		bootVersion: opts.BootVersion,
		javaVersion: opts.JavaVersion,
		template:    opts.Template,
	}

	if opts.Metadata != nil {
		m.meta = opts.Metadata
		m.state = NewPickerState(opts.Metadata, opts.PreSelected)
		// Seed boot version from metadata when not provided by caller.
		if m.bootVersion == "" {
			m.bootVersion = opts.Metadata.BootVersion.Default
		}
	} else {
		m.loading = true
	}

	return m
}

// ── Bubble Tea interface ──────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink, m.spinner.Tick}
	if m.loading {
		cmds = append(cmds, tea.EnableMouseCellMotion)
	} else {
		cmds = append(cmds, tea.EnableMouseCellMotion)
	}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ── Window resize ──────────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	// ── Metadata loaded async ──────────────────────────────────────────────
	case metadataLoadedMsg:
		m.loading = false
		m.meta = msg.meta
		m.state = NewPickerState(msg.meta, nil)
		if m.bootVersion == "" {
			m.bootVersion = msg.meta.BootVersion.Default
		}
		return m, nil

	case metadataErrMsg:
		m.loading = false
		m.canceled = true
		return m, tea.Quit

	// ── Success animation done ─────────────────────────────────────────────
	case successDoneMsg:
		return m, tea.Quit

	// ── Spinner tick ───────────────────────────────────────────────────────
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	// ── Mouse ──────────────────────────────────────────────────────────────
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			m = m.handleMouseClick(msg.X, msg.Y)
		}
		return m, nil

	// ── Keyboard ──────────────────────────────────────────────────────────
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd) {
	if m.loading {
		if msg.String() == "ctrl+c" {
			m.canceled = true
			return m, tea.Quit
		}
		return m, nil
	}

	// ── Help overlay — any key dismisses ──────────────────────────────────
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	// ── Search mode ───────────────────────────────────────────────────────
	if m.searchInput.Focused() {
		switch msg.String() {
		case "esc":
			m.searchInput.Blur()
			m.searchInput.SetValue("")
			m.state.ApplyFilter("")
			return m, nil
		case "enter":
			m.searchInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			m.state.ApplyFilter(m.searchInput.Value())
			return m, cmd
		}
	}

	// ── Normal mode ───────────────────────────────────────────────────────
	switch msg.String() {
	case "ctrl+c", "q":
		m.canceled = true
		return m, tea.Quit

	case "enter":
		m.showSuccess = true
		m.confirmed = true
		m.successStart = time.Now()
		return m, tea.Tick(successDisplayDuration, func(time.Time) tea.Msg {
			return successDoneMsg{}
		})

	case "up", "k":
		m.state.MoveCursor(-1)
	case "down", "j":
		m.state.MoveCursor(1)

	case "tab":
		m.state.TabToNextGroup()
	case "shift+tab":
		m.state.TabToPrevGroup()

	case "left", "h":
		m.state.TabToPrevGroup()
	case "right", "l":
		m.state.TabToNextGroup()

	case " ":
		m.state.ToggleCurrent()

	case "/":
		m.searchInput.Focus()
		return m, textinput.Blink

	case "esc":
		if m.state.SearchQuery != "" {
			m.searchInput.SetValue("")
			m.state.ApplyFilter("")
		}

	case "?":
		m.showHelp = true
	}

	return m, nil
}

// handleMouseClick maps a terminal click to a state mutation.
func (m model) handleMouseClick(x, y int) model {
	if m.state == nil {
		return m
	}
	// Only handle clicks in the dependency panel area; crude hit-test based on
	// the column layout computed in View(). Clicking on a visible dependency row
	// toggles it; this is best-effort since exact pixel positions require a
	// separate hit-map which we build cheaply here.
	_ = x
	_ = y
	// TODO: build a proper hitmap during View() and store it on the model.
	// For now mouse support falls back to keyboard nav only.
	return m
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.loading {
		return m.viewLoading()
	}
	if m.showSuccess {
		return m.viewSuccess()
	}
	if m.showHelp {
		return m.viewHelp()
	}
	return m.viewMain()
}

// ── Loading view ──────────────────────────────────────────────────────────────

func (m model) viewLoading() string {
	pad := strings.Repeat("\n", m.height/2-2)
	line := fmt.Sprintf("  %s  Fetching Spring Initializr metadata…", m.spinner.View())
	return pad + SpinnerStyle.Render(line) + "\n"
}

// ── Success view ──────────────────────────────────────────────────────────────

func (m model) viewSuccess() string {
	n := m.state.SelectedCount()
	noun := "dependency"
	if n != 1 {
		noun = "dependencies"
	}
	pad := strings.Repeat("\n", m.height/2-2)
	line := fmt.Sprintf("  ✔  %d %s selected — generating project…", n, noun)
	return pad + SuccessStyle.Render(line) + "\n"
}

// ── Help overlay ──────────────────────────────────────────────────────────────

func (m model) viewHelp() string {
	type entry struct{ key, desc string }
	entries := []entry{
		{"↑ / k", "Move up"},
		{"↓ / j", "Move down"},
		{"space", "Toggle selection"},
		{"tab / →", "Next group"},
		{"shift+tab / ←", "Previous group"},
		{"/ ", "Start search"},
		{"esc", "Clear search / cancel"},
		{"enter", "Confirm selection"},
		{"?", "Toggle this help"},
		{"q / ctrl+c", "Quit"},
	}

	var rows []string
	rows = append(rows, HelpTitleStyle.Render("Keyboard Shortcuts"))
	for _, e := range entries {
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			HelpKeyStyle.Render(e.key),
			HelpDescStyle.Render(e.desc),
		)
		rows = append(rows, row)
	}
	rows = append(rows, "")
	rows = append(rows, AppSubtitleStyle.Render("Press any key to close"))

	box := HelpBoxStyle.Render(strings.Join(rows, "\n"))

	// Centre in the terminal.
	padH := (m.width - lipgloss.Width(box)) / 2
	padV := (m.height - lipgloss.Height(box)) / 2
	if padH < 0 {
		padH = 0
	}
	if padV < 0 {
		padV = 0
	}
	return strings.Repeat("\n", padV) +
		strings.Repeat(" ", padH) + box + "\n"
}

// ── Main view ─────────────────────────────────────────────────────────────────

func (m model) viewMain() string {
	// Reserve rows for fixed chrome.
	const headerRows = 4 // title(1) + subtitle(1) + rule(1) + search(1)
	const footerRows = 2 // footer key-hint line + status bar
	const searchRows = 1
	contentH := m.height - headerRows - footerRows - searchRows
	if contentH < 6 {
		contentH = 6
	}

	// Column widths: groups=22, selected=26, deps=rest
	groupW := 22
	selectedW := 26
	depsW := m.width - groupW - selectedW - 6 // 6 = borders + spacing
	if depsW < 20 {
		depsW = 20
	}

	header := m.renderHeader()
	search := m.renderSearch()
	rule := HRuleStyle.Render(strings.Repeat("─", m.width))

	groupPanel := m.renderGroupPanel(groupW, contentH)
	depsPanel := m.renderDepsPanel(depsW, contentH)
	selectedPanel := m.renderSelectedPanel(selectedW, contentH)

	middle := lipgloss.JoinHorizontal(lipgloss.Top,
		groupPanel,
		" ",
		depsPanel,
		" ",
		selectedPanel,
	)

	footer := m.renderFooter()
	status := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		search,
		rule,
		middle,
		rule,
		footer,
		status,
	)
}

// ── Header ────────────────────────────────────────────────────────────────────

func (m model) renderHeader() string {
	title := AppTitleStyle.Render("springx")
	subtitle := AppSubtitleStyle.Render("Spring Boot Dependency Selection")
	rule := HRuleStyle.Render(strings.Repeat("─", m.width))
	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, rule)
}

// ── Search bar ────────────────────────────────────────────────────────────────

func (m model) renderSearch() string {
	label := SearchLabelStyle.Render("Search:")

	var input string
	if m.searchInput.Focused() {
		input = SearchActiveStyle.Render(m.searchInput.View()) +
			"  " + SearchingIndicatorStyle.Render("Searching…")
	} else if m.state.SearchQuery != "" {
		input = SearchIdleStyle.Render(
			fmt.Sprintf("%q  (esc to clear)", m.state.SearchQuery),
		)
	} else {
		input = SearchIdleStyle.Render("Press / to search")
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, "  ", label, " ", input)
}

// ── Group panel ───────────────────────────────────────────────────────────────

func (m model) renderGroupPanel(w, h int) string {
	title := PanelTitleStyle.Render("Groups")

	visible := m.state.VisibleGroupNames()
	visibleSet := make(map[string]bool, len(visible))
	for _, g := range visible {
		visibleSet[g] = true
	}

	activeGroup := ""
	groups := m.state.GetGroupNames()
	if m.state.ActiveGroupIdx() < len(groups) {
		activeGroup = groups[m.state.ActiveGroupIdx()]
	}

	var rows []string
	for _, g := range groups {
		name := truncate(g, w-3)
		if g == activeGroup {
			rows = append(rows, GroupActiveStyle.Width(w).Render(name))
		} else if visibleSet[g] {
			rows = append(rows, GroupNormalStyle.Width(w).Render(name))
		} else {
			rows = append(rows, GroupDimStyle.Width(w).Render(name))
		}
	}

	// Pad to height.
	for len(rows) < h-2 {
		rows = append(rows, strings.Repeat(" ", w))
	}

	body := strings.Join(rows, "\n")
	content := lipgloss.JoinVertical(lipgloss.Left, title, body)

	style := normalBorder.Width(w).Height(h)
	return style.Render(content)
}

// ── Dependencies panel ────────────────────────────────────────────────────────

func (m model) renderDepsPanel(w, h int) string {
	title := PanelTitleStyle.Render("Dependencies")

	var rows []string

	if len(m.state.FilteredRows) == 0 {
		rows = append(rows, EmptyStateStyle.Render("No dependencies found."))
	} else {
		activeRowIdx := -1
		if m.state.Cursor >= 0 && m.state.Cursor < len(m.state.SelectableIdx) {
			activeRowIdx = m.state.SelectableIdx[m.state.Cursor]
		}

		// Determine a scroll window so the cursor row stays visible.
		depRows := m.state.FilteredRows
		visibleH := h - 3
		scrollOffset := m.computeScrollOffset(visibleH)
		end := scrollOffset + visibleH
		if end > len(depRows) {
			end = len(depRows)
		}

		for i := scrollOffset; i < end; i++ {
			row := depRows[i]
			if row.Type == TypeHeader {
				rows = append(rows, renderGroupHeader(row.GroupName, w))
			} else {
				rows = append(rows, m.renderDepRow(row, i == activeRowIdx, w))
			}
		}
	}

	// Pad to height.
	for len(rows) < h-2 {
		rows = append(rows, "")
	}

	body := strings.Join(rows, "\n")
	content := lipgloss.JoinVertical(lipgloss.Left, title, body)

	style := focusBorder.Width(w).Height(h)
	return style.Render(content)
}

func renderGroupHeader(name string, w int) string {
	line := SectionHeaderStyle.Render("  " + name)
	rule := HRuleStyle.Render("  " + strings.Repeat("─", max(0, w-4)))
	return lipgloss.JoinVertical(lipgloss.Left, line, rule)
}

func (m model) renderDepRow(row ListRow, isCursor bool, w int) string {
	isSelected := m.state.Selected[row.ID]

	checkbox := CheckboxOffStyle.Render("[ ]")
	if isSelected {
		checkbox = CheckboxOnStyle.Render("[✓]")
	}

	// Apply search highlight to name.
	name := row.Name
	if m.state.SearchQuery != "" {
		name = HighlightMatches(name, m.state.SearchQuery)
	}

	desc := ""
	if row.Description != "" {
		maxDescW := w - len(row.Name) - 8
		if maxDescW > 12 {
			desc = DepDescStyle.Render("  " + truncate(row.Description, maxDescW))
		}
	}

	line := fmt.Sprintf(" %s %s%s", checkbox, name, desc)

	switch {
	case isCursor && isSelected:
		return DepCursorStyle.Width(w).Render(line)
	case isCursor:
		return DepCursorStyle.Width(w).Render(line)
	case isSelected:
		return DepSelectedStyle.Width(w).Render(line)
	default:
		return DepNormalStyle.Width(w).Render(line)
	}
}

// computeScrollOffset returns the first row index to render so that the
// cursor stays in view within visibleH rows.
func (m model) computeScrollOffset(visibleH int) int {
	if m.state.Cursor < 0 || len(m.state.SelectableIdx) == 0 {
		return 0
	}
	cursorRowIdx := m.state.SelectableIdx[m.state.Cursor]
	if cursorRowIdx < visibleH {
		return 0
	}
	return cursorRowIdx - visibleH/2
}

// ── Selected panel ────────────────────────────────────────────────────────────

func (m model) renderSelectedPanel(w, h int) string {
	n := m.state.SelectedCount()
	countLine := SelectedCountStyle.Render(
		fmt.Sprintf("Selected (%d)", n),
	)

	var rows []string
	for _, name := range m.state.GetSelectedNames() {
		bullet := SelectedBulletStyle.Render("✓")
		item := SelectedItemStyle.Render(" " + truncate(name, w-5))
		rows = append(rows, fmt.Sprintf(" %s%s", bullet, item))
	}

	if n == 0 {
		rows = append(rows, EmptyStateStyle.Render("None yet"))
	}

	// Pad to height.
	for len(rows) < h-3 {
		rows = append(rows, "")
	}

	body := strings.Join(rows, "\n")
	content := lipgloss.JoinVertical(lipgloss.Left, countLine, body)

	style := normalBorder.Width(w).Height(h)
	return style.Render(content)
}

// ── Footer ────────────────────────────────────────────────────────────────────

func (m model) renderFooter() string {
	sep := FooterSepStyle.Render(" • ")
	k := FooterKeyStyle

	hints := []string{
		k.Render("↑↓") + " navigate",
		k.Render("space") + " select",
		k.Render("/") + " search",
		k.Render("tab") + " next group",
		k.Render("enter") + " confirm",
		k.Render("?") + " help",
		k.Render("q") + " quit",
	}

	line := strings.Join(hints, sep)
	return FooterStyle.Render(line)
}

// ── Status bar ────────────────────────────────────────────────────────────────

func (m model) renderStatusBar() string {
	sep := StatusSepStyle.Render(" │ ")

	parts := []string{
		StatusKeyStyle.Render("Metadata") + StatusValueStyle.Render(" loaded"),
	}
	if m.bootVersion != "" {
		parts = append(parts,
			StatusKeyStyle.Render("Boot")+StatusValueStyle.Render(" "+m.bootVersion),
		)
	}
	if m.javaVersion != "" {
		parts = append(parts,
			StatusKeyStyle.Render("Java")+StatusValueStyle.Render(" "+m.javaVersion),
		)
	}
	if m.template != "" {
		parts = append(parts,
			StatusKeyStyle.Render("Template")+StatusValueStyle.Render(" "+m.template),
		)
	}

	bar := strings.Join(parts, sep)
	return StatusBarStyle.Width(m.width).Render(bar)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// truncate clips s to maxLen runes, appending "…" if clipped.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ── Public API ────────────────────────────────────────────────────────────────

// RunDependencyPicker launches the interactive TUI and returns the selected
// dependency IDs. preSelected may be nil.
func RunDependencyPicker(meta *metadata.Metadata, preSelected []string) ([]string, error) {
	return RunDependencyPickerWithOptions(PickerOptions{
		Metadata:    meta,
		PreSelected: preSelected,
	})
}

// RunDependencyPickerWithOptions is the full-featured entry point that accepts
// a PickerOptions for richer status-bar display.
func RunDependencyPickerWithOptions(opts PickerOptions) ([]string, error) {
	m := newModel(opts)
	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("dependency picker failed: %w", err)
	}

	fm, ok := finalModel.(model)
	if !ok || fm.canceled {
		return nil, fmt.Errorf("dependency selection canceled")
	}

	return fm.state.GetSelectedIDs(), nil
}
