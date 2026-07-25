package ui

import (
	"strings"

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

// MatchRange marks a [Start, End) byte range within a string as a search match.
type MatchRange struct {
	Start int
	End   int
}

// PickerState holds pure business logic and state for dependency selection &
// filtering. Keeping this entirely outside the Bubble Tea model means every
// method here is 100 % testable without TUI rendering.
type PickerState struct {
	AllRows      []ListRow       // every row built from metadata (headers + deps)
	FilteredRows []ListRow       // currently visible rows after search filter
	SelectableIdx []int          // indices into FilteredRows that are TypeDependency
	Cursor       int             // position within SelectableIdx
	Selected     map[string]bool // depID → true
	SearchQuery  string

	// Group navigation
	groupNames      []string // ordered unique group names from AllRows
	activeGroupIdx  int      // index into groupNames for the highlighted group panel row
}

// NewPickerState constructs a PickerState from metadata and optional pre-selected IDs.
func NewPickerState(meta *metadata.Metadata, preSelected []string) *PickerState {
	ps := &PickerState{
		Selected: make(map[string]bool),
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

// ─── Filtering ────────────────────────────────────────────────────────────────

// ApplyFilter filters FilteredRows to rows matching query (case-insensitive).
// Group headers are included only when they have at least one matching child.
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
				pendingHeader = &ListRow{
					Type:      TypeHeader,
					GroupName: row.GroupName,
				}
			} else {
				if matchesQuery(row, q) {
					if pendingHeader != nil {
						filtered = append(filtered, *pendingHeader)
						pendingHeader = nil
					}
					filtered = append(filtered, row)
				}
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

// matchesQuery returns true if the row's Name, ID, Description, or GroupName
// contains q (already lowercased).
func matchesQuery(row ListRow, q string) bool {
	return strings.Contains(strings.ToLower(row.Name), q) ||
		strings.Contains(strings.ToLower(row.ID), q) ||
		strings.Contains(strings.ToLower(row.Description), q) ||
		strings.Contains(strings.ToLower(row.GroupName), q)
}

// ─── Search highlighting ──────────────────────────────────────────────────────

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
// HighlightMatchStyle, and the rest of the string in normalStyle.
// normalStyle is passed by the caller so this function stays style-agnostic.
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

// ─── Cursor movement ──────────────────────────────────────────────────────────

// MoveCursor shifts the cursor by delta rows (negative = up, positive = down).
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

	// Keep the active group in sync with the cursor position.
	ps.syncActiveGroupToCursor()
}

// syncActiveGroupToCursor updates activeGroupIdx to match the group of the
// dependency currently under the cursor.
func (ps *PickerState) syncActiveGroupToCursor() {
	if ps.Cursor < 0 || ps.Cursor >= len(ps.SelectableIdx) {
		return
	}
	rowIdx := ps.SelectableIdx[ps.Cursor]
	groupName := ps.FilteredRows[rowIdx].GroupName
	for i, g := range ps.groupNames {
		if g == groupName {
			ps.activeGroupIdx = i
			return
		}
	}
}

// ─── Selection ────────────────────────────────────────────────────────────────

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

// SelectedCount returns the number of currently selected dependencies.
func (ps *PickerState) SelectedCount() int {
	return len(ps.Selected)
}

// ─── Group navigation ─────────────────────────────────────────────────────────

// GetGroupNames returns the ordered list of group names derived from metadata.
func (ps *PickerState) GetGroupNames() []string {
	return ps.groupNames
}

// ActiveGroupIdx returns the index of the currently active group.
func (ps *PickerState) ActiveGroupIdx() int {
	return ps.activeGroupIdx
}

// TabToNextGroup advances activeGroupIdx to the next group and moves the
// cursor to the first dependency in that group.
func (ps *PickerState) TabToNextGroup() {
	if len(ps.groupNames) == 0 {
		return
	}
	ps.activeGroupIdx = (ps.activeGroupIdx + 1) % len(ps.groupNames)
	ps.jumpCursorToGroup(ps.groupNames[ps.activeGroupIdx])
}

// TabToPrevGroup retreats activeGroupIdx to the previous group.
func (ps *PickerState) TabToPrevGroup() {
	if len(ps.groupNames) == 0 {
		return
	}
	ps.activeGroupIdx = (ps.activeGroupIdx - 1 + len(ps.groupNames)) % len(ps.groupNames)
	ps.jumpCursorToGroup(ps.groupNames[ps.activeGroupIdx])
}

// jumpCursorToGroup moves the dependency cursor to the first item in groupName
// within the current filtered view.
func (ps *PickerState) jumpCursorToGroup(groupName string) {
	for i, idx := range ps.SelectableIdx {
		if ps.FilteredRows[idx].GroupName == groupName {
			ps.Cursor = i
			return
		}
	}
	// Group not visible in current filter — leave cursor where it is.
}

// VisibleGroupNames returns only the group names that have at least one row
// in the current filtered view.
func (ps *PickerState) VisibleGroupNames() []string {
	seen := map[string]bool{}
	var names []string
	for _, row := range ps.FilteredRows {
		if !seen[row.GroupName] && row.GroupName != "" {
			seen[row.GroupName] = true
			names = append(names, row.GroupName)
		}
	}
	return names
}
