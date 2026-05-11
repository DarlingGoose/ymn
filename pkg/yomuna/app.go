package yomuna

import (
	"context"
	"os"

	gioapp "gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	bareicons "github.com/DarlingGoose/bare/pkg/ui/icons"
	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
	guisettings "github.com/DarlingGoose/wgl/pkg/gui/settings"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/tabs"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/layouts/sidebar"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/pages"
	"github.com/DarlingGoose/wgl/pkg/yomuna/backend"
	"github.com/DarlingGoose/wgl/pkg/yomuna/yomunapages"
	"github.com/DarlingGoose/wgl/pkg/yomuna/yomunapages/transcript"
)

type App struct {
	th    *material.Theme
	theme *theme.Client
	ctx   context.Context
	win   *gioapp.Window

	Sidebar      *sidebar.CollapsibleSidebar
	ToggleButton *components.IconButton
	Overlay      *overlay.Overlay

	Translation *yomunapages.TranslationUI
	Transcript  *transcript.TranscriptUI
	Settings    *pages.SettingsUI

	legacySettings *guisettings.Settings
	legacyTheme    barethemes.Theme
	legacyIconify  *bareicons.Iconify
}

func New(initialSource string) *App {
	th := material.NewTheme()
	tc := theme.DefaultThemeClient
	legacyTheme := barethemes.DefaultConfig().Theme(false)
	legacySettings, _ := guisettings.LoadSettings()
	if legacySettings != nil {
		legacyTheme = legacySettings.Theme()
	}
	legacyIconify := bareicons.NewIconify()

	b := backend.NewLive()
	ui := &App{
		th:             th,
		theme:          tc,
		ctx:            context.Background(),
		Overlay:        &overlay.Overlay{},
		Transcript:     transcript.NewTranscriptUI(th, tc, b),
		Translation:    yomunapages.NewTranslationUI(th, tc).WithSource(initialSource),
		Settings:       pages.NewSettingsUI(tc),
		legacySettings: legacySettings,
		legacyTheme:    legacyTheme,
		legacyIconify:  legacyIconify,
	}

	menuIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:panel-left-close")
	ui.ToggleButton = components.NewIconButton("Toggle", nil, menuIcon).WithThemeClient(tc)
	ui.ToggleButton.MinWidth = unit.Dp(0)
	ui.ToggleButton.CollapseTextBelow = unit.Dp(120)
	ui.ToggleButton.IconSize = unit.Dp(20)

	appTabs := tabs.New(
		tabs.NewTabFunc("transcript", "Transcript", "lucide:file-text", func(gtx layout.Context) layout.Dimensions {
			return ui.Transcript.Layout(gtx, ui.ctx)
		}),
		tabs.NewTabFunc("translation", "Translation", "lucide:languages", func(gtx layout.Context) layout.Dimensions {
			return ui.Translation.Layout(gtx, ui.ctx)
		}),
		tabs.NewTabFunc("settings", "Settings", "lucide:settings", func(gtx layout.Context) layout.Dimensions {
			return ui.Settings.Layout(gtx, ui.Overlay)
		}).WithPinned(true),
	)

	ui.Sidebar = sidebar.NewCollapsibleSidebar(appTabs).
		WithThemeClient(tc).
		WithTitle("Yomuna").
		WithIcon("lucide:book-open").
		WithExitButton(true)

	return ui
}

func (ui *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	errCh := make(chan error, 1)

	go func() {
		window := new(gioapp.Window)
		window.Option(gioapp.Title("Yomuna v2"), gioapp.Size(unit.Dp(1180), unit.Dp(820)))
		ui.Sidebar.WithExitFunc(func(gtx layout.Context) {
			window.Perform(system.ActionClose)
		})
		go func() {
			<-ctx.Done()
			window.Perform(system.ActionClose)
		}()

		var ops op.Ops
		for {
			switch e := window.Event().(type) {
			case gioapp.DestroyEvent:
				errCh <- e.Err
				return
			case gioapp.FrameEvent:
				ops.Reset()
				gtx := gioapp.NewContext(&ops, e)
				ui.Layout(gtx, ctx, window)
				e.Frame(gtx.Ops)
			}
		}
	}()

	go func() {
		err := <-errCh
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}()

	gioapp.Main()
	return nil
}

func (ui *App) Layout(gtx layout.Context, ctx context.Context, window *gioapp.Window) layout.Dimensions {
	if ui == nil || ui.Sidebar == nil {
		return layout.Dimensions{}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ui.ctx = ctx
	ui.win = window
	ui.syncLegacyTranscript(gtx, ctx, window)

	return ui.Overlay.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.Sidebar.Layout(
			gtx,
			ui.layoutSidebar,
			func(gtx layout.Context) layout.Dimensions {
				return panel.NewBackgroundPanel(ui.theme).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if ui.Sidebar.Tabs == nil {
						return ui.Translation.Layout(gtx, ctx)
					}
					return ui.Sidebar.Tabs.Layout(gtx)
				})
			},
		)
	})
}

func (ui *App) syncLegacyTranscript(gtx layout.Context, ctx context.Context, window *gioapp.Window) {
	//if ui == nil || ui.Transcript == nil {
	//	return
	//}
	//
	//if ui.legacySettings != nil {
	//	ui.legacyTheme = ui.legacySettings.Theme()
	//	ui.Transcript.
	//		WithTheme(ui.legacyTheme).
	//		SetTranscriptOptions(
	//			ui.legacySettings.TranscriptSize(),
	//			ui.legacySettings.TranscriptSizeLabel(),
	//			ui.legacySettings.RecentLineLimit(),
	//			ui.legacySettings.RecentLineLabel(),
	//		).
	//		SetTranscriptDisplayOptions(
	//			ui.legacySettings.ShowSpeakerOnlyTranscriptLines(),
	//			ui.legacySettings.UseCompactTranscriptTimestamps(),
	//		).
	//		SetTranslateTextOptions(
	//			ui.legacySettings.FocusedSentenceSize(),
	//			ui.legacySettings.TranslateDetailSize(),
	//		).
	//		SetTranslatorConfig(ui.legacySettings.TranslatorConfig()).
	//		SetDefaultTargetLanguage(ui.legacySettings.DefaultTranslationLanguage()).
	//		SetFocusedFuriganaDefault(ui.legacySettings.FocusedFuriganaMode()).
	//		SetAutoPlayHighlightAudio(ui.legacySettings.AutoPlayHighlightAudio()).
	//		SetColorizeHighlights(ui.legacySettings.ColorizeHighlightText())
	//} else {
	//	ui.Transcript.WithTheme(ui.legacyTheme)
	//}
	//
	//ui.Transcript.SetStatus("Transcript tab is available in guiv2. Game watching is not wired into guiv2 yet.")
	//if window != nil {
	//	ui.Transcript.HandleEvents(gtx, ctx, window)
	//}
}

func (ui *App) layoutSidebar(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.Sidebar.LayoutHeader(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if ui.ToggleButton.Clicked(gtx) {
					ui.Sidebar.Toggle(gtx.Now)
					gtx.Execute(op.InvalidateCmd{})
				}
				return ui.ToggleButton.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.Sidebar.LayoutTabButtons(gtx)
			}),
		)
	})
}
