// Package extract provides Zip Slip-safe archive extraction for springx.
package extract

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Unzip extracts the contents of the ZIP archive at src into the directory
// at dest. It preserves file permissions and protects against Zip Slip
// attacks by validating that every extracted path resolves within dest.
func Unzip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip archive %s: %w", src, err)
	}
	defer reader.Close() //nolint:errcheck

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dest, err)
	}

	for _, entry := range reader.File {
		if err := extractEntry(entry, dest); err != nil {
			return err
		}
	}

	return nil
}

// extractEntry extracts a single zip.File into the destination directory.
// It validates the path to prevent Zip Slip attacks.
func extractEntry(entry *zip.File, dest string) error {
	// Clean the entry name to remove any relative components (e.g. "../").
	// This is the first line of defense against Zip Slip.
	cleanName := filepath.Clean(entry.Name)

	target := filepath.Join(dest, cleanName)

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("failed to resolve destination path: %w", err)
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}
	rel, err := filepath.Rel(absDest, absTarget)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}
	if rel == ".." || len(rel) > 2 && rel[:3] == "../" {
		return fmt.Errorf("illegal file path: %s (zip slip detected)", entry.Name)
	}

	switch {
	case entry.FileInfo().IsDir():
		if err := os.MkdirAll(target, entry.Mode()); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", target, err)
		}

	default:
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("failed to create parent directories for %s: %w", target, err)
		}

		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("failed to open entry %s in archive: %w", entry.Name, err)
		}
		defer rc.Close() //nolint:errcheck

		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", target, err)
		}
		defer outFile.Close() //nolint:errcheck

		if _, err := io.Copy(outFile, rc); err != nil {
			return fmt.Errorf("failed to write file %s: %w", target, err)
		}
	}

	return nil
}
