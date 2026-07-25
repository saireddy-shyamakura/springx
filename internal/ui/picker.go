// Package ui implements the Bubble Tea terminal UI for springx.
// Business logic lives in state.go. Styles live in styles.go.
// This file contains only the Bubble Tea model, view rendering, and the
// two public entry-points RunDependencyPicker / RunDependencyPickerWithOptions.
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

// ── Bubble Tea messages ───────────────────────────────────────────────────────

type successDoneMsg struct{}

// ── Focus panels ─────────────────────────────────────────────────────────────

type focusPanel int

const (
	panelDeps focusPanel = iota
	panelGroups
)

// ── PickerOptions ─────────────────────────────────────────────────────────────

// PickerOptions configures the dependency picker TUI.
type PickerOptions struct {
	Metadata    *metadata.Metadata // pre-fetched; shown immediately if non-nil
	PreSelected []string           // dependency IDs to pre-check
	BootVersion string             // shown in title bar
	JavaVersion string             // shown in status bar
	Template    string             // shown in status bar
}

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	// data
	state       *PickerState
	meta        *metadata.Metadata
	bootVersion string
	javaVersion string
	template    string

	// sub-components
	searchInput textinput.Model
	spinner     spinner.Model
	focus       focusPanel

	// view state flags
	loading     bool
	showHelp    bool
	showConfirm bool
	showSuccess bool
	confirmed   bool
	canceled    bool

	// layout
	width  int
	height int

	// cached scroll offset — recomputed each frame but stored to feed
	// StickyGroupHeader without extra logic in View()
	scrollOffset int

	successStart time.Time
}

const successDuration = 700 * time.Millisecond

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
	return tea.Batch(textinput.Blink, m.spinner.Tick, tea.EnableMouseCellMotion)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case successDoneMsg:
		return m, tea.Quit

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			m = m.handleMouseClick(msg.X, msg.Y)
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// ── Key handling ──────────────────────────────────────────────────────────────

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd) {
	// Always allow quit.
	if msg.String() == "ctrl+c" {
		m.canceled = true
		return m, tea.Quit
	}

	if m.loading {
		return m, nil
	}

	// Confirmation screen.
	if m.showConfirm {
		return m.handleConfirmKey(msg)
	}

	// Help overlay — any key closes.
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	// Success flash — no input accepted.
	if m.showSuccess {
		return m, nil
	}

	// Search mode.
	if m.searchInput.Focused() {
		return m.handleSearchKey(msg)
	}

	// Normal navigation.
	return m.handleNormalKey(msg)
}

func (m model) handleSearchKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.state.ApplyFilter("")
		return m, nil
	case "ctrl+backspace", "ctrl+w":
		m.searchInput.SetValue("")
		m.state.ApplyFilter("")
		return m, nil
	case "enter":
		m.searchInput.Blur()
		return m, nil
	case "up", "ctrl+p":
		m.state.MoveCursor(-1)
		return m, nil
	case "down", "ctrl+n":
		m.state.MoveCursor(1)
		return m, nil
	case " ":
		m.state.ToggleCurrent()
		return m, nil
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.state.ApplyFilter(m.searchInput.Value())
		return m, cmd
	}
}

func (m model) handleNormalKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.canceled = true
		return m, tea.Quit

	case "enter":
		m.showConfirm = true
		return m, nil

	// Vertical navigation.
	case "up", "k":
		m.state.MoveCursor(-1)
	case "down", "j":
		m.state.MoveCursor(1)
	case "home", "g":
		m.state.MoveToFirst()
	case "end", "G":
		m.state.MoveToLast()
	case "pgup", "ctrl+u":
		m.state.PageUp()
	case "pgdown", "ctrl+d":
		m.state.PageDown()

	// Group navigation.
	case "tab", "right", "l":
		m.state.TabToNextGroup()
	case "shift+tab", "left", "h":
		m.state.TabToPrevGroup()

	// Selection.
	case " ":
		m.state.ToggleCurrent()

	// Search.
	case "/", "ctrl+f":
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

func (m model) handleConfirmKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.confirmed = true
		m.showConfirm = false
		m.showSuccess = true
		m.successStart = time.Now()
		return m, tea.Tick(successDuration, func(time.Time) tea.Msg {
			return successDoneMsg{}
		})
	case "n", "N", "esc":
		m.showConfirm = false
	}
	return m, nil
}

func (m model) handleMouseClick(_, _ int) model {
	// Hitmap-based click-to-select is a future enhancement.
	// Mouse events already enable scrolling on some terminals.
	return m
}

// ── View dispatcher ───────────────────────────────────────────────────────────

func (m model) View() string {
	switch {
	case m.loading:
		return m.viewLoading()
	case m.showSuccess:
		return m.viewSuccess()
	case m.showHelp:
		return m.viewHelp()
	case m.showConfirm:
		return m.viewConfirm()
	default:
		return m.viewMain()
	}
}

// ── Loading view ──────────────────────────────────────────────────────────────

func (m model) viewLoading() string {
	pad := strings.Repeat("\n", maxInt(m.height/2-2, 0))
	line := fmt.Sprintf("  %s  Fetching Spring Initializr metadata…", m.spinner.View())
	return pad + SpinnerStyle.Render(line) + "\n"
}

// ── Success flash ─────────────────────────────────────────────────────────────

func (m model) viewSuccess() string {
	n := m.state.SelectedCount()
	noun := "dependency"
	if n != 1 {
		noun = "dependencies"
	}
	pad := strings.Repeat("\n", maxInt(m.height/2-2, 0))
	line := fmt.Sprintf("  ✔  %d %s selected — generating project…", n, noun)
	return pad + SuccessStyle.Render(line) + "\n"
}

// ── Help overlay ──────────────────────────────────────────────────────────────

func (m model) viewHelp() string {
	type entry struct{ key, desc string }
	type section struct {
		title   string
		entries []entry
	}

	sections := []section{
		{
			"Navigation",
			[]entry{
				{"↑ / k", "Move up"},
				{"↓ / j", "Move down"},
				{"Home / g", "First dependency"},
				{"End / G", "Last dependency"},
				{"PgUp / Ctrl+U", "Page up"},
				{"PgDn / Ctrl+D", "Page down"},
				{"Tab / → / l", "Next group"},
				{"Shift+Tab / ← / h", "Previous group"},
			},
		},
		{
			"Selection",
			[]entry{
				{"Space", "Toggle selection"},
				{"Enter", "Open confirmation"},
			},
		},
		{
			"Search",
			[]entry{
				{"/ or Ctrl+F", "Open search"},
				{"Esc", "Clear search"},
				{"Ctrl+Backspace", "Clear entire query"},
			},
		},
		{
			"General",
			[]entry{
				{"?", "Toggle this help"},
				{"q / Ctrl+C", "Quit"},
			},
		},
	}

	var rows []string
	rows = append(rows, HelpTitleStyle.Render("springx — Keyboard Shortcuts"))

	for _, sec := range sections {
		rows = append(rows, HelpSectionStyle.Render(sec.title))
		for _, e := range sec.entries {
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				HelpKeyStyle.Render(e.key),
				HelpDescStyle.Render(e.desc),
			)
			rows = append(rows, row)
		}
	}

	rows = append(rows, "")
	rows = append(rows, AppSubtitleStyle.Render("Press any key to close"))

	box := HelpBoxStyle.Render(strings.Join(rows, "\n"))
	return m.centreBox(box)
}

// ── Confirmation screen ───────────────────────────────────────────────────────

func (m model) viewConfirm() string {
	items := m.state.GetSelectedItems()

	var rows []string
	rows = append(rows, ConfirmTitleStyle.Render("Confirm Generation"))

	addField := func(label, value string) {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top,
			ConfirmLabelStyle.Render(label),
			ConfirmValueStyle.Render(value),
		))
	}

	if m.bootVersion != "" {
		addField("Spring Boot:", m.bootVersion)
	}
	if m.javaVersion != "" {
		addField("Java:", m.javaVersion)
	}
	if m.template != "" {
		addField("Template:", m.template)
	}

	rows = append(rows, "")

	if len(items) == 0 {
		rows = append(rows, ConfirmLabelStyle.Render("Dependencies:")+" "+DepDescStyle.Render("none selected"))
	} else {
		rows = append(rows, ConfirmLabelStyle.Render(
			fmt.Sprintf("Dependencies (%d):", len(items)),
		))
		for _, item := range items {
			bullet := ConfirmDepStyle.Render("  ✓ " + item.Name)
			group := DepDescStyle.Render("  (" + item.GroupName + ")")
			rows = append(rows, bullet+group)
		}
	}

	rows = append(rows, "")
	rows = append(rows, ConfirmPromptStyle.Render("Generate project?  [Y/n]"))

	box := ConfirmBoxStyle.Render(strings.Join(rows, "\n"))
	return m.centreBox(box)
}

// ── Main three-panel view ─────────────────────────────────────────────────────

func (m model) viewMain() string {
	// Fixed chrome row budget.
	const titleRows  = 1
	const ruleRows   = 1
	const searchRows = 1
	const footerRows = 1
	const statusRows = 1
	const ruleRows2  = 1
	overhead := titleRows + ruleRows + searchRows + ruleRows2 + footerRows + statusRows
	contentH := m.height - overhead
	if contentH < 6 {
		contentH = 6
	}

	// Column widths — groups left, selected right, deps take the rest.
	groupW    := 22
	selectedW := 28
	borderCols := 6 // rounded borders: 2 per panel × 3 panels
	spacerCols := 2 // spaces between panels
	depsW := m.width - groupW - selectedW - borderCols - spacerCols
	if depsW < 24 {
		depsW = 24
	}

	rule := HRuleStyle.Render(strings.Repeat("─", m.width))

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderTitleBar(),
		rule,
		m.renderSearchBar(),
		rule,
		lipgloss.JoinHorizontal(lipgloss.Top,
			m.renderGroupPanel(groupW, contentH),
			" ",
			m.renderDepsPanel(depsW, contentH),
			" ",
			m.renderSelectedPanel(selectedW, contentH),
		),
		rule,
		m.renderFooter(),
		m.renderStatusBar(),
	)
}

// ── Title bar ─────────────────────────────────────────────────────────────────

func (m model) renderTitleBar() string {
	left := AppTitleStyle.Render("springx")
	right := ""
	if m.bootVersion != "" {
		right = AppVersionStyle.Render("Spring Boot " + m.bootVersion)
	}
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// ── Search bar ────────────────────────────────────────────────────────────────

func (m model) renderSearchBar() string {
	label := SearchLabelStyle.Render("Search:")
	hint  := SearchHintStyle.Render("(/ or Ctrl+F)")

	var middle string
	if m.searchInput.Focused() {
		middle = SearchActiveStyle.Render(m.searchInput.View()) +
			"  " + SearchingIndicatorStyle.Render("Searching…")
	} else if q := m.state.SearchQuery; q != "" {
		count := m.state.MatchCount()
		var countStr string
		if count == 0 {
			countStr = SearchNoResultStyle.Render("No matching dependencies")
		} else {
			countStr = SearchResultCountStyle.Render(fmt.Sprintf("Found %d %s",
				count, pluralise("dependency", "dependencies", count)))
		}
		middle = SearchIdleStyle.Render(fmt.Sprintf("%q", q)) + "  " + countStr +
			"  " + AppSubtitleStyle.Render("(esc to clear)")
	} else {
		middle = SearchIdleStyle.Render("Press / or Ctrl+F to search")
	}

	return "  " + label + " " + middle + "  " + hint
}

// ── Group panel ───────────────────────────────────────────────────────────────

func (m model) renderGroupPanel(w, h int) string {
	visible := m.state.VisibleGroupNames()
	visSet  := make(map[string]bool, len(visible))
	for _, g := range visible {
		visSet[g] = true
	}

	groups      := m.state.GetGroupNames()
	activeGroup := ""
	if idx := m.state.ActiveGroupIdx(); idx < len(groups) {
		activeGroup = groups[idx]
	}

	var rows []string
	for _, g := range groups {
		name := truncate(g, w-4)
		switch {
		case g == activeGroup:
			rows = append(rows, GroupActiveStyle.Width(w).Render("> "+name))
		case visSet[g]:
			rows = append(rows, GroupNormalStyle.Width(w).Render("  "+name))
		default:
			rows = append(rows, GroupDimStyle.Width(w).Render("  "+name))
		}
	}

	for len(rows) < h-2 {
		rows = append(rows, strings.Repeat(" ", w))
	}

	title   := PanelTitleStyle.Render("Groups")
	content := lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(rows, "\n"))
	return normalBorder.Width(w).Height(h).Render(content)
}

// ── Dependencies panel ────────────────────────────────────────────────────────

func (m model) renderDepsPanel(w, h int) string {
	title    := PanelTitleStyle.Render("Dependencies")
	visibleH := h - 4 // title + sticky header + bottom border

	var rows []string

	if len(m.state.FilteredRows) == 0 {
		rows = append(rows, EmptyStateStyle.Render("No matching dependencies."))
	} else {
		// Active row index in FilteredRows.
		activeRowIdx := -1
		if m.state.Cursor >= 0 && m.state.Cursor < len(m.state.SelectableIdx) {
			activeRowIdx = m.state.SelectableIdx[m.state.Cursor]
		}

		offset := m.computeScrollOffset(visibleH)
		m.scrollOffset = offset // cache for StickyGroupHeader
		end := offset + visibleH
		if end > len(m.state.FilteredRows) {
			end = len(m.state.FilteredRows)
		}

		for i := offset; i < end; i++ {
			row := m.state.FilteredRows[i]
			if row.Type == TypeHeader {
				rows = append(rows, m.renderInlineGroupHeader(row.GroupName, w))
			} else {
				rows = append(rows, m.renderDepRow(row, i == activeRowIdx, w))
			}
		}
	}

	for len(rows) < visibleH {
		rows = append(rows, "")
	}

	// Sticky header — shown above the scrolled content.
	sticky := m.state.StickyGroupHeader(m.scrollOffset)
	stickyLine := StickyHeaderStyle.Width(w + 2).Render("  " + truncate(sticky, w-2))

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		stickyLine,
		strings.Join(rows, "\n"),
	)
	return focusBorder.Width(w).Height(h).Render(content)
}

func (m model) renderInlineGroupHeader(name string, w int) string {
	line := SectionHeaderStyle.Render("  " + name)
	rule := HRuleStyle.Render("  " + strings.Repeat("─", maxInt(0, w-4)))
	return lipgloss.JoinVertical(lipgloss.Left, line, rule)
}

func (m model) renderDepRow(row ListRow, isCursor bool, w int) string {
	isSelected := m.state.Selected[row.ID]

	checkbox := CheckboxOffStyle.Render("[ ]")
	if isSelected {
		checkbox = CheckboxOnStyle.Render("[x]")
	}

	name := row.Name
	if m.state.SearchQuery != "" {
		name = HighlightMatches(name, m.state.SearchQuery)
	}

	// Description — truncated to avoid overflowing the panel.
	desc := ""
	if row.Description != "" {
		maxD := w - len(row.Name) - 8
		if maxD > 10 {
			desc = DepDescStyle.Render("  " + truncate(row.Description, maxD))
		}
	}

	line := fmt.Sprintf(" %s %s%s", checkbox, name, desc)

	switch {
	case isCursor && isSelected:
		return DepCursorSelectedStyle.Width(w).Render(line)
	case isCursor:
		return DepCursorStyle.Width(w).Render(line)
	case isSelected:
		return DepSelectedStyle.Width(w).Render(line)
	default:
		return DepNormalStyle.Width(w).Render(line)
	}
}

func (m model) computeScrollOffset(visibleH int) int {
	if m.state.Cursor < 0 || len(m.state.SelectableIdx) == 0 {
		return 0
	}
	cursorRow := m.state.SelectableIdx[m.state.Cursor]
	if cursorRow < visibleH {
		return 0
	}
	offset := cursorRow - visibleH/2
	maxOffset := len(m.state.FilteredRows) - visibleH
	if maxOffset < 0 {
		maxOffset = 0
	}
	return clamp(offset, 0, maxOffset)
}

// ── Selected panel ────────────────────────────────────────────────────────────

func (m model) renderSelectedPanel(w, h int) string {
	n     := m.state.SelectedCount()
	title := SelectedCountStyle.Render(fmt.Sprintf("Selected (%d)", n))

	items := m.state.GetSelectedItems()
	var rows []string
	for _, item := range items {
		bullet := SelectedBulletStyle.Render("✓")
		name   := SelectedItemStyle.Render(" " + truncate(item.Name, w-5))
		group  := SelectedGroupStyle.Render(" " + truncate(item.GroupName, w-5))
		rows = append(rows, fmt.Sprintf(" %s%s", bullet, name))
		rows = append(rows, "   "+group)
	}
	if n == 0 {
		rows = append(rows, EmptyStateStyle.Render("None yet"))
	}

	for len(rows) < h-3 {
		rows = append(rows, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(rows, "\n"))
	return normalBorder.Width(w).Height(h).Render(content)
}

// ── Footer ────────────────────────────────────────────────────────────────────

func (m model) renderFooter() string {
	sep := FooterSepStyle.Render(" • ")
	k   := FooterKeyStyle

	hints := []string{
		k.Render("↑↓") + " move",
		k.Render("←→/tab") + " groups",
		k.Render("Home/End") + " first/last",
		k.Render("PgUp/PgDn") + " page",
		k.Render("space") + " select",
		k.Render("/") + " search",
		k.Render("enter") + " confirm",
		k.Render("?") + " help",
		k.Render("q") + " quit",
	}

	line := strings.Join(hints, sep)
	return FooterStyle.Width(m.width).Render(line)
}

// ── Status bar ────────────────────────────────────────────────────────────────

func (m model) renderStatusBar() string {
	sep := StatusSepStyle.Render(" │ ")

	parts := []string{
		StatusKeyStyle.Render("Metadata") + StatusValueStyle.Render(" loaded"),
	}
	if m.bootVersion != "" {
		parts = append(parts, StatusKeyStyle.Render("Boot")+StatusValueStyle.Render(" "+m.bootVersion))
	}
	if m.javaVersion != "" {
		parts = append(parts, StatusKeyStyle.Render("Java")+StatusValueStyle.Render(" "+m.javaVersion))
	}
	if m.template != "" {
		parts = append(parts, StatusKeyStyle.Render("Template")+StatusValueStyle.Render(" "+m.template))
	}
	n := m.state.SelectedCount()
	parts = append(parts, StatusKeyStyle.Render("Selected")+StatusValueStyle.Render(fmt.Sprintf(" %d", n)))

	return StatusBarStyle.Width(m.width).Render(strings.Join(parts, sep))
}

// ── Layout helpers ────────────────────────────────────────────────────────────

// centreBox centres a pre-rendered box string in the terminal.
func (m model) centreBox(box string) string {
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
	return strings.Repeat("\n", padV) + strings.Repeat(" ", padH) + box + "\n"
}

// truncate clips s to maxLen runes, appending "…" if needed.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	if maxLen == 1 {
		return "…"
	}
	return string(r[:maxLen-1]) + "…"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pluralise(singular, plural string, n int) string {
	if n == 1 {
		return singular
	}
	return plural
}

// ── Public API ────────────────────────────────────────────────────────────────

// RunDependencyPicker launches the TUI picker and returns selected dep IDs.
func RunDependencyPicker(meta *metadata.Metadata, preSelected []string) ([]string, error) {
	return RunDependencyPickerWithOptions(PickerOptions{
		Metadata:    meta,
		PreSelected: preSelected,
	})
}

// RunDependencyPickerWithOptions is the full entry-point with all display options.
func RunDependencyPickerWithOptions(opts PickerOptions) ([]string, error) {
	m := newModel(opts)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	final, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("dependency picker failed: %w", err)
	}

	fm, ok := final.(model)
	if !ok || fm.canceled {
		return nil, fmt.Errorf("dependency selection canceled")
	}

	return fm.state.GetSelectedIDs(), nil
}
