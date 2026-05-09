package panel

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type SlidePanel struct {
	Flip *tween.Flip
	BG   *tween.ColorTween

	Width  unit.Dp
	Inset  unit.Dp
	Radius unit.Dp

	theme *theme.Client
}

func NewSlidePanel() *SlidePanel {
	tc := theme.DefaultThemeClient
	tokens := tc.GetCurrentColorToken()

	return &SlidePanel{
		theme: tc,

		Flip: tween.NewFlip(
			280*time.Millisecond,
			tween.EaseOutCubic,
		),
		BG: tween.NewColorTween(
			220*time.Millisecond,
			tween.EaseOutCubic,
			tokens.SurfaceNRGBA(),
		),

		Width:  unit.Dp(280),
		Inset:  unit.Dp(18),
		Radius: unit.Dp(18),
	}
}

func (p *SlidePanel) WithThemeClient(tc *theme.Client) *SlidePanel {
	if p == nil {
		return p
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	p.theme = tc

	tokens := tc.GetCurrentColorToken()
	if p.BG != nil {
		p.BG.JumpTo(tokens.SurfaceNRGBA())
	}

	return p
}

func (p *SlidePanel) Toggle() {
	if p == nil || p.Flip == nil {
		return
	}

	if p.Flip.Expanded() {
		p.Collapse()
		return
	}

	p.Expand()
}

func (p *SlidePanel) Expand() {
	if p == nil {
		return
	}

	if p.Flip != nil {
		p.Flip.Expand()
	}

	style := p.style()
	if p.BG != nil {
		p.BG.AnimateTo(style.OpenBG)
	}
}

func (p *SlidePanel) Collapse() {
	if p == nil {
		return
	}

	if p.Flip != nil {
		p.Flip.Collapse()
	}

	style := p.style()
	if p.BG != nil {
		p.BG.AnimateTo(style.ClosedBG)
	}
}

func (p *SlidePanel) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p == nil {
		return layout.Dimensions{}
	}

	now := time.Now()
	style := p.style()

	p.syncThemeTweens(now, style)

	progress := 0.0
	positionRunning := false
	if p.Flip != nil {
		progress, positionRunning = p.Flip.Value(now)
	}

	bg := style.ClosedBG
	colorRunning := false
	if p.BG != nil {
		bg, colorRunning = p.BG.Value(now)
	}

	if positionRunning || colorRunning {
		gtx.Execute(op.InvalidateCmd{})
	}

	if p.theme != nil && p.theme.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	panelWidth := gtx.Dp(p.Width)
	if panelWidth <= 0 {
		panelWidth = 280
	}

	x := mapInt(progress, -panelWidth, 0)

	stack := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
	defer stack.Pop()

	panelGtx := gtx
	panelGtx.Constraints.Min.X = panelWidth
	panelGtx.Constraints.Max.X = panelWidth

	return utils.SurfaceOutlined(
		panelGtx,
		bg,
		p.Radius,
		utils.SurfaceBorder{
			Color: style.Border,
			Width: unit.Dp(1),
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(p.Inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutContent(gtx, th, style)
			})
		},
	)
}

func (p *SlidePanel) syncThemeTweens(now time.Time, style slidePanelStyle) {
	if p == nil || p.BG == nil {
		return
	}

	target := style.ClosedBG
	if p.Flip != nil && p.Flip.Expanded() {
		target = style.OpenBG
	}

	p.BG.AnimateToAt(now, target)
}

func (p *SlidePanel) layoutContent(
	gtx layout.Context,
	th *material.Theme,
	style slidePanelStyle,
) layout.Dimensions {
	if th == nil {
		th = material.NewTheme()
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := material.H6(th, "Side Panel")
			title.Color = style.Title
			title.Alignment = text.Start

			theme.ApplyTypography(&title, style.Typo, theme.TextRoleH2)

			return title.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			body := material.Body1(th, "This panel uses tween.Flip for movement and theme color tweens for color.")
			body.Color = style.Body
			body.Alignment = text.Start

			theme.ApplyTypography(&body, style.Typo, theme.TextRoleBody)

			return body.Layout(gtx)
		}),
	)
}

type slidePanelStyle struct {
	Tokens *theme.ColorTokens
	Typo   theme.TypographyTokens

	ClosedBG color.NRGBA
	OpenBG   color.NRGBA
	Border   color.NRGBA

	Title color.NRGBA
	Body  color.NRGBA
}

func (p *SlidePanel) style() slidePanelStyle {
	tc := p.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		p.theme = tc
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	return slidePanelStyle{
		Tokens: tokens,
		Typo:   typo,

		ClosedBG: tokens.SurfaceNRGBA(),
		OpenBG:   tokens.SurfaceAltNRGBA(),
		Border:   tokens.BorderNRGBA(),

		Title: tokens.TextPrimaryNRGBA(),
		Body:  tokens.TextSecondaryNRGBA(),
	}
}

func mapInt(progress float64, from, to int) int {
	progress = clamp01(progress)
	return int(math.Round(float64(from) + float64(to-from)*progress))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}

	if v > 1 {
		return 1
	}

	return v
}
