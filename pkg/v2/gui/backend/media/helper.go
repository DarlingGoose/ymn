package media

import (
	"time"

	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

func unitSeconds(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

func applyButtonColors(btn *material.ButtonStyle, tc *theme.Client) {
	if btn == nil {
		return
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	tokens := tc.GetCurrentColorToken()

	btn.Background = tokens.PrimaryNRGBA()
	btn.Color = tokens.OnPrimaryNRGBA()
}
