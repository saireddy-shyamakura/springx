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
