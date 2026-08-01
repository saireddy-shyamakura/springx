// Package extract provides Zip Slip-safe archive extraction for springx.
package extract

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip extracts the contents of the ZIP archive at src into the directory
// at dest. It preserves file permissions and protects against Zip Slip
// attacks by validating that every extracted path resolves within dest.
//
// Symlinks are a classic escape vector: an archive that contains a symlink
// pointing outside dest followed by a regular file "through" that link can
// write outside dest even when the path check passes. To close that hole we
// reject any entry whose path resolves (after cleaning) through a symlink
// component inside dest. This is a deliberate trade-off: archives extracted
// by springx are Spring Initializr outputs, which never contain symlinks.
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
// It validates the path to prevent Zip Slip attacks and refuses entries
// that would resolve through a symlink (see Unzip for rationale).
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

	// Second line of defense: reject symlinks outright. A symlinked
	// directory followed by a ".." entry inside it is a classic escape.
	mode := entry.FileInfo().Mode()
	if mode&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to extract symlink entry: %s", entry.Name)
	}

	// Third line of defense: ensure no component of the target path is a
	// symlink pointing outside dest. This closes the "symlink dir + child
	// file" escape even when the child path itself looks clean.
	if err := ensureNoSymlinkComponent(absDest, absTarget); err != nil {
		return fmt.Errorf("refusing to extract %s: %w", entry.Name, err)
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

// ensureNoSymlinkComponent walks the path components from base to target
// (excluding base itself) and errors if any of them is a symlink. base and
// target must both be absolute and target must be lexically under base.
func ensureNoSymlinkComponent(base, target string) error {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return err
	}
	if rel == "." {
		return nil
	}

	walk := base
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		walk = filepath.Join(walk, part)
		fi, err := os.Lstat(walk)
		if err != nil {
			if os.IsNotExist(err) {
				// The rest of the path does not exist yet — nothing
				// symlinked below this point can be present.
				return nil
			}
			return err
		}
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path component %q is a symlink", walk)
		}
	}
	return nil
}
