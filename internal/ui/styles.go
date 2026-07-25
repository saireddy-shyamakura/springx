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
	clrSearchBg    = lipgloss.Color("236") // dark       — active search background
	clrHighlight   = lipgloss.Color("220") // gold       — matched search text
	clrPanelBorder = lipgloss.Color("238") // dark-grey  — unfocused panel border
	clrFocusBorder = lipgloss.Color("205") // hot-pink   — focused panel border
	clrSelectedBg  = lipgloss.Color("22")  // dark-green — selected row background
	clrConfirmBg   = lipgloss.Color("17")  // dark-blue  — confirmation screen bg
	clrProgressDim = lipgloss.Color("241") // grey       — pending progress steps
	clrProgressDone= lipgloss.Color("42")  // green      — completed progress steps
	clrProgressCur = lipgloss.Color("229") // yellow     — current progress step
	clrResultCount = lipgloss.Color("117") // sky-blue   — "Found N dependencies"
	clrWarning     = lipgloss.Color("214") // orange     — warnings
)

// ── Panel borders ─────────────────────────────────────────────────────────────

var (
	normalBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrPanelBorder)

	focusBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrFocusBorder).
			BorderStyle(lipgloss.ThickBorder()) // extra weight on focused panel
)

// ── Application title bar ─────────────────────────────────────────────────────

var (
	AppTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrAccent).
			PaddingLeft(1)

	// AppVersionStyle is used for the Spring Boot version in the title bar
	// right-hand column.
	AppVersionStyle = lipgloss.NewStyle().
			Foreground(clrDimText).
			PaddingRight(1)

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
				Foreground(clrYellow).
				PaddingLeft(1).
				PaddingRight(1)

	SearchIdleStyle = lipgloss.NewStyle().
				Foreground(clrDimText).
				PaddingLeft(1)

	// SearchingIndicatorStyle is shown while the user is typing in search mode.
	SearchingIndicatorStyle = lipgloss.NewStyle().
				Foreground(clrYellow).
				Bold(true)

	// SearchResultCountStyle shows "Found N dependencies" after filtering.
	SearchResultCountStyle = lipgloss.NewStyle().
				Foreground(clrResultCount).
				Bold(true)

	// SearchNoResultStyle is shown when the filter matches nothing.
	SearchNoResultStyle = lipgloss.NewStyle().
				Foreground(clrRed).
				Bold(true)

	// SearchHintStyle is the "Ctrl+F" hint in the header bar.
	SearchHintStyle = lipgloss.NewStyle().
			Foreground(clrDimText)

	// HighlightMatchStyle wraps matched characters in the dep list.
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
	// DepCursorStyle is the highlighted row under the navigation cursor.
	// Blue background distinguishes it from the purple of the group cursor.
	DepCursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrYellow).
			Background(lipgloss.Color("20")) // dark blue

	// DepCursorSelectedStyle is cursor + selected.
	DepCursorSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrGreen).
				Background(lipgloss.Color("20"))

	DepSelectedStyle = lipgloss.NewStyle().
				Foreground(clrGreen).
				Background(clrSelectedBg)

	DepNormalStyle = lipgloss.NewStyle().
			Foreground(clrNormal)

	DepDescStyle = lipgloss.NewStyle().
			Foreground(clrDimText)

	// StickyHeaderStyle is the pinned group label shown at the top of the dep
	// panel while scrolling.
	StickyHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBlue).
				Background(lipgloss.Color("234")).
				PaddingLeft(1)

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
)

// ── Selected panel ────────────────────────────────────────────────────────────

var (
	SelectedCountStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrAccent)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(clrGreen)

	SelectedGroupStyle = lipgloss.NewStyle().
				Foreground(clrDimText).
				Italic(true)

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

	HelpSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrBlue).
				MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrYellow).
			Width(22)

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
			Foreground(clrDimText).
			PaddingLeft(1).
			PaddingRight(1)

	FooterKeyStyle = lipgloss.NewStyle().
			Background(clrStatusBg).
			Foreground(clrNormal).
			Bold(true)

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

	ConfirmPromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(clrYellow).
				MarginTop(1)

	ConfirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrAccent).
			Padding(1, 3)
)

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

// ── Panel titles ──────────────────────────────────────────────────────────────

var PanelTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(clrBlue).
	PaddingLeft(1)

// ── Warning ───────────────────────────────────────────────────────────────────

var WarningStyle = lipgloss.NewStyle().
	Foreground(clrWarning).
	Bold(true)
