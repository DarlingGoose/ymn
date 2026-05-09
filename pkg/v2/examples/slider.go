package examples

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/pages"
)

type SliderAppUI struct {
	th           *material.Theme
	Background   *panel.BackgroundPanel
	ToggleButton widget.Clickable
	settings     *pages.SettingsUI
	Panel        *panel.SlidePanel
	Dropdown     *dropdowns.Dropdown
	overlay      *overlay.Overlay
}

func NewSliderAppUI(th *material.Theme) *SliderAppUI {
	return &SliderAppUI{
		th:      th,
		overlay: &overlay.Overlay{},
		Background: panel.NewBackgroundPanel(theme.DefaultThemeClient).
			WithRole(panel.BackgroundRoleBackground),
		settings: pages.NewSettingsUI(theme.DefaultThemeClient),
		Panel:    panel.NewSlidePanel(),
		Dropdown: dropdowns.NewDropdown([]dropdowns.DropdownItem{
			{Label: "Linear", Value: "linear"},
			{Label: "Ease In", Value: "ease-in"},
			{Label: "Ease Out", Value: "ease-out"},
			{Label: "Ease In Out", Value: "ease-in-out"},
		}),
	}
}

func (ui *SliderAppUI) Layout(gtx layout.Context) layout.Dimensions {
	if theme.DefaultThemeClient.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	for ui.ToggleButton.Clicked(gtx) {
		ui.Panel.Toggle()
		gtx.Execute(op.InvalidateCmd{})
	}

	return ui.Background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.overlay.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Stack{}.Layout(gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutMain(gtx)
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					return ui.Panel.Layout(gtx, ui.th)
				}),
			)
		})
	})
}

func (ui *SliderAppUI) layoutMain(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(unit.Dp(24)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Button(ui.th, &ui.ToggleButton, "Toggle Panel").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(ui.th, "Click the button to slide the side panel in and out.")
				return lbl.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.settings.Layout(gtx, ui.overlay)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.Dropdown.Layout(gtx, ui.overlay)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				item, ok := ui.Dropdown.SelectedItem()
				text := "Selected: none"
				if ok {
					text = "Selected: " + item.Label + " / " + item.Value
				}

				lbl := material.Body1(ui.th, text)
				return lbl.Layout(gtx)
			}),
		)
	})
}
