package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareui "github.com/Seann-Moser/bare/pkg/ui"
	"github.com/Seann-Moser/bare/pkg/ui/icons"
	barethemes "github.com/Seann-Moser/bare/pkg/ui/themes"
	bareutils "github.com/Seann-Moser/bare/pkg/ui/utils"
	"github.com/Seann-Moser/wgl/pkg/gui"
	"github.com/Seann-Moser/wgl/pkg/util"
)

var _ gui.EvenHandler = &Settings{}

type Settings struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	modeDropdown        bareui.Dropdown
	paletteDropdown     bareui.Dropdown
	textSizeDropdown    bareui.Dropdown
	recentLinesDropdown bareui.Dropdown

	selectedModeName        string
	selectedPaletteName     string
	selectedTextSizeName    string
	selectedRecentLinesName string
	autoPlayHighlightAudio  widget.Bool

	themeMode          barethemes.Mode
	themePalette       barethemes.PaletteName
	systemDark         bool
	transcriptTextSize unit.Sp
	recentLineLimit    atomic.Int64

	ThemeMode                   string `json:"theme_mode,omitempty"`
	ThemePalette                string `json:"theme_palette,omitempty"`
	TranscriptTextSize          string `json:"transcript_text_size,omitempty"`
	VisibleTranscript           string `json:"visible_transcript,omitempty"`
	AutoPlayHighlightPopupAudio bool   `json:"auto_play_highlight_popup_audio,omitempty"`
}

func defaultSettings() Settings {
	return Settings{
		selectedModeName:        "Dark",
		selectedPaletteName:     "Ocean",
		selectedTextSizeName:    "Medium",
		selectedRecentLinesName: "All Lines",
		themeMode:               barethemes.ModeDark,
		themePalette:            barethemes.PaletteOcean,
		transcriptTextSize:      unit.Sp(16),
	}
}
func LoadSettings() (*Settings, error) {
	var settings Settings
	data, err := os.ReadFile(guiSettingsPath())
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return nil, fmt.Errorf("decode gui settings: %w", err)
		}
	} else {
		settings = defaultSettings()
	}
	settings.applyLayout()
	settings.applyTheme()
	return &settings, nil
}

func (g *Settings) HandleEvents(gtx layout.Context, ctx context.Context, w *app.Window) {
	g.modeDropdown.Update(gtx)
	g.paletteDropdown.Update(gtx)
	g.textSizeDropdown.Update(gtx)
	g.recentLinesDropdown.Update(gtx)
	if g.autoPlayHighlightAudio.Update(gtx) {
		g.persistSettings()
	}

	for _, op := range gui.NewModeOptions() {
		opt := &op
		for opt.Clickable.Clicked(gtx) {
			g.themeMode = opt.Mode
			g.selectedModeName = opt.Label
			g.modeDropdown.Close()
			g.applyTheme()
			g.persistSettings()
		}
	}

	for _, op := range gui.NewPaletteOptions() {
		opt := &op
		for opt.Clickable.Clicked(gtx) {
			g.themePalette = opt.Palette
			g.selectedPaletteName = opt.Label
			g.paletteDropdown.Close()
			g.applyTheme()
			g.persistSettings()
		}
	}

	for _, op := range gui.NewTranscriptSizeOptions() {
		opt := &op
		for opt.Clickable.Clicked(gtx) {
			g.transcriptTextSize = opt.TextSize
			g.selectedTextSizeName = opt.Label
			g.textSizeDropdown.Close()
			g.persistSettings()
		}
	}

	for _, op := range gui.NewRecentLineOptions() {
		opt := &op
		for opt.Clickable.Clicked(gtx) {
			g.recentLineLimit.Swap(int64(opt.RecentLineLimit))
			g.selectedRecentLinesName = opt.Label
			g.recentLinesDropdown.Close()
			g.persistSettings()
		}
	}
}

func (g *Settings) applyLayout() {
	gui.NewDropDownLayout(&g.modeDropdown, "mdi:theme-light-dark")
	gui.NewDropDownLayout(&g.paletteDropdown, "mdi:palette-outline")
	gui.NewDropDownLayout(&g.textSizeDropdown, "mdi:format-size")
	gui.NewDropDownLayout(&g.recentLinesDropdown, "mdi:sort-clock-descending-outline")
}

func (g *Settings) WithIcon(icon *icons.Iconify) *Settings {
	g.iconify = icon
	return g
}

func (g *Settings) WithTheme(theme barethemes.Theme) *Settings {
	g.theme = theme
	return g
}

func (g *Settings) LayoutPage(gtx layout.Context) layout.Dimensions {
	return bareutils.Panel(gtx, g.theme.Color.Surface, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(g.theme.Gio(), "Appearance")
					lbl.Color = g.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Mode", g.selectedModeName,
						func(gtx layout.Context) layout.Dimensions {
							return g.modeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedModeName,
								gui.LayoutOptionMenu(gtx,
									gui.NewModeOptions(),
									g.selectedModeName,
									g.theme,
									g.iconify,
								),
							)
						})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Palette", g.selectedPaletteName, func(gtx layout.Context) layout.Dimensions {
						return g.paletteDropdown.Layout(gtx, g.theme, g.iconify, g.selectedPaletteName, gui.LayoutOptionMenu(gtx,
							gui.NewPaletteOptions(),
							g.selectedPaletteName,
							g.theme,
							g.iconify,
						))
					})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Transcript Size", g.selectedTextSizeName, func(gtx layout.Context) layout.Dimensions {
						return g.textSizeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedTextSizeName, gui.LayoutOptionMenu(gtx,
							gui.NewTranscriptSizeOptions(),
							g.selectedTextSizeName,
							g.theme,
							g.iconify,
						))
					})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Visible Transcript", g.selectedRecentLinesName, func(gtx layout.Context) layout.Dimensions {
						return g.recentLinesDropdown.Layout(gtx, g.theme, g.iconify, g.selectedRecentLinesName, gui.LayoutOptionMenu(gtx,
							gui.NewRecentLineOptions(),
							g.selectedRecentLinesName,
							g.theme,
							g.iconify,
						))
					})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Highlight Audio", util.BoolSettingLabel(g.autoPlayHighlightAudio.Value), func(gtx layout.Context) layout.Dimensions {
						check := material.CheckBox(g.theme.Gio(), &g.autoPlayHighlightAudio, "Auto-play audio when a highlighted word is clicked")
						check.Color = g.theme.Color.Text
						return check.Layout(gtx)
					})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(g.theme.Gio(), "Mode, transcript rendering, and highlight click behavior can be tuned without changing the watcher logic.")
					lbl.Color = g.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (g *Settings) layoutSettingRow(gtx layout.Context, label, current string, control layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, g.theme.Color.Background, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Vertical,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(g.theme.Gio(), label)
							lbl.Color = g.theme.Color.TextMuted
							return lbl.Layout(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(g.theme.Gio(), current)
							lbl.Color = g.theme.Color.Text
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(control),
			)
		})
	})
}

func (g *Settings) applyTheme() *Settings {
	g.theme = barethemes.New(g.themeMode, g.themePalette, g.systemDark)
	return g
}

func (g *Settings) persistSettings() {
	g.ThemeMode = g.selectedModeName
	g.ThemePalette = g.selectedPaletteName
	g.TranscriptTextSize = g.selectedTextSizeName
	g.VisibleTranscript = g.selectedRecentLinesName
	g.AutoPlayHighlightPopupAudio = g.autoPlayHighlightAudio.Value
	_ = g.saveSettings()
}

func (g *Settings) applySavedSettings() {
	settings, err := LoadSettings()
	if err != nil {
		return
	}

	for _, opt := range gui.NewModeOptions() {
		if opt.Label == settings.ThemeMode {
			g.themeMode = opt.Mode
			g.selectedModeName = opt.Label
			break
		}
	}

	for _, opt := range gui.NewPaletteOptions() {
		if opt.Label == settings.ThemePalette {
			g.themePalette = opt.Palette
			g.selectedPaletteName = opt.Label
			break
		}
	}

	for _, opt := range gui.NewTranscriptSizeOptions() {
		if opt.Label == settings.TranscriptTextSize {
			g.transcriptTextSize = opt.TextSize
			g.selectedTextSizeName = opt.Label
			break
		}
	}

	for _, opt := range gui.NewRecentLineOptions() {
		if opt.Label == settings.VisibleTranscript {
			g.recentLineLimit.Swap(int64(opt.RecentLineLimit))
			g.selectedRecentLinesName = opt.Label
			break
		}
	}

	g.autoPlayHighlightAudio.Value = settings.AutoPlayHighlightPopupAudio
}

func (g *Settings) saveSettings() error {
	if err := os.MkdirAll(util.ConfigBaseDir(), 0o755); err != nil {
		return fmt.Errorf("create gui settings directory: %w", err)
	}

	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return fmt.Errorf("encode gui settings: %w", err)
	}

	if err := os.WriteFile(guiSettingsPath(), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write gui settings: %w", err)
	}
	return nil
}

func guiSettingsPath() string {
	return filepath.Join(util.ConfigBaseDir(), "gui-settings.json")
}
