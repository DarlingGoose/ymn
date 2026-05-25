package backend

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/vntext/pkg/game"
)

func TestGamescopeOptionsForGameRespectsSavedFullscreen(t *testing.T) {
	cfg := gamescopeOptionsForGame(&game.Game{
		Runner: game.RunnerGamescope,
		GamescopeConfig: &gamescope.Options{
			Name:       "gamescope",
			Fullscreen: false,
		},
	})

	if cfg.Fullscreen {
		t.Fatal("Fullscreen = true, want false from saved GamescopeConfig")
	}
}

func TestLatestGameForRunPrefersCurrentConfig(t *testing.T) {
	stale := &game.Game{
		Name:   "same-game",
		Runner: game.RunnerGamescope,
		GamescopeConfig: &gamescope.Options{
			Width: 1280,
		},
	}
	current := &game.Game{
		Name:   "same-game",
		Runner: game.RunnerGamescope,
		GamescopeConfig: &gamescope.Options{
			Width: 480,
		},
	}

	b := &LiveBackend{current: current}
	got := b.latestGameForRun(stale)

	if got != current {
		t.Fatal("latestGameForRun did not return the backend's current config")
	}
	if got.GamescopeConfig == nil || got.GamescopeConfig.Width != 480 {
		t.Fatalf("Width = %v, want 480", got.GamescopeConfig)
	}
}

func TestMissingWinetrickDepsSkipsInstalledDeps(t *testing.T) {
	prefix := t.TempDir()
	logPath := filepath.Join(prefix, "winetricks.log")
	if err := os.WriteFile(logPath, []byte("fakejapanese\nvcrun2022\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := missingWinetrickDeps(prefix, []string{"vcrun2022", "corefonts", "fakejapanese"})
	want := []string{"corefonts"}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("missingWinetrickDeps() = %#v, want %#v", got, want)
	}
}

func TestMissingWinetrickDepsInstallsAllWhenLogIsMissing(t *testing.T) {
	got := missingWinetrickDeps(t.TempDir(), []string{"vcrun2022", "corefonts"})
	want := []string{"corefonts", "vcrun2022"}
	if len(got) != len(want) {
		t.Fatalf("missingWinetrickDeps() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("missingWinetrickDeps() = %#v, want %#v", got, want)
		}
	}
}

func TestGamescopeWinePrefixPrefersRunnerConfigPrefix(t *testing.T) {
	g := &game.Game{
		PrefixPath: "/game-prefix",
		GamescopeConfig: &gamescope.Options{
			DefaultWinePrefix: "/gamescope-prefix",
		},
	}
	o := gr.ApplyOptions(gr.WithWinePrefix("/runner-prefix"))

	if got := gamescopeWinePrefix(g, o); got != "/runner-prefix" {
		t.Fatalf("gamescopeWinePrefix() = %q, want /runner-prefix", got)
	}
}

func TestGamescopeWinetricksEnvSetsWinePrefix(t *testing.T) {
	env := gamescopeWinetricksEnv("/game-prefix", gr.ApplyOptions(gr.WithSystemArch("win32")))

	if !envContains(env, "WINEPREFIX=/game-prefix") {
		t.Fatalf("env missing WINEPREFIX=/game-prefix: %#v", env)
	}
	if !envContains(env, "WINEARCH=win32") {
		t.Fatalf("env missing WINEARCH=win32: %#v", env)
	}
}

func TestPrepareInstallerPathsExtractsZipInstaller(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "game.zip")
	writeZip(t, archivePath, map[string]string{
		"Game.exe": "fake exe",
	})

	installerPath, gamePath, err := prepareInstallerPaths(context.Background(), archivePath, "Game.exe")
	if err != nil {
		t.Fatal(err)
	}

	if installerPath != "" {
		t.Fatalf("installerPath = %q, want empty for archive installer", installerPath)
	}
	if !filepath.IsAbs(gamePath) {
		t.Fatalf("gamePath = %q, want absolute extracted path", gamePath)
	}
	if _, err := os.Stat(gamePath); err != nil {
		t.Fatalf("extracted game path was not created: %v", err)
	}
}

func TestPrepareInstallerPathsLeavesNonArchiveInstaller(t *testing.T) {
	dir := t.TempDir()
	installerPath := filepath.Join(dir, "setup.exe")
	if err := os.WriteFile(installerPath, []byte("MZ"), 0o644); err != nil {
		t.Fatal(err)
	}

	gotInstallerPath, gotGamePath, err := prepareInstallerPaths(context.Background(), installerPath, "Game.exe")
	if err != nil {
		t.Fatal(err)
	}

	if gotInstallerPath != installerPath {
		t.Fatalf("installerPath = %q, want %q", gotInstallerPath, installerPath)
	}
	if gotGamePath != "Game.exe" {
		t.Fatalf("gamePath = %q, want Game.exe", gotGamePath)
	}
}

func envContains(env []string, want string) bool {
	for _, item := range env {
		if item == want {
			return true
		}
	}
	return false
}

func TestLatestGameForRunPrefersReloadedSavedConfig(t *testing.T) {
	stale := &game.Game{
		Name:   "same-game",
		Runner: game.RunnerGamescope,
		GamescopeConfig: &gamescope.Options{
			Width: 1280,
		},
	}
	saved := &game.Game{
		Name:   "same-game",
		Runner: game.RunnerGamescope,
		GamescopeConfig: &gamescope.Options{
			Width: 480,
		},
	}

	b := &LiveBackend{games: []*game.Game{saved}}
	got := b.latestGameForRun(stale)

	if got != saved {
		t.Fatal("latestGameForRun did not return the saved config")
	}
	if got.GamescopeConfig == nil || got.GamescopeConfig.Width != 480 {
		t.Fatalf("Width = %v, want 480", got.GamescopeConfig)
	}
}

func TestGamePlaytimeIncludesPersistedAndActiveRun(t *testing.T) {
	startedAt := time.Now().Add(-10 * time.Minute)
	b := &LiveBackend{
		current:           &game.Game{Name: "test-game"},
		currentRun:        &gr.Process{Status: gr.StatusRunning},
		currentRunStarted: startedAt,
		activity: map[string]gameActivityEntry{
			"test-game": {PlaytimeSeconds: int64((2 * time.Hour) / time.Second)},
		},
	}

	got := b.GamePlaytime("test-game")
	if got < 2*time.Hour+9*time.Minute || got > 2*time.Hour+11*time.Minute {
		t.Fatalf("GamePlaytime() = %v, want roughly 2h10m", got)
	}
}

func TestFinalizeStoppedRunAddsPlaytimeOnce(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	startedAt := time.Now().Add(-5 * time.Minute)
	b := &LiveBackend{
		current:           &game.Game{Name: "test-game"},
		currentRun:        &gr.Process{Status: gr.StatusExited},
		currentRunStarted: startedAt,
		activity:          map[string]gameActivityEntry{},
	}

	if b.IsGameRunning() {
		t.Fatal("IsGameRunning() = true, want false for exited process")
	}
	first := b.GamePlaytime("test-game")
	second := b.GamePlaytime("test-game")
	if first < 4*time.Minute || first > 6*time.Minute {
		t.Fatalf("first GamePlaytime() = %v, want roughly 5m", first)
	}
	if second != first {
		t.Fatalf("second GamePlaytime() = %v, want %v", second, first)
	}
}

func writeZip(t *testing.T, path string, files map[string]string) {
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
