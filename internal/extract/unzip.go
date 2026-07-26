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
	// Open the ZIP archive for reading.
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip archive %s: %w", src, err)
	}
	defer reader.Close()

	// Ensure the destination directory exists.
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dest, err)
	}

	// Iterate over every entry in the archive.
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
	// This is the first line of defence against Zip Slip.
	cleanName := filepath.Clean(entry.Name)

	// Build the full target path and resolve it to an absolute path.
	target := filepath.Join(dest, cleanName)

	// Verify the resolved path is within the destination directory.
	// This prevents a malicious archive from writing files outside dest
	// by using path traversal sequences such as "../../evil.sh".
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
		// Create the directory with the same permissions as the entry.
		if err := os.MkdirAll(target, entry.Mode()); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", target, err)
		}

	default:
		// Ensure parent directories exist for the file.
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create parent directories for %s: %w", target, err)
		}

		// Open the entry for reading.
		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("failed to open entry %s in archive: %w", entry.Name, err)
		}
		defer rc.Close()

		// Create the destination file, preserving the entry's file mode.
		outFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, entry.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", target, err)
		}
		defer outFile.Close()

		// Copy the contents from the archive entry to the file.
		if _, err := io.Copy(outFile, rc); err != nil {
			return fmt.Errorf("failed to write file %s: %w", target, err)
		}
	}

	return nil
}
