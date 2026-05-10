package input

import "image/color"

type inputStyle struct {
	BG        color.NRGBA
	Border    color.NRGBA
	Text      color.NRGBA
	Muted     color.NRGBA
	Icon      color.NRGBA
	Danger    color.NRGBA
	Selection color.NRGBA
}
