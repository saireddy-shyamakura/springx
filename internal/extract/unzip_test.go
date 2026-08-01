package extract_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/extract"
)

// makeZip builds an in-memory ZIP archive with the supplied files.
// files is a map of archive path → file content.
func makeZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("zip.Create(%q): %v", name, err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatalf("zip write(%q): %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip.Close: %v", err)
	}
	return buf.Bytes()
}

// writeTempZip writes data to a temp file and returns its path.
func writeTempZip(t *testing.T, data []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.zip")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close() //nolint:errcheck
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write zip: %v", err)
	}
	return f.Name()
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestUnzip_ExtractsFiles(t *testing.T) {
	data := makeZip(t, map[string]string{
		"src/Main.java": "public class Main {}",
		"pom.xml":       "<project/>",
	})
	src := writeTempZip(t, data)
	dest := t.TempDir()

	if err := extract.Unzip(src, dest); err != nil {
		t.Fatalf("Unzip returned error: %v", err)
	}

	for _, want := range []string{
		filepath.Join(dest, "src", "Main.java"),
		filepath.Join(dest, "pom.xml"),
	} {
		if _, err := os.Stat(want); os.IsNotExist(err) {
			t.Errorf("expected file to exist: %s", want)
		}
	}
}

func TestUnzip_PreservesContent(t *testing.T) {
	const content = "public class Hello { public static void main(String[] a){} }"
	data := makeZip(t, map[string]string{"Hello.java": content})
	src := writeTempZip(t, data)
	dest := t.TempDir()

	if err := extract.Unzip(src, dest); err != nil {
		t.Fatalf("Unzip: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "Hello.java"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != content {
		t.Errorf("content mismatch\nwant: %q\n got: %q", content, string(got))
	}
}

func TestUnzip_CreatesNestedDirectories(t *testing.T) {
	data := makeZip(t, map[string]string{
		"a/b/c/deep.txt": "deep",
	})
	src := writeTempZip(t, data)
	dest := t.TempDir()

	if err := extract.Unzip(src, dest); err != nil {
		t.Fatalf("Unzip: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "a", "b", "c", "deep.txt")); os.IsNotExist(err) {
		t.Error("nested file not found after extraction")
	}
}

func TestUnzip_ZipSlipRejected(t *testing.T) {
	// Manually craft a ZIP with a path-traversal entry name.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create("../../etc/passwd")
	if err != nil {
		t.Fatalf("zip.Create: %v", err)
	}
	f.Write([]byte("root:x:0:0")) //nolint:errcheck
	w.Close()                     //nolint:errcheck

	src := writeTempZip(t, buf.Bytes())
	dest := t.TempDir()

	err = extract.Unzip(src, dest)
	if err == nil {
		t.Error("expected Zip Slip error, got nil")
	}
}

func TestUnzip_SymlinkEntryRejected(t *testing.T) {
	// An archive containing a symlink pointing outside dest must be refused
	// outright (never materialized as a link, never followed).
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	hdr := &zip.FileHeader{
		Name: "evil-link",
	}
	hdr.SetMode(os.ModeSymlink | 0o777)
	f, err := w.CreateHeader(hdr)
	if err != nil {
		t.Fatalf("zip.CreateHeader: %v", err)
	}
	f.Write([]byte("/etc")) //nolint:errcheck
	w.Close()               //nolint:errcheck

	src := writeTempZip(t, buf.Bytes())
	dest := t.TempDir()

	err = extract.Unzip(src, dest)
	if err == nil {
		t.Fatal("expected error for symlink entry, got nil")
	}

	// The link must not exist on disk.
	if _, err := os.Lstat(filepath.Join(dest, "evil-link")); !os.IsNotExist(err) {
		t.Errorf("symlink should not have been created: %v", err)
	}
}

func TestUnzip_SymlinkDirThenFileRejected(t *testing.T) {
	// Classic escape: symlink dir "link" → outside dest, then a file
	// "link/pwned" written through it. The path check alone would pass;
	// the symlink-component check must reject it.
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	linkHdr := &zip.FileHeader{Name: "link"}
	linkHdr.SetMode(os.ModeSymlink | 0o777)
	link, err := w.CreateHeader(linkHdr)
	if err != nil {
		t.Fatalf("zip.CreateHeader(link): %v", err)
	}
	link.Write([]byte("/tmp/escape-target")) //nolint:errcheck

	file, err := w.Create("link/pwned.txt")
	if err != nil {
		t.Fatalf("zip.Create(file): %v", err)
	}
	file.Write([]byte("pwned")) //nolint:errcheck
	w.Close()                   //nolint:errcheck

	src := writeTempZip(t, buf.Bytes())
	dest := t.TempDir()

	err = extract.Unzip(src, dest)
	if err == nil {
		t.Fatal("expected error for symlink-dir escape, got nil")
	}

	if _, err := os.Stat("/tmp/escape-target/pwned.txt"); !os.IsNotExist(err) {
		t.Errorf("file must not be written through symlink: %v", err)
	}
}

func TestUnzip_EmptyArchive(t *testing.T) {
	data := makeZip(t, map[string]string{})
	src := writeTempZip(t, data)
	dest := t.TempDir()

	if err := extract.Unzip(src, dest); err != nil {
		t.Errorf("empty archive should not error: %v", err)
	}
}

func TestUnzip_InvalidZipReturnsError(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "*.zip")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	f.WriteString("this is not a zip file") //nolint:errcheck
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := extract.Unzip(f.Name(), t.TempDir()); err == nil {
		t.Error("expected error for invalid zip, got nil")
	}
}

func TestUnzip_NonExistentSourceReturnsError(t *testing.T) {
	err := extract.Unzip("/does/not/exist.zip", t.TempDir())
	if err == nil {
		t.Error("expected error for missing source file, got nil")
	}
}
