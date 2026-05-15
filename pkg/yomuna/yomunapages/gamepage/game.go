package gamepage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/gr/wine"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/input"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/modal"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/tabs"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/pages/fileexplorer"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
	"github.com/DarlingGoose/ymn/pkg/yomuna/backend"
)

type GameUI struct {
	th      *material.Theme
	theme   *theme.Client
	backend backend.Backend

	list                    layout.List
	panelList               widget.List
	tabList                 layout.List
	winetrickConfiguredList layout.List
	winetrickCommonList     layout.List
	winetrickInstalledList  layout.List
	configTab               *tabs.Layout

	gameDropdown   *dropdowns.Dropdown
	runnerDropdown *dropdowns.Dropdown

	nameInput       *input.TextInput
	iconInput       *input.TextInput
	imageInput      *input.TextInput
	runnerPathInput *input.TextInput
	prefixInput     *input.TextInput
	winetrickInput  *input.TextInput

	gamescopeFields []*runnerOptionField
	wineFields      []*runnerOptionField

	saveButton               *components.IconButton
	cancelButton             *components.IconButton
	resetButton              *components.IconButton
	deleteButton             *components.IconButton
	deleteRunnerConfigButton *components.IconButton
	refreshButton            *components.IconButton
	iconBrowse               *components.IconButton
	imageBrowse              *components.IconButton
	runnerBrowse             *components.IconButton
	prefixBrowse             *components.IconButton
	addWinetrickClick        widget.Clickable
	winetrickCommonClicks    map[string]*widget.Clickable
	winetrickRemoveClicks    map[string]*widget.Clickable

	filePickerModal  *modal.Modal
	filePicker       *fileexplorer.FileExplorer
	filePickerTarget *input.TextInput
	coverPreview     *media.View
	coverPreviewPath string
	activeLayer      *overlay.Overlay

	basicSection    collapsibleSection
	runnerSection   collapsibleSection
	wineSection     collapsibleSection
	displaySection  collapsibleSection
	windowSection   collapsibleSection
	scalingSection  collapsibleSection
	advancedSection collapsibleSection

	draft        *game.Game
	previousName string
	loadedGame   string
	gamesKey     string
	dirty        bool
	status       string
}

type collapsibleSection struct {
	title       string
	summary     string
	open        bool
	click       widget.Clickable
	initialized bool
}

func newGameSection(title, summary string, open bool) collapsibleSection {
	return collapsibleSection{
		title:       title,
		summary:     summary,
		open:        open,
		initialized: true,
	}
}

func NewGameUI(th *material.Theme, tc *theme.Client, b backend.Backend) *GameUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	saveIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:save")
	cancelIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:rotate-ccw")
	resetIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:undo-2")
	deleteIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:trash-2")
	deleteConfigIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:file-x-2")
	refreshIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:refresh-cw")
	folderIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:folder-open")

	ui := &GameUI{
		th:      th,
		theme:   tc,
		backend: b,
		list:    layout.List{Axis: layout.Vertical},
		panelList: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		tabList:                 layout.List{Axis: layout.Horizontal},
		winetrickConfiguredList: layout.List{Axis: layout.Horizontal},
		winetrickCommonList:     layout.List{Axis: layout.Horizontal},
		winetrickInstalledList:  layout.List{Axis: layout.Horizontal},

		gameDropdown: dropdowns.NewDropdown(nil).
			WithThemeClient(tc).
			WithRole(theme.TextRoleLabel),
		runnerDropdown: dropdowns.NewDropdown([]dropdowns.DropdownItem{
			{Label: "Gamescope", Value: string(game.RunnerGamescope)},
			{Label: "Wine", Value: string(game.RunnerWine)},
			{Label: "Proton", Value: string(game.RunnerProton)},
			{Label: "Steam", Value: string(game.RunnerSteam)},
		}).WithThemeClient(tc).WithRole(theme.TextRoleLabel),

		nameInput:       input.NewTextInput("Name", "Game name").WithMaterialTheme(th).WithThemeClient(tc),
		iconInput:       input.NewPathInput("Icon", "/path/to/icon.png").WithMaterialTheme(th).WithThemeClient(tc),
		imageInput:      input.NewPathInput("Picture", "/path/to/image.png").WithMaterialTheme(th).WithThemeClient(tc),
		runnerPathInput: input.NewPathInput("Runner path", "Optional runner executable").WithMaterialTheme(th).WithThemeClient(tc),
		prefixInput:     input.NewPathInput("Wine prefix", "Optional Wine prefix").WithMaterialTheme(th).WithThemeClient(tc),
		winetrickInput:  input.NewTextInput("Winetricks verb", "fakejapanese").WithMaterialTheme(th).WithThemeClient(tc),

		saveButton:               components.NewIconButton("Save Game", nil, saveIcon).WithThemeClient(tc),
		cancelButton:             components.NewIconButton("Cancel", nil, cancelIcon).WithThemeClient(tc),
		resetButton:              components.NewIconButton("Reset Defaults", nil, resetIcon).WithThemeClient(tc),
		deleteButton:             components.NewIconButton("Delete", nil, deleteIcon).WithThemeClient(tc),
		deleteRunnerConfigButton: components.NewIconButton("Delete Runner Config", nil, deleteConfigIcon).WithThemeClient(tc),
		refreshButton:            components.NewIconButton("Refresh", nil, refreshIcon).WithThemeClient(tc),
		iconBrowse:               components.NewIconButton("Browse", nil, folderIcon).WithThemeClient(tc),
		imageBrowse:              components.NewIconButton("Browse", nil, folderIcon).WithThemeClient(tc),
		runnerBrowse:             components.NewIconButton("Browse", nil, folderIcon).WithThemeClient(tc),
		prefixBrowse:             components.NewIconButton("Browse", nil, folderIcon).WithThemeClient(tc),
		coverPreview:             media.NewView(media.DefaultRegistry).WithImageFit(widget.Contain),
		winetrickCommonClicks:    map[string]*widget.Clickable{},
		winetrickRemoveClicks:    map[string]*widget.Clickable{},

		basicSection:    newGameSection("General", "Name and artwork.", true),
		runnerSection:   newGameSection("Runner", "Runner type and executable paths.", true),
		wineSection:     newGameSection("Wine", "Wine prefix and process behavior.", true),
		displaySection:  newGameSection("Display", "Resolution and refresh rate.", true),
		windowSection:   newGameSection("Window Behavior", "Fullscreen, capture, and platform behavior.", true),
		scalingSection:  newGameSection("Scaling", "Upscaling and filtering choices.", true),
		advancedSection: newGameSection("Advanced", "Binary overrides and custom runner arguments.", false),
	}
	ui.configTab = tabs.New(
		tabs.NewTabFunc(configPanelGeneral, "General", "", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutScrollableConfigPanel(gtx, ui.layoutGameFields)
		}),
		tabs.NewTabFunc(configPanelRunner, "Runner", "", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutScrollableConfigPanel(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutRunnerFields(gtx, ui.activeLayer)
			})
		}),
		tabs.NewTabFunc(configPanelWine, "Wine", "", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutScrollableConfigPanel(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutWineOptions(gtx, ui.activeLayer)
			})
		}),
		tabs.NewTabFunc(configPanelDisplay, "Display", "", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutScrollableConfigPanel(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutDisplayOptions(gtx, ui.activeLayer)
			})
		}),
		tabs.NewTabFunc(configPanelWindow, "Window", "", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutScrollableConfigPanel(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutWindowBehaviorOptions(gtx, ui.activeLayer)
			})
		}),
		tabs.NewTabFunc(configPanelScaling, "Scaling", "", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutScrollableConfigPanel(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutScalingOptions(gtx, ui.activeLayer)
			})
		}),
		tabs.NewTabFunc(configPanelAdvanced, "Advanced", "", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutScrollableConfigPanel(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutAdvancedOptions(gtx, ui.activeLayer)
			})
		}),
	)
	ui.saveButton.CollapseTextBelow = unit.Dp(140)
	ui.cancelButton.CollapseTextBelow = unit.Dp(120)
	ui.cancelButton.FillWidth = false
	ui.resetButton.CollapseTextBelow = unit.Dp(170)
	ui.resetButton.FillWidth = false
	ui.deleteButton.CollapseTextBelow = unit.Dp(120)
	ui.deleteButton.FillWidth = false
	ui.deleteRunnerConfigButton.CollapseTextBelow = unit.Dp(220)
	ui.deleteRunnerConfigButton.FillWidth = false
	ui.refreshButton.CollapseTextBelow = unit.Dp(120)
	ui.refreshButton.FillWidth = false
	ui.configureBrowseButton(ui.iconBrowse)
	ui.configureBrowseButton(ui.imageBrowse)
	ui.configureBrowseButton(ui.runnerBrowse)
	ui.configureBrowseButton(ui.prefixBrowse)

	ui.configureMainInputs()

	ui.gameDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
		if !valid || ui.backend == nil {
			return
		}
		ui.backend.SelectGameByName(item.Value)
		ui.loadGame(ui.backend.CurrentGame())
	})
	ui.runnerDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
		if !valid || ui.draft == nil {
			return
		}
		nextRunner := game.RunnerType(item.Value)
		if ui.draft.Runner != nextRunner {
			ui.draft.RunnerPath = ""
			ui.runnerPathInput.SetText("")
		}
		ui.draft.Runner = nextRunner
		ui.ensureRunnerDefaults(ui.draft)
		ui.loadRunnerFields()
		ui.markDirty()
	})

	ui.loadGame(nil)
	return ui
}

func (ui *GameUI) configureMainInputs() {
	if ui == nil {
		return
	}
	configure := func(in *input.TextInput, hint string) {
		if in == nil {
			return
		}
		in.Hint = hint
		in.OnChange = func(string) {
			ui.markDirty()
		}
	}
	configure(ui.nameInput, "Required")
	configure(ui.iconInput, "Optional file path")
	configure(ui.imageInput, "Optional file path")
	configure(ui.runnerPathInput, "Leave empty to use the selected runner default")
	configure(ui.prefixInput, "Leave empty to use the game or runner default")
	configure(ui.winetrickInput, "Winetricks verb, for example fakejapanese")
	if ui.winetrickInput != nil {
		ui.winetrickInput.OnChange = nil
	}
	if ui.imageInput != nil {
		ui.imageInput.OnChange = func(path string) {
			ui.loadCoverPreview(path)
			ui.markDirty()
		}
	}
}

func (ui *GameUI) configureBrowseButton(btn *components.IconButton) {
	if btn == nil {
		return
	}
	btn.FillWidth = false
	btn.TextCollapseMode = components.TextCollapseNever
	btn.CollapseTextBelow = unit.Dp(160)
	btn.MinWidth = unit.Dp(96)
	btn.Height = unit.Dp(44)
	btn.Radius = unit.Dp(10)
}

func (ui *GameUI) markDirty() {
	if ui == nil || ui.draft == nil {
		return
	}
	ui.dirty = true
	ui.status = "Unsaved changes"
}

func (ui *GameUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}
	ui.activeLayer = layer
	ui.syncGames()

	dims := layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.layoutContentContainer(gtx, func(gtx layout.Context) layout.Dimensions {
			if gtx.Constraints.Max.X >= gtx.Dp(unit.Dp(1060)) {
				return ui.layoutWide(gtx, layer)
			}
			return ui.layoutNarrow(gtx, layer)
		})
	})
	ui.layoutFilePickerOverlay(gtx, layer)
	return dims
}

func (ui *GameUI) layoutContentContainer(gtx layout.Context, content layout.Widget) layout.Dimensions {
	if content == nil {
		return layout.Dimensions{}
	}
	maxWidth := gtx.Dp(unit.Dp(1180))
	available := gtx.Constraints.Max.X
	if available <= 0 || available <= maxWidth {
		return content(gtx)
	}
	side := (available - maxWidth) / 2
	return layout.Inset{
		Left:  unit.Dp(float32(side) / gtx.Metric.PxPerDp),
		Right: unit.Dp(float32(side) / gtx.Metric.PxPerDp),
	}.Layout(gtx, content)
}

func (ui *GameUI) layoutNarrow(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	return ui.list.Layout(gtx, 5, func(gtx layout.Context, index int) layout.Dimensions {
		switch index {
		case 0:
			return ui.layoutHeader(gtx, layer)
		case 1:
			return layout.Spacer{Height: unit.Dp(14)}.Layout(gtx)
		case 2:
			return ui.layoutConfigTabs(gtx)
		case 3:
			return layout.Spacer{Height: unit.Dp(24)}.Layout(gtx)
		case 4:
			return layout.Dimensions{}
		default:
			return layout.Dimensions{}
		}
	})
}

func (ui *GameUI) layoutWide(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	return ui.list.Layout(gtx, 5, func(gtx layout.Context, index int) layout.Dimensions {
		switch index {
		case 0:
			return ui.layoutHeader(gtx, layer)
		case 1:
			return layout.Spacer{Height: unit.Dp(14)}.Layout(gtx)
		case 2:
			return ui.layoutConfigTabs(gtx)
		case 3:
			return layout.Spacer{Height: unit.Dp(24)}.Layout(gtx)
		default:
			return layout.Dimensions{}
		}
	})
}

func (ui *GameUI) syncGames() {
	if ui.backend == nil || ui.gameDropdown == nil {
		return
	}
	current := ui.backend.CurrentGame()
	currentName := ""
	if current != nil {
		currentName = strings.TrimSpace(current.Name)
	}

	if ui.gamesKey == "" || (currentName != "" && !ui.dropdownHasValue(currentName)) {
		current = ui.refreshGameItems(currentName)
		if current != nil {
			currentName = strings.TrimSpace(current.Name)
		}
	}

	if current == nil {
		return
	}
	if currentName != "" && currentName != ui.selectedGameValue() {
		ui.gameDropdown.SelectItem(current.Name)
	}
	if ui.draft == nil || currentName != ui.loadedGame {
		ui.loadGame(current)
	}
}

func (ui *GameUI) refreshGameItems(preferredName string) *game.Game {
	if ui.backend == nil || ui.gameDropdown == nil {
		return nil
	}
	games := ui.backend.GetGames()
	items := make([]dropdowns.DropdownItem, 0, len(games))
	itemKeys := make([]string, 0, len(games))
	var first *game.Game
	var preferred *game.Game
	for _, g := range games {
		if g == nil || strings.TrimSpace(g.Name) == "" {
			continue
		}
		name := strings.TrimSpace(g.Name)
		if first == nil {
			first = g
		}
		if preferredName != "" && name == preferredName {
			preferred = g
		}
		items = append(items, dropdowns.DropdownItem{Label: name, Value: name})
		itemKeys = append(itemKeys, name)
	}
	gamesKey := strings.Join(itemKeys, "\x00")
	if gamesKey != ui.gamesKey {
		ui.gameDropdown.SetItems(items)
		ui.gamesKey = gamesKey
	}
	if preferred != nil {
		return preferred
	}
	if current := ui.backend.CurrentGame(); current != nil {
		return current
	}
	return first
}

func (ui *GameUI) loadGame(g *game.Game) {
	if g == nil && ui.backend != nil {
		g = ui.backend.CurrentGame()
	}
	if g == nil {
		ui.draft = nil
		ui.previousName = ""
		ui.loadedGame = ""
		ui.status = "Select a game to edit its config."
		return
	}

	ui.draft = cloneGame(g)
	ui.previousName = g.Name
	ui.loadedGame = strings.TrimSpace(g.Name)
	ui.ensureRunnerDefaults(ui.draft)

	ui.nameInput.SetText(ui.draft.Name)
	ui.iconInput.SetText(ui.draft.IconPath)
	ui.imageInput.SetText(ui.draft.ImagePath)
	ui.runnerPathInput.SetText(ui.draft.RunnerPath)
	ui.prefixInput.SetText(ui.draft.PrefixPath)
	ui.runnerDropdown.SelectItem(string(ui.draft.Runner))
	ui.loadRunnerFields()
	ui.loadCoverPreview(ui.draft.ImagePath)
	ui.dirty = false
	ui.status = "Saved"
}

func (ui *GameUI) loadCoverPreview(path string) {
	if ui == nil || ui.coverPreview == nil {
		return
	}
	path = strings.TrimSpace(path)
	if path == ui.coverPreviewPath {
		return
	}
	ui.coverPreviewPath = path
	if path == "" {
		_ = ui.coverPreview.Close()
		return
	}
	_ = ui.coverPreview.LoadPath(context.Background(), path)
}

func (ui *GameUI) ensureRunnerDefaults(g *game.Game) {
	if g == nil {
		return
	}
	if g.Runner == "" {
		g.Runner = game.RunnerGamescope
	}
	if !hasWineOptions(g.WineConfig) {
		cfg := defaultWineOptions(g.PrefixPath)
		g.WineConfig = &cfg
	}
	if !hasGamescopeOptions(g.GamescopeConfig) {
		cfg := defaultGamescopeOptions(g.PrefixPath)
		g.GamescopeConfig = &cfg
	}
	if !hasGRConfig(g.RunnerConfig) {
		g.RunnerConfig = defaultGRConfigForGame(g)
	}
}

func (ui *GameUI) loadRunnerFields() {
	if ui.draft == nil {
		return
	}
	ui.gamescopeFields = buildRunnerOptionFields(gamescope.Options{}, ui.th, ui.theme, ui.markDirty)
	ui.wineFields = buildRunnerOptionFields(wine.Options{}, ui.th, ui.theme, ui.markDirty)
	if ui.draft.GamescopeConfig != nil {
		setRunnerOptionFields(ui.gamescopeFields, reflect.ValueOf(*ui.draft.GamescopeConfig))
	}
	if ui.draft.WineConfig != nil {
		setRunnerOptionFields(ui.wineFields, reflect.ValueOf(*ui.draft.WineConfig))
	}
}

func (ui *GameUI) layoutHeader(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	return utils.SurfaceOutlined(gtx, ct.SurfaceAltNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Details")
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ui.layoutStatusPill(gtx)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutDivider(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(760)) {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return ui.layoutHeaderPreview(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return ui.layoutHeaderDetails(gtx, layer)
							}),
						)
					}
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ui.layoutHeaderPreview(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(26)}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return ui.layoutHeaderDetails(gtx, layer)
						}),
					)
				}),
			)
		})
	})
}

func (ui *GameUI) layoutDivider(gtx layout.Context) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	return utils.Surface(gtx, ct.BorderNRGBA(), 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Spacer{Height: unit.Dp(1)}.Layout(gtx)
	})
}

func (ui *GameUI) layoutHeaderPreview(gtx layout.Context) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	hasCover := ui.coverPreview != nil && strings.TrimSpace(ui.coverPreviewPath) != ""
	width := gtx.Dp(unit.Dp(184))
	height := gtx.Dp(unit.Dp(132))
	if hasCover {
		width = gtx.Dp(unit.Dp(260))
		height = gtx.Dp(unit.Dp(178))
	}
	if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(520)) {
		width = gtx.Constraints.Max.X
		if hasCover {
			height = gtx.Dp(unit.Dp(210))
		}
	}
	gtx.Constraints.Min.X = width
	gtx.Constraints.Max.X = width
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	return utils.SurfaceOutlined(gtx, ct.SurfaceNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		if hasCover {
			return layout.Center.Layout(gtx, ui.coverPreview.Layout)
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, "No cover image")
		})
	})
}

func (ui *GameUI) layoutHeaderDetails(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH2, theme.ThemeColorTextPrimary, ui.selectedGameTitle())
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, ui.headerSummary())
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutField(gtx, "Editing", "Choose which installed game config to edit.", func(gtx layout.Context) layout.Dimensions {
				if ui.gameDropdown == nil {
					return layout.Dimensions{}
				}
				return ui.gameDropdown.Layout(gtx, layer)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = 0
			return ui.layoutActions(gtx)
		}),
	)
}

func (ui *GameUI) selectedGameTitle() string {
	if ui != nil && ui.draft != nil {
		if name := strings.TrimSpace(ui.draft.Name); name != "" {
			return name
		}
	}
	return "Game Configuration"
}

func (ui *GameUI) headerSummary() string {
	runner := "Runner not selected"
	if ui.draft != nil {
		runner = runnerLabel(ui.draft.Runner)
		if ui.draft.Runner == game.RunnerGamescope {
			runner += " + Wine"
		}
	}
	return "Runner: " + runner
}

func (ui *GameUI) layoutStatusPill(gtx layout.Context) layout.Dimensions {
	text := "Saved"
	if ui.dirty {
		text = "Unsaved changes"
	} else if strings.TrimSpace(ui.status) != "" {
		text = ui.status
	}
	ct := ui.theme.GetCurrentColorToken()
	bg := ct.SurfaceAltNRGBA()
	if ui.dirty {
		bg = ct.PrimaryNRGBA()
	}
	fg := theme.ThemeColorTextSecondary
	if ui.dirty {
		fg = theme.ThemeColorOnPrimary
	}
	return utils.Surface(gtx, bg, unit.Dp(7), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(5), Bottom: unit.Dp(5), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, fg, text)
		})
	})
}

func (ui *GameUI) layoutRefreshButton(gtx layout.Context) layout.Dimensions {
	if ui.refreshButton.Clicked(gtx) {
		if ui.backend != nil {
			if err := ui.backend.ReloadGames(); err != nil {
				ui.status = "Refresh failed: " + err.Error()
			} else {
				ui.draft = nil
				ui.gamesKey = ""
				ui.syncGames()
				ui.status = "Game list refreshed"
			}
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	return ui.refreshButton.Layout(gtx)
}

func (ui *GameUI) layoutGameFields(gtx layout.Context) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutInputField(gtx, ui.nameInput, "Display name", "Used in game pickers and saved config filenames.")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutPathInputField(gtx, ui.iconInput, ui.iconBrowse, "Icon", "Small icon file for lists or compact game display.")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutPathInputField(gtx, ui.imageInput, ui.imageBrowse, "Cover image", "Larger picture used when the game needs artwork.")
			}),
		)
	}
	if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(720)) {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(fields),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(ui.layoutCoverPreview),
		)
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
		layout.Flexed(1, fields),
		layout.Rigid(layout.Spacer{Width: unit.Dp(14)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(300))
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(300))
			return ui.layoutCoverPreview(gtx)
		}),
	)
}

func (ui *GameUI) layoutCoverPreview(gtx layout.Context) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	return ui.layoutField(gtx, "Cover preview", "Rendered from the selected cover image path.", func(gtx layout.Context) layout.Dimensions {
		return utils.SurfaceOutlined(gtx, ct.SurfaceAltNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
			height := gtx.Dp(unit.Dp(240))
			if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(720)) {
				height = gtx.Dp(unit.Dp(220))
			}
			gtx.Constraints.Min.Y = height
			gtx.Constraints.Max.Y = height
			if ui.coverPreview == nil || strings.TrimSpace(ui.coverPreviewPath) == "" {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, "No cover image")
				})
			}
			return layout.Center.Layout(gtx, ui.coverPreview.Layout)
		})
	})
}

func (ui *GameUI) layoutRunnerFields(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutField(gtx, "Runner type", "Options below change with the selected runner.", func(gtx layout.Context) layout.Dimensions {
				return ui.runnerDropdown.Layout(gtx, layer)
			})
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutPathInputField(gtx, ui.runnerPathInput, ui.runnerBrowse, "Runner executable", "Optional executable path. Leave empty to use the selected runner's default.")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutPathInputField(gtx, ui.prefixInput, ui.prefixBrowse, "Wine prefix", "Optional Wine prefix. Leave empty to use the game or runner default.")
		}),
	)
}

func (ui *GameUI) layoutRunnerOptions(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.activeRunnerFields()
	if len(fields) == 0 {
		return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, "This runner does not have editable options yet.")
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, runnerFieldChildren(ui.th, ui.theme, fields, layer)...)
}

func (ui *GameUI) layoutWineOptions(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields("DefaultWinePrefix", "DefaultPrefix", "UseWine", "WineStartWait", "KillWineOnExit")
	children := []layout.FlexChild{}
	if len(fields) > 0 {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutFieldGrid(gtx, fields, layer)
		}))
	} else {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutMutedText(gtx, "No Wine options are available for this runner.")
		}))
	}
	children = append(children,
		layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutWinetricksOptions(gtx)
		}),
	)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (ui *GameUI) layoutWinetricksOptions(gtx layout.Context) layout.Dimensions {
	if ui == nil || ui.draft == nil {
		return layout.Dimensions{}
	}
	ct := ui.theme.GetCurrentColorToken()
	return utils.SurfaceOutlined(gtx, ct.SurfaceNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, "Winetricks dependencies")
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutMutedText(gtx, "Configured verbs are saved with this game. Installed verbs are read from the Wine prefix winetricks.log.")
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutWinetricksConfigured(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutWinetricksAdder(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutWinetricksCommon(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutInstalledWinetricks(gtx)
				}),
			)
		})
	})
}

func (ui *GameUI) layoutWinetricksConfigured(gtx layout.Context) layout.Dimensions {
	deps := ui.configuredWinetricks()
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, "Configured")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
	}
	if len(deps) == 0 {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutMutedText(gtx, "No winetricks dependencies are configured.")
		}))
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return ui.winetrickConfiguredList.Layout(gtx, len(deps), func(gtx layout.Context, index int) layout.Dimensions {
			dep := deps[index]
			return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				click := ui.winetrickRemoveClicks[dep]
				if click == nil {
					click = &widget.Clickable{}
					ui.winetrickRemoveClicks[dep] = click
				}
				for click.Clicked(gtx) {
					if ui.removeConfiguredWinetrick(dep) {
						gtx.Execute(op.InvalidateCmd{})
					}
				}
				return ui.layoutWinetrickChip(gtx, click, dep+" x", true)
			})
		})
	}))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (ui *GameUI) layoutWinetricksAdder(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.End}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if ui.winetrickInput == nil {
				return layout.Dimensions{}
			}
			oldLabel := ui.winetrickInput.Label
			ui.winetrickInput.Label = ""
			dims := ui.winetrickInput.Layout(gtx)
			ui.winetrickInput.Label = oldLabel
			return dims
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			for ui.addWinetrickClick.Clicked(gtx) {
				if ui.addConfiguredWinetrick(ui.winetrickInput.Text()) {
					ui.winetrickInput.SetText("")
					gtx.Execute(op.InvalidateCmd{})
				}
			}
			return ui.layoutTextButton(gtx, &ui.addWinetrickClick, "Add")
		}),
	)
}

func (ui *GameUI) layoutWinetricksCommon(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, "Common Japanese VN helpers")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			common := commonWinetrickVerbs()
			return ui.winetrickCommonList.Layout(gtx, len(common), func(gtx layout.Context, index int) layout.Dimensions {
				dep := common[index]
				return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					click := ui.winetrickCommonClicks[dep]
					if click == nil {
						click = &widget.Clickable{}
						ui.winetrickCommonClicks[dep] = click
					}
					for click.Clicked(gtx) {
						if ui.addConfiguredWinetrick(dep) {
							gtx.Execute(op.InvalidateCmd{})
						}
					}
					return ui.layoutWinetrickChip(gtx, click, dep, false)
				})
			})
		}),
	)
}

func (ui *GameUI) layoutInstalledWinetricks(gtx layout.Context) layout.Dimensions {
	installed, source := ui.installedWinetricks()
	label := "Installed"
	if source != "" {
		label += " from " + source
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, label)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
	}
	if len(installed) == 0 {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutMutedText(gtx, "No winetricks.log entries found for this prefix.")
		}))
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return ui.winetrickInstalledList.Layout(gtx, len(installed), func(gtx layout.Context, index int) layout.Dimensions {
			return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutWinetrickChip(gtx, nil, installed[index], false)
			})
		})
	}))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func commonWinetrickVerbs() []string {
	return []string{
		"fakejapanese", "cjkfonts", "corefonts", "riched20",
		"vcrun6", "vcrun2010", "vcrun2019", "vcrun2022",
		"d3dx9", "quartz", "gdiplus", "dotnet40",
	}
}

func (ui *GameUI) layoutWinetrickChip(gtx layout.Context, click *widget.Clickable, label string, active bool) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	bg := ct.SurfaceAltNRGBA()
	border := ct.BorderNRGBA()
	textRole := theme.ThemeColorTextSecondary
	if active {
		bg = ct.PrimaryNRGBA()
		border = ct.PrimaryNRGBA()
		textRole = theme.ThemeColorOnPrimary
	} else if click != nil && click.Hovered() {
		border = ct.PrimaryNRGBA()
		textRole = theme.ThemeColorTextPrimary
	}
	content := func(gtx layout.Context) layout.Dimensions {
		return utils.SurfaceOutlined(gtx, bg, unit.Dp(8), utils.SurfaceBorder{Color: border, Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(6), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, textRole, label)
			})
		})
	}
	if click == nil {
		return content(gtx)
	}
	return click.Layout(gtx, content)
}

func (ui *GameUI) layoutTextButton(gtx layout.Context, click *widget.Clickable, label string) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	bg := ct.PrimaryNRGBA()
	if click != nil && click.Hovered() {
		bg = ct.PrimaryHoverNRGBA()
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return utils.Surface(gtx, bg, unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(14), Right: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorOnPrimary, label)
			})
		})
	})
}

func (ui *GameUI) configuredWinetricks() []string {
	if ui == nil || ui.draft == nil {
		return nil
	}
	return normalizeWinetrickList(ui.draft.RunnerConfig.Dependencies)
}

func (ui *GameUI) addConfiguredWinetrick(dep string) bool {
	if ui == nil || ui.draft == nil {
		return false
	}
	dep = normalizeWinetrickVerb(dep)
	if dep == "" {
		return false
	}
	next := append(ui.configuredWinetricks(), dep)
	normalized := normalizeWinetrickList(next)
	if stringListEqual(normalized, ui.draft.RunnerConfig.Dependencies) {
		return false
	}
	ui.draft.RunnerConfig.Dependencies = normalized
	ui.markDirty()
	return true
}

func (ui *GameUI) removeConfiguredWinetrick(dep string) bool {
	if ui == nil || ui.draft == nil {
		return false
	}
	dep = normalizeWinetrickVerb(dep)
	if dep == "" {
		return false
	}
	current := ui.configuredWinetricks()
	next := make([]string, 0, len(current))
	removed := false
	for _, item := range current {
		if item == dep {
			removed = true
			continue
		}
		next = append(next, item)
	}
	if !removed {
		return false
	}
	ui.draft.RunnerConfig.Dependencies = next
	ui.markDirty()
	return true
}

func (ui *GameUI) installedWinetricks() ([]string, string) {
	prefix := ui.effectiveWinePrefix()
	if prefix == "" {
		return nil, ""
	}
	logPath := filepath.Join(prefix, "winetricks.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, ""
	}
	return parseWinetricksLog(string(data)), "winetricks.log"
}

func (ui *GameUI) effectiveWinePrefix() string {
	if ui == nil || ui.draft == nil {
		return ""
	}
	candidates := []string{
		ui.prefixInput.Text(),
		ui.draft.PrefixPath,
		ui.draft.RunnerConfig.WinePrefix,
	}
	if ui.draft.WineConfig != nil {
		candidates = append(candidates, ui.draft.WineConfig.DefaultPrefix)
	}
	for _, candidate := range candidates {
		if path := strings.TrimSpace(candidate); path != "" {
			return path
		}
	}
	return ""
}

func parseWinetricksLog(text string) []string {
	lines := strings.FieldsFunc(text, func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	return normalizeWinetrickList(lines)
}

func normalizeWinetrickList(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		verb := normalizeWinetrickVerb(item)
		if verb == "" {
			continue
		}
		if _, ok := seen[verb]; ok {
			continue
		}
		seen[verb] = struct{}{}
		out = append(out, verb)
	}
	sort.Strings(out)
	return out
}

func normalizeWinetrickVerb(verb string) string {
	verb = strings.TrimSpace(strings.ToLower(verb))
	verb = strings.Trim(verb, "\"'")
	if verb == "" || strings.HasPrefix(verb, "#") {
		return ""
	}
	return verb
}

func stringListEqual(a, b []string) bool {
	a = normalizeWinetrickList(a)
	b = normalizeWinetrickList(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (ui *GameUI) layoutDisplayOptions(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields(
		"Width", "Height", "RefreshRate", "OutputWidth", "OutputHeight",
	)
	if len(fields) == 0 {
		return ui.layoutMutedText(gtx, "No display options are available for this runner.")
	}
	return ui.layoutFieldGrid(gtx, fields, layer)
}

func (ui *GameUI) layoutWindowBehaviorOptions(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields("Fullscreen", "Borderless", "ForceGrab", "SteamDeckMode", "ExposeWayland")
	if len(fields) == 0 {
		return ui.layoutMutedText(gtx, "No window behavior options are available for this runner.")
	}
	return ui.layoutFieldGrid(gtx, fields, layer)
}

func (ui *GameUI) layoutScalingOptions(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields("Scaler", "Filter")
	if len(fields) == 0 {
		return ui.layoutMutedText(gtx, "No scaling options are available for this runner.")
	}
	return ui.layoutFieldGrid(gtx, fields, layer)
}

func (ui *GameUI) layoutAdvancedOptions(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields(
		"Name", "GamescopeBin", "WineBin", "WineServerBin", "WineTricksBin",
		"DefaultWinePrefix", "DefaultPrefix", "ExtraArgs",
	)
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutDeleteRunnerConfig(gtx)
		}),
	}
	if len(fields) > 0 {
		children = append(children,
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, runnerFieldChildren(ui.th, ui.theme, fields, layer)...)
			}),
		)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (ui *GameUI) filterRunnerFields(names ...string) []*runnerOptionField {
	active := ui.activeRunnerFields()
	if len(active) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(names))
	for _, name := range names {
		allowed[name] = struct{}{}
	}
	fields := make([]*runnerOptionField, 0, len(names))
	for _, field := range active {
		if field == nil {
			continue
		}
		if _, ok := allowed[field.name]; ok {
			fields = append(fields, field)
		}
	}
	return fields
}

func runnerFieldChildren(th *material.Theme, tc *theme.Client, fields []*runnerOptionField, layer *overlay.Overlay) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(fields)*2)
	for i, field := range fields {
		f := field
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.Layout(gtx, th, tc, layer)
		}))
		if i != len(fields)-1 {
			children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout))
		}
	}
	return children
}

func (ui *GameUI) layoutFieldGrid(gtx layout.Context, fields []*runnerOptionField, layer *overlay.Overlay) layout.Dimensions {
	if len(fields) == 0 {
		return layout.Dimensions{}
	}
	if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(620)) || len(fields) == 1 {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, runnerFieldChildren(ui.th, ui.theme, fields, layer)...)
	}
	rows := make([]layout.FlexChild, 0, (len(fields)+1)/2)
	for i := 0; i < len(fields); i += 2 {
		left := fields[i]
		var right *runnerOptionField
		if i+1 < len(fields) {
			right = fields[i+1]
		}
		leftField := left
		rightField := right
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return leftField.Layout(gtx, ui.th, ui.theme, layer)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if rightField == nil {
						return layout.Dimensions{}
					}
					return rightField.Layout(gtx, ui.th, ui.theme, layer)
				}),
			)
		}))
		if i+2 < len(fields) {
			rows = append(rows, layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout))
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func (ui *GameUI) layoutSection(gtx layout.Context, section *collapsibleSection, body layout.Widget) layout.Dimensions {
	if section == nil {
		return layout.Dimensions{}
	}
	canCollapse := section.title == "Advanced"
	if canCollapse {
		for section.click.Clicked(gtx) {
			section.open = !section.open
			gtx.Execute(op.InvalidateCmd{})
		}
	} else {
		section.open = true
	}
	ct := ui.theme.GetCurrentColorToken()
	return utils.SurfaceOutlined(gtx, ct.SurfaceNRGBA(), unit.Dp(10), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					header := func(gtx layout.Context) layout.Dimensions {
						icon := "+"
						if section.open {
							icon = "-"
						}
						return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextPrimary, section.title)
									}),
									layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
									layout.Rigid(func(gtx layout.Context) layout.Dimensions {
										return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, section.summary)
									}),
								)
							}),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								if !canCollapse {
									return layout.Dimensions{}
								}
								return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextSecondary, icon)
							}),
						)
					}
					if canCollapse {
						return section.click.Layout(gtx, header)
					}
					return header(gtx)
				}),
			}
			if section.open && body != nil {
				children = append(children,
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					layout.Rigid(body),
				)
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		})
	})
}

const (
	configPanelGeneral  = "general"
	configPanelRunner   = "runner"
	configPanelWine     = "wine"
	configPanelDisplay  = "display"
	configPanelWindow   = "window"
	configPanelScaling  = "scaling"
	configPanelAdvanced = "advanced"
)

func (ui *GameUI) layoutConfigTabs(gtx layout.Context) layout.Dimensions {
	if ui == nil || ui.configTab == nil {
		return layout.Dimensions{}
	}
	ui.configTab.Update(gtx)

	ct := ui.theme.GetCurrentColorToken()
	return utils.SurfaceOutlined(gtx, ct.SurfaceAltNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutConfigTabButtons(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.configTab.Layout(gtx)
				}),
			)
		})
	})
}

func (ui *GameUI) layoutConfigTabButtons(gtx layout.Context) layout.Dimensions {
	buttons := ui.configTab.Buttons()
	if len(buttons) == 0 {
		return layout.Dimensions{}
	}
	return ui.tabList.Layout(gtx, len(buttons), func(gtx layout.Context, index int) layout.Dimensions {
		btn := buttons[index]
		return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return ui.layoutConfigTabButton(gtx, btn)
		})
	})
}

func (ui *GameUI) layoutConfigTabButton(gtx layout.Context, btn tabs.Button) layout.Dimensions {
	if btn.Clickable == nil {
		return layout.Dimensions{}
	}
	for btn.Clickable.Clicked(gtx) {
		if ui.configTab != nil && ui.configTab.SwitchToID(btn.ID) {
			gtx.Execute(op.InvalidateCmd{})
		}
	}

	ct := ui.theme.GetCurrentColorToken()
	bg := ct.SurfaceNRGBA()
	border := ct.BorderNRGBA()
	textColor := theme.ThemeColorTextMuted
	if btn.Active {
		bg = ct.PrimaryNRGBA()
		border = ct.PrimaryNRGBA()
		textColor = theme.ThemeColorOnPrimary
	} else if btn.Clickable.Hovered() {
		bg = ct.SurfaceAltNRGBA()
		border = ct.PrimaryNRGBA()
		textColor = theme.ThemeColorTextPrimary
	}

	return btn.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return utils.SurfaceOutlined(gtx, bg, unit.Dp(8), utils.SurfaceBorder{Color: border, Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(9), Bottom: unit.Dp(9), Left: unit.Dp(14), Right: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, textColor, btn.Name)
			})
		})
	})
}

func (ui *GameUI) layoutScrollableConfigPanel(gtx layout.Context, body layout.Widget) layout.Dimensions {
	return ui.panelList.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
		if body == nil {
			return layout.Dimensions{}
		}
		return layout.UniformInset(unit.Dp(4)).Layout(gtx, body)
	})
}

func (ui *GameUI) layoutField(gtx layout.Context, label, help string, control layout.Widget) layout.Dimensions {
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, label)
		}),
	}
	if strings.TrimSpace(help) != "" {
		children = append(children,
			layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, help)
			}),
		)
	}
	if control != nil {
		children = append(children,
			layout.Rigid(layout.Spacer{Height: unit.Dp(7)}.Layout),
			layout.Rigid(control),
		)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (ui *GameUI) layoutInputField(gtx layout.Context, in *input.TextInput, label, help string) layout.Dimensions {
	if in == nil {
		return layout.Dimensions{}
	}
	oldLabel := in.Label
	in.Label = ""
	dims := ui.layoutField(gtx, label, help, in.Layout)
	in.Label = oldLabel
	return dims
}

func (ui *GameUI) layoutPathInputField(gtx layout.Context, in *input.TextInput, browse *components.IconButton, label, help string) layout.Dimensions {
	if in == nil {
		return layout.Dimensions{}
	}
	return ui.layoutField(gtx, label, help, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				oldLabel := in.Label
				in.Label = ""
				dims := in.Layout(gtx)
				in.Label = oldLabel
				return dims
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if browse == nil {
					return layout.Dimensions{}
				}
				if browse.Clicked(gtx) {
					ui.openFilePicker(in)
					gtx.Execute(op.InvalidateCmd{})
				}
				return browse.Layout(gtx)
			}),
		)
	})
}

func (ui *GameUI) layoutMutedText(gtx layout.Context, text string) layout.Dimensions {
	return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, text)
}

func (ui *GameUI) openFilePicker(target *input.TextInput) {
	if ui == nil || target == nil {
		return
	}
	startDir := filePickerStartDir(target.Text())
	explorer := fileexplorer.NewFileExplorer(startDir, media.DefaultRegistry, ui.theme).
		WithMaterialTheme(ui.th).
		WithThemeClient(ui.theme)
	explorer.SelectButtonText = "Use Path"
	explorer.OnChoose = func(path string) {
		if ui.filePickerTarget != nil {
			ui.filePickerTarget.SetText(path)
			ui.markDirty()
		}
		if ui.filePickerModal != nil {
			ui.filePickerModal.Dismiss()
		}
	}
	ui.filePicker = explorer
	ui.filePickerTarget = target
	ui.filePickerModal = modal.New("game-file-picker", "Select File or Folder", func(gtx layout.Context) layout.Dimensions {
		if ui.filePicker == nil {
			return layout.Dimensions{}
		}
		return ui.filePicker.Layout(gtx)
	}).WithThemeClient(ui.theme).WithMaterialTheme(ui.th).WithSize(unit.Dp(1600), unit.Dp(0))
	ui.filePickerModal.MaxWidth = 0
	ui.filePickerModal.MinHeight = unit.Dp(720)
	ui.filePickerModal.Margin = unit.Dp(16)
	ui.filePickerModal.Padding = unit.Dp(12)
	ui.filePickerModal.Radius = unit.Dp(12)
	ui.filePickerModal.Open()
}

func (ui *GameUI) layoutFilePickerOverlay(gtx layout.Context, layer *overlay.Overlay) {
	if ui == nil || layer == nil || ui.filePickerModal == nil || !ui.filePickerModal.Visible {
		return
	}
	layer.Add(gtx, ui.filePickerModal)
}

func filePickerStartDir(path string) string {
	path = strings.TrimSpace(path)
	if path != "" {
		if info, err := os.Stat(path); err == nil {
			if info.IsDir() {
				return path
			}
			return filepath.Dir(path)
		}
		if dir := filepath.Dir(path); dir != "." && dir != "" {
			if info, err := os.Stat(dir); err == nil && info.IsDir() {
				return dir
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}

func (ui *GameUI) activeRunnerFields() []*runnerOptionField {
	if ui == nil || ui.draft == nil {
		return nil
	}
	switch ui.draft.Runner {
	case game.RunnerWine:
		return ui.wineFields
	case game.RunnerGamescope:
		return ui.gamescopeFields
	default:
		return nil
	}
}

func (ui *GameUI) layoutActions(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = 0

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.saveButton.Clicked(gtx) {
				ui.save(gtx)
			}
			ui.saveButton.Disabled = ui.draft == nil
			return ui.layoutActionButton(gtx, ui.saveButton, "Save", false)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.cancelButton.Clicked(gtx) {
				ui.cancelEdits(gtx)
			}
			ui.cancelButton.Disabled = ui.draft == nil || !ui.dirty
			return ui.layoutActionButton(gtx, ui.cancelButton, "Cancel", false)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.resetButton.Clicked(gtx) {
				ui.resetRunnerDefaults(gtx)
			}
			ui.resetButton.Disabled = ui.draft == nil
			return ui.layoutActionButton(gtx, ui.resetButton, "Reset Defaults", false)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.deleteButton.Clicked(gtx) {
				ui.deleteCurrentGame(gtx)
			}
			ui.deleteButton.Disabled = ui.draft == nil
			return ui.layoutActionButton(gtx, ui.deleteButton, "Delete", true)
		}),
	)
}

func (ui *GameUI) layoutActionButton(gtx layout.Context, btn *components.IconButton, label string, danger bool) layout.Dimensions {
	if btn == nil || btn.Clickable == nil {
		return layout.Dimensions{}
	}

	gtx.Constraints.Min.X = 0
	gtx.Constraints.Min.Y = 0

	tokens := ui.theme.GetCurrentColorToken()

	bg := tokens.SurfaceAltNRGBA()
	border := tokens.BorderNRGBA()
	fg := theme.ThemeColorTextSecondary

	if danger {
		fg = theme.ThemeColorError
	}

	if btn.Disabled {
		bg = tokens.DisabledNRGBA()
		fg = theme.ThemeColorTextMuted
	} else if btn.Clickable.Hovered() {
		bg = tokens.SurfaceNRGBA()
		border = tokens.PrimaryNRGBA()
		if !danger {
			fg = theme.ThemeColorTextPrimary
		}
	}

	return btn.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min.X = 0
		gtx.Constraints.Min.Y = 0

		return utils.SurfaceOutlined(
			gtx,
			bg,
			unit.Dp(6),
			utils.SurfaceBorder{
				Color: border,
				Width: unit.Dp(1),
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(6),
					Bottom: unit.Dp(6),
					Left:   unit.Dp(10),
					Right:  unit.Dp(10),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						ui.th,
						ui.theme,
						theme.TextRoleLabel,
						fg,
						label,
					)
				})
			},
		)
	})
}

func (ui *GameUI) layoutDeleteRunnerConfig(gtx layout.Context) layout.Dimensions {
	return ui.layoutField(gtx, "Custom runner config", "Delete saved runner overrides for this game. The game will fall back to default runner settings.", func(gtx layout.Context) layout.Dimensions {
		if ui.deleteRunnerConfigButton.Clicked(gtx) {
			ui.deleteCustomRunnerConfig(gtx)
		}
		ui.deleteRunnerConfigButton.Disabled = ui.draft == nil
		return ui.layoutActionButton(gtx, ui.deleteRunnerConfigButton, "Delete Runner Config", true)
	})
}

func (ui *GameUI) layoutEmpty(gtx layout.Context) layout.Dimensions {
	return utils.Surface(gtx, ui.theme.GetCurrentColorToken().SurfaceNRGBA(), unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, "No game selected.")
		})
	})
}

func (ui *GameUI) save(gtx layout.Context) {
	if ui == nil || ui.backend == nil || ui.draft == nil {
		return
	}
	ui.applyForm()
	if strings.TrimSpace(ui.draft.Name) == "" {
		ui.status = "Game name is required"
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	if err := ui.validateRunnerFields(); err != nil {
		ui.status = err.Error()
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	if err := validateRunnerPath(ui.draft.Runner, ui.draft.RunnerPath); err != nil {
		ui.status = err.Error()
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	if err := ui.backend.SaveGameConfig(ui.draft, ui.previousName); err != nil {
		ui.status = "Save failed: " + err.Error()
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	ui.previousName = ui.draft.Name
	ui.loadedGame = strings.TrimSpace(ui.draft.Name)
	ui.gamesKey = ""
	ui.dirty = false
	ui.status = "Saved"
	gtx.Execute(op.InvalidateCmd{})
}

func (ui *GameUI) validateRunnerFields() error {
	for _, fields := range [][]*runnerOptionField{ui.gamescopeFields, ui.wineFields} {
		for _, field := range fields {
			if field == nil || field.input == nil {
				continue
			}
			if err := field.input.Validate(); err != nil {
				field.input.LastError = err
				return fmt.Errorf("%s: %s", field.label, field.input.ErrorText())
			}
		}
	}
	return nil
}

func (ui *GameUI) cancelEdits(gtx layout.Context) {
	if ui == nil || ui.backend == nil {
		return
	}
	ui.loadGame(ui.backend.CurrentGame())
	ui.status = "Changes discarded"
	gtx.Execute(op.InvalidateCmd{})
}

func (ui *GameUI) resetRunnerDefaults(gtx layout.Context) {
	if ui == nil || ui.draft == nil {
		return
	}
	switch ui.draft.Runner {
	case game.RunnerWine:
		cfg := defaultWineOptions(ui.draft.PrefixPath)
		ui.draft.WineConfig = &cfg
	case game.RunnerGamescope:
		cfg := defaultGamescopeOptions(ui.draft.PrefixPath)
		ui.draft.GamescopeConfig = &cfg
	}
	ui.draft.RunnerConfig = defaultGRConfigForGame(ui.draft)
	ui.loadRunnerFields()
	ui.markDirty()
	gtx.Execute(op.InvalidateCmd{})
}

func (ui *GameUI) deleteCurrentGame(gtx layout.Context) {
	if ui == nil || ui.backend == nil || ui.draft == nil {
		return
	}
	name := strings.TrimSpace(ui.draft.Name)
	if err := ui.backend.DeleteGameConfig(ui.draft); err != nil {
		ui.status = "Delete failed: " + err.Error()
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	ui.draft = nil
	ui.previousName = ""
	ui.loadedGame = ""
	ui.gamesKey = ""
	ui.dirty = false
	ui.syncGames()
	if name == "" {
		ui.status = "Game deleted"
	} else {
		ui.status = "Deleted " + name
	}
	gtx.Execute(op.InvalidateCmd{})
}

func (ui *GameUI) deleteCustomRunnerConfig(gtx layout.Context) {
	if ui == nil || ui.backend == nil || ui.draft == nil {
		return
	}
	if err := ui.backend.DeleteCustomRunnerConfig(ui.draft); err != nil {
		ui.status = "Delete runner config failed: " + err.Error()
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	ui.draft.RunnerPath = ""
	ui.draft.RunnerConfig = gr.Config{}
	ui.draft.WineConfig = nil
	ui.draft.GamescopeConfig = nil
	ui.ensureRunnerDefaults(ui.draft)
	ui.runnerPathInput.SetText("")
	ui.loadRunnerFields()
	ui.dirty = false
	ui.status = "Custom runner config deleted"
	gtx.Execute(op.InvalidateCmd{})
}

func (ui *GameUI) selectedGameValue() string {
	if ui == nil || ui.gameDropdown == nil {
		return ""
	}
	item, ok := ui.gameDropdown.SelectedItem()
	if !ok {
		return ""
	}
	return strings.TrimSpace(item.Value)
}

func (ui *GameUI) dropdownHasValue(value string) bool {
	if ui == nil || ui.gameDropdown == nil {
		return false
	}
	value = strings.TrimSpace(value)
	for _, item := range ui.gameDropdown.Items {
		if strings.TrimSpace(item.Value) == value {
			return true
		}
	}
	return false
}

func (ui *GameUI) applyForm() {
	if ui == nil || ui.draft == nil {
		return
	}
	ui.draft.Name = strings.TrimSpace(ui.nameInput.Text())
	ui.draft.IconPath = strings.TrimSpace(ui.iconInput.Text())
	ui.draft.ImagePath = strings.TrimSpace(ui.imageInput.Text())
	ui.draft.RunnerPath = strings.TrimSpace(ui.runnerPathInput.Text())
	ui.draft.PrefixPath = strings.TrimSpace(ui.prefixInput.Text())
	if item, ok := ui.runnerDropdown.SelectedItem(); ok {
		ui.draft.Runner = game.RunnerType(item.Value)
	}
	ui.ensureRunnerDefaults(ui.draft)

	if ui.draft.GamescopeConfig != nil {
		v := reflect.ValueOf(ui.draft.GamescopeConfig).Elem()
		applyRunnerOptionFields(ui.gamescopeFields, v)
	}
	if ui.draft.WineConfig != nil {
		v := reflect.ValueOf(ui.draft.WineConfig).Elem()
		applyRunnerOptionFields(ui.wineFields, v)
	}
}
