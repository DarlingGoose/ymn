package pages

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

type SettingsUI struct {
	ModeToggle    *toggles.ThemeModeToggle
	ThemeDropdown *dropdowns.ThemeDropdown

	theme *theme.Client
}

func NewSettingsUI(tc *theme.Client) *SettingsUI {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	return &SettingsUI{
		theme:         tc,
		ModeToggle:    toggles.NewThemeModeToggle(tc),
		ThemeDropdown: dropdowns.NewThemeDropdown(tc),
	}
}

func (ui *SettingsUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}

	if ui.ModeToggle.Update(gtx) {
		gtx.Execute(op.InvalidateCmd{})
	}

	return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(ui.ModeToggle.Layout),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.ThemeDropdown.Layout(gtx, layer)
			}),
		)
	})
}
