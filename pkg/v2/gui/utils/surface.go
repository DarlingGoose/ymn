package utils

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

func Surface(gtx layout.Context, bg color.NRGBA, radius unit.Dp, w layout.Widget) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	rect := image.Rectangle{Max: dims.Size}

	paint.FillShape(
		gtx.Ops,
		bg,
		clip.UniformRRect(rect, gtx.Dp(radius)).Op(gtx.Ops),
	)

	call.Add(gtx.Ops)

	return dims
}

func ClickableSurface(
	gtx layout.Context,
	clickable *widget.Clickable,
	bg color.NRGBA,
	radius unit.Dp,
	w layout.Widget,
) layout.Dimensions {
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return Surface(gtx, bg, radius, w)
	})
}
