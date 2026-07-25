package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/saireddy-shyamakura/springx/internal/metadata"
)

// ItemType demarcates whether a row in the list is a Group Header or a Dependency Item.
type ItemType int

const (
	TypeHeader ItemType = iota
	TypeDependency
)

// ListRow represents a flattened row in the dependency view (header or dependency).
type ListRow struct {
	Type        ItemType
	GroupName   string
	ID          string
	Name        string
	Description string
}

// PickerState holds pure business logic and state for dependency selection & filtering.
// Keeping this outside the Bubble Tea model allows 100% test coverage without TUI rendering.
type PickerState struct {
	AllRows       []ListRow       // All original rows (headers + dependencies)
	FilteredRows  []ListRow       // Currently visible rows based on search
	SelectableIdx []int           // Indices in FilteredRows that point to TypeDependency
	Cursor        int             // Index in SelectableIdx
	Selected      map[string]bool // Map of selected dependency IDs -> true
	SearchQuery   string
}

// NewPickerState constructs a new state container initialized from metadata and optional pre-selected dependency IDs.
func NewPickerState(meta *metadata.Metadata, preSelected []string) *PickerState {
	ps := &PickerState{
		Selected: make(map[string]bool),
	}

	for _, id := range preSelected {
		ps.Selected[id] = true
	}

	if meta != nil {
		for _, group := range meta.Dependencies.Values {
			ps.AllRows = append(ps.AllRows, ListRow{
				Type:      TypeHeader,
				GroupName: group.Name,
			})
			for _, dep := range group.Values {
				ps.AllRows = append(ps.AllRows, ListRow{
					Type:        TypeDependency,
					GroupName:   group.Name,
					ID:          dep.ID,
					Name:        dep.Name,
					Description: dep.Description,
				})
			}
		}
	}

	ps.ApplyFilter("")
	return ps
}

// ApplyFilter filters the displayed rows based on the search query.
func (ps *PickerState) ApplyFilter(query string) {
	ps.SearchQuery = query
	q := strings.ToLower(strings.TrimSpace(query))

	if q == "" {
		ps.FilteredRows = make([]ListRow, len(ps.AllRows))
		copy(ps.FilteredRows, ps.AllRows)
	} else {
		var filtered []ListRow
		var currentGroupRow *ListRow
		hasDepInGroup := false

		for _, row := range ps.AllRows {
			if row.Type == TypeHeader {
				currentGroupRow = &ListRow{
					Type:      TypeHeader,
					GroupName: row.GroupName,
				}
				hasDepInGroup = false
			} else {
				match := strings.Contains(strings.ToLower(row.Name), q) ||
					strings.Contains(strings.ToLower(row.ID), q) ||
					strings.Contains(strings.ToLower(row.Description), q) ||
					strings.Contains(strings.ToLower(row.GroupName), q)

				if match {
					if !hasDepInGroup && currentGroupRow != nil {
						filtered = append(filtered, *currentGroupRow)
						hasDepInGroup = true
					}
					filtered = append(filtered, row)
				}
			}
		}
		ps.FilteredRows = filtered
	}

	// Rebuild SelectableIdx
	ps.SelectableIdx = nil
	for i, row := range ps.FilteredRows {
		if row.Type == TypeDependency {
			ps.SelectableIdx = append(ps.SelectableIdx, i)
		}
	}

	if len(ps.SelectableIdx) == 0 {
		ps.Cursor = -1
	} else {
		if ps.Cursor >= len(ps.SelectableIdx) {
			ps.Cursor = len(ps.SelectableIdx) - 1
		}
		if ps.Cursor < 0 {
			ps.Cursor = 0
		}
	}
}

// MoveCursor moves the selection cursor up (negative delta) or down (positive delta).
func (ps *PickerState) MoveCursor(delta int) {
	if len(ps.SelectableIdx) == 0 {
		ps.Cursor = -1
		return
	}
	ps.Cursor += delta
	if ps.Cursor < 0 {
		ps.Cursor = 0
	}
	if ps.Cursor >= len(ps.SelectableIdx) {
		ps.Cursor = len(ps.SelectableIdx) - 1
	}
}

// ToggleCurrent toggles the selection status of the dependency under the cursor.
func (ps *PickerState) ToggleCurrent() {
	if len(ps.SelectableIdx) == 0 || ps.Cursor < 0 || ps.Cursor >= len(ps.SelectableIdx) {
		return
	}
	rowIdx := ps.SelectableIdx[ps.Cursor]
	row := ps.FilteredRows[rowIdx]
	if row.Type == TypeDependency {
		if ps.Selected[row.ID] {
			delete(ps.Selected, row.ID)
		} else {
			ps.Selected[row.ID] = true
		}
	}
}

// GetSelectedIDs returns the list of selected dependency IDs preserving original metadata order.
func (ps *PickerState) GetSelectedIDs() []string {
	var ids []string
	for _, row := range ps.AllRows {
		if row.Type == TypeDependency && ps.Selected[row.ID] {
			ids = append(ids, row.ID)
		}
	}
	return ids
}

// SelectedCount returns the total number of currently selected dependencies.
func (ps *PickerState) SelectedCount() int {
	return len(ps.Selected)
}

// Lipgloss Styles for UI Rendering
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginLeft(1).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginTop(1)

	ruleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	selectedCheckbox = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true).
				Render("[x]")

	unselectedCheckbox = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render("[ ]")

	cursorItemStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57"))

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	footerStyle = lipgloss.NewStyle().
			MarginTop(1).
			Foreground(lipgloss.Color("241"))
)

// Model represents the Bubble Tea TUI state.
type model struct {
	state       *PickerState
	searchInput textinput.Model
	isSearching bool
	width       int
	height      int
	confirmed   bool
	canceled    bool
}

func newModel(meta *metadata.Metadata, preSelected []string) model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter dependencies..."
	ti.Prompt = "/ "
	ti.CharLimit = 50

	return model{
		state:       NewPickerState(meta, preSelected),
		searchInput: ti,
		width:       80,
		height:      24,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.isSearching {
			switch msg.String() {
			case "esc":
				m.isSearching = false
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.state.ApplyFilter("")
				return m, nil
			case "enter":
				m.isSearching = false
				m.searchInput.Blur()
				return m, nil
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.state.ApplyFilter(m.searchInput.Value())
				return m, cmd
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				m.canceled = true
				return m, tea.Quit
			case "enter":
				m.confirmed = true
				return m, tea.Quit
			case "up", "k":
				m.state.MoveCursor(-1)
			case "down", "j":
				m.state.MoveCursor(1)
			case " ":
				m.state.ToggleCurrent()
			case "/":
				m.isSearching = true
				m.searchInput.Focus()
				return m, textinput.Blink
			case "esc":
				if m.state.SearchQuery != "" {
					m.searchInput.SetValue("")
					m.state.ApplyFilter("")
				}
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	var sb strings.Builder

	// Title Bar
	title := fmt.Sprintf("Spring Boot Dependency Picker (%d selected)", m.state.SelectedCount())
	sb.WriteString(titleStyle.Render(title) + "\n")

	// Search Bar
	if m.isSearching {
		sb.WriteString(m.searchInput.View() + "\n\n")
	} else if m.state.SearchQuery != "" {
		sb.WriteString(dimStyle.Render(fmt.Sprintf("Filtered by: %q (press esc to clear, / to search)\n\n", m.state.SearchQuery)))
	} else {
		sb.WriteString(dimStyle.Render("Press '/' to search dependencies\n\n"))
	}

	if len(m.state.FilteredRows) == 0 {
		sb.WriteString(dimStyle.Render("  No dependencies found matching search query.\n"))
	} else {
		// Calculate row index under cursor
		activeRowIdx := -1
		if m.state.Cursor >= 0 && m.state.Cursor < len(m.state.SelectableIdx) {
			activeRowIdx = m.state.SelectableIdx[m.state.Cursor]
		}

		for i, row := range m.state.FilteredRows {
			if row.Type == TypeHeader {
				sb.WriteString(headerStyle.Render(row.GroupName) + "\n")
				sb.WriteString(ruleStyle.Render("-----------------") + "\n")
			} else {
				chk := unselectedCheckbox
				if m.state.Selected[row.ID] {
					chk = selectedCheckbox
				}

				itemText := fmt.Sprintf(" %s %s", chk, row.Name)
				if row.Description != "" {
					itemText += dimStyle.Render(" - " + row.Description)
				}

				if i == activeRowIdx {
					sb.WriteString(cursorItemStyle.Render("> "+itemText) + "\n")
				} else {
					sb.WriteString("  " + normalItemStyle.Render(itemText) + "\n")
				}
			}
		}
	}

	// Footer Controls
	footer := "Controls: ↑/↓/k/j: navigate • space: select • /: search • enter: confirm • q: cancel"
	sb.WriteString(footerStyle.Render(footer))

	return sb.String()
}

// RunDependencyPicker executes the Bubble Tea TUI dependency selector with initial pre-selected dependency IDs.
func RunDependencyPicker(meta *metadata.Metadata, preSelected []string) ([]string, error) {
	m := newModel(meta, preSelected)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run dependency picker TUI: %w", err)
	}

	fm, ok := finalModel.(model)
	if !ok || fm.canceled {
		return nil, fmt.Errorf("dependency selection canceled")
	}

	return fm.state.GetSelectedIDs(), nil
}
