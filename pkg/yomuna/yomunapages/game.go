package yomunapages

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/gr/wine"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/input"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
	"github.com/DarlingGoose/wgl/pkg/yomuna/backend"
)

type GameUI struct {
	th      *material.Theme
	theme   *theme.Client
	backend backend.Backend

	list layout.List

	gameDropdown   *dropdowns.Dropdown
	runnerDropdown *dropdowns.Dropdown

	nameInput       *input.TextInput
	iconInput       *input.TextInput
	imageInput      *input.TextInput
	runnerPathInput *input.TextInput
	prefixInput     *input.TextInput

	gamescopeFields []*runnerOptionField
	wineFields      []*runnerOptionField

	saveButton    *components.IconButton
	cancelButton  *components.IconButton
	resetButton   *components.IconButton
	refreshButton *components.IconButton

	basicSection    collapsibleSection
	runnerSection   collapsibleSection
	wineSection     collapsibleSection
	displaySection  collapsibleSection
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

type runnerOptionField struct {
	name        string
	label       string
	description string
	kind        reflect.Kind
	isSlice     bool
	input       *input.TextInput
	toggle      *toggles.Toggle
	onChange    func()
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
	refreshIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:refresh-cw")

	ui := &GameUI{
		th:      th,
		theme:   tc,
		backend: b,
		list:    layout.List{Axis: layout.Vertical},

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

		saveButton:    components.NewIconButton("Save Game", nil, saveIcon).WithThemeClient(tc),
		cancelButton:  components.NewIconButton("Cancel", nil, cancelIcon).WithThemeClient(tc),
		resetButton:   components.NewIconButton("Reset Defaults", nil, resetIcon).WithThemeClient(tc),
		refreshButton: components.NewIconButton("Refresh", nil, refreshIcon).WithThemeClient(tc),

		basicSection:    newGameSection("General", "Name and artwork.", true),
		runnerSection:   newGameSection("Runner", "Runner type and core launch paths.", true),
		wineSection:     newGameSection("Wine behavior", "Wine process behavior for this runner.", true),
		displaySection:  newGameSection("Display", "Resolution, refresh rate, and window behavior.", true),
		advancedSection: newGameSection("Advanced", "Binary overrides and custom runner arguments.", false),
	}
	ui.saveButton.CollapseTextBelow = unit.Dp(140)
	ui.cancelButton.CollapseTextBelow = unit.Dp(120)
	ui.cancelButton.FillWidth = false
	ui.resetButton.CollapseTextBelow = unit.Dp(170)
	ui.resetButton.FillWidth = false
	ui.refreshButton.CollapseTextBelow = unit.Dp(120)
	ui.refreshButton.FillWidth = false

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
		ui.draft.Runner = game.RunnerType(item.Value)
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
	ui.syncGames()

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.list.Layout(gtx, 13, func(gtx layout.Context, index int) layout.Dimensions {
			switch index {
			case 0:
				return ui.layoutHeader(gtx, layer)
			case 1:
				return layout.Spacer{Height: unit.Dp(14)}.Layout(gtx)
			case 2:
				return ui.layoutSection(gtx, &ui.basicSection, ui.layoutGameFields)
			case 3:
				return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
			case 4:
				return ui.layoutSection(gtx, &ui.runnerSection, func(gtx layout.Context) layout.Dimensions {
					return ui.layoutRunnerFields(gtx, layer)
				})
			case 5:
				return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
			case 6:
				return ui.layoutSection(gtx, &ui.wineSection, ui.layoutWineOptions)
			case 7:
				return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
			case 8:
				return ui.layoutSection(gtx, &ui.displaySection, ui.layoutDisplayOptions)
			case 9:
				return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
			case 10:
				return ui.layoutSection(gtx, &ui.advancedSection, ui.layoutAdvancedOptions)
			case 11:
				return layout.Spacer{Height: unit.Dp(14)}.Layout(gtx)
			case 12:
				return ui.layoutActions(gtx)
			default:
				return layout.Dimensions{}
			}
		})
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
	ui.dirty = false
	ui.status = "Saved"
}

func (ui *GameUI) ensureRunnerDefaults(g *game.Game) {
	if g == nil {
		return
	}
	if g.Runner == "" {
		g.Runner = game.RunnerGamescope
	}
	if g.WineConfig == nil {
		cfg := wine.ApplyOptions()
		if strings.TrimSpace(cfg.DefaultPrefix) == "" {
			cfg.DefaultPrefix = g.PrefixPath
		}
		g.WineConfig = &cfg
	}
	if g.GamescopeConfig == nil {
		cfg := gamescope.ApplyOptions()
		cfg.UseWine = true
		cfg.Fullscreen = true
		if strings.TrimSpace(cfg.DefaultWinePrefix) == "" {
			cfg.DefaultWinePrefix = g.PrefixPath
		}
		g.GamescopeConfig = &cfg
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
	return utils.SurfaceOutlined(gtx, ct.SurfaceNRGBA(), unit.Dp(10), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH3, theme.ThemeColorTextPrimary, "Game Configuration")
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.headerSummary())
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ui.layoutStatusPill(gtx)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return ui.layoutField(gtx, "Editing", "Choose which installed game config to edit.", func(gtx layout.Context) layout.Dimensions {
								if ui.gameDropdown == nil {
									return layout.Dimensions{}
								}
								return ui.gameDropdown.Layout(gtx, layer)
							})
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ui.layoutRefreshButton(gtx)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutActions(gtx)
				}),
			)
		})
	})
}

func (ui *GameUI) headerSummary() string {
	name := "No game selected"
	runner := "Runner not selected"
	if ui.draft != nil {
		if strings.TrimSpace(ui.draft.Name) != "" {
			name = strings.TrimSpace(ui.draft.Name)
		}
		runner = runnerLabel(ui.draft.Runner)
		if ui.draft.Runner == game.RunnerGamescope {
			runner += " + Wine"
		}
	}
	return "Editing: " + name + "  |  Runner: " + runner
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
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutInputField(gtx, ui.nameInput, "Display name", "Used in game pickers and saved config filenames.")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutInputField(gtx, ui.iconInput, "Icon", "Small icon file for lists or compact game display.")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutInputField(gtx, ui.imageInput, "Cover image", "Larger picture used when the game needs artwork.")
		}),
	)
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
			return ui.layoutInputField(gtx, ui.runnerPathInput, "Runner executable", "Optional executable path. Leave empty to use the selected runner's default.")
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutInputField(gtx, ui.prefixInput, "Wine prefix", "Optional Wine prefix. Leave empty to use the game or runner default.")
		}),
	)
}

func (ui *GameUI) layoutRunnerOptions(gtx layout.Context) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.activeRunnerFields()
	if len(fields) == 0 {
		return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, "This runner does not have editable options yet.")
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, runnerFieldChildren(ui.th, ui.theme, fields)...)
}

func (ui *GameUI) layoutWineOptions(gtx layout.Context) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields("UseWine", "WineStartWait", "KillWineOnExit")
	if len(fields) == 0 {
		return ui.layoutMutedText(gtx, "No Wine behavior options are available for this runner.")
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, runnerFieldChildren(ui.th, ui.theme, fields)...)
}

func (ui *GameUI) layoutDisplayOptions(gtx layout.Context) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields(
		"Width", "Height", "RefreshRate", "OutputWidth", "OutputHeight",
		"Fullscreen", "Borderless", "ForceGrab", "SteamDeckMode", "ExposeWayland",
	)
	if len(fields) == 0 {
		return ui.layoutMutedText(gtx, "No display options are available for this runner.")
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, runnerFieldChildren(ui.th, ui.theme, fields)...)
}

func (ui *GameUI) layoutAdvancedOptions(gtx layout.Context) layout.Dimensions {
	if ui.draft == nil {
		return ui.layoutEmpty(gtx)
	}
	fields := ui.filterRunnerFields(
		"Name", "GamescopeBin", "WineBin", "WineServerBin", "WineTricksBin",
		"DefaultWinePrefix", "DefaultPrefix", "Scaler", "Filter", "ExtraArgs",
	)
	if len(fields) == 0 {
		return ui.layoutMutedText(gtx, "No advanced options are available for this runner.")
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, runnerFieldChildren(ui.th, ui.theme, fields)...)
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

func runnerFieldChildren(th *material.Theme, tc *theme.Client, fields []*runnerOptionField) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(fields)*2)
	for i, field := range fields {
		f := field
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return f.Layout(gtx, th, tc)
		}))
		if i != len(fields)-1 {
			children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout))
		}
	}
	return children
}

func (ui *GameUI) layoutSection(gtx layout.Context, section *collapsibleSection, body layout.Widget) layout.Dimensions {
	if section == nil {
		return layout.Dimensions{}
	}
	for section.click.Clicked(gtx) {
		section.open = !section.open
		gtx.Execute(op.InvalidateCmd{})
	}
	ct := ui.theme.GetCurrentColorToken()
	return utils.SurfaceOutlined(gtx, ct.SurfaceNRGBA(), unit.Dp(10), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return section.click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
								return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH4, theme.ThemeColorTextSecondary, icon)
							}),
						)
					})
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

func (ui *GameUI) layoutMutedText(gtx layout.Context, text string) layout.Dimensions {
	return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, text)
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
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.saveButton.Clicked(gtx) {
				ui.save(gtx)
			}
			ui.saveButton.Disabled = ui.draft == nil
			return ui.saveButton.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.cancelButton.Clicked(gtx) {
				ui.cancelEdits(gtx)
			}
			ui.cancelButton.Disabled = ui.draft == nil || !ui.dirty
			return ui.cancelButton.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.resetButton.Clicked(gtx) {
				ui.resetRunnerDefaults(gtx)
			}
			ui.resetButton.Disabled = ui.draft == nil
			return ui.resetButton.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if ui.status == "" {
				return layout.Dimensions{}
			}
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.status)
		}),
	)
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
		cfg := wine.ApplyOptions()
		cfg.DefaultPrefix = ui.draft.PrefixPath
		ui.draft.WineConfig = &cfg
	case game.RunnerGamescope:
		cfg := gamescope.ApplyOptions()
		cfg.UseWine = true
		cfg.Fullscreen = true
		cfg.DefaultWinePrefix = ui.draft.PrefixPath
		ui.draft.GamescopeConfig = &cfg
	}
	ui.loadRunnerFields()
	ui.markDirty()
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

func buildRunnerOptionFields(sample any, th *material.Theme, tc *theme.Client, onChange func()) []*runnerOptionField {
	t := reflect.TypeOf(sample)
	fields := make([]*runnerOptionField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		f := &runnerOptionField{
			name:        sf.Name,
			label:       runnerOptionLabel(sf.Name),
			description: runnerOptionDescription(sf.Name),
			kind:        sf.Type.Kind(),
			isSlice:     sf.Type.Kind() == reflect.Slice,
			onChange:    onChange,
		}
		switch {
		case f.kind == reflect.Bool:
			f.toggle = toggles.NewToggle("", false).WithThemeClient(tc)
		default:
			f.input = input.NewTextInput(f.label, "").WithMaterialTheme(th).WithThemeClient(tc)
			if f.kind >= reflect.Int && f.kind <= reflect.Int64 {
				f.input.Kind = input.KindInteger
				f.input.LeadingIcon = "lucide:hash"
			}
			if f.isSlice {
				f.input.Hint = "Comma-separated values"
			}
			f.input.OnChange = func(string) {
				if onChange != nil {
					onChange()
				}
			}
		}
		fields = append(fields, f)
	}
	return fields
}

func setRunnerOptionFields(fields []*runnerOptionField, v reflect.Value) {
	for _, f := range fields {
		field := v.FieldByName(f.name)
		if !field.IsValid() {
			continue
		}
		switch {
		case f.toggle != nil && field.Kind() == reflect.Bool:
			f.toggle.JumpTo(field.Bool())
		case f.input != nil && field.Kind() == reflect.String:
			f.input.SetText(field.String())
		case f.input != nil && field.Kind() >= reflect.Int && field.Kind() <= reflect.Int64:
			f.input.SetText(strconv.FormatInt(field.Int(), 10))
		case f.input != nil && field.Kind() == reflect.Slice:
			parts := make([]string, 0, field.Len())
			for i := 0; i < field.Len(); i++ {
				parts = append(parts, fmt.Sprint(field.Index(i).Interface()))
			}
			f.input.SetText(strings.Join(parts, ", "))
		}
	}
}

func applyRunnerOptionFields(fields []*runnerOptionField, v reflect.Value) {
	for _, f := range fields {
		field := v.FieldByName(f.name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		switch {
		case f.toggle != nil && field.Kind() == reflect.Bool:
			field.SetBool(f.toggle.Checked)
		case f.input != nil && field.Kind() == reflect.String:
			field.SetString(strings.TrimSpace(f.input.Text()))
		case f.input != nil && field.Kind() >= reflect.Int && field.Kind() <= reflect.Int64:
			text := strings.TrimSpace(f.input.Text())
			if text == "" {
				field.SetInt(0)
				continue
			}
			if n, err := strconv.ParseInt(text, 10, 64); err == nil {
				field.SetInt(n)
			}
		case f.input != nil && field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String:
			field.Set(reflect.ValueOf(splitCommaList(f.input.Text())))
		}
	}
}

func (f *runnerOptionField) Layout(gtx layout.Context, th *material.Theme, tc *theme.Client) layout.Dimensions {
	if f == nil {
		return layout.Dimensions{}
	}
	control := func(gtx layout.Context) layout.Dimensions {
		if f.toggle != nil {
			if f.toggle.Update(gtx) && f.onChange != nil {
				f.onChange()
			}
			return f.toggle.Layout(gtx)
		}
		if f.input != nil {
			return f.input.Layout(gtx)
		}
		return layout.Dimensions{}
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, th, tc, theme.TextRoleLabel, theme.ThemeColorTextPrimary, f.label)
		}),
	}
	if strings.TrimSpace(f.description) != "" {
		children = append(children,
			layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, th, tc, theme.TextRoleCaption, theme.ThemeColorTextMuted, f.description)
			}),
		)
	}
	children = append(children,
		layout.Rigid(layout.Spacer{Height: unit.Dp(7)}.Layout),
		layout.Rigid(control),
	)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func cloneGame(g *game.Game) *game.Game {
	if g == nil {
		return nil
	}
	cp := *g
	cp.EnvVars = append([]game.EnvVar(nil), g.EnvVars...)
	cp.TextHookFilter = append([]string(nil), g.TextHookFilter...)
	if g.WineConfig != nil {
		cfg := *g.WineConfig
		cp.WineConfig = &cfg
	}
	if g.GamescopeConfig != nil {
		cfg := *g.GamescopeConfig
		cfg.ExtraArgs = append([]string(nil), g.GamescopeConfig.ExtraArgs...)
		cp.GamescopeConfig = &cfg
	}
	return &cp
}

func splitCamel(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func splitCommaList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func runnerLabel(r game.RunnerType) string {
	switch r {
	case game.RunnerWine:
		return "Wine"
	case game.RunnerGamescope:
		return "Gamescope"
	case game.RunnerProton:
		return "Proton"
	case game.RunnerSteam:
		return "Steam"
	default:
		return "Gamescope"
	}
}

func runnerOptionLabel(name string) string {
	switch name {
	case "Width":
		return "Width (px)"
	case "Height":
		return "Height (px)"
	case "OutputWidth":
		return "Output width (px)"
	case "OutputHeight":
		return "Output height (px)"
	case "RefreshRate":
		return "Refresh rate (Hz)"
	default:
		return splitCamel(name)
	}
}

func runnerOptionDescription(name string) string {
	switch name {
	case "Name":
		return "Internal runner profile name."
	case "ExtraArgs":
		return "Additional runner arguments, separated by commas."
	case "DefaultPrefix", "DefaultWinePrefix":
		return "Used when the game does not provide a more specific prefix."
	case "UseWine":
		return "Launch the game through Wine inside the runner."
	case "WineStartWait":
		return "Wait for Wine startup before continuing."
	case "KillWineOnExit":
		return "Stop Wine processes when the game exits."
	case "Width", "Height", "OutputWidth", "OutputHeight":
		return "Pixel value passed to the runner. Use 0 to let the runner choose."
	case "RefreshRate":
		return "Target refresh rate. Use 0 for the monitor default."
	case "GamescopeBin":
		return "Binary override. Leave empty to use gamescope from PATH."
	case "WineBin":
		return "Binary override. Leave empty to use wine from PATH."
	case "WineServerBin":
		return "Binary override. Leave empty to use wineserver from PATH."
	case "WineTricksBin":
		return "Binary override. Leave empty to use winetricks from PATH."
	default:
		return ""
	}
}
