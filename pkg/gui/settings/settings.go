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
	"github.com/DarlingGoose/wgl/pkg/translation"
	"github.com/DarlingGoose/wgl/pkg/util"
)

var _ gui.EvenHandler = &Settings{}

type Settings struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	settingsList                widget.List
	textSizeDropdown            bareui.Dropdown
	focusedTextSizeDropdown     bareui.Dropdown
	translateDetailSizeDropdown bareui.Dropdown
	translatorDropdown          bareui.Dropdown
	recentLinesDropdown         bareui.Dropdown
	furiganaDropdown            bareui.Dropdown
	textSizeOptions             []gui.DropdownOption
	focusedTextSizeOptions      []gui.DropdownOption
	translateDetailSizeOptions  []gui.DropdownOption
	translatorOptions           []gui.DropdownOption
	recentLineOptions           []gui.DropdownOption
	furiganaOptions             []gui.DropdownOption
	themeSelector               *barethemes.ThemeSelector

	selectedPaletteName     string
	selectedTextSizeName    string
	selectedFocusedSizeName string
	selectedDetailSizeName  string
	selectedTranslatorName  string
	selectedRecentLinesName string
	selectedFuriganaName    string
	focusedFuriganaMode     string
	autoPlayHighlightAudio  widget.Bool
	colorizeHighlightText   widget.Bool
	openAIAPIKeyEditor      widget.Editor
	openAIModelEditor       widget.Editor
	openAIBaseURLEditor     widget.Editor
	geminiAPIKeyEditor      widget.Editor
	geminiModelEditor       widget.Editor
	geminiBaseURLEditor     widget.Editor
	ollamaModelEditor       widget.Editor
	ollamaBaseURLEditor     widget.Editor
	compatibleAPIKeyEditor  widget.Editor
	compatibleModelEditor   widget.Editor
	compatibleBaseURLEditor widget.Editor
	saveTranslatorButton    widget.Clickable

	themeConfig             barethemes.Config
	systemDark              bool
	transcriptTextSize      unit.Sp
	focusedSentenceTextSize unit.Sp
	translateDetailTextSize unit.Sp
	recentLineLimit         atomic.Int64

	TranscriptTextSize          string `json:"transcript_text_size,omitempty"`
	FocusedSentenceTextSize     string `json:"focused_sentence_text_size,omitempty"`
	TranslateDetailTextSize     string `json:"translate_detail_text_size,omitempty"`
	TranslatorProvider          string `json:"translator_provider,omitempty"`
	OpenAIAPIKey                string `json:"openai_api_key,omitempty"`
	OpenAIModel                 string `json:"openai_model,omitempty"`
	OpenAIBaseURL               string `json:"openai_base_url,omitempty"`
	GeminiAPIKey                string `json:"gemini_api_key,omitempty"`
	GeminiModel                 string `json:"gemini_model,omitempty"`
	GeminiBaseURL               string `json:"gemini_base_url,omitempty"`
	OllamaModel                 string `json:"ollama_model,omitempty"`
	OllamaBaseURL               string `json:"ollama_base_url,omitempty"`
	OpenAICompatibleAPIKey      string `json:"openai_compatible_api_key,omitempty"`
	OpenAICompatibleModel       string `json:"openai_compatible_model,omitempty"`
	OpenAICompatibleBaseURL     string `json:"openai_compatible_base_url,omitempty"`
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
		selectedFocusedSizeName: "Medium",
		selectedDetailSizeName:  "Medium",
		selectedTranslatorName:  "Ollama",
		selectedRecentLinesName: "Last 200 Lines",
		selectedFuriganaName:    "Above",
		focusedFuriganaMode:     "above",
		themeConfig:             barethemes.DefaultConfig(),
		transcriptTextSize:      unit.Sp(16),
		focusedSentenceTextSize: unit.Sp(26),
		translateDetailTextSize: unit.Sp(15),
		TranscriptTextSize:      "Medium",
		FocusedSentenceTextSize: "Medium",
		TranslateDetailTextSize: "Medium",
		TranslatorProvider:      string(translation.ProviderOllama),
		OpenAIModel:             "gpt-4.1-mini",
		GeminiModel:             "gemini-2.5-flash",
		OllamaModel:             "qwen2.5:7b",
		OllamaBaseURL:           "http://localhost:11434",
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

func (g *Settings) initEditors() {
	editors := []*widget.Editor{
		&g.openAIAPIKeyEditor,
		&g.openAIModelEditor,
		&g.openAIBaseURLEditor,
		&g.geminiAPIKeyEditor,
		&g.geminiModelEditor,
		&g.geminiBaseURLEditor,
		&g.ollamaModelEditor,
		&g.ollamaBaseURLEditor,
		&g.compatibleAPIKeyEditor,
		&g.compatibleModelEditor,
		&g.compatibleBaseURLEditor,
	}
	for _, editor := range editors {
		editor.SingleLine = true
	}
	g.openAIAPIKeyEditor.Mask = '*'
	g.geminiAPIKeyEditor.Mask = '*'
	g.compatibleAPIKeyEditor.Mask = '*'
}

func (g *Settings) HandleEvents(gtx layout.Context, ctx context.Context, w *app.Window) {
	g.textSizeDropdown.Update(gtx)
	g.focusedTextSizeDropdown.Update(gtx)
	g.translateDetailSizeDropdown.Update(gtx)
	g.translatorDropdown.Update(gtx)
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

	for i := range g.focusedTextSizeOptions {
		opt := &g.focusedTextSizeOptions[i]
		for opt.Clickable.Clicked(gtx) {
			g.focusedSentenceTextSize = opt.TextSize
			g.selectedFocusedSizeName = opt.Label
			g.focusedTextSizeDropdown.Close()
			g.persistSettings()
		}
	}

	for i := range g.translateDetailSizeOptions {
		opt := &g.translateDetailSizeOptions[i]
		for opt.Clickable.Clicked(gtx) {
			g.translateDetailTextSize = opt.TextSize
			g.selectedDetailSizeName = opt.Label
			g.translateDetailSizeDropdown.Close()
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

	for i := range g.translatorOptions {
		opt := &g.translatorOptions[i]
		for opt.Clickable.Clicked(gtx) {
			g.TranslatorProvider = opt.Value
			g.selectedTranslatorName = opt.Label
			g.translatorDropdown.Close()
			g.persistSettings()
		}
	}

	for g.saveTranslatorButton.Clicked(gtx) {
		g.persistTranslatorSettings()
	}
}

func (g *Settings) applyLayout() {
	g.settingsList.Axis = layout.Vertical
	g.initEditors()
	gui.NewDropDownLayout(&g.textSizeDropdown, "mdi:format-size")
	gui.NewDropDownLayout(&g.focusedTextSizeDropdown, "mdi:format-title")
	gui.NewDropDownLayout(&g.translateDetailSizeDropdown, "mdi:format-text")
	gui.NewDropDownLayout(&g.translatorDropdown, "mdi:translate")
	gui.NewDropDownLayout(&g.recentLinesDropdown, "mdi:sort-clock-descending-outline")
	gui.NewDropDownLayout(&g.furiganaDropdown, "mdi:ruby")
	g.textSizeOptions = gui.NewTranscriptSizeOptions()
	g.focusedTextSizeOptions = gui.NewFocusedSentenceSizeOptions()
	g.translateDetailSizeOptions = gui.NewTranslateDetailSizeOptions()
	g.translatorOptions = gui.NewTranslatorProviderOptions()
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
			rows := g.settingsRows()
			return material.List(g.theme.Gio(), &g.settingsList).Layout(gtx, len(rows), func(gtx layout.Context, index int) layout.Dimensions {
				if index < 0 || index >= len(rows) {
					return layout.Dimensions{}
				}
				return layout.Inset{Bottom: unit.Dp(14)}.Layout(gtx, rows[index])
			})
		})
	})
}

func (g *Settings) settingsRows() []layout.Widget {
	rows := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			lbl := material.H5(g.theme.Gio(), "Appearance")
			lbl.Color = g.theme.Color.Text
			return lbl.Layout(gtx)
		},
		g.layoutBareThemeSelector,
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Transcript Size", g.selectedTextSizeName, func(gtx layout.Context) layout.Dimensions {
				return g.textSizeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedTextSizeName, func(gtx layout.Context) layout.Dimensions {
					return gui.LayoutOptionMenu(gtx, g.textSizeOptions, g.selectedTextSizeName, g.theme, g.iconify)
				})
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Focused Sentence Size", g.selectedFocusedSizeName, func(gtx layout.Context) layout.Dimensions {
				return g.focusedTextSizeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedFocusedSizeName, func(gtx layout.Context) layout.Dimensions {
					return gui.LayoutOptionMenu(gtx, g.focusedTextSizeOptions, g.selectedFocusedSizeName, g.theme, g.iconify)
				})
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Translate Detail Size", g.selectedDetailSizeName, func(gtx layout.Context) layout.Dimensions {
				return g.translateDetailSizeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedDetailSizeName, func(gtx layout.Context) layout.Dimensions {
					return gui.LayoutOptionMenu(gtx, g.translateDetailSizeOptions, g.selectedDetailSizeName, g.theme, g.iconify)
				})
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Visible Transcript", g.selectedRecentLinesName, func(gtx layout.Context) layout.Dimensions {
				return g.recentLinesDropdown.Layout(gtx, g.theme, g.iconify, g.selectedRecentLinesName, func(gtx layout.Context) layout.Dimensions {
					return gui.LayoutOptionMenu(gtx, g.recentLineOptions, g.selectedRecentLinesName, g.theme, g.iconify)
				})
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Focused Furigana", g.selectedFuriganaName, func(gtx layout.Context) layout.Dimensions {
				return g.furiganaDropdown.Layout(gtx, g.theme, g.iconify, g.selectedFuriganaName, func(gtx layout.Context) layout.Dimensions {
					return gui.LayoutOptionMenu(gtx, g.furiganaOptions, g.selectedFuriganaName, g.theme, g.iconify)
				})
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Highlight Audio", util.BoolSettingLabel(g.autoPlayHighlightAudio.Value), func(gtx layout.Context) layout.Dimensions {
				check := material.CheckBox(g.theme.Gio(), &g.autoPlayHighlightAudio, "Auto-play audio when a highlighted word is clicked")
				check.Color = g.theme.Color.Text
				return check.Layout(gtx)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Word Highlight Style", util.BoolSettingLabel(g.colorizeHighlightText.Value), func(gtx layout.Context) layout.Dimensions {
				check := material.CheckBox(g.theme.Gio(), &g.colorizeHighlightText, "Use colored vocab text instead of background highlights")
				check.Color = g.theme.Color.Text
				return check.Layout(gtx)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			lbl := material.H5(g.theme.Gio(), "Translator")
			lbl.Color = g.theme.Color.Text
			return lbl.Layout(gtx)
		},
		func(gtx layout.Context) layout.Dimensions {
			return g.layoutSettingRow(gtx, "Translator Provider", g.selectedTranslatorName, func(gtx layout.Context) layout.Dimensions {
				return g.translatorDropdown.Layout(gtx, g.theme, g.iconify, g.selectedTranslatorName, func(gtx layout.Context) layout.Dimensions {
					return gui.LayoutOptionMenu(gtx, g.translatorOptions, g.selectedTranslatorName, g.theme, g.iconify)
				})
			})
		},
	}
	rows = append(rows, g.translatorSettingRows()...)
	rows = append(rows, func(gtx layout.Context) layout.Dimensions {
		saveButton := bareui.Button{
			Clickable: &g.saveTranslatorButton,
			Text:      "Save Translator Settings",
			Prefix:    "mdi:content-save-outline",
			Variant:   bareui.ButtonPrimary,
		}
		return saveButton.Layout(gtx, g.theme, g.iconify)
	})
	return rows
}

func (g *Settings) translatorSettingRows() []layout.Widget {
	switch translation.Provider(g.TranslatorProvider) {
	case translation.ProviderOpenAI:
		return []layout.Widget{
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "OpenAI API Key", &g.openAIAPIKeyEditor, "sk-...")
			},
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "OpenAI Model", &g.openAIModelEditor, "gpt-4.1-mini")
			},
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "OpenAI Base URL", &g.openAIBaseURLEditor, "https://api.openai.com/v1")
			},
		}
	case translation.ProviderGemini:
		return []layout.Widget{
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Gemini API Key", &g.geminiAPIKeyEditor, "API key")
			},
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Gemini Model", &g.geminiModelEditor, "gemini-2.5-flash")
			},
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Gemini Base URL", &g.geminiBaseURLEditor, "https://generativelanguage.googleapis.com/v1beta")
			},
		}
	case translation.ProviderOpenAICompatible:
		return []layout.Widget{
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Compatible API Key", &g.compatibleAPIKeyEditor, "optional or required by endpoint")
			},
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Compatible Model", &g.compatibleModelEditor, "model name")
			},
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Compatible Base URL", &g.compatibleBaseURLEditor, "https://host/v1")
			},
		}
	default:
		return []layout.Widget{
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Ollama Model", &g.ollamaModelEditor, "qwen2.5:7b")
			},
			func(gtx layout.Context) layout.Dimensions {
				return g.layoutEditorRow(gtx, "Ollama Base URL", &g.ollamaBaseURLEditor, "http://localhost:11434")
			},
		}
	}
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

func (g *Settings) layoutEditorRow(gtx layout.Context, label string, editor *widget.Editor, hint string) layout.Dimensions {
	return bareutils.Panel(gtx, g.theme.Color.Background, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(g.theme.Gio(), label)
					lbl.Color = g.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerW(unit.Dp(14))),
				layout.Flexed(2, func(gtx layout.Context) layout.Dimensions {
					ed := material.Editor(g.theme.Gio(), editor, hint)
					ed.Color = g.theme.Color.Text
					ed.HintColor = g.theme.Color.TextMuted
					return ed.Layout(gtx)
				}),
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
	g.FocusedSentenceTextSize = g.selectedFocusedSizeName
	g.TranslateDetailTextSize = g.selectedDetailSizeName
	g.persistTranslatorFields()
	g.VisibleTranscript = g.selectedRecentLinesName
	g.FocusedFurigana = g.selectedFuriganaName
	g.AutoPlayHighlightPopupAudio = g.autoPlayHighlightAudio.Value
	g.ColorizeHighlightWords = g.colorizeHighlightText.Value
	_ = g.saveSettings()
}

func (g *Settings) persistTranslatorSettings() {
	g.persistTranslatorFields()
	_ = g.saveSettings()
}

func (g *Settings) persistTranslatorFields() {
	g.OpenAIAPIKey = g.openAIAPIKeyEditor.Text()
	g.OpenAIModel = g.openAIModelEditor.Text()
	g.OpenAIBaseURL = g.openAIBaseURLEditor.Text()
	g.GeminiAPIKey = g.geminiAPIKeyEditor.Text()
	g.GeminiModel = g.geminiModelEditor.Text()
	g.GeminiBaseURL = g.geminiBaseURLEditor.Text()
	g.OllamaModel = g.ollamaModelEditor.Text()
	g.OllamaBaseURL = g.ollamaBaseURLEditor.Text()
	g.OpenAICompatibleAPIKey = g.compatibleAPIKeyEditor.Text()
	g.OpenAICompatibleModel = g.compatibleModelEditor.Text()
	g.OpenAICompatibleBaseURL = g.compatibleBaseURLEditor.Text()
}

func (g *Settings) applySavedSettings() {
	for _, opt := range g.textSizeOptions {
		if opt.Label == g.TranscriptTextSize {
			g.transcriptTextSize = opt.TextSize
			g.selectedTextSizeName = opt.Label
			break
		}
	}

	for _, opt := range g.focusedTextSizeOptions {
		if opt.Label == g.FocusedSentenceTextSize {
			g.focusedSentenceTextSize = opt.TextSize
			g.selectedFocusedSizeName = opt.Label
			break
		}
	}

	for _, opt := range g.translateDetailSizeOptions {
		if opt.Label == g.TranslateDetailTextSize {
			g.translateDetailTextSize = opt.TextSize
			g.selectedDetailSizeName = opt.Label
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

	for _, opt := range g.translatorOptions {
		if opt.Value == g.TranslatorProvider || opt.Label == g.TranslatorProvider {
			g.TranslatorProvider = opt.Value
			g.selectedTranslatorName = opt.Label
			break
		}
	}

	g.openAIAPIKeyEditor.SetText(g.OpenAIAPIKey)
	g.openAIModelEditor.SetText(g.OpenAIModel)
	g.openAIBaseURLEditor.SetText(g.OpenAIBaseURL)
	g.geminiAPIKeyEditor.SetText(g.GeminiAPIKey)
	g.geminiModelEditor.SetText(g.GeminiModel)
	g.geminiBaseURLEditor.SetText(g.GeminiBaseURL)
	g.ollamaModelEditor.SetText(g.OllamaModel)
	g.ollamaBaseURLEditor.SetText(g.OllamaBaseURL)
	g.compatibleAPIKeyEditor.SetText(g.OpenAICompatibleAPIKey)
	g.compatibleModelEditor.SetText(g.OpenAICompatibleModel)
	g.compatibleBaseURLEditor.SetText(g.OpenAICompatibleBaseURL)

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

func (g *Settings) FocusedSentenceSize() unit.Sp {
	return g.focusedSentenceTextSize
}

func (g *Settings) FocusedSentenceSizeLabel() string {
	return g.selectedFocusedSizeName
}

func (g *Settings) TranslateDetailSize() unit.Sp {
	return g.translateDetailTextSize
}

func (g *Settings) TranslateDetailSizeLabel() string {
	return g.selectedDetailSizeName
}

func (g *Settings) TranslatorConfig() translation.Config {
	switch translation.Provider(g.TranslatorProvider) {
	case translation.ProviderOpenAI:
		return translation.Config{
			Provider: translation.ProviderOpenAI,
			APIKey:   g.OpenAIAPIKey,
			BaseURL:  g.OpenAIBaseURL,
			Model:    g.OpenAIModel,
		}
	case translation.ProviderGemini:
		return translation.Config{
			Provider: translation.ProviderGemini,
			APIKey:   g.GeminiAPIKey,
			BaseURL:  g.GeminiBaseURL,
			Model:    g.GeminiModel,
		}
	case translation.ProviderOpenAICompatible:
		return translation.Config{
			Provider: translation.ProviderOpenAICompatible,
			APIKey:   g.OpenAICompatibleAPIKey,
			BaseURL:  g.OpenAICompatibleBaseURL,
			Model:    g.OpenAICompatibleModel,
		}
	default:
		return translation.Config{
			Provider: translation.ProviderOllama,
			BaseURL:  g.OllamaBaseURL,
			Model:    g.OllamaModel,
		}
	}
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
