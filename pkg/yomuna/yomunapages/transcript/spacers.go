package transcript

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

func spacerH(height unit.Dp) layout.Widget {
	return layout.Spacer{Height: height}.Layout
}

func spacerW(width unit.Dp) layout.Widget {
	return layout.Spacer{Width: width}.Layout
}
