package ui_test

import (
	"encoding/json"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/ui"
)

// ── Test fixtures ─────────────────────────────────────────────────────────────

const mockMetadataJSON = `{
  "dependencies": {
    "type": "hierarchical-multi-select",
    "values": [
      {
        "name": "Developer Tools",
        "values": [
          {"id": "lombok",   "name": "Lombok",              "description": "Reduce boilerplate code"},
          {"id": "devtools", "name": "Spring Boot DevTools", "description": "Fast application restarts"}
        ]
      },
      {
        "name": "Web",
        "values": [
          {"id": "web",     "name": "Spring Web",          "description": "Build web applications"},
          {"id": "graphql", "name": "Spring for GraphQL",  "description": "GraphQL support"}
        ]
      },
      {
        "name": "Data",
        "values": [
          {"id": "data-jpa",    "name": "Spring Data JPA",    "description": "Persist data in SQL"},
          {"id": "postgresql",  "name": "PostgreSQL Driver",  "description": "JDBC driver for PostgreSQL"}
        ]
      }
    ]
  }
}`

func getMockMetadata(t *testing.T) *metadata.Metadata {
	t.Helper()
	var meta metadata.Metadata
	if err := json.Unmarshal([]byte(mockMetadataJSON), &meta); err != nil {
		t.Fatalf("failed to unmarshal mock metadata: %v", err)
	}
	return &meta
}

// ── Navigation & selection ────────────────────────────────────────────────────

func TestPickerState_NavigationAndSelection(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)

	if ps.Cursor != 0 {
		t.Errorf("initial cursor: want 0, got %d", ps.Cursor)
	}

	// Toggle Lombok (cursor 0).
	ps.ToggleCurrent()
	if !ps.Selected["lombok"] {
		t.Error("lombok should be selected after toggle")
	}

	// Move down 5 positions: devtools → web → graphql → data-jpa → postgresql
	ps.MoveCursor(5)
	ps.ToggleCurrent()
	if !ps.Selected["postgresql"] {
		t.Error("postgresql should be selected after toggle")
	}

	// Clamp at last selectable (index 5).
	ps.MoveCursor(100)
	if ps.Cursor != 5 {
		t.Errorf("cursor should clamp to 5, got %d", ps.Cursor)
	}

	// Clamp at first selectable (index 0).
	ps.MoveCursor(-100)
	if ps.Cursor != 0 {
		t.Errorf("cursor should clamp to 0, got %d", ps.Cursor)
	}

	// Deselect Lombok.
	ps.ToggleCurrent()
	if ps.Selected["lombok"] {
		t.Error("lombok should be deselected after second toggle")
	}

	ids := ps.GetSelectedIDs()
	if len(ids) != 1 || ids[0] != "postgresql" {
		t.Errorf("expected [postgresql], got %v", ids)
	}
}

// ── Search filtering ──────────────────────────────────────────────────────────

func TestPickerState_SearchFiltering(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)

	ps.ApplyFilter("post")

	if len(ps.SelectableIdx) != 1 {
		t.Fatalf("expected 1 match for 'post', got %d", len(ps.SelectableIdx))
	}

	ps.ToggleCurrent()
	if !ps.Selected["postgresql"] {
		t.Error("postgresql should be selected")
	}

	ps.ApplyFilter("")
	if len(ps.SelectableIdx) != 6 {
		t.Errorf("expected 6 selectable items after clear, got %d", len(ps.SelectableIdx))
	}

	ids := ps.GetSelectedIDs()
	if len(ids) != 1 || ids[0] != "postgresql" {
		t.Errorf("selection should persist after filter clear, got %v", ids)
	}
}

func TestPickerState_SearchEmpty(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("zzz-no-match")

	if len(ps.SelectableIdx) != 0 {
		t.Errorf("expected 0 matches for non-existent query, got %d", len(ps.SelectableIdx))
	}
	if ps.Cursor != -1 {
		t.Errorf("cursor should be -1 when no rows match, got %d", ps.Cursor)
	}
}

func TestPickerState_ToggleNoOp_WhenEmpty(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("zzz-no-match")

	// Should not panic.
	ps.ToggleCurrent()
	if ps.SelectedCount() != 0 {
		t.Error("ToggleCurrent on empty filtered list should be a no-op")
	}
}

// ── Pre-selected ──────────────────────────────────────────────────────────────

func TestPickerState_PreSelected(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), []string{"web", "postgresql"})

	if !ps.Selected["web"] {
		t.Error("web should be pre-selected")
	}
	if !ps.Selected["postgresql"] {
		t.Error("postgresql should be pre-selected")
	}
	if ps.Selected["lombok"] {
		t.Error("lombok should NOT be pre-selected")
	}

	ids := ps.GetSelectedIDs()
	if len(ids) != 2 || ids[0] != "web" || ids[1] != "postgresql" {
		t.Errorf("expected [web postgresql] in metadata order, got %v", ids)
	}
}

// ── Group navigation ──────────────────────────────────────────────────────────

func TestPickerState_GetGroupNames(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	names := ps.GetGroupNames()

	want := []string{"Developer Tools", "Web", "Data"}
	if len(names) != len(want) {
		t.Fatalf("expected %d groups, got %d: %v", len(want), len(names), names)
	}
	for i, w := range want {
		if names[i] != w {
			t.Errorf("group[%d]: want %q, got %q", i, w, names[i])
		}
	}
}

func TestPickerState_TabToNextGroup(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)

	// Start in group 0 (Developer Tools).
	if ps.ActiveGroupIdx() != 0 {
		t.Errorf("initial group should be 0, got %d", ps.ActiveGroupIdx())
	}

	ps.TabToNextGroup()
	if ps.ActiveGroupIdx() != 1 {
		t.Errorf("after 1 tab, active group should be 1, got %d", ps.ActiveGroupIdx())
	}
	// Cursor should be on first dep in "Web" group (index 2 in selectables).
	cursorRow := ps.FilteredRows[ps.SelectableIdx[ps.Cursor]]
	if cursorRow.GroupName != "Web" {
		t.Errorf("cursor should be in Web group, got %q", cursorRow.GroupName)
	}

	ps.TabToNextGroup()
	ps.TabToNextGroup() // wraps back to group 0
	if ps.ActiveGroupIdx() != 0 {
		t.Errorf("after wrapping, active group should be 0, got %d", ps.ActiveGroupIdx())
	}
}

func TestPickerState_TabToPrevGroup(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)

	// Wraps from 0 → last group.
	ps.TabToPrevGroup()
	groups := ps.GetGroupNames()
	if ps.ActiveGroupIdx() != len(groups)-1 {
		t.Errorf("TabToPrevGroup from 0 should wrap to %d, got %d", len(groups)-1, ps.ActiveGroupIdx())
	}
}

func TestPickerState_CursorSyncsToActiveGroup(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)

	// Move cursor into "Data" group (indices 4 and 5 among selectables).
	ps.MoveCursor(4) // data-jpa
	if ps.ActiveGroupIdx() != 2 {
		t.Errorf("moving cursor into Data should sync activeGroupIdx to 2, got %d", ps.ActiveGroupIdx())
	}
}

// ── VisibleGroupNames ─────────────────────────────────────────────────────────

func TestPickerState_VisibleGroupNames(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("lombok") // only Developer Tools should be visible

	visible := ps.VisibleGroupNames()
	if len(visible) != 1 || visible[0] != "Developer Tools" {
		t.Errorf("expected [Developer Tools], got %v", visible)
	}
}

// ── MatchRanges ───────────────────────────────────────────────────────────────

func TestMatchRanges_EmptyQuery(t *testing.T) {
	ranges := ui.MatchRanges("Spring Web", "")
	if ranges != nil {
		t.Errorf("expected nil for empty query, got %v", ranges)
	}
}

func TestMatchRanges_NoMatch(t *testing.T) {
	ranges := ui.MatchRanges("Spring Web", "kafka")
	if len(ranges) != 0 {
		t.Errorf("expected no matches, got %v", ranges)
	}
}

func TestMatchRanges_SingleMatch(t *testing.T) {
	ranges := ui.MatchRanges("PostgreSQL Driver", "post")
	if len(ranges) != 1 {
		t.Fatalf("expected 1 match, got %d", len(ranges))
	}
	got := "PostgreSQL Driver"[ranges[0].Start:ranges[0].End]
	if got != "Post" {
		t.Errorf("expected matched text %q, got %q", "Post", got)
	}
}

func TestMatchRanges_CaseInsensitive(t *testing.T) {
	ranges := ui.MatchRanges("Spring Data JPA", "spring")
	if len(ranges) != 1 {
		t.Fatalf("expected 1 match, got %d", len(ranges))
	}
	if ranges[0].Start != 0 || ranges[0].End != 6 {
		t.Errorf("expected range [0,6), got [%d,%d)", ranges[0].Start, ranges[0].End)
	}
}

func TestMatchRanges_MultipleMatches(t *testing.T) {
	ranges := ui.MatchRanges("data data data", "data")
	if len(ranges) != 3 {
		t.Errorf("expected 3 matches, got %d", len(ranges))
	}
}

// ── HighlightMatches ──────────────────────────────────────────────────────────

func TestHighlightMatches_NoQuery(t *testing.T) {
	result := ui.HighlightMatches("Spring Web", "")
	if result != "Spring Web" {
		t.Errorf("expected unchanged string, got %q", result)
	}
}

func TestHighlightMatches_ContainsOriginalText(t *testing.T) {
	// After stripping ANSI escape codes the original text should still be present.
	result := ui.HighlightMatches("PostgreSQL Driver", "post")
	// The raw result will contain lipgloss escape codes; we check that the
	// non-highlighted suffix is still present verbatim.
	if !containsUnescaped(result, "greSQL Driver") {
		t.Errorf("highlighted string should still contain the unmatched suffix")
	}
}

// containsUnescaped checks whether plain (non-styled) substring appears in s
// after stripping ANSI codes. We use a simple check: the raw string is a
// superset of the plain text (ANSI codes only add bytes, never remove them).
func containsUnescaped(s, plain string) bool {
	// Strip obvious ESC sequences with a cheap approach: just check that all
	// bytes of plain appear in s in order (subsequence check).
	pi := 0
	for _, b := range []byte(s) {
		if pi < len(plain) && b == plain[pi] {
			pi++
		}
	}
	return pi == len(plain)
}

// ── GetSelectedNames ──────────────────────────────────────────────────────────

func TestPickerState_GetSelectedNames(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), []string{"web", "data-jpa"})
	names := ps.GetSelectedNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "Spring Web" {
		t.Errorf("expected 'Spring Web', got %q", names[0])
	}
	if names[1] != "Spring Data JPA" {
		t.Errorf("expected 'Spring Data JPA', got %q", names[1])
	}
}
