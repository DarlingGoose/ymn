package gui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Seann-Moser/wgl/pkg/util"
)

type Settings struct {
	ThemeMode                   string `json:"theme_mode,omitempty"`
	ThemePalette                string `json:"theme_palette,omitempty"`
	TranscriptTextSize          string `json:"transcript_text_size,omitempty"`
	VisibleTranscript           string `json:"visible_transcript,omitempty"`
	AutoPlayHighlightPopupAudio bool   `json:"auto_play_highlight_popup_audio,omitempty"`
}

func guiSettingsPath() string {
	return filepath.Join(util.ConfigBaseDir(), "gui-settings.json")
}

func LoadSettings() (Settings, error) {
	data, err := os.ReadFile(guiSettingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return Settings{}, nil
		}
		return Settings{}, fmt.Errorf("read gui settings: %w", err)
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, fmt.Errorf("decode gui settings: %w", err)
	}
	return settings, nil
}

func SaveSettings(settings Settings) error {
	if err := os.MkdirAll(util.ConfigBaseDir(), 0o755); err != nil {
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
