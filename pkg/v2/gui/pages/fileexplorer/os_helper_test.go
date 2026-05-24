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

func TestPathInputSearchTextUsesCurrentDirChild(t *testing.T) {
	dir := t.TempDir()
	p := &FileExplorer{CurrentDir: dir}

	got := p.pathInputSearchTextFor(filepath.Join(dir, "searching_word"))
	if got != "searching_word" {
		t.Fatalf("pathInputSearchTextFor() = %q, want searching_word", got)
	}
}

func TestPathInputSearchTextUsesFirstRelativeComponent(t *testing.T) {
	dir := t.TempDir()
	p := &FileExplorer{CurrentDir: dir}

	got := p.pathInputSearchTextFor(filepath.Join("nested", "searching_word"))
	if got != "nested" {
		t.Fatalf("pathInputSearchTextFor() = %q, want nested", got)
	}
}

func TestReloadUsesPathInputSearchText(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"searching_word.txt", "other.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	p := NewFileExplorer(dir, nil, nil)
	p.PathInput.SetText(filepath.Join(dir, "searching"))
	p.pathInputSearchActive = true
	p.reload()

	if len(p.entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1: %#v", len(p.entries), p.entries)
	}
	if p.entries[0].Name != "searching_word.txt" {
		t.Fatalf("entry = %q, want searching_word.txt", p.entries[0].Name)
	}
}

func TestSelectDoesNotUseSelectedFilePathAsSearchText(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"searching_word.txt", "other.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	p := NewFileExplorer(dir, nil, nil)
	p.pathInputSearchActive = true
	p.Select(filepath.Join(dir, "searching_word.txt"))
	p.reload()

	if len(p.entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2: %#v", len(p.entries), p.entries)
	}
}

func TestProgrammaticPathInputChangeDoesNotActivateSearch(t *testing.T) {
	dir := t.TempDir()
	p := NewFileExplorer(dir, nil, nil)
	selected := filepath.Join(dir, "searching_word.txt")

	p.setPathInputText(selected)
	p.pathInputSearchActive = true
	if !p.consumeProgrammaticPathInputChange() {
		t.Fatal("consumeProgrammaticPathInputChange() = false, want true")
	}
	if p.pathInputSearchActive {
		t.Fatal("pathInputSearchActive = true, want false")
	}
	if got := p.pathInputSearchText(); got != "" {
		t.Fatalf("pathInputSearchText() = %q, want empty", got)
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
