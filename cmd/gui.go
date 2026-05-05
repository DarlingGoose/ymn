package cmd

import (
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareui "github.com/DarlingGoose/bare/pkg/ui"
	"github.com/DarlingGoose/bare/pkg/ui/icons"
	"github.com/DarlingGoose/bare/pkg/ui/media"
	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/vntext/pkg/gameConfig"
	"github.com/DarlingGoose/vntext/pkg/runner"
	"github.com/DarlingGoose/wgl/pkg/anki"
	"github.com/DarlingGoose/wgl/pkg/flashcard"
	pkggui "github.com/DarlingGoose/wgl/pkg/gui"
	guiflashcard "github.com/DarlingGoose/wgl/pkg/gui/flashcard"
	guigame "github.com/DarlingGoose/wgl/pkg/gui/game"
	guisettings "github.com/DarlingGoose/wgl/pkg/gui/settings"
	guitoast "github.com/DarlingGoose/wgl/pkg/gui/toast"
	guitranscript "github.com/DarlingGoose/wgl/pkg/gui/transcript"
	"github.com/DarlingGoose/wgl/pkg/util"
	"github.com/spf13/cobra"
)

const (
	guiPageTranscript = "transcript"
	guiPageFlashcards = "flashcards"
	guiPageGame       = "game"
	guiPageSettings   = "settings"

	guiCompactWidth             = 1080
	guiGameRunningCheckInterval = 2 * time.Second
)

var guiGameName string
var guiPollInterval time.Duration
var guiPrintExisting bool
var guiAnkiURL string
var guiAnkiPushSync bool

var guiCmd = &cobra.Command{
	Use:   "gui [game-name]",
	Short: "Open a transcript watcher window for a saved game",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(os.Getenv("DISPLAY")) == "" && strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			return errors.New("gui mode requires a desktop session with DISPLAY or WAYLAND_DISPLAY set")
		}

		selectedName := strings.TrimSpace(guiGameName)
		if selectedName == "" && len(args) > 0 {
			selectedName = strings.TrimSpace(args[0])
		}

		configs, err := loadInstalledGameConfigs()
		if err != nil {
			return err
		}
		if selectedName != "" {
			if _, err := gameConfig.FindInstalledGame(configs, selectedName); err != nil {
				return err
			}
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		gui, err := newGUI(configs, selectedName, guiPrintExisting, guiPollInterval)
		if err != nil {
			return err
		}
		return gui.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
	guiCmd.Flags().StringVarP(&guiGameName, "game", "g", "", "name of the saved game to watch")
	guiCmd.Flags().DurationVar(&guiPollInterval, "poll-interval", 750*time.Millisecond, "how often to poll the transcript log for new text")
	guiCmd.Flags().BoolVar(&guiPrintExisting, "print-existing", true, "load the current transcript contents before waiting for new dialogue")
	guiCmd.Flags().StringVar(&guiAnkiURL, "anki-url", anki.DefaultAnkiConnectURL, "AnkiConnect URL used by the Sync Anki button")
	guiCmd.Flags().BoolVar(&guiAnkiPushSync, "sync-collection", true, "call AnkiConnect sync after uploading notes from the GUI")
}

type guiApp struct {
	configs       []*game.Game
	printExisting bool
	pollInterval  time.Duration

	shell               bareui.AppShell
	iconify             *icons.Iconify
	pageTabs            *bareui.Tabs
	gameDropdown        bareui.Dropdown
	messageModal        bareui.Modal
	exitButton          widget.Clickable
	sidebarToggleButton widget.Clickable
	toast               *guitoast.Toast

	settingsPage   *guisettings.Settings
	transcriptPage *guitranscript.Page
	flashcardPage  *guiflashcard.Page
	gamePage       *guigame.Page

	theme barethemes.Theme

	gameOptionClicks      map[string]*widget.Clickable
	sidebarTabClicks      map[string]*widget.Clickable
	sidebarGameImageViews map[string]*media.ImageView
	sidebarCollapsed      bool

	activeGameName string
	currentConfig  *game.Game
	logPath        string
	statusText     string
	messageTitle   string
	messageBody    string

	rawTranscript string
	offset        int64

	gameRunning          bool
	gameRunningPID       int
	lastGameRunningCheck time.Time

	watcherCancel     context.CancelFunc
	watcherGeneration int

	mu sync.Mutex
}

func newGUI(configs []*game.Game, selectedName string, printExisting bool, pollInterval time.Duration) (*guiApp, error) {
	settingsPage, err := guisettings.LoadSettings()
	if err != nil {
		return nil, err
	}

	activeGame := strings.TrimSpace(selectedName)
	if activeGame == "" {
		remembered := strings.TrimSpace(settingsPage.LastGame())
		if configNameExists(configs, remembered) {
			activeGame = remembered
		}
	}

	gameClicks := make(map[string]*widget.Clickable, len(configs))
	for _, cfg := range configs {
		gameClicks[cfg.Name] = new(widget.Clickable)
	}
	sidebarTabClicks := map[string]*widget.Clickable{
		guiPageTranscript: new(widget.Clickable),
		guiPageFlashcards: new(widget.Clickable),
		guiPageGame:       new(widget.Clickable),
		guiPageSettings:   new(widget.Clickable),
	}

	pageTabs := bareui.NewTabs([]bareui.TabItem{
		{ID: guiPageTranscript, Label: "Transcript", Icon: "mdi:text-box-outline"},
		{ID: guiPageFlashcards, Label: "Flashcards", Icon: "mdi:cards-outline"},
		{ID: guiPageGame, Label: "Game", Icon: "mdi:puzzle-outline"},
		{ID: guiPageSettings, Label: "Settings", Icon: "mdi:cog-outline"},
	}, guiPageTranscript)
	if activeGame == "" {
		pageTabs.Active = guiPageGame
	}
	pageTabs.Axis = layout.Vertical

	iconify := icons.NewIconify()
	settingsPage.WithIcon(iconify)
	theme := settingsPage.Theme()

	app := &guiApp{
		configs:               configs,
		printExisting:         printExisting,
		pollInterval:          pollInterval,
		shell:                 bareui.AppShell{SidebarWidth: unit.Dp(232)},
		iconify:               iconify,
		pageTabs:              pageTabs,
		gameOptionClicks:      gameClicks,
		sidebarTabClicks:      sidebarTabClicks,
		sidebarGameImageViews: map[string]*media.ImageView{},
		activeGameName:        activeGame,
		messageModal:          bareui.Modal{CloseOnScrim: true},
		toast:                 guitoast.New(),
		settingsPage:          settingsPage,
		transcriptPage:        guitranscript.New(theme).WithIcon(iconify),
		flashcardPage:         guiflashcard.New(theme).WithIcon(iconify),
		gamePage:              guigame.New(theme).WithIcon(iconify),
		theme:                 theme,
		statusText:            "Select a game to start watching its transcript.",
	}

	pkggui.NewDropDownLayout(&app.gameDropdown, "mdi:controller-classic")
	app.transcriptPage.OnError = app.showMessage
	app.transcriptPage.OnNotify = app.showToast
	app.transcriptPage.OnDeleteLog = app.deleteTranscriptLog
	app.flashcardPage.OnError = app.showMessage
	app.flashcardPage.OnNotify = app.showToast
	app.gamePage.OnError = app.showMessage
	app.gamePage.OnSaved = func(cfg *game.Game) {
		g := cfg
		if g == nil {
			return
		}
		app.reloadConfigs()
		app.currentConfig = g
		app.activeGameName = g.Name
		_ = app.settingsPage.SetLastGame(g.Name)
	}
	app.gamePage.OnSelected = func(cfg *game.Game) {
		if cfg == nil {
			return
		}
		app.currentConfig = cfg
		app.activeGameName = cfg.Name
		_ = app.settingsPage.SetLastGame(cfg.Name)
	}
	app.gamePage.OnNew = func() {
		app.currentConfig = nil
		app.activeGameName = ""
		app.pageTabs.Active = guiPageGame
		_ = app.settingsPage.SetLastGame("")
	}
	app.gamePage.OnDeleted = func(name string) {
		app.reloadConfigs()
		if strings.EqualFold(strings.TrimSpace(app.activeGameName), strings.TrimSpace(name)) {
			if cancel := app.stopWatcher(); cancel != nil {
				cancel()
			}
			app.watcherGeneration++
			app.activeGameName = ""
			app.currentConfig = nil
			app.logPath = ""
			app.rawTranscript = ""
			app.offset = 0
			app.statusText = "Select or create a game to start watching its transcript."
			app.gameRunning = false
			app.gameRunningPID = 0
			app.flashcardsFromPage(nil)
			app.transcriptPage.ClearTranscript()
			_ = app.settingsPage.SetLastGame("")
		}
		app.pageTabs.Active = guiPageGame
	}
	app.syncPages()
	return app, nil
}

func (g *guiApp) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		err := <-errCh
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	go func() {
		w := new(app.Window)
		w.Option(app.Title("Yomuna"), app.Size(unit.Dp(1440), unit.Dp(920)))

		if g.activeGameName != "" {
			g.startWatching(ctx, g.activeGameName, w)
		}

		go func() {
			<-ctx.Done()
			if cancel := g.stopWatcher(); cancel != nil {
				cancel()
			}
			w.Perform(system.ActionClose)
		}()

		var ops op.Ops
		for {
			switch e := w.Event().(type) {
			case app.DestroyEvent:
				if cancel := g.stopWatcher(); cancel != nil {
					cancel()
				}
				errCh <- e.Err
				return
			case app.FrameEvent:
				ops.Reset()
				gtx := app.NewContext(&ops, e)
				g.layout(gtx, ctx, w)
				e.Frame(gtx.Ops)
			}
		}
	}()

	app.Main()
	return nil
}

func (g *guiApp) layout(gtx layout.Context, ctx context.Context, w *app.Window) layout.Dimensions {
	g.handleEvents(gtx, ctx, w)
	g.syncPages()

	g.pageTabs.Axis = layout.Vertical
	if g.isCompactLayout(gtx) {
		g.pageTabs.Axis = layout.Horizontal
	}

	return bareutils.Surface(gtx, g.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
		if g.isCompactLayout(gtx) {
			return layout.Stack{}.Layout(gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return g.layoutTopbar(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return g.layoutMain(gtx)
						}),
					)
				}),
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return g.layoutOverlay(gtx)
				}),
			)
		}
		if g.sidebarCollapsed {
			g.shell.SidebarWidth = unit.Dp(76)
		} else {
			g.shell.SidebarWidth = unit.Dp(232)
		}
		return g.shell.Layout(gtx, g.layoutLeftSidebar, g.layoutMain, g.layoutOverlay)
	})
}

func (g *guiApp) handleEvents(gtx layout.Context, ctx context.Context, w *app.Window) {
	g.settingsPage.HandleEvents(gtx, ctx, w)
	g.toast.HandleEvents(gtx, ctx, w)
	g.gameDropdown.Update(gtx)
	for g.exitButton.Clicked(gtx) {
		w.Perform(system.ActionClose)
	}

	for name, click := range g.gameOptionClicks {
		for click.Clicked(gtx) {
			g.startWatching(ctx, name, w)
			g.gameDropdown.Close()
		}
	}

	g.transcriptPage.HandleEvents(gtx, ctx, w)
	g.flashcardPage.HandleEvents(gtx, ctx, w)
	g.gamePage.HandleEvents(gtx, ctx, w)

	switch g.pageTabs.Selected() {
	case guiPageTranscript:
		g.flashcardsFromPage(g.transcriptPage.Cards())
	case guiPageFlashcards:
		g.flashcardsFromPage(g.flashcardPage.Cards())
	}
}

func (g *guiApp) flashcardsFromPage(cards []flashcard.Flashcard) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.flashcardsFromPageLocked(cards)
}

func (g *guiApp) flashcardsFromPageLocked(cards []flashcard.Flashcard) {
	g.flashcardPage.SetCards(cards)
	g.transcriptPage.SetFlashcards(cards)
}

func (g *guiApp) syncPages() {
	g.refreshCurrentGameRunningState(false)
	g.theme = g.settingsPage.Theme()

	g.mu.Lock()
	rawTranscript := g.rawTranscript
	statusText := util.FirstNonEmpty(g.statusText, "Waiting for transcript updates.")
	g.mu.Unlock()

	g.transcriptPage.
		WithTheme(g.theme).
		SetContext(g.activeGameName, g.logPath, guiAnkiURL, g.currentConfig).
		SetPushSync(guiAnkiPushSync).
		SetRunningState(g.gameRunning, g.gameRunningPID).
		SetTranscriptOptions(
			g.settingsPage.TranscriptSize(),
			g.settingsPage.TranscriptSizeLabel(),
			g.settingsPage.RecentLineLimit(),
			g.settingsPage.RecentLineLabel(),
		).
		SetTranslateTextOptions(
			g.settingsPage.FocusedSentenceSize(),
			g.settingsPage.TranslateDetailSize(),
		).
		SetTranslatorConfig(g.settingsPage.TranslatorConfig()).
		SetFocusedFuriganaDefault(g.settingsPage.FocusedFuriganaMode()).
		SetAutoPlayHighlightAudio(g.settingsPage.AutoPlayHighlightAudio()).
		SetColorizeHighlights(g.settingsPage.ColorizeHighlightText()).
		SetStatus(statusText).
		SetRawTranscript(rawTranscript)

	g.flashcardPage.
		WithTheme(g.theme).
		SetContext(g.activeGameName, guiAnkiURL).
		SetPushSync(guiAnkiPushSync)

	g.gamePage.
		WithTheme(g.theme).
		SetConfigs(g.configs).
		SetCurrentConfig(g.currentConfig)
}

func (g *guiApp) layoutLeftSidebar(gtx layout.Context) layout.Dimensions {
	if g.sidebarCollapsed {
		return g.layoutCollapsedSidebar(gtx)
	}
	return bareutils.Panel(gtx, g.theme.Color.Surface, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.H5(g.theme.Gio(), "YMN")
							lbl.Color = g.theme.Color.Text
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							for g.sidebarToggleButton.Clicked(gtx) {
								g.sidebarCollapsed = true
							}
							btn := bareui.Button{
								Clickable: &g.sidebarToggleButton,
								Text:      "mdi:chevron-left",
								Icon:      true,
								Variant:   bareui.ButtonGhost,
							}
							return btn.Layout(gtx, g.theme, g.iconify)
						}),
					)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(g.theme.Gio(), "Transcript, flashcards, and viewer settings for saved games.")
					lbl.Color = g.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.gameDropdown.Layout(gtx, g.theme, g.iconify, g.selectedGameLabel(), g.layoutGameDropdownMenu)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.pageTabs.Layout(gtx, g.theme, g.iconify)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := bareui.Button{
						Clickable: &g.exitButton,
						Text:      "Exit",
						Prefix:    "mdi:exit-to-app",
						Variant:   bareui.ButtonGhost,
					}
					return btn.Layout(gtx, g.theme, g.iconify)
				}),
			)
		})
	})
}

func (g *guiApp) layoutCollapsedSidebar(gtx layout.Context) layout.Dimensions {
	return bareutils.Panel(gtx, g.theme.Color.Surface, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					for g.sidebarToggleButton.Clicked(gtx) {
						g.sidebarCollapsed = false
					}
					btn := bareui.Button{
						Clickable: &g.sidebarToggleButton,
						Text:      "mdi:chevron-right",
						Icon:      true,
						Variant:   bareui.ButtonGhost,
					}
					return btn.Layout(gtx, g.theme, g.iconify)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutCollapsedGamePicker(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.layoutCollapsedTabs(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Dimensions{}
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					btn := bareui.Button{
						Clickable: &g.exitButton,
						Text:      "mdi:exit-to-app",
						Icon:      true,
						Variant:   bareui.ButtonGhost,
					}
					return btn.Layout(gtx, g.theme, g.iconify)
				}),
			)
		})
	})
}

func (g *guiApp) layoutCollapsedTabs(gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(g.pageTabs.Items))
	for _, item := range g.pageTabs.Items {
		item := item
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return g.layoutCollapsedTab(gtx, item)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (g *guiApp) layoutCollapsedTab(gtx layout.Context, item bareui.TabItem) layout.Dimensions {
	click := g.sidebarTabClicks[item.ID]
	if click == nil {
		click = new(widget.Clickable)
		g.sidebarTabClicks[item.ID] = click
	}
	for click.Clicked(gtx) {
		g.pageTabs.Active = item.ID
	}
	active := item.ID == g.pageTabs.Active
	bg := g.theme.Color.Surface
	fg := g.theme.Color.TextMuted
	if active {
		bg = g.theme.Color.Primary
		fg = bareutils.ReadableOn(g.theme.Color.Primary)
	} else if click.Hovered() {
		bg = barethemes.Mix(g.theme.Color.SurfaceAlt, g.theme.Color.Surface, 0.75)
		fg = g.theme.Color.Text
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(44)), gtx.Dp(unit.Dp(44)))
		gtx.Constraints.Max = gtx.Constraints.Min
		return bareutils.RoundedSurface(gtx, bg, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if g.iconify == nil {
					return layout.Dimensions{}
				}
				return g.iconify.Layout(gtx, item.Icon, unit.Dp(20), fg)
			})
		})
	})
}

func (g *guiApp) layoutCollapsedGamePicker(gtx layout.Context) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return g.gameDropdown.Toggle.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				for g.gameDropdown.Toggle.Clicked(gtx) {
					g.gameDropdown.Open = !g.gameDropdown.Open
				}
				return g.layoutGameAvatar(gtx, g.currentConfig, g.gameDropdown.Toggle.Hovered())
			})
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if !g.gameDropdown.Open {
				return layout.Dimensions{}
			}
			macro := op.Record(gtx.Ops)
			offset := op.Offset(image.Pt(gtx.Dp(unit.Dp(56)), 0)).Push(gtx.Ops)
			menuGTX := gtx
			menuGTX.Constraints.Min = image.Point{}
			menuGTX.Constraints.Max = image.Pt(gtx.Dp(unit.Dp(260)), gtx.Dp(unit.Dp(360)))
			bareutils.Panel(menuGTX, g.theme.Color.Surface, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(8)).Layout(gtx, g.layoutGameDropdownMenu)
			})
			offset.Pop()
			op.Defer(gtx.Ops, macro.Stop())
			return layout.Dimensions{}
		}),
	)
}

func (g *guiApp) layoutGameAvatar(gtx layout.Context, cfg *game.Game, hovered bool) layout.Dimensions {
	bg := g.theme.Color.SurfaceAlt
	if hovered {
		bg = barethemes.Mix(g.theme.Color.Primary, g.theme.Color.SurfaceAlt, 0.18)
	}
	gtx.Constraints.Min = image.Pt(gtx.Dp(unit.Dp(46)), gtx.Dp(unit.Dp(46)))
	gtx.Constraints.Max = gtx.Constraints.Min
	return bareutils.RoundedSurface(gtx, bg, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		if path := gameAvatarPath(cfg); path != "" {
			view := g.sidebarGameImageView(path)
			view.Path = path
			dims := view.Draw(gtx)
			if dims.Size.X > 0 && dims.Size.Y > 0 {
				return dims
			}
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			initials := gameInitials(util.FirstNonEmpty(g.activeGameName, "YMN"))
			lbl := material.Body1(g.theme.Gio(), initials)
			lbl.Color = g.theme.Color.Text
			return lbl.Layout(gtx)
		})
	})
}

func (g *guiApp) layoutTopbar(gtx layout.Context) layout.Dimensions {
	return bareutils.Panel(gtx, g.theme.Color.Surface, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			exitButton := bareui.Button{
				Clickable: &g.exitButton,
				Text:      "Exit",
				Prefix:    "mdi:exit-to-app",
				Variant:   bareui.ButtonGhost,
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return g.gameDropdown.Layout(gtx, g.theme, g.iconify, g.selectedGameLabel(), g.layoutGameDropdownMenu)
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return exitButton.Layout(gtx, g.theme, g.iconify)
						}),
					)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.pageTabs.Layout(gtx, g.theme, g.iconify)
				}),
			)
		})
	})
}

func (g *guiApp) layoutMain(gtx layout.Context) layout.Dimensions {
	switch g.pageTabs.Selected() {
	case guiPageFlashcards:
		return g.flashcardPage.LayoutPage(gtx)
	case guiPageGame:
		return g.gamePage.LayoutPage(gtx)
	case guiPageSettings:
		return g.settingsPage.LayoutPage(gtx)
	default:
		return g.transcriptPage.LayoutPage(gtx)
	}
}

func (g *guiApp) layoutOverlay(gtx layout.Context) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return g.toast.Layout(gtx, g.theme, g.iconify)
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if !g.messageModal.Open {
				return layout.Dimensions{}
			}
			return g.messageModal.Layout(gtx, g.theme, util.FirstNonEmpty(g.messageTitle, "Message"), g.layoutModalContent)
		}),
	)
}

func (g *guiApp) layoutModalContent(gtx layout.Context) layout.Dimensions {
	lbl := material.Body1(g.theme.Gio(), g.messageBody)
	lbl.Color = g.theme.Color.Text
	return lbl.Layout(gtx)
}

func (g *guiApp) layoutGameDropdownMenu(gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(g.configs))
	for _, cfg := range g.configs {
		cfg := cfg
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := bareui.Button{
				Clickable: g.gameOptionClicks[cfg.Name],
				Text:      cfg.Name,
				Prefix:    "mdi:controller-classic",
				Variant:   dropdownButtonVariant(cfg.Name == g.activeGameName),
			}
			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return btn.Layout(gtx, g.theme, g.iconify)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (g *guiApp) sidebarGameImageView(path string) *media.ImageView {
	if g.sidebarGameImageViews == nil {
		g.sidebarGameImageViews = map[string]*media.ImageView{}
	}
	if g.sidebarGameImageViews[path] == nil {
		g.sidebarGameImageViews[path] = &media.ImageView{Path: path}
	}
	return g.sidebarGameImageViews[path]
}

func gameAvatarPath(cfg *game.Game) string {
	if cfg == nil {
		return ""
	}
	for _, path := range []string{cfg.IconPath, cfg.ImagePath} {
		path = strings.TrimSpace(path)
		if path == "" || !util.IsExistingFile(path) {
			continue
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".png", ".jpg", ".jpeg", ".webp", ".gif":
			return path
		}
	}
	return ""
}

func gameInitials(name string) string {
	fields := strings.Fields(strings.TrimSpace(name))
	if len(fields) == 0 {
		return "?"
	}
	var b strings.Builder
	for _, field := range fields {
		for _, r := range field {
			b.WriteRune(r)
			break
		}
		if b.Len() >= 2 {
			break
		}
	}
	out := strings.ToUpper(b.String())
	if strings.TrimSpace(out) == "" {
		return "?"
	}
	return out
}

func (g *guiApp) selectedGameLabel() string {
	return truncateGameDropdownLabel(util.FirstNonEmpty(g.activeGameName, "Select game"))
}

func truncateGameDropdownLabel(label string) string {
	label = strings.TrimSpace(label)
	const maxRunes = 22
	runes := []rune(label)
	if len(runes) <= maxRunes {
		return label
	}
	return string(runes[:maxRunes-1]) + "…"
}

func (g *guiApp) showMessage(title, body string) {
	g.messageTitle = title
	g.messageBody = body
	g.toast.Queue(guitoast.Notification{
		Title:   title,
		Message: body,
		Type:    guitoast.NotificationTypeError,
	})
}

func (g *guiApp) showToast(title, body string, kind guitoast.NotificationType) {
	g.toast.Queue(guitoast.Notification{
		Title:   title,
		Message: body,
		Type:    kind,
	})
}

func (g *guiApp) startWatching(ctx context.Context, gameName string, w *app.Window) {
	cfg, err := gameConfig.FindInstalledGame(g.configs, gameName)
	if err != nil {
		g.showMessage("Game Load Failed", err.Error())
		return
	}

	if cancel := g.stopWatcher(); cancel != nil {
		cancel()
	}

	raw, offset, status := initializeTranscriptState(cfg.TextHookLogFile, g.printExisting)
	g.mu.Lock()
	g.activeGameName = cfg.Name
	g.currentConfig = cfg
	g.logPath = cfg.TextHookLogFile
	g.rawTranscript = raw
	g.offset = offset
	g.statusText = status
	g.mu.Unlock()

	g.reloadConfigs()
	g.reloadFlashcards()
	g.refreshCurrentGameRunningState(true)
	_ = g.settingsPage.SetLastGame(cfg.Name)

	watcherCtx, cancel := context.WithCancel(ctx)
	g.watcherCancel = cancel
	g.watcherGeneration++
	generation := g.watcherGeneration

	go g.pollTranscript(watcherCtx, generation, cfg.TextHookLogFile, w)
}

func (g *guiApp) deleteTranscriptLog(cfg *game.Game) error {
	if cfg == nil {
		return errors.New("game config is not loaded")
	}
	if err := cfg.DeleteLog(); err != nil {
		return err
	}

	g.mu.Lock()
	if g.currentConfig != nil && strings.EqualFold(strings.TrimSpace(g.currentConfig.Name), strings.TrimSpace(cfg.Name)) {
		g.rawTranscript = ""
		g.offset = 0
		g.statusText = "Transcript log deleted; waiting for new dialogue."
	}
	g.mu.Unlock()
	return nil
}

func (g *guiApp) stopWatcher() context.CancelFunc {
	cancel := g.watcherCancel
	g.watcherCancel = nil
	return cancel
}

func (g *guiApp) pollTranscript(ctx context.Context, generation int, logPath string, w *app.Window) {
	ticker := time.NewTicker(g.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		g.mu.Lock()
		if generation != g.watcherGeneration {
			g.mu.Unlock()
			return
		}
		offset := g.offset
		g.mu.Unlock()

		delta, err := readTranscriptDelta(logPath, &offset)
		g.mu.Lock()
		if generation != g.watcherGeneration {
			g.mu.Unlock()
			return
		}
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				g.statusText = "Transcript log not found yet; start the game and advance dialogue."
				g.mu.Unlock()
				w.Invalidate()
				continue
			}
			g.statusText = err.Error()
			g.mu.Unlock()
			w.Invalidate()
			continue
		}
		g.offset = offset
		if delta != "" {
			g.rawTranscript += delta
			g.statusText = "Watching transcript log for new dialogue."
		} else if strings.TrimSpace(g.rawTranscript) == "" {
			g.statusText = "Waiting for transcript updates."
		}
		g.mu.Unlock()
		w.Invalidate()
	}
}

func (g *guiApp) reloadConfigs() {
	configs, err := loadInstalledGameConfigs()
	if err != nil {
		return
	}
	g.configs = configs
	valid := make(map[string]struct{}, len(configs))
	for _, cfg := range configs {
		valid[cfg.Name] = struct{}{}
		if g.gameOptionClicks[cfg.Name] == nil {
			g.gameOptionClicks[cfg.Name] = new(widget.Clickable)
		}
	}
	for name := range g.gameOptionClicks {
		if _, ok := valid[name]; !ok {
			delete(g.gameOptionClicks, name)
		}
	}
}

func (g *guiApp) reloadFlashcards() {
	if strings.TrimSpace(g.activeGameName) == "" {
		g.flashcardsFromPage(nil)
		return
	}
	cards, err := flashcard.LoadFlashcards(g.activeGameName)
	if err != nil {
		g.showMessage("Flashcard Load Failed", err.Error())
		return
	}
	g.flashcardsFromPage(cards)
}

func (g *guiApp) refreshCurrentGameRunningState(force bool) {
	if !force && !g.lastGameRunningCheck.IsZero() && time.Since(g.lastGameRunningCheck) < guiGameRunningCheckInterval {
		return
	}
	g.lastGameRunningCheck = time.Now()
	if g.currentConfig == nil || strings.TrimSpace(g.currentConfig.Name) == "" {
		g.gameRunning = false
		g.gameRunningPID = 0
		return
	}

	r, err := runner.New().IsRunning(g.currentConfig)
	if err == nil && r != nil && r.Status == runner.StatusRunning {
		g.gameRunning = true
		g.gameRunningPID = r.PID
		return
	}

	g.gameRunning = false
	g.gameRunningPID = 0
}

func (g *guiApp) isCompactLayout(gtx layout.Context) bool {
	return gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(guiCompactWidth))
}

func initializeTranscriptState(logPath string, printExisting bool) (raw string, offset int64, status string) {
	if printExisting {
		delta, err := readTranscriptDelta(logPath, &offset)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", 0, "Transcript log not found yet; start the game and advance dialogue."
			}
			return "", 0, err.Error()
		}
		if strings.TrimSpace(delta) == "" {
			return "", offset, "Watching transcript log for new dialogue."
		}
		return delta, offset, "Loaded transcript history and waiting for new dialogue."
	}

	info, err := os.Stat(logPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", 0, "Transcript log not found yet; start the game and advance dialogue."
		}
		return "", 0, err.Error()
	}
	return "", info.Size(), "Watching transcript log for new dialogue."
}

func dropdownButtonVariant(active bool) bareui.ButtonVariant {
	if active {
		return bareui.ButtonPrimary
	}
	return bareui.ButtonSecondary
}

func configNameExists(configs []*game.Game, name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	for _, cfg := range configs {
		if strings.EqualFold(strings.TrimSpace(cfg.Name), name) {
			return true
		}
	}
	return false
}

func loadInstalledGameConfigs() ([]*game.Game, error) {
	return gameConfig.LoadInstalledGames(gameConfigDirs()...)
}

func gameConfigDirs() []string {
	return []string{
		filepath.Join(gameConfig.ConfigBaseDir(), "games"),
		filepath.Join(util.ConfigBaseDir(), "games"),
	}
}

func readTranscriptDelta(logPath string, offset *int64) (string, error) {
	info, err := os.Stat(logPath)
	if err != nil {
		return "", err
	}
	if info.Size() < *offset {
		*offset = 0
	}
	if info.Size() == *offset {
		return "", nil
	}

	file, err := os.Open(logPath)
	if err != nil {
		return "", fmt.Errorf("open transcript log: %w", err)
	}
	defer file.Close()

	if _, err := file.Seek(*offset, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek transcript log: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read transcript log: %w", err)
	}
	*offset = info.Size()
	return string(data), nil
}
