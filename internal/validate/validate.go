// Package validate provides strict input validation for user-supplied
// project metadata before it reaches the network, filesystem, or shell.
//
// Spring Initializr is strict about these fields, and the values are used
// to build URLs, directory names, and generated file content — so allowing
// arbitrary characters is both an upstream-compatibility problem and a
// security risk (URL/SSRF injection, path traversal, markdown/HTML content
// injection into generated files).
package validate

import (
	"fmt"
	"regexp"
	"strings"
)

// MavenIdentifier matches a valid Maven groupId/artifactId: lowercase
// letters, digits, dots, hyphens, and underscores (plus the ':' separator
// some ecosystem IDs use). This is the pattern Spring Initializr accepts.
var MavenIdentifier = regexp.MustCompile(`^[a-zA-Z0-9_.:-]+$`)

// PackageName matches a Java package name as Spring Initializr accepts it:
// dot-separated identifiers, each starting with a letter or underscore and
// containing letters, digits, underscores, or hyphens. (Strict Java forbids
// hyphens in identifiers, but Initializr generates working code for them and
// springx's own defaults — e.g. "com.custom.pkg.service-demo" — use them.)
var PackageName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*(\.[a-zA-Z_][a-zA-Z0-9_-]*)*$`)

// ProjectName matches a safe directory/artifact name: letters, digits,
// dots, hyphens, underscores. It must not be a path-traversal sequence.
var ProjectName = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// JavaVersion matches version strings like "21", "17", "21.0.1".
var JavaVersion = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)

// ShellSafe matches a bare command name for $EDITOR/$VISUAL: no spaces,
// no shell metacharacters. Path separators are allowed so "/usr/bin/nano"
// or "C:\...\nano.exe" work.
var ShellSafe = regexp.MustCompile(`^[a-zA-Z0-9_./:\\-]+$`)

// ErrInvalid is returned when a field fails validation.
type ErrInvalid struct {
	Field string
	Value string
	Why   string
}

func (e *ErrInvalid) Error() string {
	return fmt.Sprintf("invalid %s %q: %s", e.Field, e.Value, e.Why)
}

// invalid is a small helper for building ErrInvalid values.
func invalid(field, value, why string) error {
	return &ErrInvalid{Field: field, Value: value, Why: why}
}

// ProjectNameValid reports whether name is a safe directory name.
func ProjectNameValid(name string) bool {
	return name != "" && ProjectName.MatchString(name) && name != "." && name != ".."
}

// GroupIDValid reports whether id is a valid Maven groupId.
func GroupIDValid(id string) bool {
	return id != "" && MavenIdentifier.MatchString(id)
}

// ArtifactIDValid reports whether id is a valid Maven artifactId.
func ArtifactIDValid(id string) bool {
	return id != "" && MavenIdentifier.MatchString(id)
}

// PackageNameValid reports whether name is a valid Java package name.
func PackageNameValid(name string) bool {
	return name != "" && PackageName.MatchString(name)
}

// JavaVersionValid reports whether v is a numeric version string.
func JavaVersionValid(v string) bool {
	return v != "" && JavaVersion.MatchString(v)
}

// BuildToolValid reports whether id looks like a Spring Initializr project
// type (e.g. "maven-project", "gradle-project-kotlin").
func BuildToolValid(id string) bool {
	return id != "" && MavenIdentifier.MatchString(id)
}

// PackagingValid reports whether v is a supported packaging value.
func PackagingValid(v string) bool {
	v = strings.ToLower(v)
	return v == "jar" || v == "war"
}

// LanguageValid reports whether v is a supported language value.
func LanguageValid(v string) bool {
	v = strings.ToLower(v)
	return v == "java" || v == "kotlin" || v == "groovy"
}

// ValidateProjectConfig validates a complete ProjectConfig and returns the
// first offending field, or nil when everything is acceptable.
func ValidateProjectConfig(projectName, groupID, artifactID, packageName, buildTool, packaging, javaVersion string) error {
	if !ProjectNameValid(projectName) {
		return invalid("project name", projectName,
			"must use only letters, digits, '.', '_' or '-', and must not be '.' or '..'")
	}
	if !GroupIDValid(groupID) {
		return invalid("group ID", groupID,
			"must use only letters, digits, '.', '_', '-' or ':'")
	}
	if !ArtifactIDValid(artifactID) {
		return invalid("artifact ID", artifactID,
			"must use only letters, digits, '.', '_', '-' or ':'")
	}
	if !PackageNameValid(packageName) {
		return invalid("package name", packageName,
			"must be a valid Java package name (dot-separated identifiers)")
	}
	if !BuildToolValid(buildTool) {
		return invalid("build tool", buildTool,
			"must be a valid Spring Initializr project type ID")
	}
	if packaging != "" && !PackagingValid(packaging) {
		return invalid("packaging", packaging, "must be 'jar' or 'war'")
	}
	if !JavaVersionValid(javaVersion) {
		return invalid("Java version", javaVersion, "must be a numeric version string")
	}
	return nil
}
