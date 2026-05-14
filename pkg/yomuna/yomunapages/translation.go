package yomunapages

import (
	"context"
	"fmt"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/ymn/pkg/translation"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/layouts/split"
)

type TranslationUI struct {
	th    *material.Theme
	theme *theme.Client

	sourceEditor      widget.Editor
	targetEditor      widget.Editor
	translationEditor widget.Editor
	gameEditor        widget.Editor

	generateButton *components.IconButton
	loadButton     *components.IconButton
	saveButton     *components.IconButton
	clearButton    *components.IconButton

	bodySplit split.SplitH

	status string
	busy   bool
	reqID  int

	results chan translationResult
}

type translationResult struct {
	reqID int
	entry translation.Entry
	err   error
}

func NewTranslationUI(th *material.Theme, tc *theme.Client) *TranslationUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	ui := &TranslationUI{
		th:    th,
		theme: tc,
		bodySplit: split.SplitH{
			Ratio:    0,
			Bar:      unit.Dp(8),
			MinRatio: -0.65,
			MaxRatio: 0.65,
		},
		status:  "Ready",
		results: make(chan translationResult, 4),
	}
	ui.sourceEditor.SingleLine = false
	ui.translationEditor.SingleLine = false
	ui.targetEditor.SingleLine = true
	ui.gameEditor.SingleLine = true
	ui.targetEditor.SetText("English")

	ui.generateButton = components.NewIconButton("Generate", nil, mustIcon("lucide:sparkles")).WithThemeClient(tc)
	ui.loadButton = components.NewIconButton("Load Cache", nil, mustIcon("lucide:database")).WithThemeClient(tc)
	ui.saveButton = components.NewIconButton("Save", nil, mustIcon("lucide:save")).WithThemeClient(tc)
	ui.clearButton = components.NewIconButton("Clear", nil, mustIcon("lucide:x")).WithThemeClient(tc)

	return ui
}

func (ui *TranslationUI) WithSource(text string) *TranslationUI {
	if ui == nil {
		return ui
	}

	ui.sourceEditor.SetText(strings.TrimSpace(text))
	return ui
}

func (ui *TranslationUI) Layout(gtx layout.Context, ctx context.Context) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ui.drainResults(gtx)
	ui.handleButtons(gtx, ctx)
	ui.syncButtonState()

	return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutHeader(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutControls(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutTranslationSplit(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutStatus(gtx)
			}),
		)
	})
}

func (ui *TranslationUI) handleButtons(gtx layout.Context, ctx context.Context) {
	if ui.generateButton.Clicked(gtx) {
		ui.generate(ctx)
	}
	if ui.loadButton.Clicked(gtx) {
		ui.loadCached()
	}
	if ui.saveButton.Clicked(gtx) {
		ui.save()
	}
	if ui.clearButton.Clicked(gtx) {
		ui.sourceEditor.SetText("")
		ui.translationEditor.SetText("")
		ui.status = "Cleared"
	}
}

func (ui *TranslationUI) drainResults(gtx layout.Context) {
	for {
		select {
		case result := <-ui.results:
			if result.reqID != ui.reqID {
				continue
			}

			ui.busy = false
			if result.err != nil {
				ui.status = result.err.Error()
				continue
			}

			ui.translationEditor.SetText(result.entry.Translation)
			ui.status = fmt.Sprintf("Generated %s translation", result.entry.TargetLanguage)
			gtx.Execute(op.InvalidateCmd{})
		default:
			return
		}
	}
}

func (ui *TranslationUI) syncButtonState() {
	source := strings.TrimSpace(ui.sourceEditor.Text())
	translated := strings.TrimSpace(ui.translationEditor.Text())

	ui.generateButton.SetLoading(ui.busy)
	ui.generateButton.Disabled = source == "" || ui.busy
	ui.loadButton.Disabled = source == "" || ui.busy
	ui.saveButton.Disabled = source == "" || translated == "" || ui.busy
	ui.clearButton.Disabled = ui.busy
}

func (ui *TranslationUI) generate(ctx context.Context) {
	source := strings.TrimSpace(ui.sourceEditor.Text())
	if source == "" {
		ui.status = "Source text is required"
		return
	}

	target := strings.TrimSpace(ui.targetEditor.Text())
	if target == "" {
		target = "English"
		ui.targetEditor.SetText(target)
	}

	ui.busy = true
	ui.reqID++
	reqID := ui.reqID
	gameName := strings.TrimSpace(ui.gameEditor.Text())
	ui.status = "Generating translation..."

	go func() {
		entry, err := translation.Generate(ctx, translation.Config{}, gameName, source, target)
		ui.results <- translationResult{reqID: reqID, entry: entry, err: err}
	}()
}

func (ui *TranslationUI) loadCached() {
	source := strings.TrimSpace(ui.sourceEditor.Text())
	if source == "" {
		ui.status = "Source text is required"
		return
	}

	target := strings.TrimSpace(ui.targetEditor.Text())
	if target == "" {
		target = "English"
		ui.targetEditor.SetText(target)
	}

	entry, ok, err := translation.Load(strings.TrimSpace(ui.gameEditor.Text()), source, target)
	if err != nil {
		ui.status = err.Error()
		return
	}
	if !ok {
		ui.status = "No cached translation found"
		return
	}

	ui.translationEditor.SetText(entry.Translation)
	ui.status = "Loaded cached translation"
}

func (ui *TranslationUI) save() {
	entry := translation.Entry{
		GameName:       strings.TrimSpace(ui.gameEditor.Text()),
		SourceText:     strings.TrimSpace(ui.sourceEditor.Text()),
		TargetLanguage: strings.TrimSpace(ui.targetEditor.Text()),
		Translation:    strings.TrimSpace(ui.translationEditor.Text()),
	}
	if entry.TargetLanguage == "" {
		entry.TargetLanguage = "English"
		ui.targetEditor.SetText(entry.TargetLanguage)
	}

	if err := translation.Save(entry); err != nil {
		ui.status = err.Error()
		return
	}

	ui.status = "Saved translation"
}

func (ui *TranslationUI) layoutHeader(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH2, theme.ThemeColorTextPrimary, "Translation")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBody, theme.ThemeColorTextSecondary, "Translate text, review cached results, and save edits.")
		}),
	)
}

func (ui *TranslationUI) layoutControls(gtx layout.Context) layout.Dimensions {
	return panel.NewBackgroundPanel(ui.theme).
		WithRole(panel.BackgroundRoleSurface).
		WithRadius(unit.Dp(8)).
		WithInset(layout.UniformInset(unit.Dp(12))).
		WithFillMax(false).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return ui.layoutInlineEditor(gtx, "Game", "Optional game name", &ui.gameEditor)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return ui.layoutInlineEditor(gtx, "Target", "English", &ui.targetEditor)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(ui.generateButton.Layout),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(ui.loadButton.Layout),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(ui.saveButton.Layout),
						layout.Flexed(1, layout.Spacer{}.Layout),
						layout.Rigid(ui.clearButton.Layout),
					)
				}),
			)
		})
}

func (ui *TranslationUI) layoutInlineEditor(gtx layout.Context, label, hint string, editor *widget.Editor) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabelSmall, theme.ThemeColorTextMuted, label)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutEditorSurface(gtx, hint, editor, unit.Dp(44))
		}),
	)
}

func (ui *TranslationUI) layoutTranslationSplit(gtx layout.Context) layout.Dimensions {
	return ui.bodySplit.Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			return ui.layoutSplitEditorBlock(gtx, "Source", "Japanese text to translate", &ui.sourceEditor)
		},
		func(gtx layout.Context) layout.Dimensions {
			return ui.layoutSplitEditorBlock(gtx, "Translation", "Generated or manually edited translation", &ui.translationEditor)
		},
	)
}

func (ui *TranslationUI) layoutSplitEditorBlock(gtx layout.Context, label, hint string, editor *widget.Editor) layout.Dimensions {
	return panel.NewBackgroundPanel(ui.theme).
		WithRole(panel.BackgroundRoleSurface).
		WithRadius(unit.Dp(8)).
		WithInset(layout.UniformInset(unit.Dp(12))).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, label)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return ui.layoutEditorSurface(gtx, hint, editor, 0)
				}),
			)
		})
}

func (ui *TranslationUI) layoutEditorSurface(gtx layout.Context, hint string, editor *widget.Editor, height unit.Dp) layout.Dimensions {
	heightPx := 0
	if height > 0 {
		heightPx = gtx.Dp(height)
	} else {
		heightPx = gtx.Constraints.Max.Y
	}
	if heightPx < 1 {
		heightPx = 1
	}

	return panel.NewBackgroundPanel(ui.theme).
		WithRole(panel.BackgroundRoleSurfaceAlt).
		WithRadius(unit.Dp(8)).
		WithInset(layout.UniformInset(unit.Dp(10))).
		WithFillMax(false).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = heightPx
			gtx.Constraints.Max.Y = heightPx

			ed := material.Editor(ui.th, editor, hint)
			tokens := ui.theme.GetCurrentColorToken()
			ed.Color = tokens.TextPrimaryNRGBA()
			ed.HintColor = tokens.TextMutedNRGBA()

			dims := ed.Layout(gtx)
			if dims.Size.Y < heightPx {
				dims.Size.Y = heightPx
			}
			return dims
		})
}

func (ui *TranslationUI) layoutStatus(gtx layout.Context) layout.Dimensions {
	return panel.NewBackgroundPanel(ui.theme).
		WithRole(panel.BackgroundRoleSurfaceAlt).
		WithRadius(unit.Dp(8)).
		WithInset(layout.UniformInset(unit.Dp(12))).
		WithFillMax(false).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			status := strings.TrimSpace(ui.status)
			if status == "" {
				status = "Ready"
			}

			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabelSmall, theme.ThemeColorTextMuted, "Status")
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextSecondary, status)
				}),
			)
		})
}

func mustIcon(name string) *iconify.SVGIcon {
	ic, err := iconify.DefaultIconify.Icon(context.Background(), name)
	if err != nil {
		return nil
	}
	return ic
}
