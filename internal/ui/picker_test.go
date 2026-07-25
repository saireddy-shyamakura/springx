package ui_test

import (
	"encoding/json"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/ui"
)

// ── Fixtures ──────────────────────────────────────────────────────────────────

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
          {"id": "data-jpa",   "name": "Spring Data JPA",   "description": "Persist data in SQL"},
          {"id": "postgresql", "name": "PostgreSQL Driver",  "description": "JDBC driver for PostgreSQL"}
        ]
      }
    ]
  }
}`

func getMockMetadata(t *testing.T) *metadata.Metadata {
	t.Helper()
	var meta metadata.Metadata
	if err := json.Unmarshal([]byte(mockMetadataJSON), &meta); err != nil {
		t.Fatalf("unmarshal mock metadata: %v", err)
	}
	return &meta
}

// ── Navigation & selection ────────────────────────────────────────────────────

func TestPickerState_NavigationAndSelection(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)

	if ps.Cursor != 0 {
		t.Errorf("initial cursor: want 0, got %d", ps.Cursor)
	}

	ps.ToggleCurrent()
	if !ps.Selected["lombok"] {
		t.Error("lombok should be selected after toggle")
	}

	ps.MoveCursor(5)
	ps.ToggleCurrent()
	if !ps.Selected["postgresql"] {
		t.Error("postgresql should be selected after toggle")
	}

	ps.MoveCursor(100)
	if ps.Cursor != 5 {
		t.Errorf("cursor clamped to 5, got %d", ps.Cursor)
	}

	ps.MoveCursor(-100)
	if ps.Cursor != 0 {
		t.Errorf("cursor clamped to 0, got %d", ps.Cursor)
	}

	ps.ToggleCurrent()
	if ps.Selected["lombok"] {
		t.Error("lombok should be deselected after second toggle")
	}

	ids := ps.GetSelectedIDs()
	if len(ids) != 1 || ids[0] != "postgresql" {
		t.Errorf("expected [postgresql], got %v", ids)
	}
}

// ── MoveToFirst / MoveToLast ──────────────────────────────────────────────────

func TestPickerState_MoveToFirst(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.MoveCursor(4)
	if ps.Cursor != 4 {
		t.Fatalf("expected cursor 4, got %d", ps.Cursor)
	}
	ps.MoveToFirst()
	if ps.Cursor != 0 {
		t.Errorf("MoveToFirst: want 0, got %d", ps.Cursor)
	}
}

func TestPickerState_MoveToLast(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.MoveToLast()
	want := len(ps.SelectableIdx) - 1
	if ps.Cursor != want {
		t.Errorf("MoveToLast: want %d, got %d", want, ps.Cursor)
	}
}

func TestPickerState_MoveToFirst_EmptyFilter(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("zzz-nothing")
	ps.MoveToFirst() // must not panic
	ps.MoveToLast()  // must not panic
}

// ── PageUp / PageDown ─────────────────────────────────────────────────────────

func TestPickerState_PageDown(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.PageDown()
	// 6 items total, pageSize=8 → should clamp to last (5).
	if ps.Cursor != 5 {
		t.Errorf("PageDown from 0 on 6-item list: want 5, got %d", ps.Cursor)
	}
}

func TestPickerState_PageUp(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.MoveToLast()
	ps.PageUp()
	// cursor was 5, pageSize=8 → clamps to 0
	if ps.Cursor != 0 {
		t.Errorf("PageUp from last on 6-item list: want 0, got %d", ps.Cursor)
	}
}

// ── MatchCount ────────────────────────────────────────────────────────────────

func TestPickerState_MatchCount_AllVisible(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	if ps.MatchCount() != 6 {
		t.Errorf("MatchCount on unfiltered list: want 6, got %d", ps.MatchCount())
	}
}

func TestPickerState_MatchCount_Filtered(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("spring")
	// Matches: Spring Boot DevTools, Spring Web, Spring for GraphQL, Spring Data JPA = 4
	if ps.MatchCount() != 4 {
		t.Errorf("MatchCount for 'spring': want 4, got %d", ps.MatchCount())
	}
}

func TestPickerState_MatchCount_NoMatch(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("zzz")
	if ps.MatchCount() != 0 {
		t.Errorf("MatchCount for no-match: want 0, got %d", ps.MatchCount())
	}
}

// ── GetSelectedItems ──────────────────────────────────────────────────────────

func TestPickerState_GetSelectedItems_Empty(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	items := ps.GetSelectedItems()
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestPickerState_GetSelectedItems_WithGroup(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), []string{"web", "postgresql"})
	items := ps.GetSelectedItems()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "web" || items[0].GroupName != "Web" {
		t.Errorf("item[0]: want {web,Web}, got {%s,%s}", items[0].ID, items[0].GroupName)
	}
	if items[1].ID != "postgresql" || items[1].GroupName != "Data" {
		t.Errorf("item[1]: want {postgresql,Data}, got {%s,%s}", items[1].ID, items[1].GroupName)
	}
}

func TestPickerState_GetSelectedItems_MetadataOrder(t *testing.T) {
	// Pre-select in reverse metadata order; result must be in forward order.
	ps := ui.NewPickerState(getMockMetadata(t), []string{"postgresql", "lombok"})
	items := ps.GetSelectedItems()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "lombok" {
		t.Errorf("expected first item to be lombok (metadata order), got %s", items[0].ID)
	}
	if items[1].ID != "postgresql" {
		t.Errorf("expected second item to be postgresql (metadata order), got %s", items[1].ID)
	}
}

// ── StickyGroupHeader ─────────────────────────────────────────────────────────

func TestPickerState_StickyGroupHeader_AtTop(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	// offset=0 means we are at the very top; first header is "Developer Tools"
	got := ps.StickyGroupHeader(0)
	if got != "Developer Tools" {
		t.Errorf("sticky header at offset 0: want 'Developer Tools', got %q", got)
	}
}

func TestPickerState_StickyGroupHeader_MidScroll(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	// FilteredRows layout:
	//  0: header "Developer Tools"
	//  1: lombok
	//  2: devtools
	//  3: header "Web"
	//  4: web
	//  5: graphql
	//  6: header "Data"
	//  7: data-jpa
	//  8: postgresql
	//
	// At offset 4 (web visible), sticky should be "Web"
	got := ps.StickyGroupHeader(4)
	if got != "Web" {
		t.Errorf("sticky header at offset 4: want 'Web', got %q", got)
	}
}

func TestPickerState_StickyGroupHeader_DataGroup(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	got := ps.StickyGroupHeader(7)
	if got != "Data" {
		t.Errorf("sticky header at offset 7: want 'Data', got %q", got)
	}
}

// ── Search filtering ──────────────────────────────────────────────────────────

func TestPickerState_SearchFiltering(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("post")

	if ps.MatchCount() != 1 {
		t.Fatalf("expected 1 match for 'post', got %d", ps.MatchCount())
	}

	ps.ToggleCurrent()
	if !ps.Selected["postgresql"] {
		t.Error("postgresql should be selected")
	}

	ps.ApplyFilter("")
	if ps.MatchCount() != 6 {
		t.Errorf("expected 6 items after clear, got %d", ps.MatchCount())
	}

	ids := ps.GetSelectedIDs()
	if len(ids) != 1 || ids[0] != "postgresql" {
		t.Errorf("selection should persist after filter clear, got %v", ids)
	}
}

func TestPickerState_SearchEmpty(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("zzz-no-match")

	if ps.MatchCount() != 0 {
		t.Errorf("expected 0 matches, got %d", ps.MatchCount())
	}
	if ps.Cursor != -1 {
		t.Errorf("cursor should be -1 when no rows match, got %d", ps.Cursor)
	}
}

func TestPickerState_ToggleNoOp_WhenEmpty(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("zzz-no-match")
	ps.ToggleCurrent() // must not panic
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
	want := []string{"Developer Tools", "Web", "Data"}
	names := ps.GetGroupNames()
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

	if ps.ActiveGroupIdx() != 0 {
		t.Errorf("initial group should be 0, got %d", ps.ActiveGroupIdx())
	}

	ps.TabToNextGroup()
	if ps.ActiveGroupIdx() != 1 {
		t.Errorf("after 1 tab, active group should be 1, got %d", ps.ActiveGroupIdx())
	}
	cursorRow := ps.FilteredRows[ps.SelectableIdx[ps.Cursor]]
	if cursorRow.GroupName != "Web" {
		t.Errorf("cursor should be in Web group, got %q", cursorRow.GroupName)
	}

	ps.TabToNextGroup()
	ps.TabToNextGroup() // wraps back to 0
	if ps.ActiveGroupIdx() != 0 {
		t.Errorf("after wrapping, active group should be 0, got %d", ps.ActiveGroupIdx())
	}
}

func TestPickerState_TabToPrevGroup(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.TabToPrevGroup()
	groups := ps.GetGroupNames()
	if ps.ActiveGroupIdx() != len(groups)-1 {
		t.Errorf("TabToPrevGroup from 0 should wrap to %d, got %d",
			len(groups)-1, ps.ActiveGroupIdx())
	}
}

func TestPickerState_CursorSyncsToActiveGroup(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.MoveCursor(4) // data-jpa
	if ps.ActiveGroupIdx() != 2 {
		t.Errorf("activeGroupIdx should be 2 for Data group, got %d", ps.ActiveGroupIdx())
	}
}

// ── VisibleGroupNames ─────────────────────────────────────────────────────────

func TestPickerState_VisibleGroupNames(t *testing.T) {
	ps := ui.NewPickerState(getMockMetadata(t), nil)
	ps.ApplyFilter("lombok")

	visible := ps.VisibleGroupNames()
	if len(visible) != 1 || visible[0] != "Developer Tools" {
		t.Errorf("expected [Developer Tools], got %v", visible)
	}
}

// ── MatchRanges ───────────────────────────────────────────────────────────────

func TestMatchRanges_EmptyQuery(t *testing.T) {
	if ui.MatchRanges("Spring Web", "") != nil {
		t.Error("expected nil for empty query")
	}
}

func TestMatchRanges_NoMatch(t *testing.T) {
	if len(ui.MatchRanges("Spring Web", "kafka")) != 0 {
		t.Error("expected no matches for 'kafka'")
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
	if len(ui.MatchRanges("data data data", "data")) != 3 {
		t.Error("expected 3 matches")
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
	result := ui.HighlightMatches("PostgreSQL Driver", "post")
	if !containsUnescaped(result, "greSQL Driver") {
		t.Error("highlighted string should still contain the unmatched suffix")
	}
}

// containsUnescaped checks via subsequence whether plain text bytes appear in s.
func containsUnescaped(s, plain string) bool {
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

// ── ProgressModel ─────────────────────────────────────────────────────────────

func TestProgressModel_InitialState(t *testing.T) {
	m := ui.NewProgressModel([]string{"Step A", "Step B", "Step C"})
	if len(m.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(m.Steps))
	}
	if m.Steps[0].Status != ui.StepRunning {
		t.Errorf("first step should be StepRunning, got %v", m.Steps[0].Status)
	}
	for i := 1; i < len(m.Steps); i++ {
		if m.Steps[i].Status != ui.StepPending {
			t.Errorf("step %d should be StepPending, got %v", i, m.Steps[i].Status)
		}
	}
}

func TestProgressModel_EmptyLabels(t *testing.T) {
	m := ui.NewProgressModel(nil)
	if len(m.Steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(m.Steps))
	}
}

func TestProgressModel_SingleStep(t *testing.T) {
	m := ui.NewProgressModel([]string{"Only step"})
	if m.Steps[0].Status != ui.StepRunning {
		t.Error("single step should start as StepRunning")
	}
}
