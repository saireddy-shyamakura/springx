// Package initializr communicates with Spring Initializr to download
// generated Spring Boot project archives.
package initializr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/saireddy-shyamakura/springx/internal/httpx"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

// MetadataBaseURL is the base URL used by BuildURL and Download.
// Override in tests to point at a local httptest.Server.
var MetadataBaseURL = "https://start.spring.io/starter.zip?"

// httpClient is used for the download. It is a package-level variable so
// tests can swap in a client that talks to an httptest.Server without
// hitting the network.
var httpClient = httpx.New(60 * time.Second)

// maxResponseBytes caps the size of a downloaded project archive to guard
// against a hostile or misconfigured server sending unbounded data.
const maxResponseBytes = 200 << 20 // 200 MiB

// BuildURL constructs the full Spring Initializr download URL from cfg.
// All user-supplied values are percent-encoded with url.Values so that
// characters like '&', '?', or '#' in project metadata cannot smuggle
// extra query parameters (SSRF/URL-injection hardening).
func BuildURL(cfg *prompt.ProjectConfig) string {
	q := url.Values{}
	q.Set("type", cfg.BuildTool)
	q.Set("language", "java")
	q.Set("groupId", cfg.GroupID)
	q.Set("artifactId", cfg.ArtifactID)
	q.Set("name", cfg.ProjectName)
	q.Set("packageName", cfg.PackageName)
	q.Set("packaging", cfg.Packaging)
	q.Set("javaVersion", cfg.JavaVersion)
	q.Set("dependencies", strings.Join(cfg.Dependencies, ","))

	base := MetadataBaseURL
	// Accept either "https://start.spring.io/starter.zip?" or a bare
	// "https://start.spring.io/starter.zip" as the base.
	if strings.Contains(base, "?") {
		return base + q.Encode()
	}
	return base + "?" + q.Encode()
}

// Download fetches a Spring Boot project ZIP from Spring Initializr and
// writes it to the current working directory. It returns the filename of the
// saved ZIP on success.
func Download(cfg *prompt.ProjectConfig) (string, error) {
	url := BuildURL(cfg)
	filename := cfg.ProjectName + ".zip"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request to Spring Initializr failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spring Initializr returned HTTP %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create output file %s: %w", filename, err)
	}
	defer file.Close() //nolint:errcheck

	// Bound the amount of data we read from the network so a malicious
	// server cannot fill the disk. The reader is an additional hard cap
	// beyond the client-level timeout.
	limited := io.LimitReader(resp.Body, maxResponseBytes+1)
	n, err := io.Copy(file, limited)
	if err != nil {
		return "", fmt.Errorf("failed to write ZIP to %s: %w", filename, err)
	}
	if n > maxResponseBytes {
		os.Remove(filename) //nolint:errcheck
		return "", fmt.Errorf("response exceeded %d bytes; refusing to save oversized archive", maxResponseBytes)
	}

	return filename, nil
}
