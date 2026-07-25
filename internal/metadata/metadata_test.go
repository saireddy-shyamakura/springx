package metadata_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
)

const mockMetadataResponse = `{
  "bootVersion": {
    "default": "3.2.0",
    "values": [
      {"id": "3.2.0", "name": "3.2.0"}
    ]
  },
  "javaVersion": {
    "default": "17",
    "values": [
      {"id": "21", "name": "21"},
      {"id": "17", "name": "17"}
    ]
  },
  "packaging": {
    "default": "jar",
    "values": [
      {"id": "jar", "name": "Jar"},
      {"id": "war", "name": "War"}
    ]
  },
  "language": {
    "default": "java",
    "values": [
      {"id": "java", "name": "Java"}
    ]
  },
  "type": {
    "default": "maven-project",
    "values": [
      {
        "id": "maven-project",
        "name": "Maven",
        "action": "/starter.zip"
      },
      {
        "id": "gradle-project",
        "name": "Gradle - Groovy",
        "action": "/starter.zip"
      },
      {
        "id": "maven-build",
        "name": "Maven POM",
        "action": "/pom.xml"
      }
    ]
  },
  "dependencies": {
    "type": "hierarchical-multi-select",
    "values": [
      {
        "name": "Developer Tools",
        "values": [
          {"id": "lombok", "name": "Lombok", "description": "Lombok library"},
          {"id": "devtools", "name": "DevTools", "description": "Developer tools"}
        ]
      },
      {
        "name": "Web",
        "values": [
          {"id": "web", "name": "Spring Web", "description": "Web support"}
        ]
      }
    ]
  }
}`

func TestFetch_SuccessAndCache(t *testing.T) {
	metadata.ResetCache()
	callCount := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Accept Header
		if r.Header.Get("Accept") != "application/vnd.initializr.v2.2+json" {
			t.Errorf("expected Accept header to be application/vnd.initializr.v2.2+json, got %q", r.Header.Get("Accept"))
		}
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockMetadataResponse))
	}))
	defer ts.Close()

	// Override URL
	oldURL := metadata.MetadataURL
	metadata.MetadataURL = ts.URL
	defer func() { metadata.MetadataURL = oldURL }()

	// First fetch should hit the server
	meta, err := metadata.Fetch()
	if err != nil {
		t.Fatalf("first Fetch failed: %v", err)
	}

	if meta.JavaVersion.Default != "17" || len(meta.JavaVersion.Values) != 2 {
		t.Errorf("unexpected javaVersion values: %+v", meta.JavaVersion)
	}

	if meta.Type.Default != "maven-project" || len(meta.Type.Values) != 3 {
		t.Errorf("unexpected type values: %+v", meta.Type)
	}

	// Second fetch should use cache
	metaCached, err := metadata.Fetch()
	if err != nil {
		t.Fatalf("second Fetch failed: %v", err)
	}

	if metaCached != meta {
		t.Error("expected cached metadata instance to be the same pointer")
	}

	if callCount != 1 {
		t.Errorf("expected server to be hit exactly once, but was hit %d times", callCount)
	}
}

func TestFetch_ErrorGraceful(t *testing.T) {
	metadata.ResetCache()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	oldURL := metadata.MetadataURL
	metadata.MetadataURL = ts.URL
	defer func() { metadata.MetadataURL = oldURL }()

	meta, err := metadata.Fetch()
	if err == nil {
		t.Fatal("expected Fetch to return error on HTTP 500, but got nil")
	}
	if meta != nil {
		t.Errorf("expected nil metadata on failure, got %+v", meta)
	}
}

func TestPrintDependencyGroups(t *testing.T) {
	metadata.ResetCache()

	// Parse the mock response manually
	var meta metadata.Metadata
	if err := json.Unmarshal([]byte(mockMetadataResponse), &meta); err != nil {
		t.Fatalf("failed to unmarshal mock: %v", err)
	}

	var buf bytes.Buffer
	meta.PrintDependencyGroups(&buf)

	expected := `Developer Tools
-------------
Lombok
DevTools

Web
-------------
Spring Web
`

	actual := buf.String()
	if strings.TrimSpace(actual) != strings.TrimSpace(expected) {
		t.Errorf("PrintDependencyGroups output mismatch:\nexpected:\n%q\ngot:\n%q", expected, actual)
	}
}
