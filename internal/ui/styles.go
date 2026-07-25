package ui

import "github.com/charmbracelet/lipgloss"

// ── Palette ───────────────────────────────────────────────────────────────────
// All colours are chosen to be legible on both dark and light terminals.
// Anything that needs to stand out uses an explicit 256-colour index or an
// adaptive colour pair; neutral chrome uses a mid-grey that disappears on
// neither background.

const (
	clrAccent      = lipgloss.Color("205") // hot-pink  – title, selected items
	clrBlue        = lipgloss.Color("39")  // cyan-blue – group headers
	clrGreen       = lipgloss.Color("42")  // green     – checked checkbox, success
	clrYellow      = lipgloss.Color("229") // yellow    – cursor row foreground
	clrPurple      = lipgloss.Color("57")  // purple    – cursor row background
	clrRed         = lipgloss.Color("196") // red       – error / failure
	clrDimText     = lipgloss.Color("243") // dark-grey – secondary text
	clrBorder      = lipgloss.Color("240") // mid-grey  – borders, rules
	clrNormal      = lipgloss.Color("252") // near-white – normal text
	clrStatusBg    = lipgloss.Color("235") // very dark – status bar background
	clrStatusFg    = lipgloss.Color("250") // light-grey – status bar text
	clrSearchBg    = lipgloss.Color("236") // dark      – search bar background
	clrHighlight   = lipgloss.Color("220") // gold      – matched search text
	clrPanelBorder = lipgloss.Color("238") // panel border
	clrFocusBorder = lipgloss.Color("205") // focused panel border (same as accent)
	clrSelectedBg  = lipgloss.Color("22")  // dark-green background for selected panel rows
)

// ── Border definitions ────────────────────────────────────────────────────────

var (
	normalBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrPanelBorder)

	focusBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrFocusBorder)
)

// ── Title & header ────────────────────────────────────────────────────────────

var (
	AppTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrAccent).
			PaddingLeft(1)

	AppSubtitleStyle = lipgloss.NewStyle().
				Foreground(clrDimText).
				PaddingLeft(1)

	SectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBlue).
				PaddingLeft(1)

	HRuleStyle = lipgloss.NewStyle().
			Foreground(clrBorder)
)

// ── Search bar ────────────────────────────────────────────────────────────────

var (
	SearchLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrAccent).
				PaddingRight(1)

	SearchActiveStyle = lipgloss.NewStyle().
				Background(clrSearchBg).
				Foreground(clrNormal).
				PaddingLeft(1).
				PaddingRight(1)

	SearchIdleStyle = lipgloss.NewStyle().
				Foreground(clrDimText).
				PaddingLeft(1)

	SearchingIndicatorStyle = lipgloss.NewStyle().
				Foreground(clrYellow).
				Bold(true)

	HighlightMatchStyle = lipgloss.NewStyle().
				Foreground(clrHighlight).
				Bold(true).
				Underline(true)
)

// ── Group panel ───────────────────────────────────────────────────────────────

var (
	GroupActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrYellow).
				Background(clrPurple).
				PaddingLeft(1).
				PaddingRight(1)

	GroupNormalStyle = lipgloss.NewStyle().
				Foreground(clrNormal).
				PaddingLeft(1)

	GroupDimStyle = lipgloss.NewStyle().
			Foreground(clrDimText).
			PaddingLeft(1)
)

// ── Dependency panel ──────────────────────────────────────────────────────────

var (
	DepCursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrYellow).
			Background(clrPurple)

	DepSelectedStyle = lipgloss.NewStyle().
				Foreground(clrGreen).
				Background(clrSelectedBg)

	DepNormalStyle = lipgloss.NewStyle().
			Foreground(clrNormal)

	DepDescStyle = lipgloss.NewStyle().
			Foreground(clrDimText)

	CheckboxOnStyle = lipgloss.NewStyle().
			Foreground(clrGreen).
			Bold(true)

	CheckboxOffStyle = lipgloss.NewStyle().
				Foreground(clrBorder)

	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(clrDimText).
			Italic(true).
			PaddingLeft(2).
			PaddingTop(1)
)

// ── Selected panel ────────────────────────────────────────────────────────────

var (
	SelectedCountStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrAccent)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(clrGreen)

	SelectedBulletStyle = lipgloss.NewStyle().
				Foreground(clrGreen).
				Bold(true)
)

// ── Status bar ────────────────────────────────────────────────────────────────

var (
	StatusBarStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrStatusFg).
			PaddingLeft(1).
			PaddingRight(1)

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

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrYellow).
			Width(18)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(clrNormal)

	HelpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrAccent).
			Padding(1, 3)
)

// ── Footer (key hints) ────────────────────────────────────────────────────────

var (
	FooterStyle = lipgloss.NewStyle().
			Foreground(clrDimText).
			PaddingLeft(1)

	FooterKeyStyle = lipgloss.NewStyle().
			Foreground(clrNormal).
			Bold(true)

	FooterSepStyle = lipgloss.NewStyle().
			Foreground(clrBorder)
)

// ── Spinner / loading ─────────────────────────────────────────────────────────

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(clrAccent).
	Bold(true)

// ── Success animation ─────────────────────────────────────────────────────────

var SuccessStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(clrGreen)

// ── Panel titles ──────────────────────────────────────────────────────────────

var PanelTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(clrBlue).
	PaddingLeft(1).
	PaddingBottom(0)
