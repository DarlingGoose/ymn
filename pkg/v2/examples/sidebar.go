package examples

import (
	"context"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/panel"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/tabs"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/layouts/sidebar"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/pages"
)

type SidebarAppUI struct {
	th *material.Theme

	Sidebar *sidebar.CollapsibleSidebar

	ToggleButton *components.IconButton
	Settings     *pages.SettingsUI
	Overlay      *overlay.Overlay
}

func NewSidebarAppUI(th *material.Theme) *SidebarAppUI {
	if th == nil {
		th = material.NewTheme()
	}

	overlayLayer := &overlay.Overlay{}
	settings := pages.NewSettingsUI(theme.DefaultThemeClient)

	trashIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "mdi:menu")

	toggleButton := components.NewIconButton("Toggle", nil, trashIcon)
	toggleButton.MinWidth = unit.Dp(0)
	toggleButton.CollapseTextBelow = unit.Dp(120)
	toggleButton.IconSize = unit.Dp(20)

	appTabs := tabs.New(
		tabs.NewTabFunc("home", "Home", "lucide:home", func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H5(th, "Home")
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						body := material.Body1(th, "This is the Home tab content.")
						return body.Layout(gtx)
					}),
				)
			})
		}),

		tabs.NewTabFunc("settings", "Settings", "lucide:settings", func(gtx layout.Context) layout.Dimensions {
			return settings.Layout(gtx, overlayLayer)
		}).WithPinned(true),

		tabs.NewTabFunc("about", "About", "lucide:info", func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						title := material.H5(th, "About")
						return title.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						body := material.Body1(th, "This tab system can later be rendered by a top bar or a sidebar.")
						return body.Layout(gtx)
					}),
				)
			})
		}),
	)

	sb := sidebar.NewCollapsibleSidebar(appTabs).
		WithThemeClient(theme.DefaultThemeClient).WithTitle("Yomuna").WithIcon("lucide:book-open")

	return &SidebarAppUI{
		th:           th,
		Sidebar:      sb,
		ToggleButton: toggleButton,
		Settings:     settings,
		Overlay:      overlayLayer,
	}
}
func (ui *SidebarAppUI) Layout(gtx layout.Context) layout.Dimensions {
	if ui == nil || ui.Sidebar == nil {
		return layout.Dimensions{}
	}

	return ui.Overlay.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.Sidebar.Layout(
			gtx,
			ui.layoutSidebar,
			func(gtx layout.Context) layout.Dimensions {
				if ui.Sidebar.Tabs == nil {
					return ui.layoutContent(gtx)
				}
				return panel.NewBackgroundPanel(theme.DefaultThemeClient).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return ui.Sidebar.Tabs.Layout(gtx)
				})
			},
		)
	})
}
func (ui *SidebarAppUI) layoutContent(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				title := material.H5(ui.th, "Main Content")
				return title.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				body := material.Body1(ui.th, "The content area automatically fills the remaining space beside the animated sidebar.")
				return body.Layout(gtx)
			}),
		)
	})
}
func (ui *SidebarAppUI) layoutSidebar(gtx layout.Context) layout.Dimensions {
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
