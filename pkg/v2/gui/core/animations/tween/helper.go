package tween

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
)

func Clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func AnimateValue(from, to, progress float64, c Curve) float64 {
	return Lerp(from, to, c.At(progress))
}

func lerpInt64(from, to int64, t float64) int64 {
	v := float64(from) + float64(to-from)*t
	return int64(math.Round(v))
}

func lerpFloat(from, to, t float64) float64 {
	return from + (to-from)*t
}

func Color(progress float64, from, to color.NRGBA) color.NRGBA {
	progress = Clamp01(progress)

	return color.NRGBA{
		R: uint8(math.Round(float64(from.R) + float64(int(to.R)-int(from.R))*progress)),
		G: uint8(math.Round(float64(from.G) + float64(int(to.G)-int(from.G))*progress)),
		B: uint8(math.Round(float64(from.B) + float64(int(to.B)-int(from.B))*progress)),
		A: uint8(math.Round(float64(from.A) + float64(int(to.A)-int(from.A))*progress)),
	}
}

func ColorRGBA(progress float64, from, to color.RGBA) color.RGBA {
	progress = Clamp01(progress)

	return color.RGBA{
		R: uint8(math.Round(float64(from.R) + float64(int(to.R)-int(from.R))*progress)),
		G: uint8(math.Round(float64(from.G) + float64(int(to.G)-int(from.G))*progress)),
		B: uint8(math.Round(float64(from.B) + float64(int(to.B)-int(from.B))*progress)),
		A: uint8(math.Round(float64(from.A) + float64(int(to.A)-int(from.A))*progress)),
	}
}

func ColorFloat(progress float64, from, to float32) float32 {
	progress = Clamp01(progress)
	return from + (to-from)*float32(progress)
}

func MapInt(progress float64, from, to int) int {
	progress = Clamp01(progress)
	return int(math.Round(float64(from) + float64(to-from)*progress))
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func RotateAroundCenter(
	gtx layout.Context,
	angleRadians float64,
	w layout.Widget,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	center := f32.Pt(
		float32(dims.Size.X)/2,
		float32(dims.Size.Y)/2,
	)

	stack := op.Affine(
		f32.Affine2D{}.
			Rotate(center, float32(angleRadians)),
	).Push(gtx.Ops)

	call.Add(gtx.Ops)

	stack.Pop()

	return dims
}

func RotateAround(
	gtx layout.Context,
	origin image.Point,
	angleRadians float64,
	w layout.Widget,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()

	stack := op.Affine(
		f32.Affine2D{}.
			Rotate(
				f32.Pt(float32(origin.X), float32(origin.Y)),
				float32(angleRadians),
			),
	).Push(gtx.Ops)

	call.Add(gtx.Ops)

	stack.Pop()

	return dims
}
