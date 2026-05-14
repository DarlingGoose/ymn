package dropdowns

import (
	"image/color"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type dropdownStyle struct {
	Tokens *theme.ColorTokens
	Typo   theme.TypographyTokens

	ButtonBG       color.NRGBA
	ButtonOpenBG   color.NRGBA
	MenuBG         color.NRGBA
	ItemHoverBG    color.NRGBA
	ItemSelectedBG color.NRGBA
	Outline        color.NRGBA

	Text  color.NRGBA
	Muted color.NRGBA
}

func (d *Dropdown) style() dropdownStyle {
	tc := d.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		d.theme = tc
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	return dropdownStyle{
		Tokens: tokens,
		Typo:   typo,

		ButtonBG:       tokens.SurfaceNRGBA(),
		ButtonOpenBG:   tokens.SurfaceAltNRGBA(),
		MenuBG:         tokens.SurfaceNRGBA(),
		ItemHoverBG:    tokens.SurfaceAltNRGBA(),
		ItemSelectedBG: tokens.SelectionNRGBA(),
		Outline:        tokens.PrimaryNRGBA(),

		Text:  tokens.TextPrimaryNRGBA(),
		Muted: tokens.TextMutedNRGBA(),
	}
}
