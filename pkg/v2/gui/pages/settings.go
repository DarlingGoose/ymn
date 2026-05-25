package pages

import (
	"fmt"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/vntext/pkg/gameConfig"
	"github.com/DarlingGoose/ymn/pkg/util"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/input"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/row"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/notifications"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

type SettingsUI struct {
	th            *material.Theme
	ModeToggle    *toggles.ThemeModeToggle
	ThemeDropdown *dropdowns.ThemeDropdown
	settingsList  widget.List
	invalidate    func()

	transcriptSettings     *TranscriptSettings
	targetLanguageDropdown *dropdowns.Dropdown
	transcriptDown         widget.Clickable
	transcriptUp           widget.Clickable
	sentenceDown           widget.Clickable
	sentenceUp             widget.Clickable
	lookupDown             widget.Clickable
	lookupUp               widget.Clickable
	maxLinesDown           widget.Clickable
	maxLinesUp             widget.Clickable
	saveTranscript         widget.Clickable
	saveTranslator         widget.Clickable
	status                 string
	translatorStatus       string

	translatorSettings *TranslatorSettings
	ollamaModelInput   *input.TextInput
	ollamaBaseURLInput *input.TextInput

	notificationSettings *NotificationSettings
	notificationDropdown *dropdowns.Dropdown
	saveNotifications    widget.Clickable
	notificationStatus   string

	appSettings     *AppSettings
	startupDropdown *dropdowns.Dropdown
	saveApp         widget.Clickable
	appStatus       string

	theme    *theme.Client
	rowCache map[string]*row.Row

	storageSizes   map[string]storageSizeState
	storageResults chan storageSizeResult
}

type storageSizeState struct {
	size       int64
	err        error
	exists     bool
	pending    bool
	measuredAt time.Time
}

type storageSizeResult struct {
	path   string
	size   int64
	err    error
	exists bool
}

type TranscriptSettings struct {
	SelectedGameName     func() string
	TargetLanguage       func() string
	TranscriptFont       func() unit.Sp
	SentenceFont         func() unit.Sp
	LookupFont           func() unit.Sp
	MaxTranscriptRows    func() int
	SetTranscriptFont    func(unit.Sp)
	SetSentenceFont      func(unit.Sp)
	SetLookupFont        func(unit.Sp)
	SetMaxTranscriptRows func(int)
	SetTargetLanguage    func(string)
	Save                 func() error
}

type TranslatorSettings struct {
	OllamaModel      func() string
	OllamaBaseURL    func() string
	SetOllamaModel   func(string)
	SetOllamaBaseURL func(string)
	Save             func() error
}

type NotificationSettings struct {
	Level    func() notifications.NotificationType
	SetLevel func(notifications.NotificationType)
	Save     func() error
}

type AppSettings struct {
	StartupPage    func() string
	SetStartupPage func(string)
	Save           func() error
}

func NewSettingsUI(tc *theme.Client) *SettingsUI {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	ui := &SettingsUI{
		th:             material.NewTheme(),
		theme:          tc,
		rowCache:       make(map[string]*row.Row),
		storageSizes:   make(map[string]storageSizeState),
		storageResults: make(chan storageSizeResult, 16),
		ModeToggle:     toggles.NewThemeModeToggle(tc),
		ThemeDropdown:  dropdowns.NewThemeDropdown(tc),
		settingsList:   widget.List{List: layout.List{Axis: layout.Vertical}},
		ollamaModelInput: input.NewTextInput("Model", "translategemma:4b").
			WithThemeClient(tc),
		ollamaBaseURLInput: input.NewTextInput("Endpoint", "http://localhost:11434").
			WithThemeClient(tc),
		notificationDropdown: dropdowns.NewDropdown(notificationLevelItems()).
			WithThemeClient(tc).
			WithRole(theme.TextRoleLabel).
			WithMenuAbove(),
		startupDropdown: dropdowns.NewDropdown(startupPageItems()).
			WithThemeClient(tc).
			WithRole(theme.TextRoleLabel),
		targetLanguageDropdown: dropdowns.NewDropdown(translationLanguageItems()).
			WithThemeClient(tc).
			WithRole(theme.TextRoleLabel),
	}
	ui.ollamaModelInput.LeadingIcon = "lucide:brain"
	ui.ollamaBaseURLInput.LeadingIcon = "lucide:server"
	ui.ollamaBaseURLInput.Normalize = func(text string) string {
		return strings.TrimRight(strings.TrimSpace(text), "/")
	}
	return ui
}

func (ui *SettingsUI) WithAppSettings(settings *AppSettings) *SettingsUI {
	if ui == nil {
		return ui
	}
	ui.appSettings = settings
	ui.syncStartupSelection()
	if ui.startupDropdown != nil {
		ui.startupDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
			if !valid {
				return
			}
			if ui.appSettings != nil && ui.appSettings.SetStartupPage != nil {
				ui.appSettings.SetStartupPage(item.Value)
			}
			ui.appStatus = "Unsaved application changes"
		})
	}
	return ui
}

func (ui *SettingsUI) WithTranscriptSettings(settings *TranscriptSettings) *SettingsUI {
	if ui == nil {
		return ui
	}
	ui.transcriptSettings = settings
	ui.syncTargetLanguageSelection()
	if ui.targetLanguageDropdown != nil {
		ui.targetLanguageDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
			if !valid {
				return
			}
			if ui.transcriptSettings != nil && ui.transcriptSettings.SetTargetLanguage != nil {
				ui.transcriptSettings.SetTargetLanguage(item.Value)
			}
			ui.status = "Unsaved transcript changes"
		})
	}
	return ui
}

func (ui *SettingsUI) WithTranslatorSettings(settings *TranslatorSettings) *SettingsUI {
	if ui == nil {
		return ui
	}
	ui.translatorSettings = settings
	ui.syncTranslatorInputs()
	if ui.ollamaModelInput != nil {
		ui.ollamaModelInput.OnChange = func(text string) {
			ui.translatorStatus = "Unsaved translator changes"
			if ui.translatorSettings != nil && ui.translatorSettings.SetOllamaModel != nil {
				ui.translatorSettings.SetOllamaModel(strings.TrimSpace(text))
			}
		}
	}
	if ui.ollamaBaseURLInput != nil {
		ui.ollamaBaseURLInput.OnChange = func(text string) {
			ui.translatorStatus = "Unsaved translator changes"
			if ui.translatorSettings != nil && ui.translatorSettings.SetOllamaBaseURL != nil {
				ui.translatorSettings.SetOllamaBaseURL(strings.TrimRight(strings.TrimSpace(text), "/"))
			}
		}
	}
	return ui
}

func (ui *SettingsUI) WithNotificationSettings(settings *NotificationSettings) *SettingsUI {
	if ui == nil {
		return ui
	}
	ui.notificationSettings = settings
	ui.syncNotificationSelection()
	if ui.notificationDropdown != nil {
		ui.notificationDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
			if !valid {
				return
			}
			level, ok := notifications.ParseLevel(item.Value)
			if !ok {
				return
			}
			if ui.notificationSettings != nil && ui.notificationSettings.SetLevel != nil {
				ui.notificationSettings.SetLevel(level)
			}
			ui.notificationStatus = "Unsaved notification changes"
		})
	}
	return ui
}

func (ui *SettingsUI) WithInvalidate(invalidate func()) *SettingsUI {
	if ui == nil {
		return ui
	}
	ui.invalidate = invalidate
	return ui
}

func (ui *SettingsUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}

	if ui.drainStorageSizeResults() {
		gtx.Execute(op.InvalidateCmd{})
	}

	if ui.ModeToggle.Update(gtx) {
		gtx.Execute(op.InvalidateCmd{})
	}

	items := []layout.Widget{
		ui.ModeToggle.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ui.ThemeDropdown.Layout(gtx, layer)
		},
		func(gtx layout.Context) layout.Dimensions {
			return ui.layoutAppSettings(gtx, layer)
		},
		func(gtx layout.Context) layout.Dimensions {
			return ui.layoutTranscriptSettings(gtx, layer)
		},
		ui.layoutTranslatorSettings,
		func(gtx layout.Context) layout.Dimensions {
			return ui.layoutNotificationSettings(gtx, layer)
		},
		ui.layoutStorageLocations,
	}

	return layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		list := material.List(ui.th, &ui.settingsList)
		return list.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
			if index < 0 || index >= len(items) {
				return layout.Dimensions{}
			}
			return layout.Inset{
				Left:   unit.Dp(16),
				Right:  unit.Dp(16),
				Bottom: settingsRowBottomSpacing(index, len(items)),
			}.Layout(gtx, items[index])
		})
	})
}

func settingsRowBottomSpacing(index, count int) unit.Dp {
	if index < 0 || index >= count-1 {
		return unit.Dp(0)
	}
	switch index {
	case 0:
		return unit.Dp(12)
	default:
		return unit.Dp(18)
	}
}

func (ui *SettingsUI) syncTranslatorInputs() {
	if ui == nil || ui.translatorSettings == nil {
		return
	}
	if ui.ollamaModelInput != nil && ui.translatorSettings.OllamaModel != nil {
		ui.ollamaModelInput.SetText(ui.translatorSettings.OllamaModel())
	}
	if ui.ollamaBaseURLInput != nil && ui.translatorSettings.OllamaBaseURL != nil {
		ui.ollamaBaseURLInput.SetText(ui.translatorSettings.OllamaBaseURL())
	}
}

func (ui *SettingsUI) syncNotificationSelection() {
	if ui == nil || ui.notificationDropdown == nil {
		return
	}
	level := notifications.NotificationTypeInfo
	if ui.notificationSettings != nil && ui.notificationSettings.Level != nil {
		level = ui.notificationSettings.Level()
	}
	ui.notificationDropdown.SelectItem(notifications.LevelValue(level))
}

func (ui *SettingsUI) syncStartupSelection() {
	if ui == nil || ui.startupDropdown == nil {
		return
	}
	page := "games"
	if ui.appSettings != nil && ui.appSettings.StartupPage != nil {
		if value := strings.TrimSpace(ui.appSettings.StartupPage()); value != "" {
			page = value
		}
	}
	ui.startupDropdown.SelectItem(page)
}

func (ui *SettingsUI) syncTargetLanguageSelection() {
	if ui == nil || ui.targetLanguageDropdown == nil {
		return
	}
	language := "english"
	if ui.transcriptSettings != nil && ui.transcriptSettings.TargetLanguage != nil {
		if value := strings.TrimSpace(ui.transcriptSettings.TargetLanguage()); value != "" {
			language = value
		}
	}
	ui.targetLanguageDropdown.SelectItem(language)
}

func (ui *SettingsUI) layoutAppSettings(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.appSettings == nil {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Application")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.settingsRow("Startup page", "Page opened when Yomuna starts.").Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if ui.startupDropdown == nil {
					return layout.Dimensions{}
				}
				ui.startupDropdown.Width = unit.Dp(240)
				return ui.startupDropdown.Layout(gtx, layer)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutSaveAppSettings(gtx)
		}),
	)
}

func (ui *SettingsUI) layoutTranscriptSettings(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.transcriptSettings == nil {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Transcript")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.settingsRow("Current game", "Saved with transcript preferences.").Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				value := "None"
				if ui.transcriptSettings.SelectedGameName != nil {
					if name := ui.transcriptSettings.SelectedGameName(); name != "" {
						value = name
					}
				}
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextSecondary, value)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.settingsRow("Translate to", "Target language used by transcript translations.").Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if ui.targetLanguageDropdown == nil {
					return layout.Dimensions{}
				}
				ui.targetLanguageDropdown.Width = unit.Dp(240)
				return ui.targetLanguageDropdown.Layout(gtx, layer)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutFontSizeRow(gtx, "Transcript font", "Live transcript row text size.", ui.transcriptSettings.TranscriptFont, ui.transcriptSettings.SetTranscriptFont, &ui.transcriptDown, &ui.transcriptUp)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutFontSizeRow(gtx, "Sentence analysis font", "Focused sentence token text size.", ui.transcriptSettings.SentenceFont, ui.transcriptSettings.SetSentenceFont, &ui.sentenceDown, &ui.sentenceUp)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutFontSizeRow(gtx, "Lookup font", "Dictionary lookup panel text size.", ui.transcriptSettings.LookupFont, ui.transcriptSettings.SetLookupFont, &ui.lookupDown, &ui.lookupUp)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutTranscriptLineLimitRow(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutSaveTranscriptSettings(gtx)
		}),
	)
}

func (ui *SettingsUI) layoutNotificationSettings(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.notificationSettings == nil {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Notifications")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.settingsRow("Notification level", "Minimum notification type shown in the app.").Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if ui.notificationDropdown == nil {
					return layout.Dimensions{}
				}
				ui.notificationDropdown.Width = unit.Dp(240)
				return ui.notificationDropdown.Layout(gtx, layer)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutSaveNotificationSettings(gtx)
		}),
	)
}

func (ui *SettingsUI) layoutStorageLocations(gtx layout.Context) layout.Dimensions {
	base := util.ConfigBaseDir()
	locations := []struct {
		label       string
		description string
		path        string
	}{
		{
			label:       "Yomuna data",
			description: "Base cache/data directory for Yomuna v2.",
			path:        base,
		},
		{
			label:       "App settings",
			description: "Notification level and app-wide v2 preferences.",
			path:        filepath.Join(base, "guiv2-app.json"),
		},
		{
			label:       "Transcript settings",
			description: "Transcript font sizes, target language, toggles, and selected game.",
			path:        filepath.Join(base, "guiv2-transcript.json"),
		},
		{
			label:       "Ollama settings",
			description: "Translator model and endpoint settings.",
			path:        filepath.Join(base, "guiv2-translator.json"),
		},
		{
			label:       "Flashcards",
			description: "Per-game flashcard JSON files.",
			path:        filepath.Join(base, "flashcards"),
		},
		{
			label:       "Anki exports",
			description: "Generated TSV exports for Anki workflows.",
			path:        filepath.Join(base, "exports"),
		},
		{
			label:       "Translations",
			description: "Cached sentence translations.",
			path:        filepath.Join(base, "translations", "sentences.json"),
		},
		{
			label:       "Dictionary",
			description: "Downloaded jpndict/JiTenDex data.",
			path:        filepath.Join(base, "jpndict"),
		},
		{
			label:       "Voices",
			description: "Voice/audio cache used by text-to-speech features.",
			path:        filepath.Join(base, "voices"),
		},
		{
			label:       "Game configs",
			description: "Installed game configuration files.",
			path:        filepath.Join(gameConfig.ConfigBaseDir(), "games"),
		},
		{
			label:       "Legacy game configs",
			description: "Older game config location still read for compatibility.",
			path:        filepath.Join(base, "games"),
		},
	}

	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Storage")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
	}
	for i, location := range locations {
		location := location
		if i > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout))
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.settingsRow(location.label, location.description).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutPathValue(gtx, location.path, ui.storageSizeLabel(location.path))
			})
		}))
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (ui *SettingsUI) layoutPathValue(gtx layout.Context, path, sizeLabel string) layout.Dimensions {
	tokens := ui.theme.GetCurrentColorToken()
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(ui.th, path)
			theme.ApplyTypography(&lbl, ui.theme.GetCurrentTypography(), theme.TextRoleCode)
			lbl.Color = tokens.TextSecondaryNRGBA()
			lbl.Alignment = text.End
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(ui.th, sizeLabel)
			theme.ApplyTypography(&lbl, ui.theme.GetCurrentTypography(), theme.TextRoleCaption)
			lbl.Color = tokens.TextMutedNRGBA()
			lbl.Alignment = text.End
			return lbl.Layout(gtx)
		}),
	)
}

func (ui *SettingsUI) drainStorageSizeResults() bool {
	if ui == nil || ui.storageResults == nil {
		return false
	}
	changed := false
	for {
		select {
		case result := <-ui.storageResults:
			if ui.storageSizes == nil {
				ui.storageSizes = make(map[string]storageSizeState)
			}
			ui.storageSizes[result.path] = storageSizeState{
				size:       result.size,
				err:        result.err,
				exists:     result.exists,
				measuredAt: time.Now(),
			}
			changed = true
		default:
			return changed
		}
	}
}

func (ui *SettingsUI) storageSizeLabel(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "Missing"
	}
	if ui.storageSizes == nil {
		ui.storageSizes = make(map[string]storageSizeState)
	}
	state := ui.storageSizes[path]
	if state.pending {
		return "Calculating..."
	}
	if state.measuredAt.IsZero() || time.Since(state.measuredAt) > 30*time.Second {
		ui.startStorageSizeScan(path)
		return "Calculating..."
	}
	if state.err != nil {
		return "Unavailable"
	}
	if !state.exists {
		return "Missing"
	}
	return formatByteSize(state.size)
}

func (ui *SettingsUI) startStorageSizeScan(path string) {
	if ui == nil || strings.TrimSpace(path) == "" {
		return
	}
	if ui.storageSizes == nil {
		ui.storageSizes = make(map[string]storageSizeState)
	}
	state := ui.storageSizes[path]
	if state.pending {
		return
	}
	state.pending = true
	ui.storageSizes[path] = state

	results := ui.storageResults
	invalidate := ui.invalidate
	if results == nil {
		results = make(chan storageSizeResult, 16)
		ui.storageResults = results
	}
	go func() {
		size, exists, err := storagePathSize(path)
		results <- storageSizeResult{path: path, size: size, exists: exists, err: err}
		if invalidate != nil {
			invalidate()
		}
	}()
}

func storagePathSize(path string) (int64, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	if !info.IsDir() {
		return info.Size(), true, nil
	}

	var total int64
	err = filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, true, err
	}
	return total, true, nil
}

func formatByteSize(size int64) string {
	if size < 0 {
		size = 0
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(size)
	unit := units[0]
	for i := 1; i < len(units) && value >= 1024; i++ {
		value /= 1024
		unit = units[i]
	}
	if unit == "B" {
		return fmt.Sprintf("%d B", size)
	}
	return fmt.Sprintf("%.1f %s", value, unit)
}

func (ui *SettingsUI) layoutSaveNotificationSettings(gtx layout.Context) layout.Dimensions {
	if ui.notificationSettings == nil || ui.notificationSettings.Save == nil {
		return layout.Dimensions{}
	}
	for ui.saveNotifications.Clicked(gtx) {
		if err := ui.notificationSettings.Save(); err != nil {
			ui.notificationStatus = "Save failed"
		} else {
			ui.notificationStatus = "Notification settings saved"
			notifications.Success(ui.notificationStatus)
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutPrimaryButton(gtx, &ui.saveNotifications, "Save Notification Settings")
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.notificationStatus == "" {
				return layout.Dimensions{}
			}
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.notificationStatus)
		}),
	)
}

func (ui *SettingsUI) layoutSaveAppSettings(gtx layout.Context) layout.Dimensions {
	if ui.appSettings == nil || ui.appSettings.Save == nil {
		return layout.Dimensions{}
	}
	for ui.saveApp.Clicked(gtx) {
		if err := ui.appSettings.Save(); err != nil {
			ui.appStatus = "Save failed"
		} else {
			ui.appStatus = "Application settings saved"
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutPrimaryButton(gtx, &ui.saveApp, "Save Application Settings")
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.appStatus == "" {
				return layout.Dimensions{}
			}
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.appStatus)
		}),
	)
}

func (ui *SettingsUI) layoutFontSizeRow(gtx layout.Context, label, description string, get func() unit.Sp, set func(unit.Sp), down, up *widget.Clickable) layout.Dimensions {
	return ui.settingsRow(label, description).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := unit.Sp(0)
		if get != nil {
			size = get()
		}
		for down.Clicked(gtx) {
			if set != nil {
				set(size - unit.Sp(1))
			}
			gtx.Execute(op.InvalidateCmd{})
		}
		for up.Clicked(gtx) {
			if set != nil {
				set(size + unit.Sp(1))
			}
			gtx.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutSmallButton(gtx, down, "-")
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := fmt.Sprintf("%.0fsp", size)
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, text)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutSmallButton(gtx, up, "+")
			}),
		)
	})
}

func notificationLevelItems() []dropdowns.DropdownItem {
	levels := []notifications.NotificationType{
		notifications.NotificationTypeDebug,
		notifications.NotificationTypeInfo,
		notifications.NotificationTypeSuccess,
		notifications.NotificationTypeWarning,
		notifications.NotificationTypeError,
		notifications.NotificationTypeOff,
	}
	items := make([]dropdowns.DropdownItem, 0, len(levels))
	for _, level := range levels {
		items = append(items, dropdowns.DropdownItem{
			Label: notifications.LevelLabel(level),
			Value: notifications.LevelValue(level),
		})
	}
	return items
}

func startupPageItems() []dropdowns.DropdownItem {
	return []dropdowns.DropdownItem{
		{Label: "Games", Value: "games"},
		{Label: "Translation", Value: "translation"},
		{Label: "Transcript", Value: "transcript"},
		{Label: "Flashcards", Value: "flashcards"},
		{Label: "Game Config", Value: "game"},
		{Label: "Add Game", Value: "add-game"},
		{Label: "Settings", Value: "settings"},
	}
}

func translationLanguageItems() []dropdowns.DropdownItem {
	languages := []string{
		"english",
		"spanish",
		"french",
		"german",
		"italian",
		"portuguese",
		"korean",
		"chinese",
	}
	items := make([]dropdowns.DropdownItem, 0, len(languages))
	for _, language := range languages {
		items = append(items, dropdowns.DropdownItem{
			Label: titleText(language),
			Value: language,
		})
	}
	return items
}

func titleText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func (ui *SettingsUI) layoutTranscriptLineLimitRow(gtx layout.Context) layout.Dimensions {
	return ui.settingsRow("Max transcript lines", "Maximum live transcript rows kept in memory.").Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		value := 200
		if ui.transcriptSettings != nil && ui.transcriptSettings.MaxTranscriptRows != nil {
			value = ui.transcriptSettings.MaxTranscriptRows()
		}
		for ui.maxLinesDown.Clicked(gtx) {
			if ui.transcriptSettings != nil && ui.transcriptSettings.SetMaxTranscriptRows != nil {
				ui.transcriptSettings.SetMaxTranscriptRows(value - 25)
			}
			gtx.Execute(op.InvalidateCmd{})
		}
		for ui.maxLinesUp.Clicked(gtx) {
			if ui.transcriptSettings != nil && ui.transcriptSettings.SetMaxTranscriptRows != nil {
				ui.transcriptSettings.SetMaxTranscriptRows(value + 25)
			}
			gtx.Execute(op.InvalidateCmd{})
		}

		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutSmallButton(gtx, &ui.maxLinesDown, "-")
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, fmt.Sprintf("%d lines", value))
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutSmallButton(gtx, &ui.maxLinesUp, "+")
			}),
		)
	})
}

func (ui *SettingsUI) layoutSaveTranscriptSettings(gtx layout.Context) layout.Dimensions {
	if ui.transcriptSettings == nil || ui.transcriptSettings.Save == nil {
		return layout.Dimensions{}
	}
	for ui.saveTranscript.Clicked(gtx) {
		if err := ui.transcriptSettings.Save(); err != nil {
			ui.status = "Save failed"
		} else {
			ui.status = "Transcript settings saved"
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutPrimaryButton(gtx, &ui.saveTranscript, "Save Transcript Settings")
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.status == "" {
				return layout.Dimensions{}
			}
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.status)
		}),
	)
}

func (ui *SettingsUI) layoutTranslatorSettings(gtx layout.Context) layout.Dimensions {
	if ui.translatorSettings == nil {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Ollama")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.settingsRow("Model", "Model name passed to Ollama for translations.").Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if ui.ollamaModelInput == nil {
					return layout.Dimensions{}
				}
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(260))
				return ui.ollamaModelInput.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.settingsRow("Endpoint", "Base URL for the Ollama server.").Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if ui.ollamaBaseURLInput == nil {
					return layout.Dimensions{}
				}
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(260))
				return ui.ollamaBaseURLInput.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutSaveTranslatorSettings(gtx)
		}),
	)
}

func (ui *SettingsUI) layoutSaveTranslatorSettings(gtx layout.Context) layout.Dimensions {
	if ui.translatorSettings == nil || ui.translatorSettings.Save == nil {
		return layout.Dimensions{}
	}
	for ui.saveTranslator.Clicked(gtx) {
		if ui.ollamaModelInput != nil && ui.translatorSettings.SetOllamaModel != nil {
			ui.translatorSettings.SetOllamaModel(strings.TrimSpace(ui.ollamaModelInput.Text()))
		}
		if ui.ollamaBaseURLInput != nil && ui.translatorSettings.SetOllamaBaseURL != nil {
			ui.translatorSettings.SetOllamaBaseURL(strings.TrimRight(strings.TrimSpace(ui.ollamaBaseURLInput.Text()), "/"))
		}
		if err := ui.translatorSettings.Save(); err != nil {
			ui.translatorStatus = "Save failed"
		} else {
			ui.translatorStatus = "Ollama settings saved"
			ui.syncTranslatorInputs()
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutPrimaryButton(gtx, &ui.saveTranslator, "Save Ollama Settings")
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.translatorStatus == "" {
				return layout.Dimensions{}
			}
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.translatorStatus)
		}),
	)
}

func (ui *SettingsUI) layoutSmallButton(gtx layout.Context, click *widget.Clickable, label string) layout.Dimensions {
	tokens := ui.theme.GetCurrentColorToken()
	bg := tokens.SurfaceNRGBA()
	if click.Hovered() {
		bg = tokens.SurfaceAltNRGBA()
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return utils.Surface(gtx, bg, unit.Dp(7), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(5), Bottom: unit.Dp(5), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, label)
			})
		})
	})
}

func (ui *SettingsUI) layoutPrimaryButton(gtx layout.Context, click *widget.Clickable, label string) layout.Dimensions {
	tokens := ui.theme.GetCurrentColorToken()
	bg := tokens.PrimaryNRGBA()
	if click.Hovered() {
		bg = tokens.PrimaryHoverNRGBA()
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return utils.Surface(gtx, bg, unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(9), Bottom: unit.Dp(9), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(ui.th, label)
				theme.ApplyTypography(&lbl, ui.theme.GetCurrentTypography(), theme.TextRoleLabel)
				lbl.Color = color.NRGBA(tokens.OnPrimaryNRGBA())
				return lbl.Layout(gtx)
			})
		})
	})
}

func (ui *SettingsUI) settingsRow(label, description string) *row.Row {
	if ui == nil {
		return row.New(label, description)
	}
	if ui.rowCache == nil {
		ui.rowCache = make(map[string]*row.Row)
	}
	key := label + "\x00" + description
	if ui.rowCache[key] == nil {
		ui.rowCache[key] = row.New(label, description).
			WithMaterialTheme(ui.th).
			WithThemeClient(ui.theme)
	}
	return ui.rowCache[key]
}
