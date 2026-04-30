package cmd

//import (
//	"context"
//	"errors"
//	"fmt"
//	"image"
//	"image/color"
//	"os"
//	"os/signal"
//	"path/filepath"
//	"slices"
//	"sort"
//	"strings"
//	"sync"
//	"syscall"
//	"time"
//
//	"gioui.org/app"
//	"gioui.org/io/system"
//	"gioui.org/layout"
//	"gioui.org/op"
//	"gioui.org/op/clip"
//	"gioui.org/op/paint"
//	"gioui.org/unit"
//	"gioui.org/widget"
//	"gioui.org/widget/material"
//	bareui "github.com/Seann-Moser/bare/pkg/ui"
//	"github.com/Seann-Moser/bare/pkg/ui/icons"
//	barethemes "github.com/Seann-Moser/bare/pkg/ui/themes"
//	bareutils "github.com/Seann-Moser/bare/pkg/ui/utils"
//	"github.com/Seann-Moser/wgl/pkg/anki"
//	"github.com/spf13/cobra"
//)
//
//var guiGameName string
//var guiPollInterval time.Duration
//var guiPrintExisting bool
//var guiAnkiURL string
//var guiAnkiPushSync bool
//
//const (
//	guiPageTranscript = "transcript"
//	guiPageFlashcards = "flashcards"
//	guiPageGame       = "game"
//	guiPageSettings   = "settings"
//)
//
//const guiGameRunningCheckInterval = 2 * time.Second
//
//const (
//	guiCompactWidth          = 1080
//	guiTranscriptStackWidth  = 1240
//	guiTranscriptMediumWidth = 1480
//)
//
//var guiCmd = &cobra.Command{
//	Use:   "gui [game-name]",
//	Short: "Open a transcript watcher window for a saved game",
//	Args:  cobra.MaximumNArgs(1),
//	RunE: func(cmd *cobra.Command, args []string) error {
//		if strings.TrimSpace(os.Getenv("DISPLAY")) == "" && strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
//			return errors.New("gui mode requires a desktop session with DISPLAY or WAYLAND_DISPLAY set")
//		}
//
//		selectedName := strings.TrimSpace(guiGameName)
//		if selectedName == "" && len(args) > 0 {
//			selectedName = strings.TrimSpace(args[0])
//		}
//
//		configs, err := listGameConfigs()
//		if err != nil {
//			return err
//		}
//		if len(configs) == 0 {
//			return fmt.Errorf("no saved games found in %s", configBaseDir())
//		}
//		if selectedName != "" {
//			if _, err := findGameConfig(selectedName); err != nil {
//				return err
//			}
//		}
//
//		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
//		defer stop()
//
//		gui := newTranscriptGUI(configs, selectedName, guiPrintExisting, guiPollInterval)
//		return gui.Run(ctx)
//	},
//}
//
//func init() {
//	rootCmd.AddCommand(guiCmd)
//	guiCmd.Flags().StringVarP(&guiGameName, "game", "g", "", "name of the saved game to watch")
//	guiCmd.Flags().DurationVar(&guiPollInterval, "poll-interval", 750*time.Millisecond, "how often to poll the transcript log for new text")
//	guiCmd.Flags().BoolVar(&guiPrintExisting, "print-existing", true, "load the current transcript contents before waiting for new dialogue")
//	guiCmd.Flags().StringVar(&guiAnkiURL, "anki-url", anki.defaultAnkiConnectURL, "AnkiConnect URL used by the Sync Anki button")
//	guiCmd.Flags().BoolVar(&guiAnkiPushSync, "sync-collection", true, "call AnkiConnect sync after uploading notes from the GUI")
//}
//
//type guiDropdownOption struct {
//	Label           string
//	Icon            string
//	Mode            barethemes.Mode
//	Palette         barethemes.PaletteName
//	TextSize        unit.Sp
//	RecentLineLimit int
//	Clickable       *widget.Clickable
//}
//
//type gamePathPreview struct {
//	ResolvedPath string
//	Executable   string
//	WorkingDir   string
//	IconPath     string
//	ImagePath    string
//	Name         string
//	Runner       string
//	SteamAppID   string
//	Verified     bool
//	Error        string
//}
//
//type browseEntry struct {
//	Name  string
//	Path  string
//	IsDir bool
//}
//
//type transcriptGUI struct {
//	configs       []GameConfig
//	printExisting bool
//	pollInterval  time.Duration
//
//	theme                 barethemes.Theme
//	iconify               *icons.Iconify
//	shell                 bareui.AppShell
//	pageTabs              *bareui.Tabs
//	gameDropdown          bareui.Dropdown
//	modeDropdown          bareui.Dropdown
//	paletteDropdown       bareui.Dropdown
//	textSizeDropdown      bareui.Dropdown
//	recentLinesDropdown   bareui.Dropdown
//	newGameRunnerDropdown bareui.Dropdown
//	messageModal          bareui.Modal
//	browseModal           bareui.Modal
//	transcriptView        widget.Selectable
//	transcriptList        widget.List
//	flashcardList         widget.List
//	lookupResultsList     widget.List
//	browseList            widget.List
//
//	gameOptionClicks        map[string]*widget.Clickable
//	modeOptions             []guiDropdownOption
//	paletteOptions          []guiDropdownOption
//	transcriptSizeOptions   []guiDropdownOption
//	recentLineOptions       []guiDropdownOption
//	newGameRunnerOptions    []guiDropdownOption
//	flashcards              []Flashcard
//	activeGameName          string
//	currentConfig           GameConfig
//	logPath                 string
//	statusText              string
//	messageTitle            string
//	messageBody             string
//	selectedModeName        string
//	selectedPaletteName     string
//	selectedTextSizeName    string
//	selectedRecentLinesName string
//	themeMode               barethemes.Mode
//	themePalette            barethemes.PaletteName
//	systemDark              bool
//	transcriptTextSize      unit.Sp
//	recentLineLimit         int
//
//	wordEditor                 widget.Editor
//	meaningEditor              widget.Editor
//	searchWordButton           widget.Clickable
//	playAudioButton            widget.Clickable
//	addAllLookupButton         widget.Clickable
//	launchGameButton           widget.Clickable
//	exitButton                 widget.Clickable
//	syncAnkiButton             widget.Clickable
//	clearButton                widget.Clickable
//	saveCardButton             widget.Clickable
//	reloadCardsButton          widget.Clickable
//	flashcardSearchEditor      widget.Editor
//	flashcardWordEditor        widget.Editor
//	flashcardMeaningEditor     widget.Editor
//	flashcardSaveButton        widget.Clickable
//	flashcardDeleteButton      widget.Clickable
//	flashcardNewButton         widget.Clickable
//	selectedFlashcardID        string
//	flashcardSelectClicks      map[string]*widget.Clickable
//	flashcardDeleteClicks      map[string]*widget.Clickable
//	transcriptHighlightClicks  map[string]*widget.Clickable
//	selectableTextStates       map[string]*widget.Selectable
//	transcriptPopupAudioButton widget.Clickable
//	popupFlashcard             *Flashcard
//	gameInstallHookButton      widget.Clickable
//	gameRefreshHookButton      widget.Clickable
//	gameHookStatus             textHookStatus
//	newGamePathEditor          widget.Editor
//	newGameSteamAppIDEditor    widget.Editor
//	newGameIconPathEditor      widget.Editor
//	newGameImagePathEditor     widget.Editor
//	newGameRequiresSteam       widget.Bool
//	autoPlayHighlightAudio     widget.Bool
//	newGameSaveButton          widget.Clickable
//	newGameAnalyzeButton       widget.Clickable
//	newGameBrowseButton        widget.Clickable
//	newGameStatus              string
//	selectedNewGameRunner      string
//	newGamePreview             gamePathPreview
//	browseUpButton             widget.Clickable
//	browseUseCurrentButton     widget.Clickable
//	browseEntryClicks          map[string]*widget.Clickable
//	browseCurrentPath          string
//	browseEntries              []browseEntry
//	browseError                string
//	lookupStatus               string
//	lookupResult               *dictionaryLookup
//	lookupResults              []dictionaryLookup
//	lookupResultAddClicks      map[string]*widget.Clickable
//	lookupResultPlayClicks     map[string]*widget.Clickable
//	gameRunning                bool
//	gameRunningPID             int
//	lastGameRunningCheck       time.Time
//
//	mu                  sync.Mutex
//	rawTranscript       string
//	displayTranscript   string
//	displayDirty        bool
//	transcriptResetView bool
//	offset              int64
//	watcherCancel       context.CancelFunc
//	watcherGeneration   int
//}
//
//func newTranscriptGUI(configs []GameConfig, selectedName string, printExisting bool, pollInterval time.Duration) *transcriptGUI {
//	activeGame := strings.TrimSpace(selectedName)
//	if activeGame == "" && len(configs) > 0 {
//		activeGame = configs[0].Name
//	}
//
//	pageTabs := bareui.NewTabs([]bareui.TabItem{
//		{ID: guiPageTranscript, Label: "Transcript", Icon: "mdi:text-box-outline"},
//		{ID: guiPageFlashcards, Label: "Flashcards", Icon: "mdi:cards-outline"},
//		{ID: guiPageGame, Label: "Game", Icon: "mdi:puzzle-outline"},
//		{ID: guiPageSettings, Label: "Settings", Icon: "mdi:cog-outline"},
//	}, guiPageTranscript)
//	pageTabs.Axis = layout.Vertical
//
//	gameOptionClicks := make(map[string]*widget.Clickable, len(configs))
//	for _, cfg := range configs {
//		gameOptionClicks[cfg.Name] = new(widget.Clickable)
//	}
//
//	gui := &transcriptGUI{
//		configs:                   configs,
//		printExisting:             printExisting,
//		pollInterval:              pollInterval,
//		iconify:                   icons.NewIconify(),
//		shell:                     bareui.AppShell{SidebarWidth: unit.Dp(232)},
//		pageTabs:                  pageTabs,
//		activeGameName:            activeGame,
//		gameOptionClicks:          gameOptionClicks,
//		flashcardSelectClicks:     make(map[string]*widget.Clickable),
//		flashcardDeleteClicks:     make(map[string]*widget.Clickable),
//		transcriptHighlightClicks: make(map[string]*widget.Clickable),
//		selectableTextStates:      make(map[string]*widget.Selectable),
//		lookupResultAddClicks:     make(map[string]*widget.Clickable),
//		lookupResultPlayClicks:    make(map[string]*widget.Clickable),
//		browseEntryClicks:         make(map[string]*widget.Clickable),
//		messageModal:              bareui.Modal{CloseOnScrim: true},
//		browseModal:               bareui.Modal{CloseOnScrim: true},
//		selectedModeName:          "Dark",
//		selectedPaletteName:       "Ocean",
//		selectedTextSizeName:      "Medium",
//		selectedRecentLinesName:   "All Lines",
//		themeMode:                 barethemes.ModeDark,
//		themePalette:              barethemes.PaletteOcean,
//		systemDark:                true,
//		transcriptTextSize:        unit.Sp(16),
//		selectedNewGameRunner:     "auto",
//		newGameStatus:             "Create a saved game config for the sidebar and launcher flows.",
//		lookupStatus:              "Dictionary lookup can fill the meaning and fetch audio for the current word.",
//	}
//	gui.wordEditor.SingleLine = true
//	gui.meaningEditor.SingleLine = false
//	gui.flashcardSearchEditor.SingleLine = true
//	gui.flashcardWordEditor.SingleLine = true
//	gui.flashcardMeaningEditor.SingleLine = false
//	gui.newGamePathEditor.SingleLine = true
//	gui.newGameSteamAppIDEditor.SingleLine = true
//	gui.newGameIconPathEditor.SingleLine = true
//	gui.newGameImagePathEditor.SingleLine = true
//	gui.transcriptList.Axis = layout.Vertical
//	gui.transcriptList.ScrollToEnd = true
//	gui.flashcardList.Axis = layout.Vertical
//	gui.lookupResultsList.Axis = layout.Vertical
//	gui.browseList.Axis = layout.Vertical
//	gui.gameDropdown.Width = unit.Dp(260)
//	gui.gameDropdown.MaxHeight = unit.Dp(320)
//	gui.gameDropdown.OffsetY = unit.Dp(52)
//	gui.gameDropdown.Prefix = "mdi:controller-classic"
//	gui.modeDropdown.Width = unit.Dp(220)
//	gui.modeDropdown.MaxHeight = unit.Dp(220)
//	gui.modeDropdown.OffsetY = unit.Dp(52)
//	gui.modeDropdown.Prefix = "mdi:theme-light-dark"
//	gui.paletteDropdown.Width = unit.Dp(220)
//	gui.paletteDropdown.MaxHeight = unit.Dp(260)
//	gui.paletteDropdown.OffsetY = unit.Dp(52)
//	gui.paletteDropdown.Prefix = "mdi:palette-outline"
//	gui.textSizeDropdown.Width = unit.Dp(220)
//	gui.textSizeDropdown.MaxHeight = unit.Dp(220)
//	gui.textSizeDropdown.OffsetY = unit.Dp(52)
//	gui.textSizeDropdown.Prefix = "mdi:format-size"
//	gui.recentLinesDropdown.Width = unit.Dp(220)
//	gui.recentLinesDropdown.MaxHeight = unit.Dp(220)
//	gui.recentLinesDropdown.OffsetY = unit.Dp(52)
//	gui.recentLinesDropdown.Prefix = "mdi:sort-clock-descending-outline"
//	gui.newGameRunnerDropdown.Width = unit.Dp(220)
//	gui.newGameRunnerDropdown.MaxHeight = unit.Dp(220)
//	gui.newGameRunnerDropdown.OffsetY = unit.Dp(52)
//	gui.newGameRunnerDropdown.Prefix = "mdi:rocket-launch-outline"
//	gui.modeOptions = newGUIModeOptions()
//	gui.paletteOptions = newGUIPaletteOptions()
//	gui.transcriptSizeOptions = newGUITranscriptSizeOptions()
//	gui.recentLineOptions = newGUIRecentLineOptions()
//	gui.newGameRunnerOptions = newGUINewGameRunnerOptions()
//	gui.applySavedSettings()
//	gui.applyTheme()
//	gui.initializeBrowsePath("")
//	gui.refreshCurrentGameHookStatus()
//	return gui
//}
//
//func (g *transcriptGUI) Run(ctx context.Context) error {
//	errCh := make(chan error, 1)
//
//	go func() {
//		err := <-errCh
//		if err != nil {
//			fmt.Fprintln(os.Stderr, err)
//			os.Exit(1)
//		}
//		os.Exit(0)
//	}()
//
//	go func() {
//		w := new(app.Window)
//		w.Option(
//			app.Title("WGL"),
//			app.Size(unit.Dp(1440), unit.Dp(920)),
//		)
//
//		if g.activeGameName != "" {
//			g.startWatching(ctx, g.activeGameName, w)
//		}
//		g.reloadFlashcards()
//
//		go func() {
//			<-ctx.Done()
//			if cancel := g.stopWatcher(); cancel != nil {
//				cancel()
//			}
//			w.Perform(system.ActionClose)
//		}()
//
//		var ops op.Ops
//		for {
//			switch e := w.Event().(type) {
//			case app.DestroyEvent:
//				if cancel := g.stopWatcher(); cancel != nil {
//					cancel()
//				}
//				errCh <- e.Err
//				return
//			case app.FrameEvent:
//				ops.Reset()
//				gtx := app.NewContext(&ops, e)
//				g.layout(gtx, ctx, w)
//				e.Frame(gtx.Ops)
//			}
//		}
//	}()
//
//	app.Main()
//
//	return nil
//}
//
//func (g *transcriptGUI) layout(gtx layout.Context, ctx context.Context, w *app.Window) layout.Dimensions {
//	g.syncTranscriptEditor()
//	g.handleGlobalEvents(gtx, ctx, w)
//	g.pageTabs.Axis = layout.Vertical
//	if g.isCompactLayout(gtx) {
//		g.pageTabs.Axis = layout.Horizontal
//	}
//
//	return bareutils.Surface(gtx, g.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
//		if g.isCompactLayout(gtx) {
//			return g.layoutCompactShell(gtx)
//		}
//		return g.shell.Layout(gtx,
//			func(gtx layout.Context) layout.Dimensions {
//				return g.layoutLeftSidebar(gtx)
//			},
//			func(gtx layout.Context) layout.Dimensions {
//				return g.layoutMain(gtx)
//			},
//			func(gtx layout.Context) layout.Dimensions {
//				return g.layoutOverlay(gtx)
//			},
//		)
//	})
//}
//
//func (g *transcriptGUI) isCompactLayout(gtx layout.Context) bool {
//	return gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(guiCompactWidth))
//}
//
//func (g *transcriptGUI) shouldStackTranscriptPage(gtx layout.Context) bool {
//	return gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(guiTranscriptStackWidth))
//}
//
//func (g *transcriptGUI) transcriptComposerWidth(gtx layout.Context) int {
//	width := gtx.Dp(unit.Dp(340))
//	if gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(guiTranscriptMediumWidth)) {
//		width = gtx.Dp(unit.Dp(300))
//	}
//	if gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(guiTranscriptStackWidth)) {
//		width = gtx.Dp(unit.Dp(280))
//	}
//	return width
//}
//
//func (g *transcriptGUI) pageInset(gtx layout.Context) unit.Dp {
//	if g.isCompactLayout(gtx) {
//		return unit.Dp(12)
//	}
//	return unit.Dp(20)
//}
//
//func (g *transcriptGUI) pageGap(gtx layout.Context) unit.Dp {
//	if g.isCompactLayout(gtx) {
//		return unit.Dp(12)
//	}
//	return unit.Dp(20)
//}
//
//func (g *transcriptGUI) compactTranscriptMinHeight(gtx layout.Context) int {
//	minHeight := gtx.Constraints.Max.Y / 2
//	floor := gtx.Dp(unit.Dp(320))
//	if minHeight < floor {
//		minHeight = floor
//	}
//	return minHeight
//}
//
//func (g *transcriptGUI) selectableTextState(key string) *widget.Selectable {
//	if g.selectableTextStates[key] == nil {
//		g.selectableTextStates[key] = new(widget.Selectable)
//	}
//	return g.selectableTextStates[key]
//}
//
//func (g *transcriptGUI) layoutSelectableBodyText(gtx layout.Context, key, text string, col color.NRGBA) layout.Dimensions {
//	lbl := material.Body1(g.theme.Gio(), text)
//	lbl.Color = col
//	lbl.State = g.selectableTextState(key)
//	return lbl.Layout(gtx)
//}
//
//func (g *transcriptGUI) layoutSelectableHeadlineText(gtx layout.Context, key, text string, col color.NRGBA) layout.Dimensions {
//	lbl := material.H6(g.theme.Gio(), text)
//	lbl.Color = col
//	lbl.State = g.selectableTextState(key)
//	return lbl.Layout(gtx)
//}
//
//func (g *transcriptGUI) layoutCompactShell(gtx layout.Context) layout.Dimensions {
//	return layout.UniformInset(g.pageInset(gtx)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return g.layoutCompactHeader(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(g.pageGap(gtx))),
//			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//				return g.layoutMain(gtx)
//			}),
//		)
//	})
//}
//
//func (g *transcriptGUI) layoutCompactHeader(gtx layout.Context) layout.Dimensions {
//	exitButton := bareui.Button{
//		Clickable: &g.exitButton,
//		Text:      "Exit",
//		Prefix:    "mdi:exit-to-app",
//		Variant:   bareui.ButtonGhost,
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{
//						Axis:      layout.Horizontal,
//						Alignment: layout.Middle,
//					}.Layout(gtx,
//						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//									return bareui.LayoutSidebarTitle(gtx, g.theme, "WGL")
//								}),
//								layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
//								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//									lbl := material.Body1(g.theme.Gio(), "Transcript, flashcards, and theme controls.")
//									lbl.Color = g.theme.Color.TextMuted
//									return lbl.Layout(gtx)
//								}),
//							)
//						}),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return exitButton.Layout(gtx, g.theme, g.iconify)
//						}),
//					)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.pageTabs.Layout(gtx, g.theme, g.iconify)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) handleGlobalEvents(gtx layout.Context, ctx context.Context, w *app.Window) {
//	g.refreshCurrentGameRunningState(false)
//	g.gameDropdown.Update(gtx)
//	g.modeDropdown.Update(gtx)
//	g.paletteDropdown.Update(gtx)
//	g.textSizeDropdown.Update(gtx)
//	g.recentLinesDropdown.Update(gtx)
//	if g.autoPlayHighlightAudio.Update(gtx) {
//		g.persistSettings()
//	}
//
//	for gameName, click := range g.gameOptionClicks {
//		for click.Clicked(gtx) {
//			g.gameDropdown.Close()
//			if gameName != g.activeGameName {
//				g.startWatching(ctx, gameName, w)
//			}
//		}
//	}
//
//	for i := range g.modeOptions {
//		opt := &g.modeOptions[i]
//		for opt.Clickable.Clicked(gtx) {
//			g.themeMode = opt.Mode
//			g.selectedModeName = opt.Label
//			g.modeDropdown.Close()
//			g.applyTheme()
//			g.persistSettings()
//		}
//	}
//
//	for i := range g.paletteOptions {
//		opt := &g.paletteOptions[i]
//		for opt.Clickable.Clicked(gtx) {
//			g.themePalette = opt.Palette
//			g.selectedPaletteName = opt.Label
//			g.paletteDropdown.Close()
//			g.applyTheme()
//			g.persistSettings()
//		}
//	}
//
//	for i := range g.transcriptSizeOptions {
//		opt := &g.transcriptSizeOptions[i]
//		for opt.Clickable.Clicked(gtx) {
//			g.transcriptTextSize = opt.TextSize
//			g.selectedTextSizeName = opt.Label
//			g.textSizeDropdown.Close()
//			g.persistSettings()
//		}
//	}
//
//	for i := range g.recentLineOptions {
//		opt := &g.recentLineOptions[i]
//		for opt.Clickable.Clicked(gtx) {
//			g.recentLineLimit = opt.RecentLineLimit
//			g.selectedRecentLinesName = opt.Label
//			g.recentLinesDropdown.Close()
//			g.mu.Lock()
//			g.updateDisplayTranscriptLocked()
//			g.mu.Unlock()
//			g.persistSettings()
//		}
//	}
//
//	for i := range g.newGameRunnerOptions {
//		opt := &g.newGameRunnerOptions[i]
//		for opt.Clickable.Clicked(gtx) {
//			g.selectedNewGameRunner = opt.Label
//			g.newGameRunnerDropdown.Close()
//		}
//	}
//
//	for g.syncAnkiButton.Clicked(gtx) {
//		g.syncCurrentGameToAnki()
//	}
//	for g.clearButton.Clicked(gtx) {
//		g.clearTranscript()
//	}
//	for g.saveCardButton.Clicked(gtx) {
//		g.saveFlashcard()
//	}
//	for g.searchWordButton.Clicked(gtx) {
//		g.lookupCurrentWord()
//	}
//	for g.playAudioButton.Clicked(gtx) {
//		g.playCurrentLookupAudio()
//	}
//	for g.addAllLookupButton.Clicked(gtx) {
//		g.addAllLookupFlashcards()
//	}
//	for g.launchGameButton.Clicked(gtx) {
//		g.launchCurrentGameInBackground()
//	}
//	for g.exitButton.Clicked(gtx) {
//		w.Perform(system.ActionClose)
//	}
//	for g.reloadCardsButton.Clicked(gtx) {
//		g.reloadFlashcards()
//	}
//	for g.flashcardSaveButton.Clicked(gtx) {
//		g.saveFlashcardFromLibrary()
//	}
//	for g.flashcardDeleteButton.Clicked(gtx) {
//		g.deleteSelectedFlashcard()
//	}
//	for g.flashcardNewButton.Clicked(gtx) {
//		g.prepareNewFlashcard()
//	}
//	for cardID, click := range g.flashcardSelectClicks {
//		for click.Clicked(gtx) {
//			g.selectFlashcard(cardID)
//		}
//	}
//	for cardID, click := range g.flashcardDeleteClicks {
//		for click.Clicked(gtx) {
//			g.deleteFlashcardByID(cardID)
//		}
//	}
//	for key, click := range g.transcriptHighlightClicks {
//		for click.Clicked(gtx) {
//			g.openTranscriptHighlightPopup(key)
//		}
//	}
//	for g.transcriptPopupAudioButton.Clicked(gtx) {
//		g.playPopupFlashcardAudio()
//	}
//	for g.gameInstallHookButton.Clicked(gtx) {
//		g.installCurrentGameTextHook()
//	}
//	for g.gameRefreshHookButton.Clicked(gtx) {
//		g.refreshCurrentGameHookStatus()
//	}
//	for g.newGameSaveButton.Clicked(gtx) {
//		g.saveNewGame(ctx, w)
//	}
//	for g.newGameAnalyzeButton.Clicked(gtx) {
//		g.analyzeNewGamePath()
//	}
//	for g.newGameBrowseButton.Clicked(gtx) {
//		g.initializeBrowsePath(strings.TrimSpace(g.newGamePathEditor.Text()))
//	}
//	for g.browseUpButton.Clicked(gtx) {
//		g.browseUp()
//	}
//	for g.browseUseCurrentButton.Clicked(gtx) {
//		g.selectCurrentBrowsePath()
//	}
//	for entryPath, click := range g.browseEntryClicks {
//		for click.Clicked(gtx) {
//			g.handleBrowseSelection(entryPath)
//		}
//	}
//	for key, click := range g.lookupResultAddClicks {
//		for click.Clicked(gtx) {
//			g.addLookupFlashcardByKey(key)
//		}
//	}
//	for key, click := range g.lookupResultPlayClicks {
//		for click.Clicked(gtx) {
//			g.playLookupAudioByKey(key)
//		}
//	}
//}
//
//func (g *transcriptGUI) layoutLeftSidebar(gtx layout.Context) layout.Dimensions {
//	exitButton := bareui.Button{
//		Clickable: &g.exitButton,
//		Text:      "Exit",
//		Prefix:    "mdi:exit-to-app",
//		Variant:   bareui.ButtonGhost,
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis: layout.Vertical,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return bareui.LayoutSidebarTitle(gtx, g.theme, "WGL")
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), "Transcript, flashcards, and theme controls.")
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(20))),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					return g.pageTabs.Layout(gtx, g.theme, g.iconify)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return exitButton.Layout(gtx, g.theme, g.iconify)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutMain(gtx layout.Context) layout.Dimensions {
//	return layout.UniformInset(g.pageInset(gtx)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{
//			Axis: layout.Vertical,
//		}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return g.layoutTopNav(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(g.pageGap(gtx))),
//			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//				switch g.pageTabs.Selected() {
//				case guiPageFlashcards:
//					return g.layoutFlashcardsPage(gtx)
//				case guiPageGame:
//					return g.layoutGamePage(gtx)
//				case guiPageSettings:
//					return g.layoutSettingsPage(gtx)
//				default:
//					return g.layoutTranscriptPage(gtx)
//				}
//			}),
//		)
//	})
//}
//
//func (g *transcriptGUI) layoutTopNav(gtx layout.Context) layout.Dimensions {
//	title := "Transcript Workspace"
//	switch g.pageTabs.Selected() {
//	case guiPageFlashcards:
//		title = "Flashcard Deck"
//	case guiPageGame:
//		title = "Game Tools"
//	case guiPageSettings:
//		title = "Settings"
//	}
//
//	gameLabel := g.activeGameName
//	if gameLabel == "" {
//		gameLabel = "Select game"
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.Surface, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			if g.isCompactLayout(gtx) {
//				return layout.Flex{
//					Axis: layout.Vertical,
//				}.Layout(gtx,
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return bareui.LayoutTopbar(gtx, g.theme, g.iconify, title, nil)
//					}),
//					layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return g.gameDropdown.Layout(gtx, g.theme, g.iconify, gameLabel, g.layoutGameDropdownMenu)
//					}),
//				)
//			}
//			return layout.Flex{
//				Axis:      layout.Horizontal,
//				Alignment: layout.Middle,
//			}.Layout(gtx,
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					return bareui.LayoutTopbar(gtx, g.theme, g.iconify, title, nil)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.gameDropdown.Layout(gtx, g.theme, g.iconify, gameLabel, g.layoutGameDropdownMenu)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutGameDropdownMenu(gtx layout.Context) layout.Dimensions {
//	children := make([]layout.FlexChild, 0, len(g.configs))
//	for _, cfg := range g.configs {
//		cfg := cfg
//		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			btn := bareui.Button{
//				Clickable: g.gameOptionClicks[cfg.Name],
//				Text:      cfg.Name,
//				Prefix:    "mdi:gamepad-variant-outline",
//				Variant:   dropdownButtonVariant(cfg.Name == g.activeGameName),
//			}
//			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return btn.Layout(gtx, g.theme, g.iconify)
//			})
//		}))
//	}
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
//}
//
//func dropdownButtonVariant(active bool) bareui.ButtonVariant {
//	if active {
//		return bareui.ButtonPrimary
//	}
//	return bareui.ButtonSecondary
//}
//
//func (g *transcriptGUI) layoutTranscriptPage(gtx layout.Context) layout.Dimensions {
//	if g.shouldStackTranscriptPage(gtx) {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//				gtx.Constraints.Min.Y = g.compactTranscriptMinHeight(gtx)
//				return g.layoutTranscriptPanel(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return g.layoutFlashcardComposer(gtx)
//			}),
//		)
//	}
//
//	return layout.Flex{
//		Axis: layout.Horizontal,
//	}.Layout(gtx,
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			return g.layoutTranscriptPanel(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			width := g.transcriptComposerWidth(gtx)
//			gtx.Constraints.Min.X = width
//			gtx.Constraints.Max.X = width
//			return g.layoutFlashcardComposer(gtx)
//		}),
//	)
//}
//
//func (g *transcriptGUI) layoutTranscriptPanel(gtx layout.Context) layout.Dimensions {
//	return bareutils.Panel(gtx, g.theme.Color.Surface, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			metaSpacing := unit.Dp(14)
//			if !g.gameRunning {
//				metaSpacing = unit.Dp(0)
//			}
//			return layout.Flex{
//				Axis: layout.Vertical,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H5(g.theme.Gio(), firstNonEmpty(g.activeGameName, "No game selected"))
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), firstNonEmpty(g.logPath, "No transcript path resolved"))
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), g.statusText)
//					lbl.Color = g.statusColor()
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutTranscriptActions(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(metaSpacing)),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					if !g.gameRunning {
//						return layout.Dimensions{}
//					}
//					return g.layoutTranscriptMeta(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(metaSpacing)),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					return bareutils.Panel(gtx, g.theme.Color.Background, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//						return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//							if !g.gameRunning {
//								return g.layoutTranscriptIdleState(gtx)
//							}
//							return g.layoutTranscriptEditor(gtx)
//						})
//					})
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutTranscriptMeta(gtx layout.Context) layout.Dimensions {
//	left := material.Body1(g.theme.Gio(), "Text Size: "+g.selectedTextSizeName)
//	left.Color = g.theme.Color.TextMuted
//
//	right := material.Body1(g.theme.Gio(), "Visible: "+g.selectedRecentLinesName)
//	right.Color = g.theme.Color.TextMuted
//
//	if g.isCompactLayout(gtx) {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(left.Layout),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//			layout.Rigid(right.Layout),
//		)
//	}
//
//	return layout.Flex{
//		Axis:      layout.Horizontal,
//		Alignment: layout.Middle,
//	}.Layout(gtx,
//		layout.Rigid(left.Layout),
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
//		layout.Rigid(right.Layout),
//	)
//}
//
//func (g *transcriptGUI) layoutTranscriptActions(gtx layout.Context) layout.Dimensions {
//	launchButton := bareui.Button{
//		Clickable: &g.launchGameButton,
//		Text:      g.transcriptLaunchButtonLabel(),
//		Prefix:    g.transcriptLaunchButtonIcon(),
//		Variant:   g.transcriptLaunchButtonVariant(),
//	}
//	syncButton := bareui.Button{
//		Clickable: &g.syncAnkiButton,
//		Text:      "Sync Anki",
//		Prefix:    "mdi:cloud-sync-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	clearButton := bareui.Button{
//		Clickable: &g.clearButton,
//		Text:      "Clear View",
//		Prefix:    "mdi:broom",
//		Variant:   bareui.ButtonGhost,
//	}
//
//	if g.isCompactLayout(gtx) {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						if g.gameRunning {
//							return launchButton.Layout(gtx.Disabled(), g.theme, g.iconify)
//						}
//						return launchButton.Layout(gtx, g.theme, g.iconify)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						return syncButton.Layout(gtx, g.theme, g.iconify)
//					}),
//				)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return clearButton.Layout(gtx, g.theme, g.iconify)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(g.theme.Gio(), g.transcriptRunningStatusText())
//				lbl.Color = g.theme.Color.TextMuted
//				return lbl.Layout(gtx)
//			}),
//		)
//	}
//
//	return layout.Flex{
//		Axis:      layout.Horizontal,
//		Alignment: layout.Middle,
//	}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if g.gameRunning {
//				return launchButton.Layout(gtx.Disabled(), g.theme, g.iconify)
//			}
//			return launchButton.Layout(gtx, g.theme, g.iconify)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return syncButton.Layout(gtx, g.theme, g.iconify) }),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return clearButton.Layout(gtx, g.theme, g.iconify) }),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body1(g.theme.Gio(), g.transcriptRunningStatusText())
//			lbl.Color = g.theme.Color.TextMuted
//			return lbl.Layout(gtx)
//		}),
//	)
//}
//
//func (g *transcriptGUI) layoutTranscriptEditor(gtx layout.Context) layout.Dimensions {
//	return material.List(g.theme.Gio(), &g.transcriptList).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
//		return layout.Stack{}.Layout(gtx,
//			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
//				label := material.Body1(g.theme.Gio(), g.displayTranscript)
//				label.Color = g.theme.Color.Text
//				label.TextSize = g.transcriptTextSize
//				label.State = &g.transcriptView
//				return label.Layout(gtx)
//			}),
//			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
//				g.paintTranscriptHighlights(gtx)
//				return layout.Dimensions{}
//			}),
//		)
//	})
//}
//
//func (g *transcriptGUI) layoutTranscriptIdleState(gtx layout.Context) layout.Dimensions {
//	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{
//			Axis:      layout.Vertical,
//			Alignment: layout.Middle,
//		}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.H6(g.theme.Gio(), "Transcript Hidden")
//				lbl.Color = g.theme.Color.Text
//				return lbl.Layout(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(g.theme.Gio(), "Start the game to show live transcript text here.")
//				lbl.Color = g.theme.Color.TextMuted
//				return lbl.Layout(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(g.theme.Gio(), "The New Flashcard composer stays on this page below the transcript panel.")
//				lbl.Color = g.theme.Color.TextMuted
//				return lbl.Layout(gtx)
//			}),
//		)
//	})
//}
//
//func (g *transcriptGUI) layoutLookupResults(gtx layout.Context) layout.Dimensions {
//	maxHeight := gtx.Dp(unit.Dp(280))
//	if g.isCompactLayout(gtx) {
//		maxHeight = gtx.Dp(unit.Dp(240))
//	}
//	gtx.Constraints.Max.Y = maxHeight
//	if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y {
//		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
//	}
//
//	return material.List(g.theme.Gio(), &g.lookupResultsList).Layout(gtx, len(g.lookupResults), func(gtx layout.Context, index int) layout.Dimensions {
//		bottom := unit.Dp(0)
//		if index < len(g.lookupResults)-1 {
//			bottom = unit.Dp(10)
//		}
//		lookup := g.lookupResults[index]
//		return layout.Inset{Bottom: bottom}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return g.layoutLookupResultCard(gtx, lookup)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutLookupResultCard(gtx layout.Context, lookup dictionaryLookup) layout.Dimensions {
//	addButton := bareui.Button{
//		Clickable: g.lookupResultAddClickable(lookupResultKey(lookup)),
//		Text:      "mdi:plus-circle-outline",
//		Icon:      true,
//		Variant:   bareui.ButtonPrimary,
//	}
//	playButton := bareui.Button{
//		Clickable: g.lookupResultPlayClickable(lookupResultKey(lookup)),
//		Text:      "mdi:play-circle-outline",
//		Icon:      true,
//		Variant:   bareui.ButtonSecondary,
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.Background, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(g.theme.Gio(), firstNonEmpty(lookup.Query, lookup.Headword))
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					if strings.TrimSpace(lookup.Reading) == "" {
//						return layout.Dimensions{}
//					}
//					lbl := material.Body1(g.theme.Gio(), "Reading: "+lookup.Reading)
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), lookup.Meaning)
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//							return addButton.Layout(gtx, g.theme, g.iconify)
//						}),
//						layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//							if strings.TrimSpace(lookup.AudioPath) == "" {
//								return playButton.Layout(gtx.Disabled(), g.theme, g.iconify)
//							}
//							return playButton.Layout(gtx, g.theme, g.iconify)
//						}),
//					)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutFlashcardComposer(gtx layout.Context) layout.Dimensions {
//	word := material.Editor(g.theme.Gio(), &g.wordEditor, "Word or phrase")
//	word.Color = g.theme.Color.Text
//	word.HintColor = g.theme.Color.TextMuted
//
//	meaning := material.Editor(g.theme.Gio(), &g.meaningEditor, "Meaning")
//	meaning.Color = g.theme.Color.Text
//	meaning.HintColor = g.theme.Color.TextMuted
//
//	searchButton := bareui.Button{
//		Clickable: &g.searchWordButton,
//		Text:      "Lookup",
//		Prefix:    "mdi:book-search-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	playButton := bareui.Button{
//		Clickable: &g.playAudioButton,
//		Text:      "mdi:play-circle-outline",
//		Icon:      true,
//		Prefix:    "mdi:play-circle-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	addAllButton := bareui.Button{
//		Clickable: &g.addAllLookupButton,
//		Text:      "Add All Matches",
//		Prefix:    "mdi:playlist-plus",
//		Variant:   bareui.ButtonSecondary,
//	}
//
//	selected := normalizeGUISelectionText(g.transcriptView.SelectedText())
//	if selected == "" {
//		selected = "Select transcript text to prefill the flashcard word."
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis: layout.Vertical,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(g.theme.Gio(), "New Flashcard")
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), selected)
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(word.Layout),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					minHeight := unit.Dp(120)
//					if g.isCompactLayout(gtx) {
//						minHeight = unit.Dp(102)
//					}
//					gtx.Constraints.Min.Y = gtx.Dp(minHeight)
//					return meaning.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return searchButton.Layout(gtx, g.theme, g.iconify)
//						}),
//						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return playButton.Layout(gtx, g.theme, g.iconify)
//						}),
//					)
//				}),
//				//	layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				//	return saveButton.Layout(gtx, g.theme, g.iconify)
//				//}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					if len(g.lookupResults) <= 1 {
//						return layout.Dimensions{}
//					}
//					return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//						return addAllButton.Layout(gtx, g.theme, g.iconify)
//					})
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				//	lbl := material.Body1(g.theme.Gio(), g.lookupStatus)
//				//	lbl.Color = g.theme.Color.TextMuted
//				//	return lbl.Layout(gtx)
//				//}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					if len(g.lookupResults) == 0 {
//						return layout.Dimensions{}
//					}
//					return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//						return g.layoutLookupResults(gtx)
//					})
//				}),
//				//layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				//	lbl := material.Body1(g.theme.Gio(), "Source line is inferred from the visible transcript when the card is saved.")
//				//	lbl.Color = g.theme.Color.TextMuted
//				//	return lbl.Layout(gtx)
//				//}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutFlashcardsPage(gtx layout.Context) layout.Dimensions {
//	return layout.Flex{
//		Axis: layout.Horizontal,
//	}.Layout(gtx,
//		layout.Flexed(0.58, func(gtx layout.Context) layout.Dimensions {
//			return g.layoutFlashcardList(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(18))),
//		layout.Flexed(0.42, func(gtx layout.Context) layout.Dimensions {
//			return g.layoutFlashcardEditorPanel(gtx)
//		}),
//	)
//}
//
//func (g *transcriptGUI) layoutFlashcardList(gtx layout.Context) layout.Dimensions {
//	return bareutils.Panel(gtx, g.theme.Color.Surface, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			filtered := g.filteredFlashcards()
//			search := material.Editor(g.theme.Gio(), &g.flashcardSearchEditor, "Search flashcards")
//			search.Color = g.theme.Color.Text
//			search.HintColor = g.theme.Color.TextMuted
//
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H5(g.theme.Gio(), fmt.Sprintf("%d Flashcards", len(filtered)))
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(search.Layout),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					if len(filtered) == 0 {
//						return g.layoutEmptyState(gtx, "No matching flashcards for this game.")
//					}
//					return material.List(g.theme.Gio(), &g.flashcardList).Layout(gtx, len(filtered), func(gtx layout.Context, index int) layout.Dimensions {
//						card := filtered[index]
//						return layout.Inset{Bottom: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//							return g.layoutFlashcardRow(gtx, card)
//						})
//					})
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutFlashcardRow(gtx layout.Context, card Flashcard) layout.Dimensions {
//	selectButton := bareui.Button{
//		Clickable: g.flashcardSelectClickable(card.ID),
//		Text:      "Edit",
//		Prefix:    "mdi:pencil-outline",
//		Variant:   g.flashcardRowButtonVariant(card.ID),
//	}
//	deleteButton := bareui.Button{
//		Clickable: g.flashcardDeleteClickable(card.ID),
//		Text:      "Delete",
//		Prefix:    "mdi:delete-outline",
//		Variant:   bareui.ButtonGhost,
//	}
//
//	surfaceColor := g.theme.Color.Background
//	if card.ID == g.selectedFlashcardID {
//		surfaceColor = g.theme.Color.SurfaceAlt
//	}
//
//	return bareutils.Panel(gtx, surfaceColor, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis: layout.Vertical,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(g.theme.Gio(), card.Text)
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), card.Meaning)
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					furigana := flashcardFuriganaText(card)
//					if furigana == "" {
//						return layout.Dimensions{}
//					}
//					lbl := material.Body1(g.theme.Gio(), "Furigana: "+furigana)
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					meta := g.flashcardMetaText(card)
//					if meta == "" {
//						return layout.Dimensions{}
//					}
//					lbl := material.Body1(g.theme.Gio(), meta)
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					source := firstNonEmpty(card.SourceLine, "No source line recorded")
//					lbl := material.Body1(g.theme.Gio(), source)
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return selectButton.Layout(gtx, g.theme, g.iconify)
//						}),
//						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return deleteButton.Layout(gtx, g.theme, g.iconify)
//						}),
//					)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutFlashcardEditorPanel(gtx layout.Context) layout.Dimensions {
//	word := material.Editor(g.theme.Gio(), &g.flashcardWordEditor, "Word or phrase")
//	word.Color = g.theme.Color.Text
//	word.HintColor = g.theme.Color.TextMuted
//
//	meaning := material.Editor(g.theme.Gio(), &g.flashcardMeaningEditor, "Meaning")
//	meaning.Color = g.theme.Color.Text
//	meaning.HintColor = g.theme.Color.TextMuted
//
//	newButton := bareui.Button{
//		Clickable: &g.flashcardNewButton,
//		Text:      "New Flashcard",
//		Prefix:    "mdi:plus-box-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	saveButton := bareui.Button{
//		Clickable: &g.flashcardSaveButton,
//		Text:      g.flashcardSaveButtonLabel(),
//		Prefix:    "mdi:content-save-outline",
//		Variant:   bareui.ButtonPrimary,
//	}
//	deleteButton := bareui.Button{
//		Clickable: &g.flashcardDeleteButton,
//		Text:      "Delete",
//		Prefix:    "mdi:delete-outline",
//		Variant:   bareui.ButtonGhost,
//	}
//	reloadButton := bareui.Button{
//		Clickable: &g.reloadCardsButton,
//		Text:      "Reload Cards",
//		Prefix:    "mdi:refresh",
//		Variant:   bareui.ButtonSecondary,
//	}
//	syncButton := bareui.Button{
//		Clickable: &g.syncAnkiButton,
//		Text:      "Sync Anki",
//		Prefix:    "mdi:cloud-upload-outline",
//		Variant:   bareui.ButtonPrimary,
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis: layout.Vertical,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(g.theme.Gio(), "Flashcard Editor")
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), g.flashcardEditorStatus())
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(word.Layout),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(220))
//					return meaning.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return newButton.Layout(gtx, g.theme, g.iconify)
//						}),
//						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return deleteButton.Layout(gtx, g.theme, g.iconify)
//						}),
//					)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return saveButton.Layout(gtx, g.theme, g.iconify)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), "Deck")
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutSelectableHeadlineText(gtx, "flashcard-editor-deck", firstNonEmpty(ankiDeckName(g.activeGameName), "No deck selected"), g.theme.Color.Text)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), "AnkiConnect URL")
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutSelectableBodyText(gtx, "flashcard-editor-anki-url", guiAnkiURL, g.theme.Color.Text)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return syncButton.Layout(gtx, g.theme, g.iconify) }),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return reloadButton.Layout(gtx, g.theme, g.iconify) }),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutGamePage(gtx layout.Context) layout.Dimensions {
//	installButton := bareui.Button{
//		Clickable: &g.gameInstallHookButton,
//		Text:      g.gameInstallHookButtonLabel(),
//		Prefix:    "mdi:puzzle-plus-outline",
//		Variant:   bareui.ButtonPrimary,
//	}
//	refreshButton := bareui.Button{
//		Clickable: &g.gameRefreshHookButton,
//		Text:      "Refresh Status",
//		Prefix:    "mdi:refresh",
//		Variant:   bareui.ButtonSecondary,
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.Surface, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H5(g.theme.Gio(), "Game Tools")
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), "Install the RPG Maker text hook and add or update game configs from the same page.")
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//						layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
//							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//									return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//										return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//											return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//													lbl := material.H6(g.theme.Gio(), "Text Hook")
//													lbl.Color = g.theme.Color.Text
//													return lbl.Layout(gtx)
//												}),
//												layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//													lbl := material.Body1(g.theme.Gio(), g.gameHookSummaryText())
//													lbl.Color = g.theme.Color.TextMuted
//													return lbl.Layout(gtx)
//												}),
//												layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//													return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//														layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//															return installButton.Layout(gtx, g.theme, g.iconify)
//														}),
//														layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//														layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//															return refreshButton.Layout(gtx, g.theme, g.iconify)
//														}),
//													)
//												}),
//												layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
//												layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//													return g.layoutGameHookDetails(gtx)
//												}),
//											)
//										})
//									})
//								}),
//								layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
//								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//									return g.layoutGameConfigPanel(gtx)
//								}),
//							)
//						}),
//						layout.Rigid(bareutils.SpacerW(unit.Dp(18))),
//						layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
//							return g.layoutBrowsePanel(gtx)
//						}),
//					)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutGameHookStatusCard(gtx layout.Context) layout.Dimensions {
//	return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(g.theme.Gio(), "Hook Status")
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), g.gameHookSummaryText())
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutGameHookDetails(gtx)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutGameHookDetails(gtx layout.Context) layout.Dimensions {
//	lines := []string{
//		"Engine: " + firstNonEmpty(g.gameHookStatus.Engine, "Unknown"),
//		"Project Root: " + firstNonEmpty(g.gameHookStatus.ProjectRoot, "Unavailable"),
//		"Plugin Path: " + firstNonEmpty(g.gameHookStatus.PluginPath, "Unavailable"),
//	}
//	if len(g.gameHookStatus.Compatibility.Findings) > 0 {
//		lines = append(lines, "Compatibility:")
//		lines = append(lines, g.gameHookStatus.Compatibility.Findings...)
//	}
//	lbl := material.Body1(g.theme.Gio(), strings.Join(lines, "\n"))
//	lbl.Color = g.theme.Color.Text
//	return lbl.Layout(gtx)
//}
//
//func (g *transcriptGUI) layoutGameConfigPanel(gtx layout.Context) layout.Dimensions {
//	pathEditor := material.Editor(g.theme.Gio(), &g.newGamePathEditor, "Path to game directory or executable")
//	pathEditor.Color = g.theme.Color.Text
//	pathEditor.HintColor = g.theme.Color.TextMuted
//
//	steamAppIDEditor := material.Editor(g.theme.Gio(), &g.newGameSteamAppIDEditor, "Steam app id (optional unless requires Steam)")
//	steamAppIDEditor.Color = g.theme.Color.Text
//	steamAppIDEditor.HintColor = g.theme.Color.TextMuted
//
//	iconPathEditor := material.Editor(g.theme.Gio(), &g.newGameIconPathEditor, "Icon path (optional)")
//	iconPathEditor.Color = g.theme.Color.Text
//	iconPathEditor.HintColor = g.theme.Color.TextMuted
//
//	imagePathEditor := material.Editor(g.theme.Gio(), &g.newGameImagePathEditor, "Image path (optional)")
//	imagePathEditor.Color = g.theme.Color.Text
//	imagePathEditor.HintColor = g.theme.Color.TextMuted
//
//	saveButton := bareui.Button{
//		Clickable: &g.newGameSaveButton,
//		Text:      "Save Game",
//		Prefix:    "mdi:content-save-outline",
//		Variant:   bareui.ButtonPrimary,
//	}
//	analyzeButton := bareui.Button{
//		Clickable: &g.newGameAnalyzeButton,
//		Text:      "Analyze Path",
//		Prefix:    "mdi:magnify-scan",
//		Variant:   bareui.ButtonSecondary,
//	}
//	browseButton := bareui.Button{
//		Clickable: &g.newGameBrowseButton,
//		Text:      "Browse Current Path",
//		Prefix:    "mdi:folder-search-outline",
//		Variant:   bareui.ButtonGhost,
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(g.theme.Gio(), "Add or Update Game")
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), g.newGameStatus)
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(pathEditor.Layout),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return analyzeButton.Layout(gtx, g.theme, g.iconify)
//						}),
//						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return browseButton.Layout(gtx, g.theme, g.iconify)
//						}),
//					)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.newGameRunnerDropdown.Layout(gtx, g.theme, g.iconify, g.selectedNewGameRunner, g.layoutNewGameRunnerDropdownMenu)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					check := material.CheckBox(g.theme.Gio(), &g.newGameRequiresSteam, "Requires Steam")
//					check.Color = g.theme.Color.Text
//					return check.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(steamAppIDEditor.Layout),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(iconPathEditor.Layout),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Rigid(imagePathEditor.Layout),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutGamePreview(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return saveButton.Layout(gtx, g.theme, g.iconify)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutNewGameRunnerDropdownMenu(gtx layout.Context) layout.Dimensions {
//	return g.layoutOptionMenu(gtx, g.newGameRunnerOptions, g.selectedNewGameRunner)
//}
//
//func (g *transcriptGUI) layoutSettingsPage(gtx layout.Context) layout.Dimensions {
//	return bareutils.Panel(gtx, g.theme.Color.Surface, unit.Dp(g.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis: layout.Vertical,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H5(g.theme.Gio(), "Appearance")
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutSettingRow(gtx, "Mode", g.selectedModeName, func(gtx layout.Context) layout.Dimensions {
//						return g.modeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedModeName, g.layoutModeDropdownMenu)
//					})
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutSettingRow(gtx, "Palette", g.selectedPaletteName, func(gtx layout.Context) layout.Dimensions {
//						return g.paletteDropdown.Layout(gtx, g.theme, g.iconify, g.selectedPaletteName, g.layoutPaletteDropdownMenu)
//					})
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutSettingRow(gtx, "Transcript Size", g.selectedTextSizeName, func(gtx layout.Context) layout.Dimensions {
//						return g.textSizeDropdown.Layout(gtx, g.theme, g.iconify, g.selectedTextSizeName, g.layoutTranscriptSizeDropdownMenu)
//					})
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutSettingRow(gtx, "Visible Transcript", g.selectedRecentLinesName, func(gtx layout.Context) layout.Dimensions {
//						return g.recentLinesDropdown.Layout(gtx, g.theme, g.iconify, g.selectedRecentLinesName, g.layoutRecentLinesDropdownMenu)
//					})
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return g.layoutSettingRow(gtx, "Highlight Audio", boolSettingLabel(g.autoPlayHighlightAudio.Value), func(gtx layout.Context) layout.Dimensions {
//						check := material.CheckBox(g.theme.Gio(), &g.autoPlayHighlightAudio, "Auto-play audio when a highlighted word is clicked")
//						check.Color = g.theme.Color.Text
//						return check.Layout(gtx)
//					})
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), "Mode, transcript rendering, and highlight click behavior can be tuned without changing the watcher logic.")
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutSettingRow(gtx layout.Context, label, current string, control layout.Widget) layout.Dimensions {
//	return bareutils.Panel(gtx, g.theme.Color.Background, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis:      layout.Horizontal,
//				Alignment: layout.Middle,
//			}.Layout(gtx,
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{
//						Axis: layout.Vertical,
//					}.Layout(gtx,
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							lbl := material.Body1(g.theme.Gio(), label)
//							lbl.Color = g.theme.Color.TextMuted
//							return lbl.Layout(gtx)
//						}),
//						layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							lbl := material.H6(g.theme.Gio(), current)
//							lbl.Color = g.theme.Color.Text
//							return lbl.Layout(gtx)
//						}),
//					)
//				}),
//				layout.Rigid(control),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutModeDropdownMenu(gtx layout.Context) layout.Dimensions {
//	return g.layoutOptionMenu(gtx, g.modeOptions, g.selectedModeName)
//}
//
//func (g *transcriptGUI) layoutPaletteDropdownMenu(gtx layout.Context) layout.Dimensions {
//	return g.layoutOptionMenu(gtx, g.paletteOptions, g.selectedPaletteName)
//}
//
//func (g *transcriptGUI) layoutTranscriptSizeDropdownMenu(gtx layout.Context) layout.Dimensions {
//	return g.layoutOptionMenu(gtx, g.transcriptSizeOptions, g.selectedTextSizeName)
//}
//
//func (g *transcriptGUI) layoutRecentLinesDropdownMenu(gtx layout.Context) layout.Dimensions {
//	return g.layoutOptionMenu(gtx, g.recentLineOptions, g.selectedRecentLinesName)
//}
//
//func (g *transcriptGUI) layoutOptionMenu(gtx layout.Context, options []guiDropdownOption, selected string) layout.Dimensions {
//	children := make([]layout.FlexChild, 0, len(options))
//	for i := range options {
//		opt := options[i]
//		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			btn := bareui.Button{
//				Clickable: opt.Clickable,
//				Text:      opt.Label,
//				Prefix:    opt.Icon,
//				Variant:   dropdownButtonVariant(opt.Label == selected),
//			}
//			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return btn.Layout(gtx, g.theme, g.iconify)
//			})
//		}))
//	}
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
//}
//
//func (g *transcriptGUI) layoutEmptyState(gtx layout.Context, text string) layout.Dimensions {
//	lbl := material.Body1(g.theme.Gio(), text)
//	lbl.Color = g.theme.Color.TextMuted
//	return lbl.Layout(gtx)
//}
//
//func (g *transcriptGUI) layoutOverlay(gtx layout.Context) layout.Dimensions {
//	return layout.Stack{}.Layout(gtx,
//		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
//			if !g.messageModal.Open {
//				return layout.Dimensions{}
//			}
//			return g.messageModal.Layout(gtx, g.theme, g.messageTitle, g.layoutMessageModalContent)
//		}),
//		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
//			if !g.browseModal.Open {
//				return layout.Dimensions{}
//			}
//			return g.browseModal.Layout(gtx, g.theme, "Browse Game Path", g.layoutBrowseModalContent)
//		}),
//	)
//}
//
//func (g *transcriptGUI) layoutMessageModalContent(gtx layout.Context) layout.Dimensions {
//	if g.popupFlashcard != nil {
//		return g.layoutTranscriptFlashcardPopup(gtx, *g.popupFlashcard)
//	}
//	return g.layoutSelectableBodyText(gtx, "message-modal-body", g.messageBody, g.theme.Color.Text)
//}
//
//func (g *transcriptGUI) layoutTranscriptFlashcardPopup(gtx layout.Context, card Flashcard) layout.Dimensions {
//	audioButton := bareui.Button{
//		Clickable: &g.transcriptPopupAudioButton,
//		Text:      "Play Audio",
//		Prefix:    "mdi:play-circle-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return g.layoutSelectableHeadlineText(gtx, "popup-card-text-"+card.ID, card.Text, g.theme.Color.Text)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return g.layoutSelectableBodyText(gtx, "popup-card-meaning-"+card.ID, card.Meaning, g.theme.Color.Text)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			var lines []string
//			if furigana := strings.TrimSpace(flashcardFuriganaText(card)); furigana != "" {
//				lines = append(lines, "Furigana: "+furigana)
//			}
//			if reading := strings.TrimSpace(card.Reading); reading != "" {
//				lines = append(lines, "Reading: "+reading)
//			}
//			if pronunciation := strings.TrimSpace(card.PronunciationText); pronunciation != "" {
//				lines = append(lines, "Pronunciation: "+pronunciation)
//			}
//			meta := strings.Join(lines, "\n")
//			if meta == "" {
//				return layout.Dimensions{}
//			}
//			return g.layoutSelectableBodyText(gtx, "popup-card-meta-"+card.ID, meta, g.theme.Color.TextMuted)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if !isExistingFile(card.AudioPath) {
//				return layout.Dimensions{}
//			}
//			return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return audioButton.Layout(gtx, g.theme, g.iconify)
//			})
//		}),
//	)
//}
//
//func (g *transcriptGUI) layoutBrowseModalContent(gtx layout.Context) layout.Dimensions {
//	return g.layoutBrowsePanelContents(gtx, true)
//}
//
//func (g *transcriptGUI) layoutBrowsePanel(gtx layout.Context) layout.Dimensions {
//	return bareutils.Panel(gtx, g.theme.Color.SurfaceAlt, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(g.theme.Gio(), "File Browser")
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), "Pick a game folder or select a specific `.exe` directly.")
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					return g.layoutBrowsePanelContents(gtx, false)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) layoutBrowsePanelContents(gtx layout.Context, showTitle bool) layout.Dimensions {
//	upButton := bareui.Button{
//		Clickable: &g.browseUpButton,
//		Text:      "Up",
//		Prefix:    "mdi:arrow-up",
//		Variant:   bareui.ButtonSecondary,
//	}
//	selectButton := bareui.Button{
//		Clickable: &g.browseUseCurrentButton,
//		Text:      "Use This Folder",
//		Prefix:    "mdi:check-bold",
//		Variant:   bareui.ButtonPrimary,
//	}
//
//	pathLabel := material.Body1(g.theme.Gio(), firstNonEmpty(g.browseCurrentPath, "No directory loaded"))
//	pathLabel.Color = g.theme.Color.Text
//
//	children := []layout.FlexChild{}
//	if showTitle {
//		children = append(children,
//			layout.Rigid(pathLabel.Layout),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//		)
//	} else {
//		children = append(children,
//			layout.Rigid(pathLabel.Layout),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//		)
//	}
//
//	children = append(children,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return upButton.Layout(gtx, g.theme, g.iconify)
//				}),
//				layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return selectButton.Layout(gtx, g.theme, g.iconify)
//				}),
//			)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if strings.TrimSpace(g.browseError) == "" {
//				return layout.Dimensions{}
//			}
//			lbl := material.Body1(g.theme.Gio(), g.browseError)
//			lbl.Color = g.theme.Color.TextMuted
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			if len(g.browseEntries) == 0 {
//				return g.layoutEmptyState(gtx, "No folders or `.exe` files found here.")
//			}
//			return material.List(g.theme.Gio(), &g.browseList).Layout(gtx, len(g.browseEntries), func(gtx layout.Context, index int) layout.Dimensions {
//				entry := g.browseEntries[index]
//				return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//					return g.layoutBrowseEntry(gtx, entry)
//				})
//			})
//		}),
//	)
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
//}
//
//func (g *transcriptGUI) layoutBrowseEntry(gtx layout.Context, entry browseEntry) layout.Dimensions {
//	btn := bareui.Button{
//		Clickable: g.browseEntryClickable(entry.Path),
//		Text:      entry.Name,
//		Prefix:    browseEntryIcon(entry),
//		Variant:   bareui.ButtonSecondary,
//	}
//	return btn.Layout(gtx, g.theme, g.iconify)
//}
//
//func (g *transcriptGUI) layoutGamePreview(gtx layout.Context) layout.Dimensions {
//	preview := g.newGamePreview
//	lines := []string{}
//	switch {
//	case strings.TrimSpace(preview.Error) != "":
//		lines = append(lines, "Error: "+preview.Error)
//	case strings.TrimSpace(preview.Name) == "":
//		lines = append(lines, "Analyze a path to preview the detected executable, assets, and runner.")
//	default:
//		lines = append(lines, "Name: "+preview.Name)
//		lines = append(lines, "Resolved Path: "+firstNonEmpty(preview.ResolvedPath, "Unavailable"))
//		lines = append(lines, "Executable: "+firstNonEmpty(preview.Executable, "Unavailable"))
//		lines = append(lines, "Working Dir: "+firstNonEmpty(preview.WorkingDir, "Unavailable"))
//		lines = append(lines, "Runner: "+firstNonEmpty(preview.Runner, "Unavailable"))
//		if strings.TrimSpace(preview.SteamAppID) != "" {
//			lines = append(lines, "Steam App ID: "+preview.SteamAppID)
//		}
//		if strings.TrimSpace(preview.IconPath) != "" {
//			lines = append(lines, "Icon: "+preview.IconPath)
//		}
//		if strings.TrimSpace(preview.ImagePath) != "" {
//			lines = append(lines, "Image: "+preview.ImagePath)
//		}
//		if preview.Verified {
//			lines = append(lines, "Verification: passed")
//		}
//	}
//
//	return bareutils.Panel(gtx, g.theme.Color.Background, unit.Dp(g.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), "Preview")
//					lbl.Color = g.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(g.theme.Gio(), strings.Join(lines, "\n"))
//					lbl.Color = g.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//			)
//		})
//	})
//}
//
//func (g *transcriptGUI) browseEntryClickable(path string) *widget.Clickable {
//	if g.browseEntryClicks[path] == nil {
//		g.browseEntryClicks[path] = new(widget.Clickable)
//	}
//	return g.browseEntryClicks[path]
//}
//
//func browseEntryIcon(entry browseEntry) string {
//	if entry.IsDir {
//		return "mdi:folder-outline"
//	}
//	return "mdi:application-outline"
//}
//
//func (g *transcriptGUI) initializeBrowsePath(input string) {
//	target := strings.TrimSpace(input)
//	if target == "" {
//		if home, err := os.UserHomeDir(); err == nil {
//			target = home
//		}
//	}
//	if target == "" {
//		target = "."
//	}
//	g.loadBrowseEntries(target)
//}
//
//func (g *transcriptGUI) loadBrowseEntries(input string) {
//	resolved, err := filepath.Abs(strings.TrimSpace(input))
//	if err != nil {
//		g.browseError = err.Error()
//		return
//	}
//
//	info, err := os.Stat(resolved)
//	if err != nil {
//		g.browseError = err.Error()
//		return
//	}
//
//	targetDir := resolved
//	if !info.IsDir() {
//		targetDir = filepath.Dir(resolved)
//	}
//
//	entries, err := os.ReadDir(targetDir)
//	if err != nil {
//		g.browseError = err.Error()
//		return
//	}
//
//	results := make([]browseEntry, 0, len(entries))
//	for _, entry := range entries {
//		fullPath := filepath.Join(targetDir, entry.Name())
//		if entry.IsDir() {
//			results = append(results, browseEntry{
//				Name:  entry.Name(),
//				Path:  fullPath,
//				IsDir: true,
//			})
//			continue
//		}
//		if isExeFile(fullPath) {
//			results = append(results, browseEntry{
//				Name:  entry.Name(),
//				Path:  fullPath,
//				IsDir: false,
//			})
//		}
//	}
//
//	sort.Slice(results, func(i, j int) bool {
//		if results[i].IsDir != results[j].IsDir {
//			return results[i].IsDir
//		}
//		return strings.ToLower(results[i].Name) < strings.ToLower(results[j].Name)
//	})
//
//	g.browseCurrentPath = targetDir
//	g.browseEntries = results
//	g.browseError = ""
//
//	valid := make(map[string]struct{}, len(results))
//	for _, entry := range results {
//		valid[entry.Path] = struct{}{}
//		if g.browseEntryClicks[entry.Path] == nil {
//			g.browseEntryClicks[entry.Path] = new(widget.Clickable)
//		}
//	}
//	for path := range g.browseEntryClicks {
//		if _, ok := valid[path]; !ok {
//			delete(g.browseEntryClicks, path)
//		}
//	}
//}
//
//func (g *transcriptGUI) browseUp() {
//	current := strings.TrimSpace(g.browseCurrentPath)
//	if current == "" {
//		g.initializeBrowsePath("")
//		return
//	}
//	parent := filepath.Dir(current)
//	g.loadBrowseEntries(parent)
//}
//
//func (g *transcriptGUI) selectCurrentBrowsePath() {
//	if strings.TrimSpace(g.browseCurrentPath) == "" {
//		g.newGameStatus = "Browse to a folder before selecting it."
//		return
//	}
//	g.newGamePathEditor.SetText(g.browseCurrentPath)
//	g.newGameStatus = "Selected folder from browser."
//	g.analyzeNewGamePath()
//}
//
//func (g *transcriptGUI) handleBrowseSelection(path string) {
//	for _, entry := range g.browseEntries {
//		if entry.Path != path {
//			continue
//		}
//		if entry.IsDir {
//			g.loadBrowseEntries(entry.Path)
//			return
//		}
//		g.newGamePathEditor.SetText(entry.Path)
//		g.newGameStatus = "Selected executable from browser."
//		g.analyzeNewGamePath()
//		return
//	}
//}
//
//func (g *transcriptGUI) reloadFlashcards() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.flashcards = nil
//		g.prepareNewFlashcard()
//		return
//	}
//	cards, err := loadFlashcards(g.activeGameName)
//	if err != nil {
//		g.showMessage("Flashcard Load Failed", err.Error())
//		return
//	}
//	sort.Slice(cards, func(i, j int) bool {
//		return cards[i].UpdatedAt.After(cards[j].UpdatedAt)
//	})
//	g.flashcards = cards
//	g.syncFlashcardRowState()
//	if g.selectedFlashcardID != "" {
//		for _, card := range cards {
//			if card.ID == g.selectedFlashcardID {
//				g.loadFlashcardIntoEditor(card)
//				return
//			}
//		}
//	}
//
//	g.prepareNewFlashcard()
//}
//
//func (g *transcriptGUI) gameHookSummaryText() string {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		return "Select a game to inspect text hook support."
//	}
//	if !g.gameHookStatus.Supported {
//		return firstNonEmpty(g.gameHookStatus.Message, "This game does not look like a supported RPG Maker MV/MZ project.")
//	}
//	return g.gameHookStatus.Message
//}
//
//func (g *transcriptGUI) gameInstallHookButtonLabel() string {
//	if g.gameHookStatus.Installed {
//		return "Reinstall Text Hook"
//	}
//	return "Install Text Hook"
//}
//
//func (g *transcriptGUI) refreshCurrentGameHookStatus() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.gameHookStatus = textHookStatus{Message: "Select a game to inspect text hook support."}
//		return
//	}
//	inputPath := firstNonEmpty(g.currentConfig.GamePath, g.currentConfig.Executable)
//	if strings.TrimSpace(inputPath) == "" {
//		g.gameHookStatus = textHookStatus{Message: "Selected game does not have a configured path or executable."}
//		return
//	}
//	status, err := inspectRPGMakerClipboardHook(inputPath)
//	if err != nil {
//		g.gameHookStatus = textHookStatus{Message: err.Error()}
//		return
//	}
//	g.gameHookStatus = status
//}
//
//func (g *transcriptGUI) installCurrentGameTextHook() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.showMessage("Text Hook Install Failed", "Select a game before installing the text hook.")
//		return
//	}
//	inputPath := firstNonEmpty(g.currentConfig.GamePath, g.currentConfig.Executable)
//	if strings.TrimSpace(inputPath) == "" {
//		g.showMessage("Text Hook Install Failed", "Selected game does not have a configured path or executable.")
//		return
//	}
//	result, err := installRPGMakerClipboardHook(inputPath)
//	if err != nil {
//		g.showMessage("Text Hook Install Failed", err.Error())
//		return
//	}
//	g.gameHookStatus = textHookStatus{
//		Supported:         true,
//		Installed:         true,
//		Engine:            result.Engine,
//		ProjectRoot:       result.Compatibility.ProjectRoot,
//		PluginPath:        result.PluginPath,
//		PluginsConfigPath: result.PluginsConfigPath,
//		Compatibility:     result.Compatibility,
//		Message:           "Text hook plugin is installed and enabled.",
//	}
//	g.showMessage("Text Hook Installed", "Installed the clipboard text hook plugin for this game.")
//}
//
//func (g *transcriptGUI) launchCurrentGameInBackground() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.showMessage("Launch Failed", "Select a game before launching it.")
//		return
//	}
//	if strings.TrimSpace(g.currentConfig.Name) == "" {
//		g.showMessage("Launch Failed", "The selected game configuration is not loaded yet.")
//		return
//	}
//	g.refreshCurrentGameRunningState(true)
//	if g.gameRunning {
//		g.statusText = g.transcriptRunningStatusText()
//		return
//	}
//	if err := launchGameInBackground(g.currentConfig); err != nil {
//		g.statusText = err.Error()
//		g.showMessage("Launch Failed", err.Error())
//		return
//	}
//	g.statusText = fmt.Sprintf("Launching %s in the background.", g.currentConfig.Name)
//	g.lastGameRunningCheck = time.Time{}
//	g.refreshCurrentGameRunningState(true)
//}
//
//func (g *transcriptGUI) saveNewGame(ctx context.Context, w *app.Window) {
//	cfg, err := g.buildNewGameConfig()
//	if err != nil {
//		g.newGameStatus = err.Error()
//		g.newGamePreview.Error = err.Error()
//		g.showMessage("Save Game Failed", err.Error())
//		return
//	}
//
//	g.newGameStatus = "Verifying launch settings..."
//	verifiedCfg, err := verifyAndAutofixGameConfig(cfg)
//	if err != nil {
//		g.newGamePreview = g.previewFromConfig(cfg, err)
//		g.newGameStatus = err.Error()
//		g.showMessage("Save Game Failed", err.Error())
//		return
//	}
//
//	if _, err := saveGameConfig(verifiedCfg); err != nil {
//		g.newGameStatus = err.Error()
//		g.newGamePreview.Error = err.Error()
//		g.showMessage("Save Game Failed", err.Error())
//		return
//	}
//
//	configs, err := listGameConfigs()
//	if err != nil {
//		g.newGameStatus = err.Error()
//		g.showMessage("Save Game Failed", err.Error())
//		return
//	}
//
//	g.configs = configs
//	if g.gameOptionClicks == nil {
//		g.gameOptionClicks = make(map[string]*widget.Clickable)
//	}
//	for _, saved := range configs {
//		if g.gameOptionClicks[saved.Name] == nil {
//			g.gameOptionClicks[saved.Name] = new(widget.Clickable)
//		}
//	}
//
//	g.newGamePreview = g.previewFromConfig(verifiedCfg, nil)
//	g.newGamePreview.Verified = true
//	g.newGameStatus = fmt.Sprintf("Saved %q. Switched to the game.", verifiedCfg.Name)
//	g.pageTabs.Active = guiPageTranscript
//	g.startWatching(ctx, verifiedCfg.Name, w)
//	g.resetNewGameForm()
//	g.showMessage("Game Saved", fmt.Sprintf("Saved %q and added it to the game list.", verifiedCfg.Name))
//}
//
//func (g *transcriptGUI) analyzeNewGamePath() {
//	cfg, err := g.buildNewGameConfig()
//	if err != nil {
//		g.newGamePreview = gamePathPreview{Error: err.Error()}
//		g.newGameStatus = err.Error()
//		return
//	}
//	g.newGamePreview = g.previewFromConfig(cfg, nil)
//	g.newGameStatus = fmt.Sprintf("Ready to save %q.", cfg.Name)
//}
//
//func (g *transcriptGUI) buildNewGameConfig() (GameConfig, error) {
//	inputPath := strings.TrimSpace(g.newGamePathEditor.Text())
//	if inputPath == "" {
//		return GameConfig{}, errors.New("game path is required")
//	}
//	return buildGameConfig(
//		inputPath,
//		g.selectedNewGameRunner,
//		g.newGameRequiresSteam.Value,
//		strings.TrimSpace(g.newGameSteamAppIDEditor.Text()),
//		strings.TrimSpace(g.newGameIconPathEditor.Text()),
//		strings.TrimSpace(g.newGameImagePathEditor.Text()),
//	)
//}
//
//func (g *transcriptGUI) previewFromConfig(cfg GameConfig, err error) gamePathPreview {
//	preview := gamePathPreview{
//		ResolvedPath: cfg.GamePath,
//		Executable:   cfg.Executable,
//		WorkingDir:   cfg.WorkingDir,
//		IconPath:     cfg.IconPath,
//		ImagePath:    cfg.ImagePath,
//		Name:         cfg.Name,
//		Runner:       string(cfg.Runner),
//		SteamAppID:   cfg.SteamAppID,
//		Verified:     cfg.Verification.Verified,
//	}
//	if err != nil {
//		preview.Error = err.Error()
//	}
//	return preview
//}
//
//func (g *transcriptGUI) resetNewGameForm() {
//	g.newGamePathEditor.SetText("")
//	g.newGameSteamAppIDEditor.SetText("")
//	g.newGameIconPathEditor.SetText("")
//	g.newGameImagePathEditor.SetText("")
//	g.newGameRequiresSteam.Value = false
//	g.selectedNewGameRunner = "auto"
//	g.newGamePreview = gamePathPreview{}
//	if home, err := os.UserHomeDir(); err == nil {
//		g.loadBrowseEntries(home)
//	}
//}
//
//func (g *transcriptGUI) syncFlashcardRowState() {
//	valid := make(map[string]struct{}, len(g.flashcards))
//	for _, card := range g.flashcards {
//		valid[card.ID] = struct{}{}
//		if g.flashcardSelectClicks[card.ID] == nil {
//			g.flashcardSelectClicks[card.ID] = new(widget.Clickable)
//		}
//		if g.flashcardDeleteClicks[card.ID] == nil {
//			g.flashcardDeleteClicks[card.ID] = new(widget.Clickable)
//		}
//	}
//	for id := range g.flashcardSelectClicks {
//		if _, ok := valid[id]; !ok {
//			delete(g.flashcardSelectClicks, id)
//		}
//	}
//	for id := range g.flashcardDeleteClicks {
//		if _, ok := valid[id]; !ok {
//			delete(g.flashcardDeleteClicks, id)
//		}
//	}
//}
//
//func (g *transcriptGUI) filteredFlashcards() []Flashcard {
//	query := strings.TrimSpace(strings.ToLower(g.flashcardSearchEditor.Text()))
//	if query == "" {
//		return g.flashcards
//	}
//	filtered := make([]Flashcard, 0, len(g.flashcards))
//	for _, card := range g.flashcards {
//		haystack := strings.ToLower(strings.Join([]string{
//			card.Text,
//			card.Meaning,
//			card.Reading,
//			card.SourceLine,
//		}, "\n"))
//		if strings.Contains(haystack, query) {
//			filtered = append(filtered, card)
//		}
//	}
//	return filtered
//}
//
//func (g *transcriptGUI) flashcardSelectClickable(cardID string) *widget.Clickable {
//	if g.flashcardSelectClicks[cardID] == nil {
//		g.flashcardSelectClicks[cardID] = new(widget.Clickable)
//	}
//	return g.flashcardSelectClicks[cardID]
//}
//
//func (g *transcriptGUI) flashcardDeleteClickable(cardID string) *widget.Clickable {
//	if g.flashcardDeleteClicks[cardID] == nil {
//		g.flashcardDeleteClicks[cardID] = new(widget.Clickable)
//	}
//	return g.flashcardDeleteClicks[cardID]
//}
//
//func (g *transcriptGUI) flashcardRowButtonVariant(cardID string) bareui.ButtonVariant {
//	if cardID == g.selectedFlashcardID {
//		return bareui.ButtonPrimary
//	}
//	return bareui.ButtonSecondary
//}
//
//func (g *transcriptGUI) flashcardSaveButtonLabel() string {
//	if strings.TrimSpace(g.selectedFlashcardID) != "" {
//		return "Save Changes"
//	}
//	return "Create Flashcard"
//}
//
//func (g *transcriptGUI) flashcardEditorStatus() string {
//	if strings.TrimSpace(g.selectedFlashcardID) != "" {
//		return "Editing selected flashcard."
//	}
//	return "Create a new card or pick one from the list to edit."
//}
//
//func (g *transcriptGUI) prepareNewFlashcard() {
//	g.selectedFlashcardID = ""
//	g.flashcardWordEditor.SetText("")
//	g.flashcardMeaningEditor.SetText("")
//}
//
//func (g *transcriptGUI) selectFlashcard(cardID string) {
//	for _, card := range g.flashcards {
//		if card.ID == cardID {
//			g.loadFlashcardIntoEditor(card)
//			return
//		}
//	}
//}
//
//func (g *transcriptGUI) loadFlashcardIntoEditor(card Flashcard) {
//	g.selectedFlashcardID = card.ID
//	g.flashcardWordEditor.SetText(card.Text)
//	g.flashcardMeaningEditor.SetText(card.Meaning)
//}
//
//func (g *transcriptGUI) saveFlashcard() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.showMessage("Create Flashcard Failed", "Select a game before creating a flashcard.")
//		return
//	}
//
//	if selected := normalizeGUISelectionText(g.transcriptView.SelectedText()); selected != "" && strings.TrimSpace(g.wordEditor.Text()) == "" {
//		g.wordEditor.SetText(selected)
//	}
//
//	word := normalizeGUISelectionText(g.wordEditor.Text())
//	if word == "" {
//		g.showMessage("Create Flashcard Failed", "Flashcard word cannot be empty.")
//		return
//	}
//
//	meaning := strings.TrimSpace(g.meaningEditor.Text())
//	if meaning == "" {
//		g.showMessage("Create Flashcard Failed", "Flashcard meaning cannot be empty.")
//		return
//	}
//
//	card := Flashcard{
//		GameName:   g.currentConfig.Name,
//		Text:       word,
//		Meaning:    meaning,
//		SourcePath: g.logPath,
//		SourceLine: findFlashcardSourceLine(g.displayTranscript, word),
//	}
//	if g.lookupResult != nil && strings.TrimSpace(g.lookupResult.Query) == word {
//		card.Reading = g.lookupResult.Reading
//		card.PronunciationText = g.lookupResult.PronunciationText
//		card.PronunciationPitch = g.lookupResult.Pitch
//		card.AudioPath = g.lookupResult.AudioPath
//	}
//	if err := addFlashcard(card); err != nil {
//		g.showMessage("Create Flashcard Failed", err.Error())
//		return
//	}
//
//	g.wordEditor.SetText("")
//	g.meaningEditor.SetText("")
//	g.lookupResult = nil
//	g.lookupResults = nil
//	g.lookupStatus = "Dictionary lookup can fill the meaning and fetch audio for the current word."
//	g.reloadFlashcards()
//
//	message := fmt.Sprintf("Saved %q to deck %q.", card.Text, ankiDeckName(card.GameName))
//	if strings.TrimSpace(card.SourceLine) != "" {
//		message += "\nSource: " + card.SourceLine
//	}
//	g.showMessage("Flashcard Saved", message)
//}
//
//func (g *transcriptGUI) saveFlashcardFromLibrary() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.showMessage("Save Flashcard Failed", "Select a game before editing flashcards.")
//		return
//	}
//
//	word := normalizeGUISelectionText(g.flashcardWordEditor.Text())
//	meaning := strings.TrimSpace(g.flashcardMeaningEditor.Text())
//	if word == "" {
//		g.showMessage("Save Flashcard Failed", "Flashcard word cannot be empty.")
//		return
//	}
//	if meaning == "" {
//		g.showMessage("Save Flashcard Failed", "Flashcard meaning cannot be empty.")
//		return
//	}
//
//	if strings.TrimSpace(g.selectedFlashcardID) == "" {
//		card := Flashcard{
//			GameName:   g.currentConfig.Name,
//			Text:       word,
//			Meaning:    meaning,
//			SourcePath: g.logPath,
//			SourceLine: findFlashcardSourceLine(g.displayTranscript, word),
//		}
//		if err := addFlashcard(card); err != nil {
//			g.showMessage("Save Flashcard Failed", err.Error())
//			return
//		}
//		g.reloadFlashcards()
//		g.showMessage("Flashcard Saved", fmt.Sprintf("Created %q in deck %q.", card.Text, ankiDeckName(card.GameName)))
//		return
//	}
//
//	for _, existing := range g.flashcards {
//		if existing.ID != g.selectedFlashcardID {
//			continue
//		}
//		existing.Text = word
//		existing.Meaning = meaning
//		existing.UpdatedAt = time.Now().UTC()
//		if strings.TrimSpace(existing.SourceLine) == "" {
//			existing.SourceLine = findFlashcardSourceLine(g.displayTranscript, word)
//		}
//		if err := updateFlashcard(existing); err != nil {
//			g.showMessage("Save Flashcard Failed", err.Error())
//			return
//		}
//		g.reloadFlashcards()
//		g.showMessage("Flashcard Saved", fmt.Sprintf("Updated %q.", existing.Text))
//		return
//	}
//
//	g.showMessage("Save Flashcard Failed", "Selected flashcard could not be found.")
//}
//
//func (g *transcriptGUI) deleteSelectedFlashcard() {
//	if strings.TrimSpace(g.selectedFlashcardID) == "" {
//		g.showMessage("Delete Flashcard Failed", "Select a flashcard before deleting it.")
//		return
//	}
//	g.deleteFlashcardByID(g.selectedFlashcardID)
//}
//
//func (g *transcriptGUI) deleteFlashcardByID(cardID string) {
//	for _, card := range g.flashcards {
//		if card.ID != cardID {
//			continue
//		}
//		if err := anki.deleteFlashcardFromAnki(card, guiAnkiURL, guiAnkiPushSync); err != nil {
//			g.showMessage("Delete Flashcard Failed", err.Error()+"\n\nMake sure the AnkiConnect add-on is installed:\nhttps://ankiweb.net/shared/info/2055492159")
//			return
//		}
//		if err := deleteFlashcard(g.currentConfig.Name, cardID); err != nil {
//			g.showMessage("Delete Flashcard Failed", err.Error())
//			return
//		}
//		if g.selectedFlashcardID == cardID {
//			g.prepareNewFlashcard()
//		}
//		g.reloadFlashcards()
//		g.showMessage("Flashcard Deleted", fmt.Sprintf("Deleted %q.", card.Text))
//		return
//	}
//	g.showMessage("Delete Flashcard Failed", "Selected flashcard could not be found.")
//}
//
//func (g *transcriptGUI) syncCurrentGameToAnki() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.showMessage("Anki Sync Failed", "Select a game before syncing Anki.")
//		return
//	}
//
//	result, err := anki.syncFlashcardsToAnki(g.currentConfig.Name, guiAnkiURL, guiAnkiPushSync)
//	if err != nil {
//		g.showMessage("Anki Sync Failed", err.Error()+"\n\nMake sure the AnkiConnect add-on is installed:\nhttps://ankiweb.net/shared/info/2055492159")
//		return
//	}
//
//	g.reloadFlashcards()
//	g.showMessage("Anki Sync Complete", fmt.Sprintf(
//		"Synced %d cards to deck %q.\nCreated: %d\nUpdated: %d",
//		result.Total,
//		result.DeckName,
//		result.Created,
//		result.Updated,
//	))
//}
//
//func (g *transcriptGUI) lookupCurrentWord() {
//	if selected := normalizeGUISelectionText(g.transcriptView.SelectedText()); selected != "" {
//		g.wordEditor.SetText(selected)
//	}
//	g.lookupResult = nil
//	g.lookupResults = nil
//	g.lookupStatus = "Looking up word..."
//	g.meaningEditor.SetText("")
//
//	word := normalizeGUISelectionText(g.wordEditor.Text())
//	if word == "" {
//		g.lookupStatus = "Dictionary lookup can fill the meaning and fetch audio for the current word."
//		g.showMessage("Dictionary Lookup Failed", "Flashcard word cannot be empty.")
//		return
//	}
//
//	lookups, err := lookupDictionaryWords(word)
//	if err != nil {
//		g.lookupResult = nil
//		g.lookupResults = nil
//		g.lookupStatus = "Dictionary lookup failed."
//		g.showMessage("Dictionary Lookup Failed", err.Error())
//		return
//	}
//
//	g.lookupResults = lookups
//	g.lookupResult = &lookups[0]
//	g.wordEditor.SetText(firstNonEmpty(lookups[0].Query, lookups[0].Key, lookups[0].Headword))
//	g.meaningEditor.SetText(lookups[0].Meaning)
//	g.lookupStatus = g.lookupStatusText(lookups)
//}
//
//func (g *transcriptGUI) playCurrentLookupAudio() {
//	if g.lookupResult == nil || strings.TrimSpace(g.lookupResult.AudioPath) == "" {
//		g.showMessage("Audio Playback Failed", "No audio is available for the current lookup.")
//		return
//	}
//	if err := playAudioFile(g.lookupResult.AudioPath); err != nil {
//		g.showMessage("Audio Playback Failed", err.Error())
//	}
//}
//
//func (g *transcriptGUI) lookupStatusText(lookups []dictionaryLookup) string {
//	if len(lookups) == 0 {
//		return "Dictionary lookup can fill the meaning and fetch audio for the current word."
//	}
//
//	lookup := lookups[0]
//	parts := make([]string, 0, 5)
//	if len(lookups) > 1 {
//		parts = append(parts, fmt.Sprintf("%d matches ready to add.", len(lookups)))
//	}
//	if firstNonEmpty(lookup.Query, lookup.Headword, lookup.Key) != "" {
//		parts = append(parts, "Primary: "+firstNonEmpty(lookup.Query, lookup.Headword, lookup.Key))
//	}
//	if lookup.Reading != "" {
//		parts = append(parts, "Reading: "+lookup.Reading)
//	}
//	if lookup.PronunciationText != "" {
//		pronunciation := "Pronunciation: " + lookup.PronunciationText
//		if lookup.Pitch != "" {
//			pronunciation += " (" + lookup.Pitch + ")"
//		}
//		parts = append(parts, pronunciation)
//	}
//	if lookup.AudioPath != "" {
//		parts = append(parts, "Audio ready")
//	} else {
//		parts = append(parts, "Audio unavailable")
//	}
//	return strings.Join(parts, "\n")
//}
//
//func (g *transcriptGUI) flashcardMetaText(card Flashcard) string {
//	parts := make([]string, 0, 3)
//	if strings.TrimSpace(card.Reading) != "" {
//		parts = append(parts, "Reading: "+card.Reading)
//	}
//	if strings.TrimSpace(card.PronunciationText) != "" {
//		line := "Pronunciation: " + card.PronunciationText
//		if strings.TrimSpace(card.PronunciationPitch) != "" {
//			line += " (" + card.PronunciationPitch + ")"
//		}
//		parts = append(parts, line)
//	}
//	if isExistingFile(card.AudioPath) {
//		parts = append(parts, "Audio cached")
//	}
//	return strings.Join(parts, "\n")
//}
//
//func (g *transcriptGUI) lookupResultAddClickable(key string) *widget.Clickable {
//	if g.lookupResultAddClicks[key] == nil {
//		g.lookupResultAddClicks[key] = new(widget.Clickable)
//	}
//	return g.lookupResultAddClicks[key]
//}
//
//func (g *transcriptGUI) lookupResultPlayClickable(key string) *widget.Clickable {
//	if g.lookupResultPlayClicks[key] == nil {
//		g.lookupResultPlayClicks[key] = new(widget.Clickable)
//	}
//	return g.lookupResultPlayClicks[key]
//}
//
//func lookupResultKey(lookup dictionaryLookup) string {
//	return firstNonEmpty(lookup.Key, lookup.Query, lookup.Headword)
//}
//
//func (g *transcriptGUI) flashcardFromLookup(lookup dictionaryLookup) Flashcard {
//	word := firstNonEmpty(lookup.Query, lookup.Headword, lookup.Key)
//	return Flashcard{
//		GameName:           g.currentConfig.Name,
//		Text:               word,
//		Meaning:            lookup.Meaning,
//		Reading:            lookup.Reading,
//		PronunciationText:  lookup.PronunciationText,
//		PronunciationPitch: lookup.Pitch,
//		AudioPath:          lookup.AudioPath,
//		SourcePath:         g.logPath,
//		SourceLine:         findFlashcardSourceLine(g.displayTranscript, word),
//	}
//}
//
//func (g *transcriptGUI) addLookupFlashcardByKey(key string) {
//	key = strings.TrimSpace(key)
//	if key == "" {
//		return
//	}
//	for _, lookup := range g.lookupResults {
//		if lookupResultKey(lookup) != key {
//			continue
//		}
//		card := g.flashcardFromLookup(lookup)
//		if err := addFlashcard(card); err != nil {
//			g.showMessage("Create Flashcard Failed", err.Error())
//			return
//		}
//		g.reloadFlashcards()
//		message := fmt.Sprintf("Saved %q to deck %q.", card.Text, ankiDeckName(card.GameName))
//		if strings.TrimSpace(card.SourceLine) != "" {
//			message += "\nSource: " + card.SourceLine
//		}
//		g.showMessage("Flashcard Saved", message)
//		return
//	}
//}
//
//func (g *transcriptGUI) playLookupAudioByKey(key string) {
//	key = strings.TrimSpace(key)
//	if key == "" {
//		return
//	}
//	for _, lookup := range g.lookupResults {
//		if lookupResultKey(lookup) != key {
//			continue
//		}
//		if strings.TrimSpace(lookup.AudioPath) == "" {
//			g.showMessage("Audio Playback Failed", "No audio is available for this lookup result.")
//			return
//		}
//		if err := playAudioFile(lookup.AudioPath); err != nil {
//			g.showMessage("Audio Playback Failed", err.Error())
//		}
//		return
//	}
//}
//
//func (g *transcriptGUI) addAllLookupFlashcards() {
//	if strings.TrimSpace(g.activeGameName) == "" {
//		g.showMessage("Create Flashcards Failed", "Select a game before creating flashcards.")
//		return
//	}
//	if len(g.lookupResults) == 0 {
//		g.showMessage("Create Flashcards Failed", "Run Dictionary Lookup first.")
//		return
//	}
//
//	cards := make([]Flashcard, 0, len(g.lookupResults))
//	for _, lookup := range g.lookupResults {
//		cards = append(cards, g.flashcardFromLookup(lookup))
//	}
//	added, skipped, err := addFlashcards(g.currentConfig.Name, cards)
//	if err != nil {
//		g.showMessage("Create Flashcards Failed", err.Error())
//		return
//	}
//	g.reloadFlashcards()
//
//	message := fmt.Sprintf("Added %d flashcards to deck %q.", added, ankiDeckName(g.currentConfig.Name))
//	if skipped > 0 {
//		message += fmt.Sprintf("\nSkipped %d duplicate matches.", skipped)
//	}
//	g.showMessage("Flashcards Added", message)
//}
//
//func (g *transcriptGUI) clearTranscript() {
//	g.mu.Lock()
//	defer g.mu.Unlock()
//
//	g.rawTranscript = ""
//	g.displayTranscript = ""
//	g.displayDirty = true
//	g.transcriptResetView = true
//
//	if info, err := os.Stat(g.logPath); err == nil {
//		g.offset = info.Size()
//	} else {
//		g.offset = 0
//	}
//	g.statusText = "Transcript view cleared; waiting for new dialogue."
//}
//
//func (g *transcriptGUI) showMessage(title, body string) {
//	g.popupFlashcard = nil
//	g.messageTitle = title
//	g.messageBody = body
//	g.messageModal.Open = true
//}
//
//func (g *transcriptGUI) showFlashcardPopup(card Flashcard) {
//	cardCopy := card
//	g.popupFlashcard = &cardCopy
//	g.messageTitle = card.Text
//	g.messageBody = ""
//	g.messageModal.Open = true
//}
//
//func (g *transcriptGUI) playPopupFlashcardAudio() {
//	if g.popupFlashcard == nil {
//		g.showMessage("Audio Playback Failed", "No flashcard is selected.")
//		return
//	}
//	g.playFlashcardAudio(*g.popupFlashcard, true)
//}
//
//func (g *transcriptGUI) playFlashcardAudio(card Flashcard, notifyMissing bool) {
//	if !isExistingFile(card.AudioPath) {
//		if notifyMissing {
//			g.showMessage("Audio Playback Failed", "No audio is available for this flashcard.")
//		}
//		return
//	}
//	if err := playAudioFile(card.AudioPath); err != nil {
//		g.showMessage("Audio Playback Failed", err.Error())
//	}
//}
//
//func (g *transcriptGUI) syncTranscriptEditor() {
//	g.mu.Lock()
//	defer g.mu.Unlock()
//
//	if !g.displayDirty {
//		return
//	}
//
//	g.transcriptView.SetText(g.displayTranscript)
//	if g.transcriptResetView {
//		g.transcriptList.Position = layout.Position{}
//		runes := len([]rune(g.displayTranscript))
//		g.transcriptView.SetCaret(runes, runes)
//		g.transcriptResetView = false
//	}
//	g.displayDirty = false
//}
//
//func (g *transcriptGUI) startWatching(parent context.Context, gameName string, w *app.Window) {
//	cfg, err := findGameConfig(gameName)
//	if err != nil {
//		g.setFailedGameState(gameName, err)
//		if w != nil {
//			w.Invalidate()
//		}
//		return
//	}
//
//	logPath, err := resolveRPGMakerTranscriptPath(firstNonEmpty(cfg.GamePath, cfg.Executable, cfg.WorkingDir))
//	if err != nil {
//		g.activeGameName = gameName
//		g.currentConfig = cfg
//		g.logPath = ""
//		g.statusText = err.Error()
//		g.lastGameRunningCheck = time.Time{}
//		g.refreshCurrentGameRunningState(true)
//		g.refreshCurrentGameHookStatus()
//		g.mu.Lock()
//		g.rawTranscript = ""
//		g.displayTranscript = ""
//		g.displayDirty = true
//		g.transcriptResetView = true
//		g.offset = 0
//		g.mu.Unlock()
//		if w != nil {
//			w.Invalidate()
//		}
//		return
//	}
//
//	if cancel := g.stopWatcher(); cancel != nil {
//		cancel()
//	}
//
//	g.activeGameName = gameName
//	g.currentConfig = cfg
//	g.logPath = logPath
//	g.statusText = "Watching transcript."
//	g.lastGameRunningCheck = time.Time{}
//	g.refreshCurrentGameRunningState(true)
//	g.reloadFlashcards()
//	g.refreshCurrentGameHookStatus()
//
//	g.mu.Lock()
//	raw, offset, status := initializeTranscriptState(logPath, g.printExisting)
//	g.rawTranscript = raw
//	g.updateDisplayTranscriptLocked()
//	g.transcriptResetView = true
//	g.offset = offset
//	if status != "" {
//		g.statusText = status
//	}
//	g.watcherGeneration++
//	generation := g.watcherGeneration
//	g.mu.Unlock()
//
//	watchCtx, cancel := context.WithCancel(parent)
//	g.mu.Lock()
//	g.watcherCancel = cancel
//	g.mu.Unlock()
//
//	go g.watchLoop(watchCtx, generation, logPath, w)
//	if w != nil {
//		w.Invalidate()
//	}
//}
//
//func (g *transcriptGUI) setFailedGameState(gameName string, err error) {
//	g.activeGameName = gameName
//	g.currentConfig = GameConfig{Name: gameName}
//	g.logPath = ""
//	g.statusText = err.Error()
//	g.gameRunning = false
//	g.gameRunningPID = 0
//	g.lastGameRunningCheck = time.Time{}
//	g.flashcards = nil
//	g.gameHookStatus = textHookStatus{Message: err.Error()}
//	g.mu.Lock()
//	g.rawTranscript = ""
//	g.displayTranscript = ""
//	g.displayDirty = true
//	g.transcriptResetView = true
//	g.offset = 0
//	g.mu.Unlock()
//}
//
//func (g *transcriptGUI) stopWatcher() context.CancelFunc {
//	g.mu.Lock()
//	defer g.mu.Unlock()
//	cancel := g.watcherCancel
//	g.watcherCancel = nil
//	return cancel
//}
//
//func initializeTranscriptState(logPath string, printExisting bool) (raw string, offset int64, status string) {
//	if printExisting {
//		delta, err := readTranscriptDelta(logPath, &offset)
//		if err == nil {
//			raw = delta
//			if strings.TrimSpace(raw) == "" {
//				status = "Watching transcript."
//			}
//			return raw, offset, status
//		}
//		if errors.Is(err, os.ErrNotExist) {
//			return "", 0, "Transcript log not found yet; start the game and advance dialogue."
//		}
//		return "", 0, err.Error()
//	}
//
//	info, err := os.Stat(logPath)
//	if err == nil {
//		offset = info.Size()
//		return "", offset, "Waiting for new dialogue."
//	}
//	if errors.Is(err, os.ErrNotExist) {
//		return "", 0, "Transcript log not found yet; start the game and advance dialogue."
//	}
//	return "", 0, err.Error()
//}
//
//func (g *transcriptGUI) watchLoop(ctx context.Context, generation int, logPath string, w *app.Window) {
//	interval := g.pollInterval
//	if interval <= 0 {
//		interval = 750 * time.Millisecond
//	}
//
//	ticker := time.NewTicker(interval)
//	defer ticker.Stop()
//
//	for {
//		select {
//		case <-ctx.Done():
//			return
//		case <-ticker.C:
//			g.pollTranscript(generation, logPath)
//			if w != nil {
//				w.Invalidate()
//			}
//		}
//	}
//}
//
//func (g *transcriptGUI) pollTranscript(generation int, logPath string) {
//	g.refreshCurrentGameRunningState(false)
//	g.mu.Lock()
//	if generation != g.watcherGeneration {
//		g.mu.Unlock()
//		return
//	}
//	offset := g.offset
//	g.mu.Unlock()
//
//	delta, err := readTranscriptDelta(logPath, &offset)
//	if err != nil {
//		g.mu.Lock()
//		defer g.mu.Unlock()
//		if generation != g.watcherGeneration {
//			return
//		}
//		if errors.Is(err, os.ErrNotExist) {
//			g.statusText = "Transcript log not found yet; start the game and advance dialogue."
//			return
//		}
//		g.statusText = err.Error()
//		return
//	}
//
//	g.mu.Lock()
//	defer g.mu.Unlock()
//	if generation != g.watcherGeneration {
//		return
//	}
//
//	g.offset = offset
//	if delta == "" {
//		if strings.TrimSpace(g.rawTranscript) == "" {
//			g.statusText = "Waiting for new dialogue."
//		} else {
//			g.statusText = "Watching transcript."
//		}
//		return
//	}
//
//	g.rawTranscript += delta
//	g.updateDisplayTranscriptLocked()
//	g.statusText = "Watching transcript."
//}
//
//func (g *transcriptGUI) applyTheme() {
//	g.theme = barethemes.New(g.themeMode, g.themePalette, g.systemDark)
//}
//
//func (g *transcriptGUI) applySavedSettings() {
//	settings, err := loadGUISettings()
//	if err != nil {
//		return
//	}
//
//	for _, opt := range g.modeOptions {
//		if opt.Label == settings.ThemeMode {
//			g.themeMode = opt.Mode
//			g.selectedModeName = opt.Label
//			break
//		}
//	}
//
//	for _, opt := range g.paletteOptions {
//		if opt.Label == settings.ThemePalette {
//			g.themePalette = opt.Palette
//			g.selectedPaletteName = opt.Label
//			break
//		}
//	}
//
//	for _, opt := range g.transcriptSizeOptions {
//		if opt.Label == settings.TranscriptTextSize {
//			g.transcriptTextSize = opt.TextSize
//			g.selectedTextSizeName = opt.Label
//			break
//		}
//	}
//
//	for _, opt := range g.recentLineOptions {
//		if opt.Label == settings.VisibleTranscript {
//			g.recentLineLimit = opt.RecentLineLimit
//			g.selectedRecentLinesName = opt.Label
//			break
//		}
//	}
//
//	g.autoPlayHighlightAudio.Value = settings.AutoPlayHighlightPopupAudio
//}
//
//func (g *transcriptGUI) persistSettings() {
//	_ = saveGUISettings(guiSettings{
//		ThemeMode:                   g.selectedModeName,
//		ThemePalette:                g.selectedPaletteName,
//		TranscriptTextSize:          g.selectedTextSizeName,
//		VisibleTranscript:           g.selectedRecentLinesName,
//		AutoPlayHighlightPopupAudio: g.autoPlayHighlightAudio.Value,
//	})
//}
//
//func (g *transcriptGUI) updateDisplayTranscriptLocked() {
//	sanitized := sanitizeTranscriptForDisplay(g.rawTranscript)
//	g.displayTranscript = limitTranscriptLines(sanitized, g.recentLineLimit)
//	g.displayDirty = true
//}
//
//func (g *transcriptGUI) paintTranscriptHighlights(gtx layout.Context) {
//	highlights := g.transcriptHighlights()
//	if len(highlights) == 0 || strings.TrimSpace(g.displayTranscript) == "" {
//		return
//	}
//
//	colorMacro := op.Record(gtx.Ops)
//	paint.ColorOp{Color: transcriptHighlightColor(g.theme.Color.Primary)}.Add(gtx.Ops)
//	fill := colorMacro.Stop()
//
//	regions := make([]widget.Region, 0, 8)
//	validClicks := make(map[string]struct{}, len(highlights))
//	for _, match := range highlights {
//		validClicks[match.Key] = struct{}{}
//		if g.transcriptHighlightClicks[match.Key] == nil {
//			g.transcriptHighlightClicks[match.Key] = new(widget.Clickable)
//		}
//		regions = g.transcriptView.Regions(match.StartRune, match.EndRune, regions[:0])
//		for _, region := range regions {
//			stack := clip.Rect(region.Bounds).Push(gtx.Ops)
//			fill.Add(gtx.Ops)
//			paint.PaintOp{}.Add(gtx.Ops)
//			stack.Pop()
//
//			offset := op.Offset(image.Pt(region.Bounds.Min.X, region.Bounds.Min.Y)).Push(gtx.Ops)
//			gtx.Constraints.Min = region.Bounds.Size()
//			gtx.Constraints.Max = region.Bounds.Size()
//			g.transcriptHighlightClicks[match.Key].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return layout.Dimensions{Size: region.Bounds.Size()}
//			})
//			offset.Pop()
//		}
//	}
//	for key := range g.transcriptHighlightClicks {
//		if _, ok := validClicks[key]; !ok {
//			delete(g.transcriptHighlightClicks, key)
//		}
//	}
//}
//
//func (g *transcriptGUI) transcriptHighlights() []transcriptMatch {
//	seen := make(map[string]Flashcard, len(g.flashcards))
//	words := make([]string, 0, len(g.flashcards))
//	for _, card := range g.flashcards {
//		word := strings.TrimSpace(card.Text)
//		if word == "" {
//			continue
//		}
//		if _, ok := seen[word]; ok {
//			continue
//		}
//		seen[word] = card
//		words = append(words, word)
//	}
//	sort.SliceStable(words, func(i, j int) bool {
//		return len([]rune(words[i])) > len([]rune(words[j]))
//	})
//
//	matches := findTranscriptMatches(g.displayTranscript, words)
//	for i := range matches {
//		matches[i].Card = seen[matches[i].Word]
//		matches[i].Key = fmt.Sprintf("%s-%d-%d", sanitizeName(matches[i].Card.ID), matches[i].StartRune, matches[i].EndRune)
//	}
//	return matches
//}
//
//
//func transcriptHighlightColor(base color.NRGBA) color.NRGBA {
//	return color.NRGBA{
//		R: base.R,
//		G: base.G,
//		B: base.B,
//		A: 72,
//	}
//}
//
//func (g *transcriptGUI) openTranscriptHighlightPopup(key string) {
//	for _, match := range g.transcriptHighlights() {
//		if match.Key == key {
//			g.showFlashcardPopup(match.Card)
//			if g.autoPlayHighlightAudio.Value {
//				g.playFlashcardAudio(match.Card, false)
//			}
//			return
//		}
//	}
//}
//
//func (g *transcriptGUI) statusColor() color.NRGBA {
//	status := strings.ToLower(g.statusText)
//	switch {
//	case strings.Contains(status, "failed"), strings.Contains(status, "error"):
//		return g.theme.Color.Error
//	case strings.Contains(status, "not found"):
//		return g.theme.Color.Warning
//	default:
//		return g.theme.Color.TextMuted
//	}
//}
//
//func (g *transcriptGUI) refreshCurrentGameRunningState(force bool) {
//	if !force && !g.lastGameRunningCheck.IsZero() && time.Since(g.lastGameRunningCheck) < guiGameRunningCheckInterval {
//		return
//	}
//	g.lastGameRunningCheck = time.Now()
//
//	if strings.TrimSpace(g.currentConfig.Name) == "" {
//		g.gameRunning = false
//		g.gameRunningPID = 0
//		return
//	}
//
//	processes, err := listProcesses()
//	if err != nil {
//		g.gameRunning = false
//		g.gameRunningPID = 0
//		return
//	}
//
//	matches := rankProcessMatches(g.currentConfig, processes)
//	if len(matches) == 0 {
//		g.gameRunning = false
//		g.gameRunningPID = 0
//		return
//	}
//
//	g.gameRunning = true
//	g.gameRunningPID = matches[0].PID
//}
//
//func (g *transcriptGUI) transcriptLaunchButtonLabel() string {
//	if g.gameRunning {
//		return "Game Running"
//	}
//	return "Launch Game"
//}
//
//func (g *transcriptGUI) transcriptLaunchButtonIcon() string {
//	if g.gameRunning {
//		return "mdi:check-circle-outline"
//	}
//	return "mdi:play-box-outline"
//}
//
//func (g *transcriptGUI) transcriptLaunchButtonVariant() bareui.ButtonVariant {
//	if g.gameRunning {
//		return bareui.ButtonSecondary
//	}
//	return bareui.ButtonPrimary
//}
//
//func (g *transcriptGUI) transcriptRunningStatusText() string {
//	if g.gameRunning {
//		if g.gameRunningPID > 0 {
//			return fmt.Sprintf("Detected running game process (pid %d).", g.gameRunningPID)
//		}
//		return "Detected running game process."
//	}
//	return "Game process not detected."
//}
