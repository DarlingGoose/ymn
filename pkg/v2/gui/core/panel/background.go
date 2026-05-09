package panel

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

type BackgroundRole int

const (
	BackgroundRoleBackground BackgroundRole = iota
	BackgroundRoleSurface
	BackgroundRoleSurfaceAlt
	BackgroundRolePrimary
	BackgroundRoleSecondary
)

type BackgroundPanel struct {
	Theme  *theme.Client
	Role   BackgroundRole
	Radius unit.Dp
	Inset  layout.Inset

	FillMax bool
}

func NewBackgroundPanel(tc *theme.Client) *BackgroundPanel {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	return &BackgroundPanel{
		Theme:   tc,
		Role:    BackgroundRoleBackground,
		Radius:  unit.Dp(0),
		FillMax: true,
	}
}

func (p *BackgroundPanel) WithRole(role BackgroundRole) *BackgroundPanel {
	if p == nil {
		return p
	}

	p.Role = role
	return p
}

func (p *BackgroundPanel) WithRadius(radius unit.Dp) *BackgroundPanel {
	if p == nil {
		return p
	}

	p.Radius = radius
	return p
}

func (p *BackgroundPanel) WithInset(inset layout.Inset) *BackgroundPanel {
	if p == nil {
		return p
	}

	p.Inset = inset
	return p
}

func (p *BackgroundPanel) Layout(gtx layout.Context, w layout.Widget) layout.Dimensions {
	if p == nil {
		return w(gtx)
	}

	tc := p.Theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		p.Theme = tc
	}

	if tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	tokens := tc.GetCurrentColorToken()
	bg := panelColor(tokens, p.Role)

	macro := op.Record(gtx.Ops)

	contentDims := p.Inset.Layout(gtx, w)

	call := macro.Stop()

	size := contentDims.Size
	if p.FillMax {
		if gtx.Constraints.Max.X > 0 {
			size.X = gtx.Constraints.Max.X
		}
		if gtx.Constraints.Max.Y > 0 {
			size.Y = gtx.Constraints.Max.Y
		}
	}

	if size.X <= 0 || size.Y <= 0 {
		call.Add(gtx.Ops)
		return contentDims
	}

	rect := image.Rectangle{Max: size}

	paint.FillShape(
		gtx.Ops,
		bg,
		clip.UniformRRect(rect, gtx.Dp(p.Radius)).Op(gtx.Ops),
	)

	call.Add(gtx.Ops)

	if p.FillMax {
		return layout.Dimensions{Size: size}
	}

	return contentDims
}

func panelColor(tokens *theme.ColorTokens, role BackgroundRole) color.NRGBA {
	if tokens == nil {
		return color.NRGBA{A: 255}
	}

	switch role {
	case BackgroundRoleSurface:
		return tokens.SurfaceNRGBA()
	case BackgroundRoleSurfaceAlt:
		return tokens.SurfaceAltNRGBA()
	case BackgroundRolePrimary:
		return tokens.PrimaryNRGBA()
	case BackgroundRoleSecondary:
		return tokens.SecondaryNRGBA()
	case BackgroundRoleBackground:
		fallthrough
	default:
		return tokens.BackgroundNRGBA()
	}
}
