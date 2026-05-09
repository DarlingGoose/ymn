package panel

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type SlidePanel struct {
	Flip *tween.Flip
	BG   *tween.ColorTween

	Width  unit.Dp
	Inset  unit.Dp
	Radius unit.Dp

	ClosedBG color.NRGBA
	OpenBG   color.NRGBA
}

func NewSlidePanel() *SlidePanel {
	closedBG := color.NRGBA{R: 32, G: 34, B: 42, A: 255}
	openBG := color.NRGBA{R: 45, G: 64, B: 115, A: 255}

	return &SlidePanel{
		Flip: tween.NewFlip(
			280*time.Millisecond,
			tween.EaseOutCubic,
		),
		BG: tween.NewColorTween(
			220*time.Millisecond,
			tween.EaseOutCubic,
			closedBG,
		),

		Width:  unit.Dp(280),
		Inset:  unit.Dp(18),
		Radius: unit.Dp(18),

		ClosedBG: closedBG,
		OpenBG:   openBG,
	}
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

	if p.BG != nil {
		p.BG.AnimateTo(p.OpenBG)
	}
}

func (p *SlidePanel) Collapse() {
	if p == nil {
		return
	}

	if p.Flip != nil {
		p.Flip.Collapse()
	}

	if p.BG != nil {
		p.BG.AnimateTo(p.ClosedBG)
	}
}

func (p *SlidePanel) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p == nil {
		return layout.Dimensions{}
	}

	now := time.Now()

	progress := 0.0
	positionRunning := false
	if p.Flip != nil {
		progress, positionRunning = p.Flip.Value(now)
	}

	bg := p.ClosedBG
	colorRunning := false
	if p.BG != nil {
		bg, colorRunning = p.BG.Value(now)
	}

	if positionRunning || colorRunning {
		gtx.Execute(op.InvalidateCmd{})
	}

	panelWidth := gtx.Dp(p.Width)

	x := mapInt(progress, -panelWidth, 0)

	stack := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
	defer stack.Pop()

	gtx.Constraints.Min.X = panelWidth
	gtx.Constraints.Max.X = panelWidth

	return utils.Surface(gtx, bg, p.Radius, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(p.Inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.layoutContent(gtx, th)
		})
	})
}

func (p *SlidePanel) layoutContent(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := material.H6(th, "Side Panel")
			title.Color = color.NRGBA{R: 245, G: 247, B: 255, A: 255}
			return title.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			body := material.Body1(th, "This panel uses tween.Flip for movement and tween.ColorTween for color.")
			body.Color = color.NRGBA{R: 220, G: 225, B: 240, A: 255}
			return body.Layout(gtx)
		}),
	)
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
