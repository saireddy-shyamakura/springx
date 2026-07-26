package ui

import (
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
)

// ── Row types ─────────────────────────────────────────────────────────────────

// ItemType demarcates whether a row in the list is a Group Header or a
// Dependency Item.
type ItemType int

const (
	TypeHeader ItemType = iota
	TypeDependency
)

// ListRow represents a flattened row in the dependency view (header or dep).
type ListRow struct {
	Type        ItemType
	GroupName   string
	ID          string
	Name        string
	Description string
}

// SelectedItem carries the display data for a single confirmed selection,
// including its group so the selected-panel can show the group label.
type SelectedItem struct {
	ID        string
	Name      string
	GroupName string
}

// ── Search highlighting ───────────────────────────────────────────────────────

// MatchRange marks a [Start, End) byte range within a string as a search hit.
type MatchRange struct {
	Start int
	End   int
}

// MatchRanges returns the byte ranges within s where query appears
// (case-insensitive). Returns nil when query is empty or not found.
func MatchRanges(s, query string) []MatchRange {
	if query == "" {
		return nil
	}
	sl := strings.ToLower(s)
	ql := strings.ToLower(query)
	var ranges []MatchRange
	start := 0
	for {
		idx := strings.Index(sl[start:], ql)
		if idx < 0 {
			break
		}
		abs := start + idx
		ranges = append(ranges, MatchRange{Start: abs, End: abs + len(ql)})
		start = abs + len(ql)
	}
	return ranges
}

// HighlightMatches returns s with every occurrence of query wrapped in
// HighlightMatchStyle. The surrounding text is left unstyled so the caller's
// panel style applies to it naturally.
func HighlightMatches(s, query string) string {
	ranges := MatchRanges(s, query)
	if len(ranges) == 0 {
		return s
	}
	var sb strings.Builder
	prev := 0
	for _, r := range ranges {
		if r.Start > prev {
			sb.WriteString(s[prev:r.Start])
		}
		sb.WriteString(HighlightMatchStyle.Render(s[r.Start:r.End]))
		prev = r.End
	}
	if prev < len(s) {
		sb.WriteString(s[prev:])
	}
	return sb.String()
}

// ── PickerState ───────────────────────────────────────────────────────────────

// PickerState holds pure business logic and state for dependency selection &
// filtering. Every method here is testable without a TUI renderer.
type PickerState struct {
	// Immutable after construction.
	AllRows    []ListRow // all rows built from metadata (headers + deps)
	groupNames []string  // ordered unique group names

	// Mutable during interaction.
	FilteredRows  []ListRow       // visible rows after search
	SelectableIdx []int           // indices into FilteredRows that are TypeDependency
	Cursor        int             // position within SelectableIdx (-1 = none)
	Selected      map[string]bool // depID → true
	SearchQuery   string

	// Pre-search cursor saved so ESC restores exact position.
	preCursor    int
	preGroupIdx  int
	searchActive bool // true while a search is in effect

	// Group navigation.
	activeGroupIdx int // index into groupNames
}

// NewPickerState builds a PickerState from metadata and optional pre-selected IDs.
func NewPickerState(meta *metadata.Metadata, preSelected []string) *PickerState {
	ps := &PickerState{
		Selected:    make(map[string]bool),
		preCursor:   0,
		preGroupIdx: 0,
	}
	for _, id := range preSelected {
		ps.Selected[id] = true
	}
	if meta != nil {
		seen := map[string]bool{}
		for _, group := range meta.Dependencies.Values {
			ps.AllRows = append(ps.AllRows, ListRow{
				Type:      TypeHeader,
				GroupName: group.Name,
			})
			if !seen[group.Name] {
				ps.groupNames = append(ps.groupNames, group.Name)
				seen[group.Name] = true
			}
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

// ── Filtering ─────────────────────────────────────────────────────────────────

// ApplyFilter updates FilteredRows to only include rows matching query
// (case-insensitive). Group headers are included only when they have at least
// one matching child. Cursor is clamped after the filter changes.
func (ps *PickerState) ApplyFilter(query string) {
	ps.SearchQuery = query
	q := strings.ToLower(strings.TrimSpace(query))

	if q == "" {
		ps.FilteredRows = make([]ListRow, len(ps.AllRows))
		copy(ps.FilteredRows, ps.AllRows)
	} else {
		var filtered []ListRow
		var pendingHeader *ListRow
		for _, row := range ps.AllRows {
			if row.Type == TypeHeader {
				pendingHeader = &ListRow{Type: TypeHeader, GroupName: row.GroupName}
			} else if matchesQuery(row, q) {
				if pendingHeader != nil {
					filtered = append(filtered, *pendingHeader)
					pendingHeader = nil
				}
				filtered = append(filtered, row)
			}
		}
		ps.FilteredRows = filtered
	}

	// Rebuild selectable index.
	ps.SelectableIdx = nil
	for i, row := range ps.FilteredRows {
		if row.Type == TypeDependency {
			ps.SelectableIdx = append(ps.SelectableIdx, i)
		}
	}

	// Clamp cursor.
	n := len(ps.SelectableIdx)
	switch {
	case n == 0:
		ps.Cursor = -1
	case ps.Cursor >= n:
		ps.Cursor = n - 1
	case ps.Cursor < 0:
		ps.Cursor = 0
	}
}

// BeginSearch saves the current cursor position so it can be restored when
// search is cleared. Should be called exactly once when search mode starts.
func (ps *PickerState) BeginSearch() {
	if !ps.searchActive {
		ps.preCursor = ps.Cursor
		ps.preGroupIdx = ps.activeGroupIdx
		ps.searchActive = true
	}
}

// ClearSearch exits search mode, clears the filter, and restores the cursor
// to the position it was at before search began.
func (ps *PickerState) ClearSearch() {
	ps.searchActive = false
	ps.SearchQuery = ""
	ps.ApplyFilter("")
	// Restore pre-search position (clamped to new list bounds).
	n := len(ps.SelectableIdx)
	if n == 0 {
		ps.Cursor = -1
		return
	}
	ps.Cursor = clamp(ps.preCursor, 0, n-1)
	ps.activeGroupIdx = clamp(ps.preGroupIdx, 0, len(ps.groupNames)-1)
}

// IsSearchActive reports whether the user has an active filter applied.
func (ps *PickerState) IsSearchActive() bool {
	return ps.searchActive
}

// MatchCount returns the number of selectable dependencies in the current
// filtered view. Used to display "Found N dependencies".
func (ps *PickerState) MatchCount() int {
	return len(ps.SelectableIdx)
}

func matchesQuery(row ListRow, q string) bool {
	return strings.Contains(strings.ToLower(row.Name), q) ||
		strings.Contains(strings.ToLower(row.ID), q) ||
		strings.Contains(strings.ToLower(row.Description), q) ||
		strings.Contains(strings.ToLower(row.GroupName), q)
}

// ── Cursor movement ───────────────────────────────────────────────────────────

const pageSize = 8 // rows moved by PageUp / PageDown

// MoveCursor shifts the cursor by delta (negative = up, positive = down).
func (ps *PickerState) MoveCursor(delta int) {
	if len(ps.SelectableIdx) == 0 {
		ps.Cursor = -1
		return
	}
	ps.Cursor = clamp(ps.Cursor+delta, 0, len(ps.SelectableIdx)-1)
	ps.syncActiveGroupToCursor()
}

// MoveToFirst moves the cursor to the first selectable row.
func (ps *PickerState) MoveToFirst() {
	if len(ps.SelectableIdx) == 0 {
		return
	}
	ps.Cursor = 0
	ps.syncActiveGroupToCursor()
}

// MoveToLast moves the cursor to the last selectable row.
func (ps *PickerState) MoveToLast() {
	if len(ps.SelectableIdx) == 0 {
		return
	}
	ps.Cursor = len(ps.SelectableIdx) - 1
	ps.syncActiveGroupToCursor()
}

// PageUp moves the cursor up by pageSize rows.
func (ps *PickerState) PageUp() { ps.MoveCursor(-pageSize) }

// PageDown moves the cursor down by pageSize rows.
func (ps *PickerState) PageDown() { ps.MoveCursor(pageSize) }

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (ps *PickerState) syncActiveGroupToCursor() {
	if ps.Cursor < 0 || ps.Cursor >= len(ps.SelectableIdx) {
		return
	}
	groupName := ps.FilteredRows[ps.SelectableIdx[ps.Cursor]].GroupName
	for i, g := range ps.groupNames {
		if g == groupName {
			ps.activeGroupIdx = i
			return
		}
	}
}

// ── Selection ─────────────────────────────────────────────────────────────────

// ToggleCurrent toggles the dependency under the cursor.
func (ps *PickerState) ToggleCurrent() {
	if len(ps.SelectableIdx) == 0 || ps.Cursor < 0 || ps.Cursor >= len(ps.SelectableIdx) {
		return
	}
	row := ps.FilteredRows[ps.SelectableIdx[ps.Cursor]]
	if row.Type != TypeDependency {
		return
	}
	if ps.Selected[row.ID] {
		delete(ps.Selected, row.ID)
	} else {
		ps.Selected[row.ID] = true
	}
}

// GetSelectedIDs returns selected dependency IDs in original metadata order.
func (ps *PickerState) GetSelectedIDs() []string {
	var ids []string
	for _, row := range ps.AllRows {
		if row.Type == TypeDependency && ps.Selected[row.ID] {
			ids = append(ids, row.ID)
		}
	}
	return ids
}

// GetSelectedNames returns the display names of selected dependencies in
// original metadata order.
func (ps *PickerState) GetSelectedNames() []string {
	var names []string
	for _, row := range ps.AllRows {
		if row.Type == TypeDependency && ps.Selected[row.ID] {
			names = append(names, row.Name)
		}
	}
	return names
}

// GetSelectedItems returns rich SelectedItem values (ID, Name, GroupName) in
// original metadata order. Used by the confirmation screen and selected panel.
func (ps *PickerState) GetSelectedItems() []SelectedItem {
	var items []SelectedItem
	for _, row := range ps.AllRows {
		if row.Type == TypeDependency && ps.Selected[row.ID] {
			items = append(items, SelectedItem{
				ID:        row.ID,
				Name:      row.Name,
				GroupName: row.GroupName,
			})
		}
	}
	return items
}

// GetSelectedByGroup returns selected items keyed by group name, preserving
// the original metadata group order. Only groups that have at least one
// selected item are included. Used by the selected panel and confirmation screen.
func (ps *PickerState) GetSelectedByGroup() []SelectedGroup {
	groupOrder := make([]string, 0)
	seen := make(map[string]bool)
	byGroup := make(map[string][]SelectedItem)

	for _, row := range ps.AllRows {
		if row.Type == TypeDependency && ps.Selected[row.ID] {
			if !seen[row.GroupName] {
				groupOrder = append(groupOrder, row.GroupName)
				seen[row.GroupName] = true
			}
			byGroup[row.GroupName] = append(byGroup[row.GroupName], SelectedItem{
				ID:        row.ID,
				Name:      row.Name,
				GroupName: row.GroupName,
			})
		}
	}

	result := make([]SelectedGroup, 0, len(groupOrder))
	for _, g := range groupOrder {
		result = append(result, SelectedGroup{
			Name:  g,
			Items: byGroup[g],
		})
	}
	return result
}

// SelectedGroup bundles a group name with its selected dependencies.
type SelectedGroup struct {
	Name  string
	Items []SelectedItem
}

// SelectedCount returns the number of currently selected dependencies.
func (ps *PickerState) SelectedCount() int {
	return len(ps.Selected)
}

// ── Group navigation ──────────────────────────────────────────────────────────

// GetGroupNames returns the ordered list of group names from metadata.
func (ps *PickerState) GetGroupNames() []string { return ps.groupNames }

// ActiveGroupIdx returns the index of the currently active group.
func (ps *PickerState) ActiveGroupIdx() int { return ps.activeGroupIdx }

// TabToNextGroup advances to the next group, wrapping around.
func (ps *PickerState) TabToNextGroup() {
	if len(ps.groupNames) == 0 {
		return
	}
	ps.activeGroupIdx = (ps.activeGroupIdx + 1) % len(ps.groupNames)
	ps.jumpCursorToGroup(ps.groupNames[ps.activeGroupIdx])
}

// TabToPrevGroup retreats to the previous group, wrapping around.
func (ps *PickerState) TabToPrevGroup() {
	if len(ps.groupNames) == 0 {
		return
	}
	ps.activeGroupIdx = (ps.activeGroupIdx - 1 + len(ps.groupNames)) % len(ps.groupNames)
	ps.jumpCursorToGroup(ps.groupNames[ps.activeGroupIdx])
}

func (ps *PickerState) jumpCursorToGroup(groupName string) {
	for i, idx := range ps.SelectableIdx {
		if ps.FilteredRows[idx].GroupName == groupName {
			ps.Cursor = i
			return
		}
	}
}

// VisibleGroupNames returns only the group names that have rows in the current
// filtered view.
func (ps *PickerState) VisibleGroupNames() []string {
	seen := map[string]bool{}
	var names []string
	for _, row := range ps.FilteredRows {
		if row.GroupName != "" && !seen[row.GroupName] {
			seen[row.GroupName] = true
			names = append(names, row.GroupName)
		}
	}
	return names
}

// StickyGroupHeader returns the group name that should be shown as a "pinned"
// header at the top of the deps panel for the given scrollOffset. It is the
// group of the first visible header row at or before scrollOffset, so the
// user always knows which group they are browsing.
func (ps *PickerState) StickyGroupHeader(scrollOffset int) string {
	sticky := ""
	for i := 0; i <= scrollOffset && i < len(ps.FilteredRows); i++ {
		if ps.FilteredRows[i].Type == TypeHeader {
			sticky = ps.FilteredRows[i].GroupName
		}
	}
	return sticky
}
