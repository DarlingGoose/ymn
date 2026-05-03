package components

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
)

type PillKind string

const (
	PillNeutral PillKind = "neutral"
	PillLive    PillKind = "live"
	PillWarning PillKind = "warning"
	PillError   PillKind = "error"
)

func PagePanel(gtx layout.Context, theme barethemes.Theme, child layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, theme.Color.Surface, unit.Dp(theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, child)
	})
}

func Section(gtx layout.Context, theme barethemes.Theme, child layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, theme.Color.Background, unit.Dp(theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, child)
	})
}

func QuietSection(gtx layout.Context, theme barethemes.Theme, child layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, theme.Color.SurfaceAlt, unit.Dp(theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, child)
	})
}

func SectionHeader(gtx layout.Context, theme barethemes.Theme, title, subtitle string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(theme.Gio(), title)
			lbl.Color = theme.Color.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if subtitle == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(theme.Gio(), subtitle)
				lbl.Color = theme.Color.TextMuted
				return lbl.Layout(gtx)
			})
		}),
	)
}

func StatusPill(gtx layout.Context, theme barethemes.Theme, text string, kind PillKind) layout.Dimensions {
	bg, fg := pillColors(theme, kind)
	return bareutils.RoundedSurface(gtx, bg, unit.Dp(theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(7),
			Bottom: unit.Dp(7),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(theme.Gio(), text)
			lbl.Color = fg
			return lbl.Layout(gtx)
		})
	})
}

func Chip(gtx layout.Context, theme barethemes.Theme, text string) layout.Dimensions {
	return bareutils.RoundedSurface(gtx, theme.Color.Surface, unit.Dp(theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(theme.Gio(), text)
			lbl.Color = theme.Color.Text
			return lbl.Layout(gtx)
		})
	})
}

func MutedBody(gtx layout.Context, theme barethemes.Theme, text string) layout.Dimensions {
	lbl := material.Body1(theme.Gio(), text)
	lbl.Color = theme.Color.TextMuted
	return lbl.Layout(gtx)
}

func pillColors(theme barethemes.Theme, kind PillKind) (color.NRGBA, color.NRGBA) {
	switch kind {
	case PillLive:
		return alpha(theme.Color.Success, 42), theme.Color.Success
	case PillWarning:
		return alpha(theme.Color.Warning, 42), theme.Color.Warning
	case PillError:
		return alpha(theme.Color.Error, 42), theme.Color.Error
	default:
		return theme.Color.SurfaceAlt, theme.Color.TextMuted
	}
}

func alpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}
