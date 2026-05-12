package pages

import (
	"fmt"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/row"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type SettingsUI struct {
	ModeToggle    *toggles.ThemeModeToggle
	ThemeDropdown *dropdowns.ThemeDropdown

	transcriptSettings *TranscriptSettings
	transcriptDown     widget.Clickable
	transcriptUp       widget.Clickable
	sentenceDown       widget.Clickable
	sentenceUp         widget.Clickable
	lookupDown         widget.Clickable
	lookupUp           widget.Clickable
	saveTranscript     widget.Clickable
	status             string

	theme *theme.Client
}

type TranscriptSettings struct {
	SelectedGameName  func() string
	TranscriptFont    func() unit.Sp
	SentenceFont      func() unit.Sp
	LookupFont        func() unit.Sp
	SetTranscriptFont func(unit.Sp)
	SetSentenceFont   func(unit.Sp)
	SetLookupFont     func(unit.Sp)
	Save              func() error
}

func NewSettingsUI(tc *theme.Client) *SettingsUI {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	return &SettingsUI{
		theme:         tc,
		ModeToggle:    toggles.NewThemeModeToggle(tc),
		ThemeDropdown: dropdowns.NewThemeDropdown(tc),
	}
}

func (ui *SettingsUI) WithTranscriptSettings(settings *TranscriptSettings) *SettingsUI {
	if ui == nil {
		return ui
	}
	ui.transcriptSettings = settings
	return ui
}

func (ui *SettingsUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}

	if ui.ModeToggle.Update(gtx) {
		gtx.Execute(op.InvalidateCmd{})
	}

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(ui.ModeToggle.Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.ThemeDropdown.Layout(gtx, layer)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutTranscriptSettings(gtx)
			}),
		)
	})
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
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutSaveTranscriptSettings(gtx)
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
