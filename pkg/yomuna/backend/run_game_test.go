package backend

import (
	"os"
	"path/filepath"
	"testing"

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
