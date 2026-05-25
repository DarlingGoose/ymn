package gamepage

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/modal"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	layoutgrid "github.com/DarlingGoose/ymn/pkg/v2/gui/layouts/grid"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
	"github.com/DarlingGoose/ymn/pkg/yomuna/backend"
)

const (
	gameSortLastPlayed = "last_played"
	gameSortNewest     = "newest"
	gameSortName       = "name"
)

type GameSelectorUI struct {
	th      *material.Theme
	theme   *theme.Client
	backend backend.Backend

	list         layout.List
	grid         layoutgrid.Grid
	sortDropdown *dropdowns.Dropdown
	refresh      *components.IconButton
	addGame      *components.IconButton
	playIcon     *iconify.SVGIcon
	settingsIcon *iconify.SVGIcon
	configUI     *GameUI
	addGameUI    *AddGameUI

	cards          map[string]*widget.Clickable
	playClicks     map[string]*widget.Clickable
	settingsClicks map[string]*widget.Clickable
	coverViews     map[string]*media.View
	settingsModal  *modal.Modal
	addGameModal   *modal.Modal
	status         string
	activeLayer    *overlay.Overlay
	playAction     func(*game.Game)
}

func NewGameSelectorUI(th *material.Theme, tc *theme.Client, b backend.Backend) *GameSelectorUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	refreshIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:refresh-cw")
	plusIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:plus")
	playIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:play")
	settingsIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:settings")
	ui := &GameSelectorUI{
		th:      th,
		theme:   tc,
		backend: b,
		list:    layout.List{Axis: layout.Vertical},
		grid: layoutgrid.Grid{
			MinCellWidth: unit.Dp(178),
			Gap:          unit.Dp(12),
			MaxColumns:   6,
		},
		sortDropdown: dropdowns.NewDropdown([]dropdowns.DropdownItem{
			{Label: "Last played", Value: gameSortLastPlayed},
			{Label: "Newest", Value: gameSortNewest},
			{Label: "Name", Value: gameSortName},
		}).WithThemeClient(tc).WithRole(theme.TextRoleLabel),
		refresh:        components.NewIconButton("Refresh", nil, refreshIcon).WithThemeClient(tc),
		addGame:        components.NewIconButton("Add Game", nil, plusIcon).WithThemeClient(tc),
		playIcon:       playIcon,
		settingsIcon:   settingsIcon,
		cards:          map[string]*widget.Clickable{},
		playClicks:     map[string]*widget.Clickable{},
		settingsClicks: map[string]*widget.Clickable{},
		coverViews:     map[string]*media.View{},
		status:         "Select a game to make it active.",
	}
	ui.refresh.FillWidth = false
	ui.refresh.TextCollapseMode = components.TextCollapseNever
	ui.refresh.CollapseTextBelow = unit.Dp(140)
	ui.refresh.MinWidth = unit.Dp(104)
	ui.refresh.Height = unit.Dp(44)
	ui.refresh.Radius = unit.Dp(10)
	ui.addGame.FillWidth = false
	ui.addGame.TextCollapseMode = components.TextCollapseNever
	ui.addGame.CollapseTextBelow = unit.Dp(150)
	ui.addGame.MinWidth = unit.Dp(116)
	ui.addGame.Height = unit.Dp(44)
	ui.addGame.Radius = unit.Dp(10)
	return ui
}

func (ui *GameSelectorUI) WithConfigUI(configUI *GameUI) *GameSelectorUI {
	if ui == nil {
		return ui
	}
	ui.configUI = configUI
	return ui
}

func (ui *GameSelectorUI) WithAddGameUI(addGameUI *AddGameUI) *GameSelectorUI {
	if ui == nil {
		return ui
	}
	ui.addGameUI = addGameUI
	return ui
}

func (ui *GameSelectorUI) WithPlayAction(fn func(*game.Game)) *GameSelectorUI {
	if ui == nil {
		return ui
	}
	ui.playAction = fn
	return ui
}

func (ui *GameSelectorUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}
	ui.activeLayer = layer
	dims := layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.layoutContentContainer(gtx, func(gtx layout.Context) layout.Dimensions {
			games := ui.sortedGames()
			if ui.backend != nil && ui.backend.IsGameRunning() {
				gtx.Execute(op.InvalidateCmd{})
			}
			return ui.list.Layout(gtx, 4, func(gtx layout.Context, index int) layout.Dimensions {
				switch index {
				case 0:
					return ui.layoutToolbar(gtx, layer, len(games))
				case 1:
					return layout.Spacer{Height: unit.Dp(14)}.Layout(gtx)
				case 2:
					return ui.layoutGames(gtx, games)
				case 3:
					return layout.Spacer{Height: unit.Dp(24)}.Layout(gtx)
				default:
					return layout.Dimensions{}
				}
			})
		})
	})
	ui.layoutSettingsOverlay(gtx, layer)
	ui.layoutAddGameOverlay(gtx, layer)
	return dims
}

func (ui *GameSelectorUI) layoutContentContainer(gtx layout.Context, content layout.Widget) layout.Dimensions {
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

func (ui *GameSelectorUI) layoutToolbar(gtx layout.Context, layer *overlay.Overlay, count int) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	return utils.SurfaceOutlined(gtx, ct.SurfaceAltNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH2, theme.ThemeColorTextPrimary, "Games")
								}),
								layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, ui.selectorSummary(count))
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return ui.layoutToolbarActions(gtx)
						}),
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(520)) {
						ui.sortDropdown.Width = unit.Dp(float32(gtx.Constraints.Max.X) / gtx.Metric.PxPerDp)
					} else {
						ui.sortDropdown.Width = unit.Dp(180)
					}
					return ui.sortDropdown.Layout(gtx, layer)
				}),
			)
		})
	})
}

func (ui *GameSelectorUI) layoutToolbarActions(gtx layout.Context) layout.Dimensions {
	if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(520)) {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.End}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutAddGameButton(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutRefresh(gtx)
			}),
		)
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutAddGameButton(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutRefresh(gtx)
		}),
	)
}

func (ui *GameSelectorUI) layoutAddGameButton(gtx layout.Context) layout.Dimensions {
	if ui.addGame.Clicked(gtx) {
		ui.openAddGameModal()
		gtx.Execute(op.InvalidateCmd{})
	}
	return ui.addGame.Layout(gtx)
}

func (ui *GameSelectorUI) selectorSummary(count int) string {
	prefix := "No games installed"
	if count == 1 {
		prefix = "1 game installed"
	} else if count > 1 {
		prefix = "Sorting " + strconv.Itoa(count) + " games"
	}
	if strings.TrimSpace(ui.status) == "" {
		return prefix
	}
	return prefix + " - " + ui.status
}

func (ui *GameSelectorUI) layoutRefresh(gtx layout.Context) layout.Dimensions {
	if ui.refresh.Clicked(gtx) {
		if ui.backend != nil {
			if err := ui.backend.ReloadGames(); err != nil {
				ui.status = "Refresh failed: " + err.Error()
			} else {
				ui.status = "Game list refreshed"
			}
		}
		gtx.Execute(op.InvalidateCmd{})
	}
	return ui.refresh.Layout(gtx)
}

func (ui *GameSelectorUI) layoutGames(gtx layout.Context, games []*game.Game) layout.Dimensions {
	if len(games) == 0 {
		ct := ui.theme.GetCurrentColorToken()
		return utils.SurfaceOutlined(gtx, ct.SurfaceNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(180))
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, "No games found")
			})
		})
	}
	return ui.grid.Layout(gtx, len(games), func(gtx layout.Context, index int) layout.Dimensions {
		return ui.layoutGameCard(gtx, games[index])
	})
}

func (ui *GameSelectorUI) layoutGameCard(gtx layout.Context, g *game.Game) layout.Dimensions {
	if g == nil {
		return layout.Dimensions{}
	}
	name := strings.TrimSpace(g.Name)
	if name == "" {
		name = "Untitled Game"
	}
	click := ui.cardClick(name)
	for click.Clicked(gtx) {
		if ui.backend != nil {
			ui.backend.SelectGame(g)
		}
		ui.status = "Selected " + name
		gtx.Execute(op.InvalidateCmd{})
	}

	ct := ui.theme.GetCurrentColorToken()
	borderColor := ct.BorderNRGBA()
	if current := ui.currentName(); current != "" && strings.EqualFold(current, strings.TrimSpace(g.Name)) {
		borderColor = ct.PrimaryNRGBA()
	}
	return utils.ClickableSurfaceOutlined(gtx, click, ct.SurfaceNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: borderColor, Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutCardCover(gtx, g)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, name)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, ui.cardMeta(g))
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutCardActions(gtx, g)
				}),
			)
		})
	})
}

func (ui *GameSelectorUI) layoutCardActions(gtx layout.Context, g *game.Game) layout.Dimensions {
	name := strings.TrimSpace(g.Name)
	play := ui.iconButtonFor(ui.playClicks, name)
	settings := ui.iconButtonFor(ui.settingsClicks, name)

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if play.Clicked(gtx) {
				ui.selectGame(g)
				if ui.playAction != nil {
					ui.playAction(g)
				}
				gtx.Execute(op.InvalidateCmd{})
			}
			return ui.layoutSmallAction(gtx, play, ui.playIcon, "Play")
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if settings.Clicked(gtx) {
				ui.selectGame(g)
				ui.openSettingsModal(g)
				gtx.Execute(op.InvalidateCmd{})
			}
			return ui.layoutSmallAction(gtx, settings, ui.settingsIcon, "Config")
		}),
	)
}

func (ui *GameSelectorUI) layoutSmallAction(gtx layout.Context, click *widget.Clickable, icon *iconify.SVGIcon, label string) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		bg := ct.SurfaceAltNRGBA()
		border := ct.BorderNRGBA()
		if click.Hovered() {
			bg = ct.SurfaceNRGBA()
			border = ct.PrimaryNRGBA()
		}
		return utils.SurfaceOutlined(gtx, bg, unit.Dp(6), utils.SurfaceBorder{Color: border, Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(7), Bottom: unit.Dp(7), Left: unit.Dp(9), Right: unit.Dp(9)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if icon == nil {
							return layout.Dimensions{}
						}
						return icon.Layout(gtx, unit.Dp(16), ct.TextSecondaryNRGBA())
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabelSmall, theme.ThemeColorTextSecondary, label)
					}),
				)
			})
		})
	})
}

func (ui *GameSelectorUI) layoutCardCover(gtx layout.Context, g *game.Game) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	width := gtx.Constraints.Max.X
	if width <= 0 {
		width = gtx.Dp(unit.Dp(178))
	}
	height := width * 16 / 9
	minHeight := gtx.Dp(unit.Dp(220))
	maxHeight := gtx.Dp(unit.Dp(340))
	if height < minHeight {
		height = minHeight
	}
	if height > maxHeight {
		height = maxHeight
	}
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height
	return utils.SurfaceOutlined(gtx, ct.SurfaceAltNRGBA(), unit.Dp(7), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		path := strings.TrimSpace(g.ImagePath)
		if path == "" {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleCaption, theme.ThemeColorTextMuted, "No cover")
			})
		}
		view := ui.coverView(path)
		return layout.Center.Layout(gtx, view.Layout)
	})
}

func (ui *GameSelectorUI) sortedGames() []*game.Game {
	if ui == nil || ui.backend == nil {
		return nil
	}
	games := ui.backend.GetGames()
	out := make([]*game.Game, 0, len(games))
	for _, g := range games {
		if g != nil {
			out = append(out, g)
		}
	}
	sortKey := gameSortLastPlayed
	if item, ok := ui.sortDropdown.SelectedItem(); ok {
		sortKey = item.Value
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch sortKey {
		case gameSortNewest:
			return a.CreatedAt.After(b.CreatedAt)
		case gameSortName:
			return strings.ToLower(strings.TrimSpace(a.Name)) < strings.ToLower(strings.TrimSpace(b.Name))
		case gameSortLastPlayed:
			fallthrough
		default:
			at := ui.backend.GameLastPlayed(a.Name)
			bt := ui.backend.GameLastPlayed(b.Name)
			if !at.Equal(bt) {
				return at.After(bt)
			}
			return strings.ToLower(strings.TrimSpace(a.Name)) < strings.ToLower(strings.TrimSpace(b.Name))
		}
	})
	return out
}

func (ui *GameSelectorUI) cardClick(name string) *widget.Clickable {
	if ui.cards == nil {
		ui.cards = map[string]*widget.Clickable{}
	}
	key := strings.TrimSpace(name)
	if key == "" {
		key = "untitled"
	}
	if ui.cards[key] == nil {
		ui.cards[key] = &widget.Clickable{}
	}
	return ui.cards[key]
}

func (ui *GameSelectorUI) iconButtonFor(store map[string]*widget.Clickable, name string) *widget.Clickable {
	key := strings.TrimSpace(name)
	if key == "" {
		key = "untitled"
	}
	if store[key] == nil {
		store[key] = &widget.Clickable{}
	}
	return store[key]
}

func (ui *GameSelectorUI) selectGame(g *game.Game) {
	if ui == nil || ui.backend == nil || g == nil {
		return
	}
	ui.backend.SelectGame(g)
	name := strings.TrimSpace(g.Name)
	if name == "" {
		name = "game"
	}
	ui.status = "Selected " + name
}

func (ui *GameSelectorUI) openSettingsModal(g *game.Game) {
	if ui == nil || g == nil {
		return
	}
	if ui.configUI != nil {
		ui.configUI.loadGame(g)
	}
	title := strings.TrimSpace(g.Name)
	if title == "" {
		title = "Game Config"
	}
	ui.settingsModal = modal.New("game-settings", title, ui.layoutConfigModalContent).
		WithThemeClient(ui.theme).
		WithMaterialTheme(ui.th).
		WithSize(unit.Dp(1120), unit.Dp(760))
	ui.settingsModal.MaxWidth = 0
	ui.settingsModal.MinHeight = unit.Dp(620)
	ui.settingsModal.Margin = unit.Dp(16)
	ui.settingsModal.Padding = unit.Dp(0)
	ui.settingsModal.Radius = unit.Dp(12)
	ui.settingsModal.Description = "Edit the same game details shown in the sidebar."
	ui.settingsModal.Open()
}

func (ui *GameSelectorUI) layoutSettingsOverlay(gtx layout.Context, layer *overlay.Overlay) {
	if ui == nil || layer == nil || ui.settingsModal == nil || !ui.settingsModal.Visible {
		return
	}
	layer.Add(gtx, ui.settingsModal)
}

func (ui *GameSelectorUI) layoutAddGameOverlay(gtx layout.Context, layer *overlay.Overlay) {
	if ui == nil || layer == nil || ui.addGameModal == nil || !ui.addGameModal.Visible {
		return
	}
	layer.Add(gtx, ui.addGameModal)
}

func (ui *GameSelectorUI) layoutConfigModalContent(gtx layout.Context) layout.Dimensions {
	if ui.configUI == nil {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, "Game config is unavailable.")
		})
	}
	ui.configUI.hideGameDropdown = true
	defer func() {
		ui.configUI.hideGameDropdown = false
	}()
	return ui.configUI.Layout(gtx, ui.activeLayer)
}

func (ui *GameSelectorUI) openAddGameModal() {
	if ui == nil {
		return
	}
	ui.addGameModal = modal.New("game-add", "Add Game", ui.layoutAddGameModalContent).
		WithThemeClient(ui.theme).
		WithMaterialTheme(ui.th).
		WithSize(unit.Dp(980), unit.Dp(680))
	ui.addGameModal.MaxWidth = 0
	ui.addGameModal.MinHeight = unit.Dp(420)
	ui.addGameModal.Margin = unit.Dp(16)
	ui.addGameModal.Padding = unit.Dp(0)
	ui.addGameModal.Radius = unit.Dp(12)
	ui.addGameModal.Description = "Create a new game config."
	ui.addGameModal.Open()
}

func (ui *GameSelectorUI) layoutAddGameModalContent(gtx layout.Context) layout.Dimensions {
	if ui.addGameUI == nil {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, "Add game is unavailable.")
		})
	}
	return ui.addGameUI.Layout(gtx, ui.activeLayer)
}

func (ui *GameSelectorUI) coverView(path string) *media.View {
	if ui.coverViews == nil {
		ui.coverViews = map[string]*media.View{}
	}
	view := ui.coverViews[path]
	if view == nil {
		view = media.NewView(media.DefaultRegistry).WithImageFit(widget.Cover)
		ui.coverViews[path] = view
	}
	if view.Source.Path != path {
		_ = view.LoadPath(context.Background(), path)
	}
	return view
}

func (ui *GameSelectorUI) cardMeta(g *game.Game) string {
	if ui == nil || g == nil || ui.backend == nil {
		return ""
	}
	playtime := ui.backend.GamePlaytime(g.Name)
	last := ui.backend.GameLastPlayed(g.Name)
	prefix := "Played " + formatPlaytime(playtime)
	if !last.IsZero() {
		return prefix + " - Last played " + last.Format("Jan 2, 2006")
	}
	if !g.CreatedAt.IsZero() {
		return prefix + " - Added " + g.CreatedAt.Format("Jan 2, 2006")
	}
	return prefix
}

func (ui *GameSelectorUI) currentName() string {
	if ui == nil || ui.backend == nil {
		return ""
	}
	if current := ui.backend.CurrentGame(); current != nil {
		return strings.TrimSpace(current.Name)
	}
	return ""
}

func formatPlaytime(d time.Duration) string {
	if d < time.Minute {
		return "0m"
	}
	d = d.Round(time.Minute)
	hours := int(d / time.Hour)
	minutes := int((d % time.Hour) / time.Minute)
	if hours <= 0 {
		return strconv.Itoa(minutes) + "m"
	}
	if minutes == 0 {
		return strconv.Itoa(hours) + "h"
	}
	return strconv.Itoa(hours) + "h " + strconv.Itoa(minutes) + "m"
}
