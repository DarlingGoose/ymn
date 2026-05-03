package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareui "github.com/DarlingGoose/bare/pkg/ui"
	"github.com/DarlingGoose/bare/pkg/ui/icons"
	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/wgl/pkg/gui"
	"github.com/DarlingGoose/wgl/pkg/util"
)

var _ gui.EvenHandler = &Settings{}

type Settings struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	textSizeDropdown    bareui.Dropdown
	recentLinesDropdown bareui.Dropdown
	furiganaDropdown    bareui.Dropdown
	textSizeOptions     []gui.DropdownOption
	recentLineOptions   []gui.DropdownOption
	furiganaOptions     []gui.DropdownOption
	themeSelector       *barethemes.ThemeSelector

	selectedPaletteName     string
	selectedTextSizeName    string
	selectedRecentLinesName string
	selectedFuriganaName    string
	focusedFuriganaMode     string
	autoPlayHighlightAudio  widget.Bool
	colorizeHighlightText   widget.Bool

	themeConfig        barethemes.Config
	systemDark         bool
	transcriptTextSize unit.Sp
	recentLineLimit    atomic.Int64

	TranscriptTextSize          string `json:"transcript_text_size,omitempty"`
	VisibleTranscript           string `json:"visible_transcript,omitempty"`
	FocusedFurigana             string `json:"focused_furigana,omitempty"`
	AutoPlayHighlightPopupAudio bool   `json:"auto_play_highlight_popup_audio,omitempty"`
	ColorizeHighlightWords      bool   `json:"colorize_highlight_text,omitempty"`
	LastSelectedGame            string `json:"last_selected_game,omitempty"`
}

func defaultSettings() Settings {
	return Settings{
		selectedPaletteName:     "Moonlit Library",
		selectedTextSizeName:    "Medium",
		selectedRecentLinesName: "Last 200 Lines",
		selectedFuriganaName:    "Above",
		focusedFuriganaMode:     "above",
		themeConfig:             barethemes.DefaultConfig(),
		transcriptTextSize:      unit.Sp(16),
		TranscriptTextSize:      "Medium",
		VisibleTranscript:       "Last 200 Lines",
		FocusedFurigana:         "Above",
	}
}
func LoadSettings() (*Settings, error) {
	settings := defaultSettings()
	data, err := os.ReadFile(guiSettingsPath())
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return nil, fmt.Errorf("decode gui settings: %w", err)
		}
	}
	err = barethemes.LoadCustomThemes()
	if err != nil {
		slog.Error("failed loading custom themes", "err", err)
	}
	config, err := barethemes.LoadConfig()
	if err != nil {
		slog.Error("failed loading theme config", "err", err)
		config = barethemes.DefaultConfig()
	}
	settings.themeConfig = config
	settings.themeSelector = barethemes.NewThemeSelector()

	settings.applyLayout()
	settings.applySavedSettings()
	settings.applyTheme()
	return &settings, nil
}

func (g *Settings) HandleEvents(gtx layout.Context, ctx context.Context, w *app.Window) {
	g.textSizeDropdown.Update(gtx)
	g.recentLinesDropdown.Update(gtx)
	g.furiganaDropdown.Update(gtx)
	if g.autoPlayHighlightAudio.Update(gtx) {
		g.persistSettings()
	}
	if g.colorizeHighlightText.Update(gtx) {
		g.persistSettings()
	}

	for i := range g.textSizeOptions {
		opt := &g.textSizeOptions[i]
		for opt.Clickable.Clicked(gtx) {
			g.transcriptTextSize = opt.TextSize
			g.selectedTextSizeName = opt.Label
			g.textSizeDropdown.Close()
			g.persistSettings()
		}
	}

	for i := range g.recentLineOptions {
		opt := &g.recentLineOptions[i]
		for opt.Clickable.Clicked(gtx) {
			g.recentLineLimit.Swap(int64(opt.RecentLineLimit))
			g.selectedRecentLinesName = opt.Label
			g.recentLinesDropdown.Close()
			g.persistSettings()
		}
	}

	for i := range g.furiganaOptions {
		opt := &g.furiganaOptions[i]
		for opt.Clickable.Clicked(gtx) {
			g.focusedFuriganaMode = opt.Value
			g.selectedFuriganaName = opt.Label
			g.furiganaDropdown.Close()
			g.persistSettings()
		}
	}
}

func (g *Settings) applyLayout() {
	gui.NewDropDownLayout(&g.textSizeDropdown, "mdi:format-size")
	gui.NewDropDownLayout(&g.recentLinesDropdown, "mdi:sort-clock-descending-outline")
	gui.NewDropDownLayout(&g.furiganaDropdown, "mdi:ruby")
	g.textSizeOptions = gui.NewTranscriptSizeOptions()
	g.recentLineOptions = gui.NewRecentLineOptions()
	g.furiganaOptions = gui.NewFuriganaModeOptions()
	if g.themeSelector == nil {
		g.themeSelector = barethemes.NewThemeSelector()
	}
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
	if g.iconify == nil {
		g.iconify = icons.NewIconify() //terrible please set elseware but fine as a fallback
	}
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
					return g.layoutBareThemeSelector(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Transcript Size", g.selectedTextSizeName, func(gtx layout.Context) layout.Dimensions {
						return g.textSizeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedTextSizeName, func(gtx layout.Context) layout.Dimensions {
							return gui.LayoutOptionMenu(gtx, g.textSizeOptions, g.selectedTextSizeName, g.theme, g.iconify)
						})
					})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Visible Transcript", g.selectedRecentLinesName, func(gtx layout.Context) layout.Dimensions {
						return g.recentLinesDropdown.Layout(gtx, g.theme, g.iconify, g.selectedRecentLinesName, func(gtx layout.Context) layout.Dimensions {
							return gui.LayoutOptionMenu(gtx, g.recentLineOptions, g.selectedRecentLinesName, g.theme, g.iconify)
						})
					})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Focused Furigana", g.selectedFuriganaName, func(gtx layout.Context) layout.Dimensions {
						return g.furiganaDropdown.Layout(gtx, g.theme, g.iconify, g.selectedFuriganaName, func(gtx layout.Context) layout.Dimensions {
							return gui.LayoutOptionMenu(gtx, g.furiganaOptions, g.selectedFuriganaName, g.theme, g.iconify)
						})
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
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutSettingRow(gtx, "Word Highlight Style", util.BoolSettingLabel(g.colorizeHighlightText.Value), func(gtx layout.Context) layout.Dimensions {
						check := material.CheckBox(g.theme.Gio(), &g.colorizeHighlightText, "Use colored vocab text instead of background highlights")
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

func (g *Settings) layoutBareThemeSelector(gtx layout.Context) layout.Dimensions {
	if g.themeSelector == nil {
		g.themeSelector = barethemes.NewThemeSelector()
	}
	nextTheme, dims := g.themeSelector.LayoutThemeSelector(gtx, g.theme, g.systemDark)
	if nextTheme.Mode != g.theme.Mode || nextTheme.Palette != g.theme.Palette {
		g.theme = nextTheme
		g.themeConfig = barethemes.ConfigFromTheme(nextTheme)
		if err := barethemes.SaveConfig(g.themeConfig); err != nil {
			slog.Error("failed saving theme config", "err", err)
		}
	}
	return dims
}

func (g *Settings) applyTheme() *Settings {
	g.theme = g.themeConfig.Theme(g.systemDark)
	return g
}

func (g *Settings) persistSettings() {
	g.TranscriptTextSize = g.selectedTextSizeName
	g.VisibleTranscript = g.selectedRecentLinesName
	g.FocusedFurigana = g.selectedFuriganaName
	g.AutoPlayHighlightPopupAudio = g.autoPlayHighlightAudio.Value
	g.ColorizeHighlightWords = g.colorizeHighlightText.Value
	_ = g.saveSettings()
}

func (g *Settings) applySavedSettings() {
	for _, opt := range g.textSizeOptions {
		if opt.Label == g.TranscriptTextSize {
			g.transcriptTextSize = opt.TextSize
			g.selectedTextSizeName = opt.Label
			break
		}
	}

	for _, opt := range g.recentLineOptions {
		if opt.Label == g.VisibleTranscript {
			g.recentLineLimit.Swap(int64(opt.RecentLineLimit))
			g.selectedRecentLinesName = opt.Label
			break
		}
	}

	for _, opt := range g.furiganaOptions {
		if opt.Label == g.FocusedFurigana || opt.Value == g.FocusedFurigana {
			g.focusedFuriganaMode = opt.Value
			g.selectedFuriganaName = opt.Label
			break
		}
	}

	g.autoPlayHighlightAudio.Value = g.AutoPlayHighlightPopupAudio
	g.colorizeHighlightText.Value = g.ColorizeHighlightWords
}

func (g *Settings) Theme() barethemes.Theme {
	return g.theme
}

func (g *Settings) TranscriptSize() unit.Sp {
	return g.transcriptTextSize
}

func (g *Settings) TranscriptSizeLabel() string {
	return g.selectedTextSizeName
}

func (g *Settings) RecentLineLimit() int {
	return int(g.recentLineLimit.Load())
}

func (g *Settings) RecentLineLabel() string {
	return g.selectedRecentLinesName
}

func (g *Settings) FocusedFuriganaMode() string {
	return g.focusedFuriganaMode
}

func (g *Settings) FocusedFuriganaLabel() string {
	return g.selectedFuriganaName
}

func (g *Settings) AutoPlayHighlightAudio() bool {
	return g.autoPlayHighlightAudio.Value
}

func (g *Settings) ColorizeHighlightText() bool {
	return g.colorizeHighlightText.Value
}

func (g *Settings) LastGame() string {
	return g.LastSelectedGame
}

func (g *Settings) SetLastGame(name string) error {
	g.LastSelectedGame = name
	return g.saveSettings()
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
