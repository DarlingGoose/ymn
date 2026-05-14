package transcript

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/unit"
	"github.com/DarlingGoose/ymn/pkg/util"
)

type transcriptPreferences struct {
	SelectedGameName     string  `json:"selected_game_name,omitempty"`
	LookupFontSizeSp     float32 `json:"lookup_font_size_sp,omitempty"`
	SentenceFontSizeSp   float32 `json:"sentence_font_size_sp,omitempty"`
	TranscriptFontSizeSp float32 `json:"transcript_font_size_sp,omitempty"`
	MaxTranscriptRows    int     `json:"max_transcript_rows,omitempty"`
	ShowLanguageOnly     bool    `json:"show_language_only,omitempty"`
}

func defaultTranscriptPreferences() transcriptPreferences {
	return transcriptPreferences{
		LookupFontSizeSp:     14,
		SentenceFontSizeSp:   24,
		TranscriptFontSizeSp: 22,
		MaxTranscriptRows:    200,
	}
}

func loadTranscriptPreferences() transcriptPreferences {
	prefs := defaultTranscriptPreferences()
	data, err := os.ReadFile(transcriptPreferencesPath())
	if err != nil {
		return prefs
	}
	if err := json.Unmarshal(data, &prefs); err != nil {
		return defaultTranscriptPreferences()
	}
	prefs.normalize()
	return prefs
}

func saveTranscriptPreferences(prefs transcriptPreferences) error {
	prefs.normalize()
	if err := os.MkdirAll(filepath.Dir(transcriptPreferencesPath()), 0o755); err != nil {
		return fmt.Errorf("create transcript preferences dir: %w", err)
	}
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode transcript preferences: %w", err)
	}
	if err := os.WriteFile(transcriptPreferencesPath(), data, 0o644); err != nil {
		return fmt.Errorf("write transcript preferences: %w", err)
	}
	return nil
}

func transcriptPreferencesPath() string {
	return filepath.Join(util.ConfigBaseDir(), "guiv2-transcript.json")
}

func (p *transcriptPreferences) normalize() {
	p.SelectedGameName = strings.TrimSpace(p.SelectedGameName)
	if p.LookupFontSizeSp < 11 || p.LookupFontSizeSp > 24 {
		p.LookupFontSizeSp = 14
	}
	if p.SentenceFontSizeSp < 16 || p.SentenceFontSizeSp > 40 {
		p.SentenceFontSizeSp = 24
	}
	if p.TranscriptFontSizeSp < 14 || p.TranscriptFontSizeSp > 34 {
		p.TranscriptFontSizeSp = 22
	}
	p.MaxTranscriptRows = clampTranscriptRowLimit(p.MaxTranscriptRows)
}

func spToFloat(value unit.Sp) float32 {
	return float32(value)
}
