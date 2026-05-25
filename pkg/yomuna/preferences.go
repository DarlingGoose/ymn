package yomuna

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DarlingGoose/ymn/pkg/util"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/notifications"
)

type appPreferences struct {
	NotificationLevel string `json:"notification_level,omitempty"`
	StartupTab        string `json:"startup_tab,omitempty"`
}

func defaultAppPreferences() appPreferences {
	return appPreferences{
		NotificationLevel: notifications.LevelValue(notifications.NotificationTypeInfo),
		StartupTab:        "games",
	}
}

func loadAppPreferences() appPreferences {
	prefs := defaultAppPreferences()
	data, err := os.ReadFile(appPreferencesPath())
	if err != nil {
		return prefs
	}
	if err := json.Unmarshal(data, &prefs); err != nil {
		return defaultAppPreferences()
	}
	prefs.normalize()
	return prefs
}

func saveAppPreferences(prefs appPreferences) error {
	prefs.normalize()
	if err := os.MkdirAll(filepath.Dir(appPreferencesPath()), 0o755); err != nil {
		return fmt.Errorf("create app preferences dir: %w", err)
	}
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode app preferences: %w", err)
	}
	if err := os.WriteFile(appPreferencesPath(), data, 0o644); err != nil {
		return fmt.Errorf("write app preferences: %w", err)
	}
	return nil
}

func appPreferencesPath() string {
	return filepath.Join(util.ConfigBaseDir(), "guiv2-app.json")
}

func (p *appPreferences) normalize() {
	level, ok := notifications.ParseLevel(p.NotificationLevel)
	if !ok {
		level = notifications.NotificationTypeInfo
	}
	p.NotificationLevel = notifications.LevelValue(level)
	p.StartupTab = normalizeStartupTab(p.StartupTab)
}

func (p appPreferences) notificationLevel() notifications.NotificationType {
	level, ok := notifications.ParseLevel(p.NotificationLevel)
	if !ok {
		return notifications.NotificationTypeInfo
	}
	return level
}

func normalizeStartupTab(tab string) string {
	switch tab {
	case "games", "translation", "transcript", "flashcards", "game", "add-game", "settings":
		return tab
	default:
		return "games"
	}
}
