package ui_test

import (
	"encoding/json"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/ui"
)

const mockMetadataJSON = `{
  "dependencies": {
    "type": "hierarchical-multi-select",
    "values": [
      {
        "name": "Developer Tools",
        "values": [
          {"id": "lombok", "name": "Lombok", "description": "Reduce boilerplate code"},
          {"id": "devtools", "name": "Spring Boot DevTools", "description": "Fast application restarts"}
        ]
      },
      {
        "name": "Web",
        "values": [
          {"id": "web", "name": "Spring Web", "description": "Build web applications"},
          {"id": "graphql", "name": "Spring for GraphQL", "description": "GraphQL support"}
        ]
      },
      {
        "name": "Data",
        "values": [
          {"id": "data-jpa", "name": "Spring Data JPA", "description": "Persist data in SQL"},
          {"id": "postgresql", "name": "PostgreSQL Driver", "description": "JDBC driver for PostgreSQL"}
        ]
      }
    ]
  }
}`

func getMockMetadata(t *testing.T) *metadata.Metadata {
	var meta metadata.Metadata
	if err := json.Unmarshal([]byte(mockMetadataJSON), &meta); err != nil {
		t.Fatalf("failed to unmarshal mock metadata: %v", err)
	}
	return &meta
}

func TestPickerState_NavigationAndSelection(t *testing.T) {
	meta := getMockMetadata(t)
	ps := ui.NewPickerState(meta)

	// Initial cursor should be at index 0 (Lombok)
	if ps.Cursor != 0 {
		t.Errorf("expected initial cursor to be 0, got %d", ps.Cursor)
	}

	// Toggle Lombok (item 0)
	ps.ToggleCurrent()
	if !ps.Selected["lombok"] {
		t.Error("expected 'lombok' to be selected")
	}

	// Move cursor down 5 steps (Lombok -> devtools -> web -> graphql -> data-jpa -> postgresql)
	ps.MoveCursor(5)
	// Toggle PostgreSQL Driver
	ps.ToggleCurrent()
	if !ps.Selected["postgresql"] {
		t.Error("expected 'postgresql' to be selected")
	}

	// Moving cursor past end should clamp to last element (index 5)
	ps.MoveCursor(100)
	if ps.Cursor != 5 {
		t.Errorf("expected cursor to be clamped at 5, got %d", ps.Cursor)
	}

	// Moving cursor negative should clamp to 0
	ps.MoveCursor(-100)
	if ps.Cursor != 0 {
		t.Errorf("expected cursor to be clamped at 0, got %d", ps.Cursor)
	}

	// Toggle Lombok off
	ps.ToggleCurrent()
	if ps.Selected["lombok"] {
		t.Error("expected 'lombok' to be deselected")
	}

	selectedIDs := ps.GetSelectedIDs()
	if len(selectedIDs) != 1 || selectedIDs[0] != "postgresql" {
		t.Errorf("expected selected IDs to be ['postgresql'], got %v", selectedIDs)
	}
}

func TestPickerState_SearchFiltering(t *testing.T) {
	meta := getMockMetadata(t)
	ps := ui.NewPickerState(meta)

	// Search "post"
	ps.ApplyFilter("post")

	// Filtered result should contain Data header and PostgreSQL Driver (1 selectable dependency)
	if len(ps.SelectableIdx) != 1 {
		t.Fatalf("expected 1 selectable dependency for 'post', got %d", len(ps.SelectableIdx))
	}

	// Select PostgreSQL Driver
	ps.ToggleCurrent()
	if !ps.Selected["postgresql"] {
		t.Error("expected 'postgresql' to be selected")
	}

	// Clear filter
	ps.ApplyFilter("")

	if len(ps.SelectableIdx) != 6 {
		t.Errorf("expected 6 selectable dependencies after clearing filter, got %d", len(ps.SelectableIdx))
	}

	// PostgreSQL should remain selected after clearing filter
	selectedIDs := ps.GetSelectedIDs()
	if len(selectedIDs) != 1 || selectedIDs[0] != "postgresql" {
		t.Errorf("expected selected IDs to retain 'postgresql', got %v", selectedIDs)
	}
}
