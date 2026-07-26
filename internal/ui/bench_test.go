package ui_test

import (
	"encoding/json"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/ui"
)

// BenchmarkPickerState_NewAndFilter measures the cost of constructing a
// PickerState from a realistic metadata payload and running a filter query.
func BenchmarkPickerState_NewAndFilter(b *testing.B) {
	meta := loadBenchMeta(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps := ui.NewPickerState(meta, nil)
		ps.ApplyFilter("spring")
	}
}

// BenchmarkPickerState_ToggleAndSelect measures selection toggles.
func BenchmarkPickerState_ToggleAndSelect(b *testing.B) {
	meta := loadBenchMeta(b)
	ps := ui.NewPickerState(meta, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.MoveCursor(1)
		ps.ToggleCurrent()
		ps.MoveCursor(-1)
		ps.ToggleCurrent()
	}
}

// BenchmarkHighlightMatches measures the per-row highlighting cost.
func BenchmarkHighlightMatches(b *testing.B) {
	s := "Spring Data JPA — support for JPA-based repositories"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.HighlightMatches(s, "jpa")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func loadBenchMeta(b *testing.B) *metadata.Metadata {
	b.Helper()
	raw := `{
		"bootVersion":{"default":"3.5.0","values":[{"id":"3.5.0","name":"3.5.0"}]},
		"javaVersion":{"default":"21","values":[{"id":"21","name":"21"},{"id":"17","name":"17"}]},
		"packaging":{"default":"jar","values":[{"id":"jar","name":"Jar"}]},
		"language":{"default":"java","values":[{"id":"java","name":"Java"}]},
		"type":{"default":"maven-project","values":[{"id":"maven-project","name":"Maven","action":"/starter.zip"}]},
		"dependencies":{"type":"hierarchical-multi-select","values":[
			{"name":"Web","values":[
				{"id":"web","name":"Spring Web","description":"Build web, including RESTful, applications"},
				{"id":"webflux","name":"Spring Reactive Web","description":"Build reactive web applications"},
				{"id":"graphql","name":"Spring for GraphQL","description":"GraphQL support"}
			]},
			{"name":"Data","values":[
				{"id":"data-jpa","name":"Spring Data JPA","description":"Persist data in SQL stores"},
				{"id":"postgresql","name":"PostgreSQL Driver","description":"A JDBC and R2DBC driver"},
				{"id":"flyway","name":"Flyway Migration","description":"Version control for your database"},
				{"id":"redis","name":"Spring Data Redis","description":"Advanced key-value store"}
			]},
			{"name":"Security","values":[
				{"id":"security","name":"Spring Security","description":"Highly customizable authentication"},
				{"id":"oauth2-client","name":"OAuth2 Client","description":"Spring Boot integration for OAuth 2.0"}
			]},
			{"name":"Developer Tools","values":[
				{"id":"lombok","name":"Lombok","description":"Java annotation library"},
				{"id":"devtools","name":"Spring Boot DevTools","description":"Fast application restarts"},
				{"id":"actuator","name":"Spring Boot Actuator","description":"Production ready features"}
			]}
		]}
	}`
	var meta metadata.Metadata
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		b.Fatalf("failed to parse benchmark metadata: %v", err)
	}
	return &meta
}
