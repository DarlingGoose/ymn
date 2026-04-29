package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type guiSettings struct {
	ThemeMode          string `json:"theme_mode,omitempty"`
	ThemePalette       string `json:"theme_palette,omitempty"`
	TranscriptTextSize string `json:"transcript_text_size,omitempty"`
	VisibleTranscript  string `json:"visible_transcript,omitempty"`
}

func guiSettingsPath() string {
	return filepath.Join(configBaseDir(), "gui-settings.json")
}

func loadGUISettings() (guiSettings, error) {
	data, err := os.ReadFile(guiSettingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return guiSettings{}, nil
		}
		return guiSettings{}, fmt.Errorf("read gui settings: %w", err)
	}

	var settings guiSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return guiSettings{}, fmt.Errorf("decode gui settings: %w", err)
	}
	return settings, nil
}

func saveGUISettings(settings guiSettings) error {
	if err := os.MkdirAll(configBaseDir(), 0o755); err != nil {
		return fmt.Errorf("create gui settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode gui settings: %w", err)
	}

	if err := os.WriteFile(guiSettingsPath(), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write gui settings: %w", err)
	}
	return nil
}
