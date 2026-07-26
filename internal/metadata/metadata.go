package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// MetadataURL points to the Spring Initializr client metadata endpoint.
// It can be overridden in tests to point to a mock server.
var MetadataURL = "https://start.spring.io/metadata/client"

// MetadataSelectValue represents a selectable option in the metadata.
type MetadataSelectValue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MetadataSelect represents a single-select metadata field.
type MetadataSelect struct {
	Default string                `json:"default"`
	Values  []MetadataSelectValue `json:"values"`
}

// TypeValue represents a project type/build tool option in the metadata.
type TypeValue struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Action      string            `json:"action"`
	Tags        map[string]string `json:"tags"`
}

// TypeMetadata represents the metadata for project types.
type TypeMetadata struct {
	Default string      `json:"default"`
	Values  []TypeValue `json:"values"`
}

// DependencyValue represents a single dependency option.
type DependencyValue struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DependencyGroup represents a group of dependencies.
type DependencyGroup struct {
	Name   string            `json:"name"`
	Values []DependencyValue `json:"values"`
}

// DependencyMetadata represents the metadata for all dependencies.
type DependencyMetadata struct {
	Type   string            `json:"type"`
	Values []DependencyGroup `json:"values"`
}

// Metadata models the Spring Initializr client metadata response.
type Metadata struct {
	BootVersion  MetadataSelect     `json:"bootVersion"`
	JavaVersion  MetadataSelect     `json:"javaVersion"`
	Packaging    MetadataSelect     `json:"packaging"`
	Language     MetadataSelect     `json:"language"`
	Type         TypeMetadata       `json:"type"`
	Dependencies DependencyMetadata `json:"dependencies"`
}

var (
	cacheMutex     sync.Mutex
	cachedMetadata *Metadata
)

// Fetch retrieves Spring Initializr client metadata. If metadata has already
// been successfully fetched during the process's lifetime, it returns the cached
// metadata to avoid multiple HTTP requests.
func Fetch() (*Metadata, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if cachedMetadata != nil {
		return cachedMetadata, nil
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, MetadataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	// Request Initializr client metadata v2.2 JSON format.
	req.Header.Set("Accept", "application/vnd.initializr.v2.2+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to Spring Initializr failed: %w. Please check your internet connection.", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Spring Initializr returned HTTP status %s", resp.Status)
	}

	var meta Metadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("failed to parse Spring Initializr metadata JSON: %w", err)
	}

	cachedMetadata = &meta
	return cachedMetadata, nil
}

// ResetCache clears the cached metadata. This is mainly useful for unit tests.
func ResetCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cachedMetadata = nil
}

// PrintDependencyGroups prints dependency groups and their items to the provided writer.
// Formatting matches the requirement of printing groups separated by horizontal rules and spacing.
func (m *Metadata) PrintDependencyGroups(w io.Writer) {
	for i, group := range m.Dependencies.Values {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, group.Name)
		fmt.Fprintln(w, "-------------")
		for _, dep := range group.Values {
			fmt.Fprintln(w, dep.Name)
		}
	}
}
