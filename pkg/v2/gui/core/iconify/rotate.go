package iconify

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func LayoutRotatedIcon(
	gtx layout.Context,
	ic *SVGIcon,
	size unit.Dp,
	col color.NRGBA,
	angle float32,
) layout.Dimensions {
	if ic == nil {
		return layout.Dimensions{}
	}

	px := gtx.Dp(size)
	if px <= 0 {
		return layout.Dimensions{}
	}

	iconGtx := gtx
	iconGtx.Constraints.Min.X = px
	iconGtx.Constraints.Max.X = px
	iconGtx.Constraints.Min.Y = px
	iconGtx.Constraints.Max.Y = px

	// Record the icon untransformed first.
	macro := op.Record(gtx.Ops)
	dims := ic.Layout(iconGtx, size, col)
	call := macro.Stop()

	// Force a stable square box so flex layout does not jitter while rotating.
	if dims.Size.X <= 0 {
		dims.Size.X = px
	}
	if dims.Size.Y <= 0 {
		dims.Size.Y = px
	}

	center := f32.Pt(
		float32(dims.Size.X)/2,
		float32(dims.Size.Y)/2,
	)

	stack := op.Affine(
		f32.Affine2D{}.Rotate(center, angle),
	).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()

	return layout.Dimensions{
		Size: image.Pt(px, px),
	}
}
