package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/ymn/pkg/util"
)

type gameActivityEntry struct {
	LastPlayed      time.Time `json:"last_played,omitempty"`
	PlaytimeSeconds int64     `json:"playtime_seconds,omitempty"`
}

func (b *LiveBackend) GameLastPlayed(name string) time.Time {
	if b == nil {
		return time.Time{}
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return time.Time{}
	}
	b.activityMu.Lock()
	defer b.activityMu.Unlock()
	b.ensureActivityLoaded()
	return b.activity[name].LastPlayed
}

func (b *LiveBackend) GamePlaytime(name string) time.Duration {
	if b == nil {
		return 0
	}
	b.finalizeStoppedRunIfNeeded()
	name = strings.TrimSpace(name)
	if name == "" {
		return 0
	}

	b.activityMu.Lock()
	b.ensureActivityLoaded()
	total := time.Duration(b.activity[name].PlaytimeSeconds) * time.Second
	b.activityMu.Unlock()

	b.gameMu.RLock()
	currentName := ""
	if b.current != nil {
		currentName = strings.TrimSpace(b.current.Name)
	}
	startedAt := b.currentRunStarted
	running := b.currentRun != nil
	b.gameMu.RUnlock()

	if running && strings.EqualFold(currentName, name) && !startedAt.IsZero() {
		total += time.Since(startedAt).Round(time.Second)
	}
	return total
}

func (b *LiveBackend) recordGameStarted(name string, startedAt time.Time) {
	if b == nil {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	b.activityMu.Lock()
	defer b.activityMu.Unlock()
	b.ensureActivityLoaded()
	entry := b.activity[name]
	entry.LastPlayed = startedAt
	b.activity[name] = entry
	_ = saveGameActivity(b.activity)
}

func (b *LiveBackend) recordGameStopped(name string, startedAt, stoppedAt time.Time) {
	if b == nil {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" || startedAt.IsZero() || stoppedAt.Before(startedAt) {
		return
	}
	elapsed := stoppedAt.Sub(startedAt).Round(time.Second)
	if elapsed <= 0 {
		return
	}

	b.activityMu.Lock()
	defer b.activityMu.Unlock()
	b.ensureActivityLoaded()
	entry := b.activity[name]
	entry.PlaytimeSeconds += int64(elapsed / time.Second)
	if entry.LastPlayed.IsZero() || startedAt.After(entry.LastPlayed) {
		entry.LastPlayed = startedAt
	}
	b.activity[name] = entry
	_ = saveGameActivity(b.activity)
}

func (b *LiveBackend) finalizeStoppedRunIfNeeded() {
	if b == nil {
		return
	}
	b.gameMu.RLock()
	g := b.current
	proc := b.currentRun
	startedAt := b.currentRunStarted
	stopped := proc != nil && (proc.Status == gr.StatusExited || proc.Status == gr.StatusStopped)
	b.gameMu.RUnlock()
	if !stopped {
		return
	}

	b.gameMu.Lock()
	if b.currentRun == proc {
		b.currentRun = nil
		b.currentRunStarted = time.Time{}
	}
	b.gameMu.Unlock()
	if g != nil {
		b.recordGameStopped(g.Name, startedAt, time.Now())
	}
}

func (b *LiveBackend) ensureActivityLoaded() {
	if b.activity != nil {
		return
	}
	b.activity = loadGameActivity()
}

func loadGameActivity() map[string]gameActivityEntry {
	out := map[string]gameActivityEntry{}
	data, err := os.ReadFile(gameActivityPath())
	if err != nil {
		return out
	}
	if err := json.Unmarshal(data, &out); err == nil {
		return out
	}
	legacy := map[string]time.Time{}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return map[string]gameActivityEntry{}
	}
	for name, lastPlayed := range legacy {
		out[name] = gameActivityEntry{LastPlayed: lastPlayed}
	}
	return out
}

func saveGameActivity(activity map[string]gameActivityEntry) error {
	if err := os.MkdirAll(filepath.Dir(gameActivityPath()), 0o755); err != nil {
		return fmt.Errorf("create game activity dir: %w", err)
	}
	data, err := json.MarshalIndent(activity, "", "  ")
	if err != nil {
		return fmt.Errorf("encode game activity: %w", err)
	}
	if err := os.WriteFile(gameActivityPath(), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write game activity: %w", err)
	}
	return nil
}

func gameActivityPath() string {
	return filepath.Join(util.ConfigBaseDir(), "guiv2-game-activity.json")
}
