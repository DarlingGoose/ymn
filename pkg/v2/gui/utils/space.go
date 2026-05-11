package utils

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
)

func Spacer(dp unit.Dp) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		size := gtx.Dp(dp)
		return layout.Dimensions{Size: image.Pt(size, size)}
	}
}

func SpacerH(dp unit.Dp) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(0, gtx.Dp(dp))}
	}
}

func SpacerW(dp unit.Dp) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(gtx.Dp(dp), 0)}
	}
}
