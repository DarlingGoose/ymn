package backend

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DarlingGoose/tr/pkg/textractor"
	"github.com/DarlingGoose/vntext/pkg/engine"
	"github.com/DarlingGoose/vntext/pkg/game"
)

const textHookHistoryFile = "yomuna-textractor-hooks.jsonl"

type textHookHistoryEntry struct {
	Time    time.Time `json:"time"`
	RawTime string    `json:"raw_time,omitempty"`
	Hook    string    `json:"hook,omitempty"`
	Speaker string    `json:"speaker,omitempty"`
	Text    string    `json:"text,omitempty"`
	Raw     string    `json:"raw,omitempty"`
	Voice   string    `json:"voice,omitempty"`
}

func (b *LiveBackend) recordGameTextHookHistory(ctx context.Context, g *game.Game, in <-chan engine.Line) chan engine.Line {
	out := make(chan engine.Line)
	go func() {
		defer close(out)
		var lastKey string
		haveLast := false
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-in:
				if !ok {
					return
				}
				key := consecutiveHookLineDedupeKey(line)
				if haveLast && key != "" && key == lastKey {
					continue
				}
				lastKey = key
				haveLast = true

				_ = b.appendGameTextHookHistory(g, line)
				select {
				case <-ctx.Done():
					return
				case out <- line:
				}
			}
		}
	}()
	return out
}

func consecutiveHookLineDedupeKey(line engine.Line) string {
	text := line.Text
	if text == "" {
		text = line.Raw
	}
	if text == "" {
		return ""
	}
	return strings.TrimSpace(line.Speaker) + "\x00" + text
}

func (b *LiveBackend) appendGameTextHookHistory(g *game.Game, line engine.Line) error {
	path, err := gameTextHookHistoryPath(g)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create hook history dir: %w", err)
	}

	when := line.Time
	if when.IsZero() {
		when = time.Now()
	}
	entry := textHookHistoryEntry{
		Time:    when,
		RawTime: strings.TrimSpace(line.RawTime),
		Hook:    strings.TrimSpace(line.Hook),
		Speaker: strings.TrimSpace(line.Speaker),
		Text:    strings.TrimSpace(line.Text),
		Raw:     line.Raw,
		Voice:   strings.TrimSpace(line.Voice),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode hook history: %w", err)
	}

	b.hookHistoryMu.Lock()
	defer b.hookHistoryMu.Unlock()

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open hook history: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write hook history: %w", err)
	}
	return nil
}

func (b *LiveBackend) ReadGameTextHookHistory(g *game.Game, filters []string, maxLines int) ([]engine.Line, error) {
	path, err := gameTextHookHistoryPath(g)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open hook history: %w", err)
	}
	defer f.Close()

	normalized := normalizeTextHookFilters(filters)
	out := make([]engine.Line, 0)
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		var entry textHookHistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if !textHookHistoryMatchesFilters(entry.Hook, normalized) {
			continue
		}
		out = append(out, engine.Line{
			Raw:     entry.Raw,
			Hook:    entry.Hook,
			Text:    entry.Text,
			Speaker: entry.Speaker,
			Time:    entry.Time,
			RawTime: entry.RawTime,
			Voice:   entry.Voice,
		})
		if maxLines > 0 && len(out) > maxLines {
			out = out[len(out)-maxLines:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read hook history: %w", err)
	}
	return out, nil
}

func textHookHistoryMatchesFilters(hook string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	hook = strings.TrimSpace(textractor.HookGroup(hook))
	if hook == "" {
		return false
	}
	for _, filter := range filters {
		if strings.EqualFold(filter, hook) {
			return true
		}
	}
	return false
}

func gameTextHookHistoryPath(g *game.Game) (string, error) {
	if g == nil {
		return "", fmt.Errorf("game is required")
	}
	dir := strings.TrimSpace(g.WorkingDir)
	if dir == "" {
		dir = strings.TrimSpace(g.GamePath)
	}
	if dir == "" && strings.TrimSpace(g.Executable) != "" {
		dir = filepath.Dir(g.Executable)
	}
	if dir == "" {
		return "", fmt.Errorf("game directory is required")
	}
	return filepath.Join(dir, textHookHistoryFile), nil
}
