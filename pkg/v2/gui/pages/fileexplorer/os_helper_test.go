package fileexplorer

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractZipArchiveExtractsToSiblingDirectory(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "sample.zip")
	writeTestZip(t, archivePath, map[string]string{
		"nested/file.txt": "contents",
	})

	dest, err := extractZipArchive(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	if dest != filepath.Join(dir, "sample") {
		t.Fatalf("dest = %q, want sibling sample dir", dest)
	}

	got, err := os.ReadFile(filepath.Join(dest, "nested", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "contents" {
		t.Fatalf("extracted content = %q, want contents", got)
	}
}

func TestExtractZipArchiveUsesUniqueDirectory(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "sample.zip")
	writeTestZip(t, archivePath, map[string]string{
		"file.txt": "contents",
	})
	if err := os.Mkdir(filepath.Join(dir, "sample"), 0o755); err != nil {
		t.Fatal(err)
	}

	dest, err := extractZipArchive(archivePath)
	if err != nil {
		t.Fatal(err)
	}

	if dest != filepath.Join(dir, "sample 2") {
		t.Fatalf("dest = %q, want sample 2", dest)
	}
}

func TestExtractZipArchiveRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "bad.zip")
	writeTestZip(t, archivePath, map[string]string{
		"../escape.txt": "bad",
	})

	if _, err := extractZipArchive(archivePath); err == nil {
		t.Fatal("extractZipArchive() err = nil, want traversal error")
	}
}

func TestReadDirSortsByExtension(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"z.log", "a.txt", "noext"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := readDir(dir, "", false, SortByExtension, false)
	if err != nil {
		t.Fatal(err)
	}

	got := []string{entries[0].Name, entries[1].Name, entries[2].Name}
	want := []string{"z.log", "a.txt", "noext"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entries = %#v, want %#v", got, want)
		}
	}
}

func writeTestZip(t *testing.T, path string, files map[string]string) {
	t.Helper()

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for name, contents := range files {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
	}
}
