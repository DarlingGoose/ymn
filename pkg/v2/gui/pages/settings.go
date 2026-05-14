package pages

import (
	"fmt"
	"image/color"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

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
	ModeToggle    *toggles.ThemeModeToggle
	ThemeDropdown *dropdowns.ThemeDropdown
	settingsList  widget.List

	transcriptSettings *TranscriptSettings
	transcriptDown     widget.Clickable
	transcriptUp       widget.Clickable
	sentenceDown       widget.Clickable
	sentenceUp         widget.Clickable
	lookupDown         widget.Clickable
	lookupUp           widget.Clickable
	maxLinesDown       widget.Clickable
	maxLinesUp         widget.Clickable
	saveTranscript     widget.Clickable
	saveTranslator     widget.Clickable
	status             string
	translatorStatus   string

	translatorSettings *TranslatorSettings
	ollamaModelInput   *input.TextInput
	ollamaBaseURLInput *input.TextInput

	notificationSettings *NotificationSettings
	notificationDropdown *dropdowns.Dropdown
	saveNotifications    widget.Clickable
	notificationStatus   string

	theme *theme.Client
}

type TranscriptSettings struct {
	SelectedGameName     func() string
	TranscriptFont       func() unit.Sp
	SentenceFont         func() unit.Sp
	LookupFont           func() unit.Sp
	MaxTranscriptRows    func() int
	SetTranscriptFont    func(unit.Sp)
	SetSentenceFont      func(unit.Sp)
	SetLookupFont        func(unit.Sp)
	SetMaxTranscriptRows func(int)
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

func NewSettingsUI(tc *theme.Client) *SettingsUI {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	ui := &SettingsUI{
		theme:         tc,
		ModeToggle:    toggles.NewThemeModeToggle(tc),
		ThemeDropdown: dropdowns.NewThemeDropdown(tc),
		settingsList:  widget.List{List: layout.List{Axis: layout.Vertical}},
		ollamaModelInput: input.NewTextInput("Model", "translategemma:4b").
			WithThemeClient(tc),
		ollamaBaseURLInput: input.NewTextInput("Endpoint", "http://localhost:11434").
			WithThemeClient(tc),
		notificationDropdown: dropdowns.NewDropdown(notificationLevelItems()).
			WithThemeClient(tc).
			WithRole(theme.TextRoleLabel).
			WithMenuAbove(),
	}
	ui.ollamaModelInput.LeadingIcon = "lucide:brain"
	ui.ollamaBaseURLInput.LeadingIcon = "lucide:server"
	ui.ollamaBaseURLInput.Normalize = func(text string) string {
		return strings.TrimRight(strings.TrimSpace(text), "/")
	}
	return ui
}

func (ui *SettingsUI) WithTranscriptSettings(settings *TranscriptSettings) *SettingsUI {
	if ui == nil {
		return ui
	}
	ui.transcriptSettings = settings
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

func (ui *SettingsUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}

	if ui.ModeToggle.Update(gtx) {
		gtx.Execute(op.InvalidateCmd{})
	}

	items := []layout.Widget{
		ui.ModeToggle.Layout,
		func(gtx layout.Context) layout.Dimensions {
			return ui.ThemeDropdown.Layout(gtx, layer)
		},
		ui.layoutTranscriptSettings,
		ui.layoutTranslatorSettings,
		func(gtx layout.Context) layout.Dimensions {
			return ui.layoutNotificationSettings(gtx, layer)
		},
	}

	return layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		list := material.List(material.NewTheme(), &ui.settingsList)
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

func (ui *SettingsUI) layoutTranscriptSettings(gtx layout.Context) layout.Dimensions {
	if ui.transcriptSettings == nil {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Transcript")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return row.New("Current game", "Saved with transcript preferences.").WithThemeClient(ui.theme).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				value := "None"
				if ui.transcriptSettings.SelectedGameName != nil {
					if name := ui.transcriptSettings.SelectedGameName(); name != "" {
						value = name
					}
				}
				return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextSecondary, value)
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
	ui.syncNotificationSelection()
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Notifications")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return row.New("Notification level", "Minimum notification type shown in the app.").WithThemeClient(ui.theme).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
			return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.notificationStatus)
		}),
	)
}

func (ui *SettingsUI) layoutFontSizeRow(gtx layout.Context, label, description string, get func() unit.Sp, set func(unit.Sp), down, up *widget.Clickable) layout.Dimensions {
	return row.New(label, description).WithThemeClient(ui.theme).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
				return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, text)
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

func (ui *SettingsUI) layoutTranscriptLineLimitRow(gtx layout.Context) layout.Dimensions {
	return row.New("Max transcript lines", "Maximum live transcript rows kept in memory.").WithThemeClient(ui.theme).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
				return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, fmt.Sprintf("%d lines", value))
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
			return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.status)
		}),
	)
}

func (ui *SettingsUI) layoutTranslatorSettings(gtx layout.Context) layout.Dimensions {
	if ui.translatorSettings == nil {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Ollama")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return row.New("Model", "Model name passed to Ollama for translations.").WithThemeClient(ui.theme).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if ui.ollamaModelInput == nil {
					return layout.Dimensions{}
				}
				gtx.Constraints.Min.X = gtx.Dp(unit.Dp(260))
				return ui.ollamaModelInput.Layout(gtx)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return row.New("Endpoint", "Base URL for the Ollama server.").WithThemeClient(ui.theme).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
			return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.translatorStatus)
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
				return theme.ThemedLabel(gtx, material.NewTheme(), ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, label)
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
				lbl := material.Body2(material.NewTheme(), label)
				theme.ApplyTypography(&lbl, ui.theme.GetCurrentTypography(), theme.TextRoleLabel)
				lbl.Color = color.NRGBA(tokens.OnPrimaryNRGBA())
				return lbl.Layout(gtx)
			})
		})
	})
}
