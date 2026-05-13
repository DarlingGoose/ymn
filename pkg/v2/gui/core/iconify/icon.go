package iconify

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

type SVGIcon struct {
	src      []byte
	op       paint.ImageOp
	imgSize  int
	imgColor color.NRGBA
}

func (ic *SVGIcon) Layout(
	gtx layout.Context,
	size unit.Dp,
	col color.NRGBA,
) layout.Dimensions {
	px := gtx.Dp(size)
	if px <= 0 {
		return layout.Dimensions{}
	}

	dims := image.Pt(px, px)
	defer clip.Rect{Max: dims}.Push(gtx.Ops).Pop()

	op := ic.image(px, col)
	op.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Dimensions{Size: dims}
}

// LayoutRotated renders the icon rotated around its center.
// angle is in radians.
func (ic *SVGIcon) LayoutRotated(
	gtx layout.Context,
	size unit.Dp,
	col color.NRGBA,
	angle float64,
) layout.Dimensions {
	px := gtx.Dp(size)
	if px <= 0 {
		return layout.Dimensions{}
	}

	dims := image.Pt(px, px)

	macro := op.Record(gtx.Ops)
	ic.Layout(gtx, size, col)
	call := macro.Stop()

	center := f32.Pt(
		float32(dims.X)/2,
		float32(dims.Y)/2,
	)

	stack := op.Affine(
		f32.Affine2D{}.
			Rotate(center, float32(angle)),
	).Push(gtx.Ops)

	call.Add(gtx.Ops)

	stack.Pop()

	return layout.Dimensions{Size: dims}
}

// LayoutRotatedDeg renders the icon rotated around its center.
// angleDegrees is in degrees.
func (ic *SVGIcon) LayoutRotatedDeg(
	gtx layout.Context,
	size unit.Dp,
	col color.NRGBA,
	angleDegrees float64,
) layout.Dimensions {
	return ic.LayoutRotated(gtx, size, col, degToRad(angleDegrees))
}

func degToRad(v float64) float64 {
	return v * math.Pi / 180
}

func (ic *SVGIcon) image(sz int, col color.NRGBA) paint.ImageOp {
	if sz == ic.imgSize && col == ic.imgColor {
		return ic.op
	}

	img := image.NewRGBA(image.Rect(0, 0, sz, sz))
	icon, err := oksvg.ReadIconStream(bytes.NewReader(replaceCurrentColor(ic.src, col)))
	if err != nil {
		ic.op = paint.NewImageOp(img)
		ic.imgSize = sz
		ic.imgColor = col
		return ic.op
	}

	icon.SetTarget(0, 0, float64(sz), float64(sz))
	scanner := rasterx.NewScannerGV(sz, sz, img, img.Bounds())
	dasher := rasterx.NewDasher(sz, sz, scanner)
	icon.Draw(dasher, 1)

	ic.op = paint.NewImageOp(img)
	ic.imgSize = sz
	ic.imgColor = col
	return ic.op
}

func replaceCurrentColor(src []byte, col color.NRGBA) []byte {
	hex := fmt.Sprintf("#%02x%02x%02x", col.R, col.G, col.B)
	replacer := strings.NewReplacer(
		"currentColor", hex,
		"currentcolor", hex,
	)
	return []byte(replacer.Replace(string(src)))
}
