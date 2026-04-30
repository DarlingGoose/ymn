package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
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
	bareui "github.com/Seann-Moser/bare/pkg/ui"
	"github.com/Seann-Moser/bare/pkg/ui/icons"
	barethemes "github.com/Seann-Moser/bare/pkg/ui/themes"
	bareutils "github.com/Seann-Moser/bare/pkg/ui/utils"
	"github.com/Seann-Moser/wgl/pkg/anki"
	"github.com/Seann-Moser/wgl/pkg/flashcard"
	"github.com/Seann-Moser/wgl/pkg/game/gameconfig"
	pkggui "github.com/Seann-Moser/wgl/pkg/gui"
	guiflashcard "github.com/Seann-Moser/wgl/pkg/gui/flashcard"
	guigame "github.com/Seann-Moser/wgl/pkg/gui/game"
	guisettings "github.com/Seann-Moser/wgl/pkg/gui/settings"
	guitoast "github.com/Seann-Moser/wgl/pkg/gui/toast"
	guitranscript "github.com/Seann-Moser/wgl/pkg/gui/transcript"
	"github.com/Seann-Moser/wgl/pkg/util"
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

		configs, err := gameconfig.ListConfigs()
		if err != nil {
			return err
		}
		if selectedName != "" {
			if _, err := gameconfig.FindConfig(selectedName); err != nil {
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
	configs       []gameconfig.GameConfig
	printExisting bool
	pollInterval  time.Duration

	shell        bareui.AppShell
	iconify      *icons.Iconify
	pageTabs     *bareui.Tabs
	gameDropdown bareui.Dropdown
	messageModal bareui.Modal
	exitButton   widget.Clickable
	toast        *guitoast.Toast

	settingsPage   *guisettings.Settings
	transcriptPage *guitranscript.Page
	flashcardPage  *guiflashcard.Page
	gamePage       *guigame.Page

	theme barethemes.Theme

	gameOptionClicks map[string]*widget.Clickable

	activeGameName string
	currentConfig  *gameconfig.GameConfig
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

func newGUI(configs []gameconfig.GameConfig, selectedName string, printExisting bool, pollInterval time.Duration) (*guiApp, error) {
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
		configs:          configs,
		printExisting:    printExisting,
		pollInterval:     pollInterval,
		shell:            bareui.AppShell{SidebarWidth: unit.Dp(232)},
		iconify:          iconify,
		pageTabs:         pageTabs,
		gameOptionClicks: gameClicks,
		activeGameName:   activeGame,
		messageModal:     bareui.Modal{CloseOnScrim: true},
		toast:            guitoast.New(),
		settingsPage:     settingsPage,
		transcriptPage:   guitranscript.New(theme).WithIcon(iconify),
		flashcardPage:    guiflashcard.New(theme).WithIcon(iconify),
		gamePage:         guigame.New(theme).WithIcon(iconify),
		theme:            theme,
		statusText:       "Select a game to start watching its transcript.",
	}

	pkggui.NewDropDownLayout(&app.gameDropdown, "mdi:controller-classic")
	app.transcriptPage.OnError = app.showMessage
	app.transcriptPage.OnNotify = app.showToast
	app.flashcardPage.OnError = app.showMessage
	app.flashcardPage.OnNotify = app.showToast
	app.gamePage.OnError = app.showMessage
	app.gamePage.OnSaved = func(cfg *gameconfig.GameConfig) {
		g := cfg
		if g == nil {
			return
		}
		app.reloadConfigs()
		app.currentConfig = g
		app.activeGameName = g.Name
		_ = app.settingsPage.SetLastGame(g.Name)
	}
	app.gamePage.OnSelected = func(cfg *gameconfig.GameConfig) {
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
		w.Option(app.Title("WGL"), app.Size(unit.Dp(1440), unit.Dp(920)))

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
	return bareutils.Panel(gtx, g.theme.Color.Surface, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(g.theme.Gio(), "WGL")
					lbl.Color = g.theme.Color.Text
					return lbl.Layout(gtx)
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

func (g *guiApp) selectedGameLabel() string {
	return util.FirstNonEmpty(g.activeGameName, "Select game")
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
	cfg, err := gameconfig.FindConfig(gameName)
	if err != nil {
		g.showMessage("Game Load Failed", err.Error())
		return
	}

	logPath, err := resolveRPGMakerTranscriptPath(util.FirstNonEmpty(cfg.GamePath, cfg.Executable, cfg.WorkingDir))
	if err != nil {
		g.showMessage("Transcript Setup Failed", err.Error())
		return
	}

	if cancel := g.stopWatcher(); cancel != nil {
		cancel()
	}

	raw, offset, status := initializeTranscriptState(logPath, g.printExisting)
	g.mu.Lock()
	g.activeGameName = cfg.Name
	g.currentConfig = cfg
	g.logPath = logPath
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

	go g.pollTranscript(watcherCtx, generation, logPath, w)
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
	configs, err := gameconfig.ListConfigs()
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

	processes, err := listProcesses()
	if err != nil {
		g.gameRunning = false
		g.gameRunningPID = 0
		return
	}
	matches := rankProcessMatches(*g.currentConfig, processes)
	if len(matches) == 0 {
		g.gameRunning = false
		g.gameRunningPID = 0
		return
	}
	g.gameRunning = true
	g.gameRunningPID = matches[0].PID
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

func configNameExists(configs []gameconfig.GameConfig, name string) bool {
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
