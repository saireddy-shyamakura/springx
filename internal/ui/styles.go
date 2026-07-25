package ui

import "github.com/charmbracelet/lipgloss"

// ── Palette ───────────────────────────────────────────────────────────────────
// All colours use explicit 256-colour indices so they remain legible on both
// dark and light terminals without adaptation magic.

const (
	clrAccent      = lipgloss.Color("205") // hot-pink   — title, accents
	clrBlue        = lipgloss.Color("39")  // cyan-blue  — group headers
	clrGreen       = lipgloss.Color("42")  // green      — checked items, success
	clrYellow      = lipgloss.Color("229") // yellow     — cursor foreground, search
	clrPurple      = lipgloss.Color("57")  // purple     — cursor background
	clrRed         = lipgloss.Color("196") // red        — errors
	clrDimText     = lipgloss.Color("243") // mid-grey   — secondary text
	clrBorder      = lipgloss.Color("240") // mid-grey   — borders, rules, separators
	clrNormal      = lipgloss.Color("252") // near-white — body text
	clrStatusBg    = lipgloss.Color("235") // very dark  — status / header bars
	clrStatusFg    = lipgloss.Color("250") // light-grey — status bar text
	clrSearchBg    = lipgloss.Color("17")  // dark blue  — active search background
	clrSearchFg    = lipgloss.Color("231") // white      — active search text
	clrSearchBdr   = lipgloss.Color("39")  // cyan-blue  — focused search border
	clrHighlight   = lipgloss.Color("220") // gold       — matched search text
	clrHighlightBg = lipgloss.Color("52")  // dark red   — matched search background

	// Panel border colours — each panel has its own identity.
	clrGroupBorder    = lipgloss.Color("238") // dark-grey  — groups panel (unfocused)
	clrDepsBorder     = lipgloss.Color("238") // dark-grey  — deps panel (unfocused)
	clrSelectedBorder = lipgloss.Color("238") // dark-grey  — selected panel (unfocused)

	// Focused border colours — vivid, immediately visible.
	clrGroupFocus    = lipgloss.Color("33")  // bright blue — groups focused
	clrDepsFocus     = lipgloss.Color("205") // hot-pink    — deps focused (primary)
	clrSelectedFocus = lipgloss.Color("42")  // green       — selected focused

	clrSelectedBg  = lipgloss.Color("22")  // dark-green — selected row background
	clrConfirmBg   = lipgloss.Color("17")  // dark-blue  — confirmation screen bg
	clrProgressDim = lipgloss.Color("241") // grey       — pending progress steps
	clrProgressDone= lipgloss.Color("42")  // green      — completed progress steps
	clrProgressCur = lipgloss.Color("229") // yellow     — current progress step
	clrResultCount = lipgloss.Color("117") // sky-blue   — "Found N dependencies"
	clrWarning     = lipgloss.Color("214") // orange     — warnings

	// Confirmation button colours.
	clrBtnYesBg = lipgloss.Color("28")  // deep green  — [Y] button background
	clrBtnYesFg = lipgloss.Color("231") // white       — [Y] button foreground
	clrBtnNoBg  = lipgloss.Color("88")  // deep red    — [N] button background
	clrBtnNoFg  = lipgloss.Color("231") // white       — [N] button foreground
	clrBtnFocus = lipgloss.Color("229") // yellow      — focused button border
)

// ── Panel borders ─────────────────────────────────────────────────────────────
// Each panel uses a distinct border character set so panels are visually
// differentiated even before focus. The focused variant swaps to a thick
// border with a vivid colour.

// groupPanelBorder returns the border style for the Groups panel.
// Unfocused: normal rounded, dark grey. Focused: thick rounded, bright blue.
func groupPanelBorder(focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(clrGroupFocus)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(clrGroupBorder)
}

// depsPanelBorder returns the border style for the Dependencies panel.
// Unfocused: normal rounded, dark grey. Focused: double border, hot-pink.
func depsPanelBorder(focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(clrDepsFocus)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(clrDepsBorder)
}

// selectedPanelBorder returns the border style for the Selected panel.
// Unfocused: normal rounded, dark grey. Focused: thick, green.
func selectedPanelBorder(focused bool) lipgloss.Style {
	if focused {
		return lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(clrSelectedFocus)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(clrSelectedBorder)
}

// searchBoxBorder returns the border style for the search input box.
// Active: double border, cyan. Idle: rounded, dark grey.
func searchBoxBorder(active bool) lipgloss.Style {
	if active {
		return lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(clrSearchBdr)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(clrBorder)
}

// ── Application title bar ─────────────────────────────────────────────────────

var (
	AppTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrAccent)

	AppVersionStyle = lipgloss.NewStyle().
			Foreground(clrDimText)

	AppSubtitleStyle = lipgloss.NewStyle().
				Foreground(clrDimText)

	SectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBlue).
				PaddingLeft(1)

	HRuleStyle = lipgloss.NewStyle().
			Foreground(clrBorder)
)

// ── Search bar ────────────────────────────────────────────────────────────────

var (
	// SearchLabelStyle: bold label above the search box.
	SearchLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrAccent)

	// SearchInputActiveStyle: the text inside the box while focused.
	SearchInputActiveStyle = lipgloss.NewStyle().
				Background(clrSearchBg).
				Foreground(clrSearchFg)

	// SearchInputIdleStyle: box content while not focused but has a query.
	SearchInputIdleStyle = lipgloss.NewStyle().
				Foreground(clrYellow)

	// SearchInputEmptyStyle: box content when empty and not focused.
	SearchInputEmptyStyle = lipgloss.NewStyle().
				Foreground(clrDimText)

	// SearchingIndicatorStyle: "Searching for: <query>" label.
	SearchingIndicatorStyle = lipgloss.NewStyle().
				Foreground(clrYellow).
				Bold(true)

	// SearchResultCountStyle: "Found N dependencies" label.
	SearchResultCountStyle = lipgloss.NewStyle().
				Foreground(clrResultCount).
				Bold(true)

	// SearchNoResultStyle: shown when filter matches nothing.
	SearchNoResultStyle = lipgloss.NewStyle().
				Foreground(clrRed).
				Bold(true)

	// SearchHintStyle: "Ctrl+F" shortcut key shown in search row.
	SearchHintStyle = lipgloss.NewStyle().
			Foreground(clrDimText)

	// HighlightMatchStyle wraps matched characters in the dep list.
	HighlightMatchStyle = lipgloss.NewStyle().
				Foreground(clrHighlight).
				Background(clrHighlightBg).
				Bold(true)
)

// ── Group panel ───────────────────────────────────────────────────────────────

var (
	// GroupCursorStyle: the actively highlighted group row.
	GroupCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrYellow).
				Background(clrPurple)

	// GroupNormalStyle: a group visible in the current filter.
	GroupNormalStyle = lipgloss.NewStyle().
				Foreground(clrNormal)

	// GroupDimStyle: a group not in current filter results.
	GroupDimStyle = lipgloss.NewStyle().
			Foreground(clrDimText)

	// GroupHasSelectionStyle: a group with at least one selected dep.
	GroupHasSelectionStyle = lipgloss.NewStyle().
				Foreground(clrGreen)
)

// ── Dependency panel ──────────────────────────────────────────────────────────

var (
	// DepCursorStyle: highlighted row under the navigation cursor.
	DepCursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrYellow).
			Background(lipgloss.Color("20")) // dark blue

	// DepCursorSelectedStyle: cursor row that is also selected.
	DepCursorSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrGreen).
				Background(lipgloss.Color("20"))

	// DepSelectedStyle: selected but not under cursor.
	DepSelectedStyle = lipgloss.NewStyle().
				Foreground(clrGreen).
				Background(clrSelectedBg)

	// DepNormalStyle: normal unselected, non-cursor row.
	DepNormalStyle = lipgloss.NewStyle().
			Foreground(clrNormal)

	// DepDescStyle: dimmed description text beside a dependency name.
	DepDescStyle = lipgloss.NewStyle().
			Foreground(clrDimText)

	// StickyHeaderStyle: pinned group label at top of deps panel while scrolling.
	StickyHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBlue).
				Background(lipgloss.Color("234")).
				PaddingLeft(1)

	// EmptyStateStyle: shown when no dependencies match a search.
	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(clrDimText).
			Italic(true).
			PaddingLeft(2).
			PaddingTop(1)

	CheckboxOnStyle = lipgloss.NewStyle().
			Foreground(clrGreen).
			Bold(true)

	CheckboxOffStyle = lipgloss.NewStyle().
				Foreground(clrBorder)

	// CursorArrowStyle: the "❯" glyph preceding the focused dep row.
	CursorArrowStyle = lipgloss.NewStyle().
				Foreground(clrYellow).
				Bold(true)
)

// ── Selected panel ────────────────────────────────────────────────────────────

var (
	// SelectedPanelTitleStyle: "Selected (N)" heading.
	SelectedPanelTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrGreen)

	// SelectedGroupLabelStyle: group name inside the selected panel.
	SelectedGroupLabelStyle = lipgloss.NewStyle().
				Foreground(clrBlue).
				Bold(true)

	// SelectedItemStyle: individual selected dep name.
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(clrGreen)

	// SelectedBulletStyle: the "✓" bullet preceding each selected dep.
	SelectedBulletStyle = lipgloss.NewStyle().
				Foreground(clrGreen).
				Bold(true)

	// SelectedEmptyStyle: shown when nothing is selected yet.
	SelectedEmptyStyle = lipgloss.NewStyle().
				Foreground(clrDimText).
				Italic(true)
)

// ── Panel titles ──────────────────────────────────────────────────────────────

// PanelTitleStyle is the title row inside any panel box.
var PanelTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(clrBlue).
	PaddingLeft(1)

// FocusedPanelTitleStyle is the title row for the currently focused panel.
var FocusedPanelTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(clrAccent).
	PaddingLeft(1)

// ── Status bar ────────────────────────────────────────────────────────────────

var (
	StatusBarStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrStatusFg)

	StatusKeyStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrAccent).
			Bold(true)

	StatusValueStyle = lipgloss.NewStyle().
				Background(clrStatusBg).
				Foreground(clrStatusFg)

	StatusSepStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrBorder)
)

// ── Help overlay ──────────────────────────────────────────────────────────────

var (
	HelpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrAccent).
			MarginBottom(1)

	HelpSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBlue).
				MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrYellow).
			Width(24)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(clrNormal)

	HelpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrAccent).
			Padding(1, 3)
)

// ── Footer ────────────────────────────────────────────────────────────────────

var (
	FooterStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrDimText)

	FooterKeyStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrNormal).
			Bold(true)

	FooterDescStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrDimText)

	FooterSepStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrBorder)
)

// ── Spinner / loading ─────────────────────────────────────────────────────────

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(clrAccent).
	Bold(true)

// ── Success ───────────────────────────────────────────────────────────────────

var SuccessStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(clrGreen)

// ── Error screen ─────────────────────────────────────────────────────────────

var (
	ErrorTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrRed).
			MarginBottom(1)

	ErrorReasonStyle = lipgloss.NewStyle().
				Foreground(clrNormal).
				PaddingLeft(2)

	ErrorSuggestionStyle = lipgloss.NewStyle().
				Foreground(clrDimText).
				PaddingLeft(4)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrRed).
			Padding(1, 3)
)

// ── Confirmation screen ───────────────────────────────────────────────────────

var (
	ConfirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrAccent).
				MarginBottom(1)

	ConfirmLabelStyle = lipgloss.NewStyle().
				Foreground(clrDimText).
				Width(18)

	ConfirmValueStyle = lipgloss.NewStyle().
				Foreground(clrNormal).
				Bold(true)

	ConfirmDepStyle = lipgloss.NewStyle().
			Foreground(clrGreen)

	ConfirmGroupStyle = lipgloss.NewStyle().
				Foreground(clrBlue).
				Bold(true)

	ConfirmPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrYellow).
				MarginTop(1)

	ConfirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrAccent).
			Padding(1, 3)

	// ConfirmBtnYes: the [Y] button, focused state.
	ConfirmBtnYesFocused = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBtnYesFg).
				Background(clrBtnYesBg).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(clrBtnFocus).
				Padding(0, 2)

	// ConfirmBtnYes: the [Y] button, unfocused state.
	ConfirmBtnYesNormal = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBtnYesFg).
				Background(clrBtnYesBg).
				Border(lipgloss.HiddenBorder()).
				Padding(0, 2)

	// ConfirmBtnNo: the [N] button, focused state.
	ConfirmBtnNoFocused = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBtnNoFg).
				Background(clrBtnNoBg).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(clrBtnFocus).
				Padding(0, 2)

	// ConfirmBtnNo: the [N] button, unfocused state.
	ConfirmBtnNoNormal = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBtnNoFg).
				Background(clrBtnNoBg).
				Border(lipgloss.HiddenBorder()).
				Padding(0, 2)
)

// ── Warning ───────────────────────────────────────────────────────────────────

var WarningStyle = lipgloss.NewStyle().
	Foreground(clrWarning).
	Bold(true)

// ── Progress pipeline ─────────────────────────────────────────────────────────

var (
	ProgressTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrAccent).
				MarginBottom(1)

	ProgressDoneStyle = lipgloss.NewStyle().
				Foreground(clrProgressDone).
				Bold(true)

	ProgressCurrentStyle = lipgloss.NewStyle().
				Foreground(clrProgressCur).
				Bold(true)

	ProgressPendingStyle = lipgloss.NewStyle().
				Foreground(clrProgressDim)

	ProgressErrorStyle = lipgloss.NewStyle().
				Foreground(clrRed).
				Bold(true)

	ProgressBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrAccent).
			Padding(1, 4)
)
