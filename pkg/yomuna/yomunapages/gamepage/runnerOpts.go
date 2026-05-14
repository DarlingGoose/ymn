package gamepage

import (
	"reflect"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/input"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type runnerOptionField struct {
	name        string
	label       string
	description string
	kind        reflect.Kind
	isSlice     bool
	input       *input.TextInput
	toggle      *toggles.Toggle
	dropdown    *dropdowns.Dropdown
	onChange    func()
}

func (f *runnerOptionField) Layout(gtx layout.Context, th *material.Theme, tc *theme.Client, layer *overlay.Overlay) layout.Dimensions {
	if f == nil {
		return layout.Dimensions{}
	}
	if f.toggle != nil {
		if f.toggle.Update(gtx) && f.onChange != nil {
			f.onChange()
		}
		children := []layout.FlexChild{
			layout.Rigid(f.toggle.Layout),
		}
		if strings.TrimSpace(f.description) != "" {
			children = append(children,
				layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, th, tc, theme.TextRoleCaption, theme.ThemeColorTextMuted, f.description)
				}),
			)
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	}
	control := func(gtx layout.Context) layout.Dimensions {
		if f.input != nil {
			return f.input.Layout(gtx)
		}
		if f.dropdown != nil {
			return f.dropdown.Layout(gtx, layer)
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
