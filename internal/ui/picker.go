// Package ui implements the Bubble Tea terminal UI for springx.
//
// Architecture:
//   - state.go  — pure business logic and state (PickerState)
//   - styles.go — all Lipgloss style definitions and palette
//   - picker.go — Bubble Tea model, view rendering, key handling, public API
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

// ── Messages ──────────────────────────────────────────────────────────────────

type successDoneMsg struct{}

// ── Focus panels ─────────────────────────────────────────────────────────────

type focusPanel int

const (
	panelGroups focusPanel = iota
	panelDeps
	panelSelected
)

// ── Confirmation button focus ─────────────────────────────────────────────────

type confirmBtn int

const (
	btnYes confirmBtn = iota
	btnNo
)

// ── PickerOptions ─────────────────────────────────────────────────────────────

// PickerOptions configures the dependency picker TUI.
type PickerOptions struct {
	Metadata    *metadata.Metadata // pre-fetched; shown immediately if non-nil
	PreSelected []string           // dependency IDs to pre-check
	BootVersion string             // shown in title bar and status bar
	JavaVersion string             // shown in status bar
	Template    string             // shown in status bar
}

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	// data
	state        *PickerState
	meta         *metadata.Metadata
	successStart time.Time

	// sub-components
	searchInput textinput.Model
	spinner     spinner.Model

	// strings
	bootVersion string
	javaVersion string
	template    string

	// view state flags
	focus        focusPanel
	confirmFocus confirmBtn

	// layout
	width  int
	height int

	// cached scroll offset for StickyGroupHeader
	scrollOffset int

	loading     bool
	showHelp    bool
	showConfirm bool
	showSuccess bool
	confirmed   bool
	canceled    bool
}

const successDuration = 700 * time.Millisecond

// ── Constructor ───────────────────────────────────────────────────────────────

func newModel(opts PickerOptions) model {
	ti := textinput.New()
	ti.Placeholder = "type to filter…"
	ti.Prompt = ""
	ti.CharLimit = 64

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = SpinnerStyle

	m := model{
		searchInput: ti,
		spinner:     sp,
		focus:       panelDeps,
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

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// ── Key handling ──────────────────────────────────────────────────────────────

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.canceled = true
		return m, tea.Quit
	}
	if m.loading {
		return m, nil
	}
	if m.showConfirm {
		return m.handleConfirmKey(msg)
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}
	if m.showSuccess {
		return m, nil
	}
	if m.searchInput.Focused() {
		return m.handleSearchKey(msg)
	}
	return m.handleNormalKey(msg)
}

func (m model) handleSearchKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// ESC exits search and restores pre-search cursor position.
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.state.ClearSearch()
		return m, nil

	case "ctrl+backspace", "ctrl+w":
		m.searchInput.SetValue("")
		m.state.ApplyFilter("")
		return m, nil

	case "enter":
		// Enter exits the search input but stays in the picker.
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

	// F5 — open the confirmation screen to generate.
	// Enter alone is intentionally a no-op; generation must be explicit.
	case "f5":
		m.showConfirm = true
		m.confirmFocus = btnYes
		return m, nil

	case "enter":
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

	// Panel focus cycling: left/right or Tab.
	case "right", "l":
		m.focus = (m.focus + 1) % 3
	case "left", "h":
		m.focus = (m.focus + 2) % 3
	case "tab":
		m.state.TabToNextGroup()
	case "shift+tab":
		m.state.TabToPrevGroup()

	// Selection — Space is the only toggle key.
	case " ":
		m.state.ToggleCurrent()

	// Search.
	case "/", "ctrl+f":
		m.state.BeginSearch()
		m.searchInput.Focus()
		return m, textinput.Blink

	// ESC in normal mode clears an active filter.
	case "esc":
		if m.state.SearchQuery != "" {
			m.searchInput.SetValue("")
			m.state.ClearSearch()
		}

	case "?":
		m.showHelp = true
	}

	return m, nil
}

func (m model) handleConfirmKey(msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	// Tab / arrow keys move focus between [Y] and [N].
	case "tab", "right", "l":
		if m.confirmFocus == btnYes {
			m.confirmFocus = btnNo
		} else {
			m.confirmFocus = btnYes
		}
	case "shift+tab", "left", "h":
		if m.confirmFocus == btnNo {
			m.confirmFocus = btnYes
		} else {
			m.confirmFocus = btnNo
		}

	// Y / y always confirms regardless of button focus.
	case "y", "Y":
		m.confirmed = true
		m.showConfirm = false
		m.showSuccess = true
		m.successStart = time.Now()
		return m, tea.Tick(successDuration, func(time.Time) tea.Msg {
			return successDoneMsg{}
		})

	// N / n / ESC always cancels back to picker.
	case "n", "N", "esc":
		m.showConfirm = false

	// Enter confirms the focused button.
	case "enter":
		if m.confirmFocus == btnYes {
			m.confirmed = true
			m.showConfirm = false
			m.showSuccess = true
			m.successStart = time.Now()
			return m, tea.Tick(successDuration, func(time.Time) tea.Msg {
				return successDoneMsg{}
			})
		}
		m.showConfirm = false
	}
	return m, nil
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
		{"Navigation", []entry{
			{"↑ / k", "Move up"},
			{"↓ / j", "Move down"},
			{"Home / g", "First dependency"},
			{"End / G", "Last dependency"},
			{"PgUp / Ctrl+U", "Page up"},
			{"PgDn / Ctrl+D", "Page down"},
			{"Tab", "Jump to next group"},
			{"Shift+Tab", "Jump to previous group"},
			{"← / h  → / l", "Switch panel focus"},
		}},
		{"Selection", []entry{
			{"Space", "Toggle dependency on/off"},
			{"F5", "Open confirm & generate"},
		}},
		{"Search", []entry{
			{"/ or Ctrl+F", "Open search box"},
			{"Esc", "Clear search, restore cursor"},
			{"Ctrl+Backspace", "Clear entire query"},
			{"Enter (in search)", "Exit search, keep results"},
		}},
		{"General", []entry{
			{"?", "Toggle this help screen"},
			{"q / Ctrl+C", "Quit without generating"},
		}},
	}

	var rows []string
	rows = append(rows, HelpTitleStyle.Render("springx — Keyboard Reference"))

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

	rows = append(rows, "", AppSubtitleStyle.Render("Press any key to close"))

	box := HelpBoxStyle.Render(strings.Join(rows, "\n"))
	return m.centreBox(box)
}

// ── Confirmation screen ───────────────────────────────────────────────────────

func (m model) viewConfirm() string {
	var rows []string
	rows = append(rows,
		ConfirmTitleStyle.Render("  Review & Confirm Generation"),
		HRuleStyle.Render(strings.Repeat("─", 44)),
		"",
	)

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

	// Dependencies grouped by category.
	groups := m.state.GetSelectedByGroup()
	if len(groups) == 0 {
		rows = append(rows, ConfirmLabelStyle.Render("Dependencies:")+" "+
			DepDescStyle.Render("none selected"))
	} else {
		total := m.state.SelectedCount()
		rows = append(rows, ConfirmLabelStyle.Render(
			fmt.Sprintf("Dependencies (%d):", total),
		))
		for _, g := range groups {
			rows = append(rows, "  "+ConfirmGroupStyle.Render(g.Name))
			for _, item := range g.Items {
				rows = append(rows, "    "+ConfirmDepStyle.Render("✓ "+item.Name))
			}
		}
	}

	rows = append(rows, "", ConfirmPromptStyle.Render("Generate project?"), "")

	// [Y] / [N] buttons with focus indicator.
	btnY := ConfirmBtnYesNormal.Render("  Y — Generate  ")
	btnN := ConfirmBtnNoNormal.Render("  N — Cancel  ")
	if m.confirmFocus == btnYes {
		btnY = ConfirmBtnYesFocused.Render("  Y — Generate  ")
	} else {
		btnN = ConfirmBtnNoFocused.Render("  N — Cancel  ")
	}
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, btnY, "   ", btnN), "", AppSubtitleStyle.Render(
		"Tab / ← →  move focus    Enter  confirm    Esc / n  go back"))
	box := ConfirmBoxStyle.Render(strings.Join(rows, "\n"))
	return m.centreBox(box)
}

// ── Main dashboard view ───────────────────────────────────────────────────────
//
// Layout (matches the design spec):
//
//  ┌──────────────────────────────────────────────────────────────────────────┐
//  │ springx                                          Spring Boot 3.5.4       │
//  ├──────────────────────────────────────────────────────────────────────────┤
//  │ Search                                                          Ctrl+F    │
//  │ ╔══════════════════════════════════════════════╗                          │
//  │ ║ > postgres█                                  ║  Found 8 dependencies   │
//  │ ╚══════════════════════════════════════════════╝                          │
//  ├────────────────┬──────────────────────────────┬─────────────────────────┤
//  │ Groups         │ Dependencies                  │ Selected (4)            │
//  │ ❯ Web          │ ❯ [x] Spring Web              │ Web                     │
//  │   Data         │   [ ] GraphQL                 │   ✓ Spring Web          │
//  │   Security     │   [x] PostgreSQL Driver        │ Data                    │
//  ├────────────────┴──────────────────────────────┴─────────────────────────┤
//  │ Template: REST API   Java:21   Boot:3.5.4   Selected:4                  │
//  ├──────────────────────────────────────────────────────────────────────────┤
//  │ ↑↓ Move  ←→ Panels  Space Toggle  / Search  Esc Clear  Ctrl+↵ Generate  │
//  └──────────────────────────────────────────────────────────────────────────┘

func (m model) viewMain() string {
	// ── Budget fixed chrome rows ──────────────────────────────────────────
	// Title bar: 1 row
	// Separator: 1 row
	// Search label: 1 row
	// Search box (border=2 + input=1): 3 rows
	// Separator: 1 row
	// Three-panel row: variable
	// Separator: 1 row
	// Status bar: 1 row
	// Footer: 1 row
	const fixedOverhead = 10
	contentH := m.height - fixedOverhead
	if contentH < 6 {
		contentH = 6
	}

	// ── Column widths ─────────────────────────────────────────────────────
	// Groups panel: fixed 22 cols (inner content, border adds 2 each side)
	// Selected panel: fixed 26 cols
	// Deps panel: remainder
	const groupInner = 20
	const selectedInner = 24
	const borderEach = 2 // one border character each side
	const panelGap = 0   // panels sit directly adjacent, borders touch

	totalBorders := (borderEach * 3) + panelGap
	depsInner := m.width - groupInner - selectedInner - totalBorders - 4
	if depsInner < 20 {
		depsInner = 20
	}

	sep := HRuleStyle.Render(strings.Repeat("─", m.width))

	title := m.renderTitleBar()
	search := m.renderSearchSection()
	panels := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderGroupPanel(groupInner, contentH),
		m.renderDepsPanel(depsInner, contentH),
		m.renderSelectedPanel(selectedInner, contentH),
	)
	status := m.renderStatusBar()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		sep,
		search,
		sep,
		panels,
		sep,
		status,
		footer,
	)
}

// ── Title bar ─────────────────────────────────────────────────────────────────

func (m model) renderTitleBar() string {
	left := AppTitleStyle.Render(" springx")
	right := ""
	if m.bootVersion != "" {
		right = AppVersionStyle.Render("Spring Boot " + m.bootVersion + " ")
	}
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// ── Search section ────────────────────────────────────────────────────────────
//
// Renders:
//   Search                                                    Ctrl+F
//   ╔══════════════════════════════╗
//   ║ > postgres█                  ║   Searching for: postgres  Found 8
//   ╚══════════════════════════════╝

func (m model) renderSearchSection() string {
	active := m.searchInput.Focused()
	query := m.state.SearchQuery

	// Row 1 — label + hint.
	label := SearchLabelStyle.Render(" Search")
	hint := SearchHintStyle.Render("Ctrl+F ")
	labelGap := m.width - lipgloss.Width(label) - lipgloss.Width(hint)
	if labelGap < 1 {
		labelGap = 1
	}
	labelRow := label + strings.Repeat(" ", labelGap) + hint

	// Input box width: roughly half the terminal width, min 32.
	boxW := m.width/2 - 4
	if boxW < 32 {
		boxW = 32
	}
	if boxW > m.width-4 {
		boxW = m.width - 4
	}

	// Build the text shown inside the box.
	var innerText string
	switch {
	case active:
		innerText = SearchInputActiveStyle.Width(boxW).Render(
			"❯ " + m.searchInput.View(),
		)
	case query != "":
		innerText = SearchInputIdleStyle.Width(boxW).Render(
			"❯ " + query,
		)
	default:
		innerText = SearchInputEmptyStyle.Width(boxW).Render(
			"  type to filter…",
		)
	}

	box := searchBoxBorder(active).
		Padding(0, 1).
		Render(innerText)

	// Status text beside the box.
	var statusText string
	switch {
	case active && query != "":
		count := m.state.MatchCount()
		searching := SearchingIndicatorStyle.Render("Searching for: " + query)
		var countStr string
		if count == 0 {
			countStr = SearchNoResultStyle.Render("  No dependencies found.")
		} else {
			countStr = SearchResultCountStyle.Render(
				fmt.Sprintf("  Found %d %s",
					count, pluralise("dependency", "dependencies", count)))
		}
		statusText = searching + countStr
	case !active && query != "":
		count := m.state.MatchCount()
		if count == 0 {
			statusText = SearchNoResultStyle.Render("No dependencies found.  ") +
				AppSubtitleStyle.Render("Press Esc to clear search.")
		} else {
			statusText = SearchResultCountStyle.Render(
				fmt.Sprintf("Found %d %s",
					count, pluralise("dependency", "dependencies", count))) +
				AppSubtitleStyle.Render("  Esc to clear")
		}
	default:
		statusText = AppSubtitleStyle.Render("  Press / or Ctrl+F to search")
	}

	boxRow := lipgloss.JoinHorizontal(lipgloss.Center, box, "  ", statusText)
	return lipgloss.JoinVertical(lipgloss.Left, labelRow, boxRow)
}

// ── Group panel ───────────────────────────────────────────────────────────────

func (m model) renderGroupPanel(innerW, h int) string {
	focused := m.focus == panelGroups
	groups := m.state.GetGroupNames()
	visSet := make(map[string]bool)
	for _, g := range m.state.VisibleGroupNames() {
		visSet[g] = true
	}

	activeGroup := ""
	if idx := m.state.ActiveGroupIdx(); idx < len(groups) {
		activeGroup = groups[idx]
	}

	// Build group rows — check if any dep in that group is selected.
	selectedGroups := make(map[string]bool)
	for _, sg := range m.state.GetSelectedByGroup() {
		selectedGroups[sg.Name] = true
	}

	var rows []string
	for _, g := range groups {
		name := truncate(g, innerW-4)
		hasSel := selectedGroups[g]
		switch {
		case g == activeGroup && focused:
			marker := "❯ "
			rows = append(rows, GroupCursorStyle.Width(innerW).Render(marker+name))
		case g == activeGroup:
			marker := "❯ "
			rows = append(rows, GroupNormalStyle.Width(innerW).Render(marker+name))
		case hasSel:
			rows = append(rows, GroupHasSelectionStyle.Width(innerW).Render("  "+name))
		case visSet[g]:
			rows = append(rows, GroupNormalStyle.Width(innerW).Render("  "+name))
		default:
			rows = append(rows, GroupDimStyle.Width(innerW).Render("  "+name))
		}
	}

	// Pad to fill panel height.
	innerH := h - 4 // border + title + padding
	for len(rows) < innerH {
		rows = append(rows, strings.Repeat(" ", innerW))
	}

	titleStyle := PanelTitleStyle
	if focused {
		titleStyle = FocusedPanelTitleStyle
	}
	title := titleStyle.Render("Groups")
	content := lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(rows, "\n"))

	return groupPanelBorder(focused).Width(innerW).Height(h).Render(content)
}

// ── Dependencies panel ────────────────────────────────────────────────────────

func (m model) renderDepsPanel(innerW, h int) string {
	focused := m.focus == panelDeps
	// visibleH: rows available for list content (h minus border top/bottom,
	// title row, sticky-header row).
	visibleH := h - 4

	var rows []string

	if len(m.state.FilteredRows) == 0 {
		rows = append(rows,
			EmptyStateStyle.Render("No dependencies found."),
			EmptyStateStyle.Render("Press Esc to clear search."),
		)
	} else {
		activeRowIdx := -1
		if m.state.Cursor >= 0 && m.state.Cursor < len(m.state.SelectableIdx) {
			activeRowIdx = m.state.SelectableIdx[m.state.Cursor]
		}

		offset := m.computeScrollOffset(visibleH)
		m.scrollOffset = offset
		end := offset + visibleH
		if end > len(m.state.FilteredRows) {
			end = len(m.state.FilteredRows)
		}

		for i := offset; i < end; i++ {
			row := m.state.FilteredRows[i]
			if row.Type == TypeHeader {
				rows = append(rows, m.renderInlineGroupHeader(row.GroupName, innerW))
			} else {
				rows = append(rows, m.renderDepRow(row, i == activeRowIdx, innerW))
			}
		}
	}

	for len(rows) < visibleH {
		rows = append(rows, "")
	}

	// Sticky header — always shows which group is currently visible.
	sticky := m.state.StickyGroupHeader(m.scrollOffset)
	stickyLine := StickyHeaderStyle.Width(innerW + 2).Render("  " + truncate(sticky, innerW-2))

	titleStyle := PanelTitleStyle
	if focused {
		titleStyle = FocusedPanelTitleStyle
	}

	title := titleStyle.Render("Dependencies")
	if m.state.SearchQuery != "" {
		count := m.state.MatchCount()
		countLabel := SearchResultCountStyle.Render(
			fmt.Sprintf(" (%d)", count))
		title = titleStyle.Render("Dependencies") + countLabel
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		stickyLine,
		strings.Join(rows, "\n"),
	)
	return depsPanelBorder(focused).Width(innerW).Height(h).Render(content)
}

func (m model) renderInlineGroupHeader(name string, w int) string {
	line := SectionHeaderStyle.Render("  " + name)
	rule := HRuleStyle.Render("  " + strings.Repeat("─", maxInt(0, w-4)))
	return lipgloss.JoinVertical(lipgloss.Left, line, rule)
}

func (m model) renderDepRow(row ListRow, isCursor bool, w int) string {
	isSelected := m.state.Selected[row.ID]

	var checkbox string
	if isSelected {
		checkbox = CheckboxOnStyle.Render("[x]")
	} else {
		checkbox = CheckboxOffStyle.Render("[ ]")
	}

	name := row.Name
	if m.state.SearchQuery != "" {
		name = HighlightMatches(name, m.state.SearchQuery)
	}

	// Cursor arrow — makes the focused row unambiguous.
	cursor := "  "
	if isCursor {
		cursor = CursorArrowStyle.Render("❯ ")
	}

	// Description — only if there's room.
	desc := ""
	rawNameLen := len([]rune(row.Name))
	maxDesc := w - rawNameLen - 10
	if row.Description != "" && maxDesc > 8 {
		desc = DepDescStyle.Render("  " + truncate(row.Description, maxDesc))
	}

	line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, desc)

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
	return clamp(offset, maxOffset)
}

// ── Selected panel ────────────────────────────────────────────────────────────

func (m model) renderSelectedPanel(innerW, h int) string {
	focused := m.focus == panelSelected
	n := m.state.SelectedCount()

	titleStyle := PanelTitleStyle
	if focused {
		titleStyle = FocusedPanelTitleStyle
	}
	title := titleStyle.Render(fmt.Sprintf("Selected (%d)", n))

	groups := m.state.GetSelectedByGroup()
	var rows []string

	if n == 0 {
		rows = append(rows, SelectedEmptyStyle.Width(innerW).Render(" None yet"))
	} else {
		for _, g := range groups {
			// Group label.
			rows = append(rows, SelectedGroupLabelStyle.Width(innerW).Render(
				" "+truncate(g.Name, innerW-3)))
			for _, item := range g.Items {
				bullet := SelectedBulletStyle.Render("✓")
				name := SelectedItemStyle.Render(" " + truncate(item.Name, innerW-5))
				rows = append(rows, fmt.Sprintf("  %s%s", bullet, name))
			}
		}
	}

	innerH := h - 4
	for len(rows) < innerH {
		rows = append(rows, "")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, title, strings.Join(rows, "\n"))
	return selectedPanelBorder(focused).Width(innerW).Height(h).Render(content)
}

// ── Status bar ────────────────────────────────────────────────────────────────

func (m model) renderStatusBar() string {
	sep := StatusSepStyle.Render(" │ ")

	var parts []string

	if m.template != "" {
		parts = append(parts, StatusKeyStyle.Render("Template:")+
			StatusValueStyle.Render(" "+m.template))
	}
	if m.bootVersion != "" {
		parts = append(parts, StatusKeyStyle.Render("Boot:")+
			StatusValueStyle.Render(" "+m.bootVersion))
	}
	if m.javaVersion != "" {
		parts = append(parts, StatusKeyStyle.Render("Java:")+
			StatusValueStyle.Render(" "+m.javaVersion))
	}

	n := m.state.SelectedCount()
	parts = append(parts, StatusKeyStyle.Render("Selected:")+
		StatusValueStyle.Render(fmt.Sprintf(" %d", n)))

	if m.state.SearchQuery != "" {
		parts = append(parts, StatusKeyStyle.Render("Filter:")+
			StatusValueStyle.Render(" "+m.state.SearchQuery))
	} else {
		parts = append(parts, StatusValueStyle.Render("Metadata: loaded"))
	}

	line := " " + strings.Join(parts, sep) + " "
	return StatusBarStyle.Width(m.width).Render(line)
}

// ── Footer ────────────────────────────────────────────────────────────────────

func (m model) renderFooter() string {
	sep := FooterSepStyle.Render("  ")
	k := FooterKeyStyle
	d := FooterDescStyle

	hints := []string{
		k.Render("↑↓") + d.Render(" Move"),
		k.Render("←→") + d.Render(" Panels"),
		k.Render("Tab") + d.Render(" Group"),
		k.Render("Space") + d.Render(" Toggle"),
		k.Render("/") + d.Render(" Search"),
		k.Render("Esc") + d.Render(" Clear"),
		k.Render("?") + d.Render(" Help"),
		k.Render("F5") + d.Render(" Generate"),
		k.Render("q") + d.Render(" Quit"),
	}

	line := " " + strings.Join(hints, sep) + " "
	return FooterStyle.Width(m.width).Render(line)
}

// ── Layout helpers ────────────────────────────────────────────────────────────

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
