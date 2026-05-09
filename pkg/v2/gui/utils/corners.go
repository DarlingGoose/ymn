package utils

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

type CornerRadius struct {
	TopLeft     unit.Dp
	TopRight    unit.Dp
	BottomRight unit.Dp
	BottomLeft  unit.Dp
}

func UniformCorners(radius unit.Dp) CornerRadius {
	return CornerRadius{
		TopLeft:     radius,
		TopRight:    radius,
		BottomRight: radius,
		BottomLeft:  radius,
	}
}

func SurfaceCorners(
	gtx layout.Context,
	bg color.NRGBA,
	radius CornerRadius,
	w layout.Widget,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	if bg.A > 0 && dims.Size.X > 0 && dims.Size.Y > 0 {
		rect := image.Rectangle{Max: dims.Size}
		path := rrectCorners(gtx, rect, radius)
		paint.FillShape(gtx.Ops, bg, clip.Outline{Path: path.End()}.Op())
	}

	call.Add(gtx.Ops)

	return dims
}

func ClickableSurfaceCorners(
	gtx layout.Context,
	clickable *widget.Clickable,
	bg color.NRGBA,
	radius CornerRadius,
	w layout.Widget,
) layout.Dimensions {
	if clickable == nil {
		return SurfaceCorners(gtx, bg, radius, w)
	}

	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return SurfaceCorners(gtx, bg, radius, w)
	})
}

func rrectCorners(gtx layout.Context, rect image.Rectangle, radius CornerRadius) clip.Path {
	w := float32(rect.Dx())
	h := float32(rect.Dy())

	tl := clampRadius(float32(gtx.Dp(radius.TopLeft)), w, h)
	tr := clampRadius(float32(gtx.Dp(radius.TopRight)), w, h)
	br := clampRadius(float32(gtx.Dp(radius.BottomRight)), w, h)
	bl := clampRadius(float32(gtx.Dp(radius.BottomLeft)), w, h)

	var p clip.Path
	p.Begin(gtx.Ops)

	// Start after top-left corner.
	p.MoveTo(f32.Pt(tl, 0))

	// Top edge + top-right.
	p.LineTo(f32.Pt(w-tr, 0))
	if tr > 0 {
		p.QuadTo(f32.Pt(w, 0), f32.Pt(w, tr))
	}

	// Right edge + bottom-right.
	p.LineTo(f32.Pt(w, h-br))
	if br > 0 {
		p.QuadTo(f32.Pt(w, h), f32.Pt(w-br, h))
	}

	// Bottom edge + bottom-left.
	p.LineTo(f32.Pt(bl, h))
	if bl > 0 {
		p.QuadTo(f32.Pt(0, h), f32.Pt(0, h-bl))
	}

	// Left edge + top-left.
	p.LineTo(f32.Pt(0, tl))
	if tl > 0 {
		p.QuadTo(f32.Pt(0, 0), f32.Pt(tl, 0))
	}

	p.Close()
	return p
}

func clampRadius(radius, w, h float32) float32 {
	if radius < 0 {
		return 0
	}

	max := min(w, h) / 2
	if radius > max {
		return max
	}

	return radius
}
