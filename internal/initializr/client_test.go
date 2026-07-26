package initializr_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/initializr"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

func minimalCfg() *prompt.ProjectConfig {
	return &prompt.ProjectConfig{
		ProjectName:  "demo",
		GroupID:      "com.example",
		ArtifactID:   "demo",
		PackageName:  "com.example.demo",
		BuildTool:    "maven-project",
		Packaging:    "jar",
		JavaVersion:  "21",
		Dependencies: []string{"web"},
	}
}

// ── BuildURL ──────────────────────────────────────────────────────────────────

func TestBuildURL_ContainsAllFields(t *testing.T) {
	cfg := minimalCfg()
	u := initializr.BuildURL(cfg)

	for _, want := range []string{
		"type=maven-project",
		"language=java",
		"groupId=com.example",
		"artifactId=demo",
		"packageName=com.example.demo",
		"packaging=jar",
		"javaVersion=21",
		"dependencies=web",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("BuildURL missing %q\nURL: %s", want, u)
		}
	}
}

func TestBuildURL_MultiDependencies(t *testing.T) {
	cfg := minimalCfg()
	cfg.Dependencies = []string{"web", "data-jpa", "postgresql"}
	u := initializr.BuildURL(cfg)
	// URL may or may not encode the commas depending on http.Get encoding.
	if !strings.Contains(u, "web") || !strings.Contains(u, "data-jpa") || !strings.Contains(u, "postgresql") {
		t.Errorf("URL does not contain all dependencies\nURL: %s", u)
	}
}

func TestBuildURL_EmptyDependencies(t *testing.T) {
	cfg := minimalCfg()
	cfg.Dependencies = nil
	u := initializr.BuildURL(cfg)
	if !strings.Contains(u, "dependencies=") {
		t.Errorf("URL should contain dependencies= parameter\nURL: %s", u)
	}
}

func TestBuildURL_StartsWithBase(t *testing.T) {
	origBase := initializr.MetadataBaseURL
	defer func() { initializr.MetadataBaseURL = origBase }()
	initializr.MetadataBaseURL = "https://start.spring.io/starter.zip?"

	u := initializr.BuildURL(minimalCfg())
	if !strings.HasPrefix(u, "https://start.spring.io/") {
		t.Errorf("expected HTTPS URL, got: %s", u)
	}
}

// ── Download ──────────────────────────────────────────────────────────────────

// emptyZip is a valid empty ZIP archive (22-byte end-of-central-directory record).
var emptyZip = []byte{
	0x50, 0x4B, 0x05, 0x06,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00,
	0x00, 0x00,
}

func withTestServer(t *testing.T, handler http.HandlerFunc) (cleanup func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	orig := initializr.MetadataBaseURL
	initializr.MetadataBaseURL = srv.URL + "/starter.zip?"
	return func() {
		initializr.MetadataBaseURL = orig
		srv.Close()
	}
}

func inTempDir(t *testing.T) (cleanup func()) {
	t.Helper()
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir to temp dir: %v", err)
	}
	return func() { os.Chdir(orig) } //nolint:errcheck
}

func TestDownload_Success(t *testing.T) {
	cleanup := withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/zip")
		w.Write(emptyZip) //nolint:errcheck
	})
	defer cleanup()
	defer inTempDir(t)()

	got, err := initializr.Download(minimalCfg())
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if !strings.HasSuffix(got, ".zip") {
		t.Errorf("expected .zip filename, got %q", got)
	}
	if _, err := os.Stat(got); os.IsNotExist(err) {
		t.Errorf("downloaded file does not exist: %s", got)
	}
}

func TestDownload_ServerReturns500(t *testing.T) {
	cleanup := withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	defer cleanup()
	defer inTempDir(t)()

	_, err := initializr.Download(minimalCfg())
	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention HTTP status, got: %v", err)
	}
}

func TestDownload_NetworkRefused(t *testing.T) {
	orig := initializr.MetadataBaseURL
	initializr.MetadataBaseURL = "http://127.0.0.1:1/starter.zip?"
	defer func() { initializr.MetadataBaseURL = orig }()
	defer inTempDir(t)()

	_, err := initializr.Download(minimalCfg())
	if err == nil {
		t.Error("expected network error, got nil")
	}
}

func TestDownload_FilenameMatchesProjectName(t *testing.T) {
	cleanup := withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(emptyZip) //nolint:errcheck
	})
	defer cleanup()
	defer inTempDir(t)()

	cfg := minimalCfg()
	cfg.ProjectName = "my-service"
	got, err := initializr.Download(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "my-service.zip" {
		t.Errorf("expected filename 'my-service.zip', got %q", got)
	}
}
