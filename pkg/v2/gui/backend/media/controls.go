package media

import (
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media/player"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type Controls struct {
	PlayPause widget.Clickable
	Stop      widget.Clickable
	Seek      widget.Float
	Volume    widget.Float

	Theme *material.Theme
	tc    *theme.Client
}

func NewControls() *Controls {
	return &Controls{
		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,
		Volume: widget.Float{
			Value: 0.8,
		},
	}
}

func (c *Controls) WithThemeClient(tc *theme.Client) *Controls {
	if c == nil {
		return c
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	c.tc = tc
	return c
}

func (c *Controls) WithMaterialTheme(th *material.Theme) *Controls {
	if c == nil {
		return c
	}
	if th != nil {
		c.Theme = th
	}
	return c
}

func (c *Controls) Layout(gtx layout.Context, p player.Playable) layout.Dimensions {
	if c == nil || p == nil {
		return layout.Dimensions{}
	}

	if c.Theme == nil {
		c.Theme = material.NewTheme()
	}
	if c.tc == nil {
		c.tc = theme.DefaultThemeClient
	}
	if c.tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	for c.PlayPause.Clicked(gtx) {
		switch p.State() {
		case player.StatePlaying:
			_ = p.Pause()
		default:
			_ = p.Play()
		}
	}

	for c.Stop.Clicked(gtx) {
		_ = p.Stop()
	}

	if c.Volume.Dragging() {
		_ = p.SetVolume(c.Volume.Value)
	}

	// Optional: only seek when the value changes.
	if c.Seek.Dragging() {
		duration := p.Duration()
		if duration > 0 {
			pos := duration.Seconds() * float64(c.Seek.Value)
			_ = p.Seek(unitSeconds(pos))
		}
	}

	playText := "Play"
	if p.State() == player.StatePlaying {
		playText = "Pause"
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(c.Theme, &c.PlayPause, playText)
			applyButtonColors(&btn, c.tc)
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := material.Button(c.Theme, &c.Stop, "Stop")
			applyButtonColors(&btn, c.tc)
			return btn.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return material.Slider(c.Theme, &c.Seek).Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				c.Theme,
				c.tc,
				theme.TextRoleCaption,
				theme.ThemeColorTextMuted,
				"Volume",
			)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(96))
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(96))
			return material.Slider(c.Theme, &c.Volume).Layout(gtx)
		}),
	)
}
