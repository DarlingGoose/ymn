package tooltip

import (
	"image"
	"image/color"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

type Placement int

const (
	Top Placement = iota
	Bottom
	Left
	Right
)

type Tooltip struct {
	Text string

	Theme       *material.Theme
	ThemeClient *theme.Client

	Placement Placement
	Delay     time.Duration
	Flip      *tween.Flip

	MaxWidth unit.Dp
	Padding  layout.Inset
	Gap      unit.Dp
	Radius   unit.Dp
	Nudge    unit.Dp

	Role theme.TextRole

	clickable    widget.Clickable
	hoverStarted time.Time
	open         bool
	anchorSize   image.Point
}

func New(text string) *Tooltip {
	return &Tooltip{
		Text:        text,
		Theme:       material.NewTheme(),
		ThemeClient: theme.DefaultThemeClient,
		Placement:   Top,
		Delay:       450 * time.Millisecond,
		Flip:        tween.NewFlip(120*time.Millisecond, tween.EaseOutCubic),
		MaxWidth:    unit.Dp(280),
		Padding: layout.Inset{
			Top:    unit.Dp(7),
			Bottom: unit.Dp(7),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		},
		Gap:    unit.Dp(8),
		Radius: unit.Dp(8),
		Nudge:  unit.Dp(5),
		Role:   theme.TextRoleCaption,
	}
}

func (t *Tooltip) WithThemeClient(tc *theme.Client) *Tooltip {
	if t == nil {
		return t
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	t.ThemeClient = tc
	return t
}

func (t *Tooltip) WithMaterialTheme(th *material.Theme) *Tooltip {
	if t == nil {
		return t
	}
	if th != nil {
		t.Theme = th
	}
	return t
}

func (t *Tooltip) WithPlacement(placement Placement) *Tooltip {
	if t == nil {
		return t
	}
	t.Placement = placement
	return t
}

func (t *Tooltip) WithDelay(delay time.Duration) *Tooltip {
	if t == nil {
		return t
	}
	t.Delay = delay
	return t
}

func (t *Tooltip) Layout(gtx layout.Context, layer *overlay.Overlay, target layout.Widget) layout.Dimensions {
	if t == nil {
		if target == nil {
			return layout.Dimensions{}
		}
		return target(gtx)
	}
	if target == nil {
		target = func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}
	}

	return t.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		dims := target(gtx)
		t.LayoutFor(gtx, layer, t.clickable.Hovered(), dims)
		return dims
	})
}

func (t *Tooltip) LayoutFor(
	gtx layout.Context,
	layer *overlay.Overlay,
	hovered bool,
	anchor layout.Dimensions,
) {
	if t == nil {
		return
	}

	t.ensureDefaults()
	t.anchorSize = anchor.Size
	t.update(gtx, hovered)

	progress, running := t.Flip.Value(time.Now())
	if running || hovered || t.open {
		gtx.Execute(op.InvalidateCmd{})
	}

	if progress <= 0 || strings.TrimSpace(t.Text) == "" {
		return
	}

	renderer := renderer{
		tooltip:  t,
		anchor:   anchor.Size,
		progress: progress,
	}

	if layer != nil {
		layer.Add(gtx, renderer)
		return
	}

	renderer.OverlayLayout(gtx)
}

func (t *Tooltip) Open(gtx layout.Context, layer *overlay.Overlay, anchor layout.Dimensions) {
	t.LayoutFor(gtx, layer, true, anchor)
}

func (t *Tooltip) Close() {
	if t == nil {
		return
	}
	t.open = false
	t.hoverStarted = time.Time{}
	if t.Flip != nil {
		t.Flip.Collapse()
	}
}

func (t *Tooltip) ensureDefaults() {
	if t.Theme == nil {
		t.Theme = material.NewTheme()
	}
	if t.ThemeClient == nil {
		t.ThemeClient = theme.DefaultThemeClient
	}
	if t.Flip == nil {
		t.Flip = tween.NewFlip(120*time.Millisecond, tween.EaseOutCubic)
	}
	if t.MaxWidth <= 0 {
		t.MaxWidth = unit.Dp(280)
	}
	if t.Radius <= 0 {
		t.Radius = unit.Dp(8)
	}
	if t.Gap <= 0 {
		t.Gap = unit.Dp(8)
	}
	if t.Nudge <= 0 {
		t.Nudge = unit.Dp(5)
	}
}

func (t *Tooltip) update(gtx layout.Context, hovered bool) {
	now := time.Now()

	if hovered {
		if t.hoverStarted.IsZero() {
			t.hoverStarted = now
		}

		if t.Delay <= 0 || now.Sub(t.hoverStarted) >= t.Delay {
			if !t.open {
				t.open = true
				t.Flip.Expand()
			}
			return
		}

		gtx.Execute(op.InvalidateCmd{})
		return
	}

	t.hoverStarted = time.Time{}
	if t.open {
		t.open = false
		t.Flip.Collapse()
	}
}

type renderer struct {
	tooltip  *Tooltip
	anchor   image.Point
	progress float64
}

func (r renderer) OverlayLayout(gtx layout.Context) {
	t := r.tooltip
	if t == nil || r.progress <= 0 {
		return
	}
	if t.ThemeClient.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	maxWidth := gtx.Dp(t.MaxWidth)
	if maxWidth <= 0 {
		maxWidth = 280
	}
	if gtx.Constraints.Max.X > 0 && maxWidth > gtx.Constraints.Max.X {
		maxWidth = gtx.Constraints.Max.X
	}

	cardGtx := gtx
	cardGtx.Constraints.Min.X = 0
	cardGtx.Constraints.Max.X = maxWidth

	macro := op.Record(gtx.Ops)
	dims := t.layoutBubble(cardGtx, r.progress)
	call := macro.Stop()

	offset := t.offset(gtx, r.anchor, dims.Size, r.progress)
	stack := op.Offset(offset).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()
}

func (t *Tooltip) layoutBubble(gtx layout.Context, progress float64) layout.Dimensions {
	tokens := t.ThemeClient.GetCurrentColorToken()
	bg := alpha(tokens.TextPrimaryNRGBA(), uint8(tween.MapInt(progress, 0, 238)))
	border := alpha(tokens.BorderNRGBA(), uint8(tween.MapInt(progress, 0, 210)))
	text := alpha(tokens.TextInverseNRGBA(), uint8(tween.MapInt(progress, 0, 255)))

	return utils.SurfaceOutlined(
		gtx,
		bg,
		t.Radius,
		utils.SurfaceBorder{Color: border, Width: unit.Dp(1)},
		func(gtx layout.Context) layout.Dimensions {
			return t.Padding.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(t.Theme, t.Text)
				theme.ApplyTypography(&lbl, t.ThemeClient.GetCurrentTypography(), t.Role)
				lbl.Color = text
				return lbl.Layout(gtx)
			})
		},
	)
}

func (t *Tooltip) offset(gtx layout.Context, anchor, bubble image.Point, progress float64) image.Point {
	gap := gtx.Dp(t.Gap)
	nudge := tween.MapInt(1-progress, 0, gtx.Dp(t.Nudge))

	switch t.Placement {
	case Bottom:
		return image.Pt((anchor.X-bubble.X)/2, anchor.Y+gap+nudge)
	case Left:
		return image.Pt(-bubble.X-gap-nudge, (anchor.Y-bubble.Y)/2)
	case Right:
		return image.Pt(anchor.X+gap+nudge, (anchor.Y-bubble.Y)/2)
	case Top:
		fallthrough
	default:
		return image.Pt((anchor.X-bubble.X)/2, -bubble.Y-gap-nudge)
	}
}

func alpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}
