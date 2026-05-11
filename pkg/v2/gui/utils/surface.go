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

type SurfaceBorder struct {
	Color color.NRGBA
	Width unit.Dp
}

func Surface(
	gtx layout.Context,
	bg color.NRGBA,
	radius unit.Dp,
	w layout.Widget,
) layout.Dimensions {
	return SurfaceOutlined(gtx, bg, radius, SurfaceBorder{}, w)
}

func SurfaceOutlined(
	gtx layout.Context,
	bg color.NRGBA,
	radius unit.Dp,
	border SurfaceBorder,
	w layout.Widget,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	rect := image.Rectangle{Max: dims.Size}
	r := gtx.Dp(radius)

	if bg.A > 0 {
		paint.FillShape(
			gtx.Ops,
			bg,
			clip.UniformRRect(rect, r).Op(gtx.Ops),
		)
	}

	drawBorder(gtx, rect, r, border)

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
	return ClickableSurfaceOutlined(gtx, clickable, bg, radius, SurfaceBorder{}, w)
}

func ClickableSurfaceOutlined(
	gtx layout.Context,
	clickable *widget.Clickable,
	bg color.NRGBA,
	radius unit.Dp,
	border SurfaceBorder,
	w layout.Widget,
) layout.Dimensions {
	if clickable == nil {
		return SurfaceOutlined(gtx, bg, radius, border, w)
	}

	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return SurfaceOutlined(gtx, bg, radius, border, w)
	})
}

func drawBorder(
	gtx layout.Context,
	rect image.Rectangle,
	radiusPx int,
	border SurfaceBorder,
) {
	if border.Color.A == 0 || border.Width <= 0 {
		return
	}

	widthPx := gtx.Dp(border.Width)
	if widthPx <= 0 {
		widthPx = 1
	}

	// Inset by half the stroke width so the border is drawn inside the surface bounds
	// instead of being clipped at the edges.
	inset := widthPx / 2
	if inset < 1 {
		inset = 1
	}

	borderRect := rect.Inset(inset)
	if borderRect.Empty() {
		return
	}

	borderRadius := radiusPx - inset
	if borderRadius < 0 {
		borderRadius = 0
	}

	rr := clip.UniformRRect(borderRect, borderRadius)

	paint.FillShape(
		gtx.Ops,
		border.Color,
		clip.Stroke{
			Path:  rr.Path(gtx.Ops),
			Width: float32(widthPx),
		}.Op(),
	)
}

func PxToDp(gtx layout.Context, px int) unit.Dp {
	if gtx.Metric.PxPerDp <= 0 {
		return unit.Dp(px)
	}

	return unit.Dp(float32(px) / gtx.Metric.PxPerDp)
}

func RoundedSurface(
	gtx layout.Context,
	radius unit.Dp,
	fill color.NRGBA,
	w layout.Widget,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	rect := image.Rectangle{
		Max: dims.Size,
	}

	rr := clip.RRect{
		Rect: rect,
		NE:   gtx.Dp(radius),
		NW:   gtx.Dp(radius),
		SE:   gtx.Dp(radius),
		SW:   gtx.Dp(radius),
	}

	clipStack := rr.Push(gtx.Ops)
	paint.Fill(gtx.Ops, fill)
	clipStack.Pop()

	call.Add(gtx.Ops)

	return dims
}

//func LayoutStatusPill(gtx layout.Context, th *material.Theme, ct *theme.ColorTokens, text string, live bool) layout.Dimensions {
//	bg := ct.SurfaceAltNRGBA()
//	fg := ct.TextMutedNRGBA()
//
//	if live {
//		fg = ct.OnPrimaryNRGBA()
//		bg = ct.PrimaryNRGBA()
//	}
//
//	return RoundedSurfaceWrap(
//		gtx,
//		bg,
//		unit.Dp(20),
//		func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{
//				Top:    unit.Dp(7),
//				Bottom: unit.Dp(7),
//				Left:   unit.Dp(10),
//				Right:  unit.Dp(10),
//			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body2(th, text)
//				lbl.Color = fg
//				return lbl.Layout(gtx)
//			})
//		},
//	)
//}

func RoundedSurfaceWrap(
	gtx layout.Context,
	bg color.NRGBA,
	radius unit.Dp,
	w layout.Widget,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)

	dims := w(gtx)

	call := macro.Stop()

	rr := clip.RRect{
		Rect: image.Rectangle{
			Max: dims.Size,
		},
		NE: int(gtx.Dp(radius)),
		NW: int(gtx.Dp(radius)),
		SE: int(gtx.Dp(radius)),
		SW: int(gtx.Dp(radius)),
	}

	paint.FillShape(gtx.Ops, bg, rr.Op(gtx.Ops))
	call.Add(gtx.Ops)

	return dims
}
