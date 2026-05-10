package row

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type Row struct {
	Label       string
	Description string

	Theme *material.Theme
	tc    *theme.Client

	// Layout.
	MinHeight unit.Dp
	Padding   unit.Dp
	Gap       unit.Dp
	Radius    unit.Dp

	// Left side.
	LabelWidth unit.Dp

	// If true, the left label area stays fixed-width.
	// If false, label and control share space more naturally.
	FixedLabelWidth bool

	// Surface styling.
	ShowSurface bool
	ShowBorder  bool

	// Typography.
	LabelRole       theme.TextRole
	DescriptionRole theme.TextRole

	// Colors.
	LabelColor       theme.TextColorRole
	DescriptionColor theme.TextColorRole
}

func New(label, description string) *Row {
	return &Row{
		Label:       label,
		Description: description,

		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,

		MinHeight: unit.Dp(52),
		Padding:   unit.Dp(12),
		Gap:       unit.Dp(16),
		Radius:    unit.Dp(12),

		LabelWidth:       unit.Dp(180),
		FixedLabelWidth:  true,
		ShowSurface:      false,
		ShowBorder:       false,
		LabelRole:        theme.TextRoleLabel,
		DescriptionRole:  theme.TextRoleCaption,
		LabelColor:       theme.ThemeColorTextPrimary,
		DescriptionColor: theme.ThemeColorTextMuted,
	}
}

func (r *Row) WithThemeClient(tc *theme.Client) *Row {
	if r == nil {
		return r
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	r.tc = tc
	return r
}

func (r *Row) WithMaterialTheme(th *material.Theme) *Row {
	if r == nil {
		return r
	}
	if th != nil {
		r.Theme = th
	}
	return r
}

func (r *Row) WithLabelWidth(width unit.Dp) *Row {
	if r == nil {
		return r
	}
	r.LabelWidth = width
	r.FixedLabelWidth = width > 0
	return r
}

func (r *Row) WithSurface(enabled bool) *Row {
	if r == nil {
		return r
	}
	r.ShowSurface = enabled
	return r
}

func (r *Row) WithBorder(enabled bool) *Row {
	if r == nil {
		return r
	}
	r.ShowBorder = enabled
	return r
}

func (r *Row) Layout(gtx layout.Context, control layout.Widget) layout.Dimensions {
	if r == nil {
		if control == nil {
			return layout.Dimensions{}
		}
		return control(gtx)
	}
	if r.Theme == nil {
		r.Theme = material.NewTheme()
	}
	if r.tc == nil {
		r.tc = theme.DefaultThemeClient
	}
	if r.tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}
	if control == nil {
		control = func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}
	}

	body := func(gtx layout.Context) layout.Dimensions {
		minHeight := gtx.Dp(r.MinHeight)
		if minHeight > 0 {
			gtx.Constraints.Min.Y = minHeight
		}

		return layout.UniformInset(r.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return r.layoutLabel(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: r.Gap}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.E.Layout(gtx, control)
				}),
			)
		})
	}

	if !r.ShowSurface && !r.ShowBorder {
		return body(gtx)
	}

	tokens := r.tc.GetCurrentColorToken()

	border := utils.SurfaceBorder{}
	if r.ShowBorder {
		border = utils.SurfaceBorder{
			Color: tokens.BorderNRGBA(),
			Width: unit.Dp(1),
		}
	}

	return utils.SurfaceOutlined(
		gtx,
		tokens.SurfaceNRGBA(),
		r.Radius,
		border,
		body,
	)
}

func (r *Row) layoutLabel(gtx layout.Context) layout.Dimensions {
	if r.FixedLabelWidth && r.LabelWidth > 0 {
		w := gtx.Dp(r.LabelWidth)
		gtx.Constraints.Min.X = w
		gtx.Constraints.Max.X = w
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				r.Theme,
				r.tc,
				r.LabelRole,
				r.LabelColor,
				r.Label,
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if r.Description == "" {
				return layout.Dimensions{}
			}

			return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(
					gtx,
					r.Theme,
					r.tc,
					r.DescriptionRole,
					r.DescriptionColor,
					r.Description,
				)
			})
		}),
	)
}

func rgbaWithAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}
