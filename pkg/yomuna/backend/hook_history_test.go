package backend

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DarlingGoose/vntext/pkg/engine"
	"github.com/DarlingGoose/vntext/pkg/game"
)

func TestTextHookHistoryPersistsInGameDirAndFiltersByHook(t *testing.T) {
	dir := t.TempDir()
	g := &game.Game{WorkingDir: dir}
	b := &LiveBackend{}

	firstTime := time.Date(2026, 5, 15, 9, 0, 0, 0, time.UTC)
	if err := b.appendGameTextHookHistory(g, engine.Line{
		Time:    firstTime,
		Hook:    "Thread A@0",
		Speaker: "Alice",
		Text:    "first",
		Raw:     "[Alice] first",
	}); err != nil {
		t.Fatal(err)
	}
	if err := b.appendGameTextHookHistory(g, engine.Line{
		Time: time.Date(2026, 5, 15, 9, 1, 0, 0, time.UTC),
		Hook: "Thread B@1",
		Text: "second",
		Raw:  "second",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, textHookHistoryFile)); err != nil {
		t.Fatalf("expected hook history in game dir: %v", err)
	}

	lines, err := b.ReadGameTextHookHistory(g, []string{"@0"}, 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected one filtered history line, got %d", len(lines))
	}
	if lines[0].Hook != "Thread A@0" || lines[0].Speaker != "Alice" || lines[0].Text != "first" || !lines[0].Time.Equal(firstTime) {
		t.Fatalf("unexpected history line: %+v", lines[0])
	}
}

func TestTextHookHistoryMaxLinesKeepsNewestMatches(t *testing.T) {
	dir := t.TempDir()
	g := &game.Game{WorkingDir: dir}
	b := &LiveBackend{}

	for _, text := range []string{"one", "two", "three"} {
		if err := b.appendGameTextHookHistory(g, engine.Line{Hook: "@0", Text: text}); err != nil {
			t.Fatal(err)
		}
	}

	lines, err := b.ReadGameTextHookHistory(g, nil, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0].Text != "two" || lines[1].Text != "three" {
		t.Fatalf("expected newest two lines, got %+v", lines)
	}
}
