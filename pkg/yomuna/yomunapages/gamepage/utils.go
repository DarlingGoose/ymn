package gamepage

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/input"
)

// String Utils
func splitCamel(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func splitCommaList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func runnerLabel(r game.RunnerType) string {
	switch r {
	case game.RunnerWine:
		return "Wine"
	case game.RunnerGamescope:
		return "Gamescope"
	case game.RunnerProton:
		return "Proton"
	case game.RunnerSteam:
		return "Steam"
	default:
		return "Gamescope"
	}
}

func validateRunnerPath(r game.RunnerType, path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	name := strings.ToLower(filepath.Base(path))
	switch r {
	case game.RunnerGamescope:
		if strings.Contains(name, "wine") || strings.Contains(name, "proton") {
			return fmt.Errorf("runner executable must point to gamescope for Gamescope configs, not %s", filepath.Base(path))
		}
	case game.RunnerWine:
		if strings.Contains(name, "gamescope") {
			return fmt.Errorf("runner executable must point to wine for Wine configs, not %s", filepath.Base(path))
		}
	}
	return nil
}

func nonNegativeIntegerRule(name string) input.Rule {
	_ = name
	return func(text string) error {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return fmt.Errorf("enter a whole number")
		}
		if value < 0 {
			return fmt.Errorf("enter 0 or a positive whole number")
		}
		return nil
	}
}

//Game Utils

func cloneGame(g *game.Game) *game.Game {
	if g == nil {
		return nil
	}
	cp := *g
	cp.EnvVars = append([]game.EnvVar(nil), g.EnvVars...)
	cp.TextHookFilter = append([]string(nil), g.TextHookFilter...)
	if g.WineConfig != nil {
		cfg := *g.WineConfig
		cp.WineConfig = &cfg
	}
	if g.GamescopeConfig != nil {
		cfg := *g.GamescopeConfig
		cfg.ExtraArgs = append([]string(nil), g.GamescopeConfig.ExtraArgs...)
		cp.GamescopeConfig = &cfg
	}
	return &cp
}
