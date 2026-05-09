package examples

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/panel"
)

type SliderAppUI struct {
	th *material.Theme

	ToggleButton widget.Clickable
	Panel        *panel.SlidePanel
	Dropdown     *components.Dropdown
}

func NewSliderAppUI(th *material.Theme) *SliderAppUI {
	return &SliderAppUI{
		th:    th,
		Panel: panel.NewSlidePanel(),
		Dropdown: components.NewDropdown([]components.DropdownItem{
			{Label: "Linear", Value: "linear"},
			{Label: "Ease In", Value: "ease-in"},
			{Label: "Ease Out", Value: "ease-out"},
			{Label: "Ease In Out", Value: "ease-in-out"},
		}),
	}
}

func (ui *SliderAppUI) Layout(gtx layout.Context) layout.Dimensions {
	for ui.ToggleButton.Clicked(gtx) {
		ui.Panel.Toggle()

		// Newer Gio invalidation API.
		gtx.Execute(op.InvalidateCmd{})
	}

	return layout.Stack{}.Layout(gtx,
		// Main content.
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutMain(gtx)
		}),

		// Sliding panel overlay.
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return ui.Panel.Layout(gtx, ui.th)
		}),
	)
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
