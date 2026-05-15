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
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/tabs"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/notifications"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/layouts/sidebar"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/pages"
	"github.com/DarlingGoose/ymn/pkg/yomuna/assets"
	"github.com/DarlingGoose/ymn/pkg/yomuna/backend"
	"github.com/DarlingGoose/ymn/pkg/yomuna/yomunapages"
	"github.com/DarlingGoose/ymn/pkg/yomuna/yomunapages/gamepage"
	"github.com/DarlingGoose/ymn/pkg/yomuna/yomunapages/transcript"
)

type App struct {
	th    *material.Theme
	theme *theme.Client
	ctx   context.Context
	win   *gioapp.Window

	Sidebar       *sidebar.CollapsibleSidebar
	ToggleButton  *components.IconButton
	Overlay       *overlay.Overlay
	Notifications *notifications.Client

	Translation *yomunapages.TranslationUI
	Flashcards  *yomunapages.FlashcardsUI
	Transcript  *transcript.TranscriptUI
	Game        *gamepage.GameUI
	AddGame     *gamepage.AddGameUI
	Settings    *pages.SettingsUI
}

func New(initialSource string) *App {
	th := material.NewTheme()
	tc := theme.DefaultThemeClient

	b := backend.NewLive()
	appPrefs := loadAppPreferences()
	notificationClient := notifications.DefaultNotificationClient.WithThemeClient(tc).WithMaterialTheme(th)
	notificationClient.NotificationLevel = appPrefs.notificationLevel()
	ui := &App{
		th:            th,
		theme:         tc,
		ctx:           context.Background(),
		Overlay:       &overlay.Overlay{},
		Notifications: notificationClient,
		Transcript:    transcript.NewTranscriptUI(th, tc, b),
		Translation:   yomunapages.NewTranslationUI(th, tc).WithSource(initialSource),
		Flashcards:    yomunapages.NewFlashcardsUI(th, tc, b),
		Game:          gamepage.NewGameUI(th, tc, b),
		AddGame:       gamepage.NewAddGameUI(th, tc, b),
		Settings:      pages.NewSettingsUI(tc),
	}
	ui.Settings.WithTranscriptSettings(&pages.TranscriptSettings{
		SelectedGameName:     ui.Transcript.SelectedGameName,
		TargetLanguage:       ui.Transcript.TargetLanguage,
		TranscriptFont:       ui.Transcript.TranscriptFontSize,
		SentenceFont:         ui.Transcript.SentenceFontSize,
		LookupFont:           ui.Transcript.LookupFontSize,
		MaxTranscriptRows:    ui.Transcript.MaxTranscriptRows,
		SetTranscriptFont:    ui.Transcript.SetTranscriptFontSize,
		SetSentenceFont:      ui.Transcript.SetSentenceFontSize,
		SetLookupFont:        ui.Transcript.SetLookupFontSize,
		SetMaxTranscriptRows: ui.Transcript.SetMaxTranscriptRows,
		SetTargetLanguage:    ui.Transcript.SetTargetLanguage,
		Save:                 ui.Transcript.SavePreferences,
	})
	translatorCfg := b.TranslatorConfig()
	ui.Settings.WithTranslatorSettings(&pages.TranslatorSettings{
		OllamaModel: func() string {
			return translatorCfg.OllamaModel
		},
		OllamaBaseURL: func() string {
			return translatorCfg.OllamaBaseURL
		},
		SetOllamaModel: func(model string) {
			translatorCfg.OllamaModel = model
		},
		SetOllamaBaseURL: func(baseURL string) {
			translatorCfg.OllamaBaseURL = baseURL
		},
		Save: func() error {
			if err := b.SaveTranslatorConfig(translatorCfg); err != nil {
				return err
			}
			translatorCfg = b.TranslatorConfig()
			return nil
		},
	})
	ui.Settings.WithNotificationSettings(&pages.NotificationSettings{
		Level: func() notifications.NotificationType {
			return notificationClient.NotificationLevel
		},
		SetLevel: func(level notifications.NotificationType) {
			notificationClient.NotificationLevel = level
			appPrefs.NotificationLevel = notifications.LevelValue(level)
		},
		Save: func() error {
			appPrefs.NotificationLevel = notifications.LevelValue(notificationClient.NotificationLevel)
			if err := saveAppPreferences(appPrefs); err != nil {
				return err
			}
			appPrefs = loadAppPreferences()
			notificationClient.NotificationLevel = appPrefs.notificationLevel()
			return nil
		},
	})

	menuIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:panel-left-close")
	ui.ToggleButton = components.NewIconButton("Toggle", nil, menuIcon).WithThemeClient(tc)
	ui.ToggleButton.MinWidth = unit.Dp(0)
	ui.ToggleButton.CollapseTextBelow = unit.Dp(120)
	ui.ToggleButton.IconSize = unit.Dp(20)

	appTabs := tabs.New(
		tabs.NewTabFunc("transcript", "Transcript", "lucide:file-text", func(gtx layout.Context) layout.Dimensions {
			return ui.Transcript.Layout(gtx, ui.ctx)
		}),
		//tabs.NewTabFunc("translation", "Translation", "lucide:languages", func(gtx layout.Context) layout.Dimensions {
		//	return ui.Translation.Layout(gtx, ui.ctx)
		//}),
		tabs.NewTabFunc("flashcards", "Flashcards", "lucide:library", func(gtx layout.Context) layout.Dimensions {
			return ui.Flashcards.Layout(gtx, ui.Overlay)
		}),
		tabs.NewTabFunc("game", "Game", "lucide:gamepad-2", func(gtx layout.Context) layout.Dimensions {
			return ui.Game.Layout(gtx, ui.Overlay)
		}),
		tabs.NewTabFunc("add-game", "Add Game", "lucide:plus", func(gtx layout.Context) layout.Dimensions {
			return ui.AddGame.Layout(gtx, ui.Overlay)
		}),
		tabs.NewTabFunc("settings", "Settings", "lucide:settings", func(gtx layout.Context) layout.Dimensions {
			return ui.Settings.Layout(gtx, ui.Overlay)
		}).WithPinned(true),
	)

	ui.Sidebar = sidebar.NewCollapsibleSidebar(appTabs).
		WithThemeClient(tc).
		WithTitle("Yomuna").
		WithImage(assets.YomunaLogo()).
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
	if ui.Transcript != nil && window != nil {
		ui.Transcript.WithInvalidate(window.Invalidate)
	}
	if ui.Notifications != nil && window != nil {
		ui.Notifications.WithInvalidate(window.Invalidate)
	}
	if ui.Settings != nil && window != nil {
		ui.Settings.WithInvalidate(window.Invalidate)
	}
	return ui.Overlay.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := ui.Sidebar.Layout(
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
		if ui.Notifications != nil {
			ui.Overlay.Add(gtx, ui.Notifications)
		}
		return dims
	})
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
