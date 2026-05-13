package examples

import (
	"context"
	"fmt"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/tabs"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/layouts/grid"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/layouts/sidebar"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/pages"
)

type GridCard struct {
	Title string
	Body  string
}

type SidebarGridAppUI struct {
	th *material.Theme

	Sidebar *sidebar.CollapsibleSidebar

	ToggleButton *components.IconButton
	Settings     *pages.SettingsUI
	Overlay      *overlay.Overlay

	Grid grid.ScrollGrid

	Cards []GridCard

	AddButton    widget.Clickable
	RemoveButton widget.Clickable
}

func NewSidebarGridAppUI(th *material.Theme) *SidebarGridAppUI {
	if th == nil {
		th = material.NewTheme()
	}

	ui := &SidebarGridAppUI{
		th:       th,
		Settings: pages.NewSettingsUI(theme.DefaultThemeClient),
		Overlay:  &overlay.Overlay{},
		Grid: grid.ScrollGrid{
			Grid: grid.Grid{
				MinCellWidth: unit.Dp(220),
				Gap:          unit.Dp(12),
				MinColumns:   1,
				MaxColumns:   4,
				Inset:        layout.UniformInset(unit.Dp(16)),
			},
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		Cards: []GridCard{
			{Title: "Library", Body: "Browse installed visual novels and launchers."},
			{Title: "Hooks", Body: "Manage Textractor, Wine hooks, and captured text."},
			{Title: "Themes", Body: "Preview dark and light color palettes."},
			{Title: "Prefixes", Body: "Inspect Wine prefixes and installed games."},
			{Title: "Logs", Body: "View runner logs, hook logs, and extraction output."},
			{Title: "Tools", Body: "Quick access to helper commands and diagnostics."},
		},
	}

	menuIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "mdi:menu")

	toggleButton := components.NewIconButton("Toggle", nil, menuIcon)
	toggleButton.MinWidth = unit.Dp(0)
	toggleButton.CollapseTextBelow = unit.Dp(120)
	toggleButton.IconSize = unit.Dp(20)
	ui.ToggleButton = toggleButton

	appTabs := tabs.New(
		tabs.NewTabFunc("home", "Home", "lucide:home", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutGridHome(gtx)
		}),

		tabs.NewTabFunc("settings", "Settings", "lucide:settings", func(gtx layout.Context) layout.Dimensions {
			return ui.Settings.Layout(gtx, ui.Overlay)
		}).WithPinned(true),

		tabs.NewTabFunc("about", "About", "lucide:info", func(gtx layout.Context) layout.Dimensions {
			return ui.layoutAbout(gtx)
		}),
	)

	ui.Sidebar = sidebar.NewCollapsibleSidebar(appTabs).
		WithThemeClient(theme.DefaultThemeClient).
		WithTitle("Yomuna").
		WithIcon("lucide:book-open")

	return ui
}

func (ui *SidebarGridAppUI) Layout(gtx layout.Context) layout.Dimensions {
	if ui == nil || ui.Sidebar == nil {
		return layout.Dimensions{}
	}

	return ui.Overlay.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.Sidebar.Layout(
			gtx,
			ui.layoutSidebar,
			func(gtx layout.Context) layout.Dimensions {
				if ui.Sidebar.Tabs == nil {
					return ui.layoutGridHome(gtx)
				}

				return panel.NewBackgroundPanel(theme.DefaultThemeClient).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ui.Sidebar.Tabs.Layout(gtx)
				})
			},
		)
	})
}

func (ui *SidebarGridAppUI) layoutSidebar(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.Sidebar.LayoutHeader(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if ui.ToggleButton.Clicked(gtx) {
					ui.Sidebar.Toggle(time.Now())
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

func (ui *SidebarGridAppUI) layoutGridHome(gtx layout.Context) layout.Dimensions {
	for ui.AddButton.Clicked(gtx) {
		next := len(ui.Cards) + 1
		ui.Cards = append(ui.Cards, GridCard{
			Title: fmt.Sprintf("Card %d", next),
			Body:  "This card was added dynamically.",
		})
		gtx.Execute(op.InvalidateCmd{})
	}

	for ui.RemoveButton.Clicked(gtx) {
		if len(ui.Cards) > 0 {
			ui.Cards = ui.Cards[:len(ui.Cards)-1]
			gtx.Execute(op.InvalidateCmd{})
		}
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutGridToolbar(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return grid.LayoutScrollSlice(gtx, &ui.Grid, ui.Cards, func(gtx layout.Context, item GridCard, index int) layout.Dimensions {
				return ui.layoutCard(gtx, item, index)
			})
		}),
	)
}

func (ui *SidebarGridAppUI) layoutGridToolbar(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.H5(ui.th, "Dashboard").Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Body2(ui.th, fmt.Sprintf("%d cards", len(ui.Cards))).Layout(gtx)
					}),
				)
			}),
			layout.Flexed(1, layout.Spacer{}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(ui.th, &ui.RemoveButton, "Remove").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(ui.th, &ui.AddButton, "Add").Layout(gtx)
			}),
		)
	})
}

func (ui *SidebarGridAppUI) layoutCard(gtx layout.Context, item GridCard, index int) layout.Dimensions {
	tc := theme.DefaultThemeClient

	gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(120))

	return panel.NewBackgroundPanel(tc).
		WithRole(panel.BackgroundRoleSurfaceAlt).
		WithRadius(unit.Dp(12)).
		WithInset(layout.UniformInset(unit.Dp(14))).
		WithFillMax(false).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						ui.th,
						tc,
						theme.TextRoleH4,
						theme.ThemeColorTextPrimary,
						item.Title,
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						ui.th,
						tc,
						theme.TextRoleBodySmall,
						theme.ThemeColorTextSecondary,
						item.Body,
					)
				}),
			)
		})
}
func (ui *SidebarGridAppUI) layoutAbout(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.H5(ui.th, "About").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(ui.th, "This example renders a responsive dynamic grid inside the sidebar app content area.").Layout(gtx)
			}),
		)
	})
}
