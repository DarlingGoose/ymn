package yomunapages

import (
	"context"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/layouts/split"
)

type TranscriptUI struct {
	th    *material.Theme
	theme *theme.Client

	bodySplit split.SplitH
}

func NewTranscriptUI(th *material.Theme, tc *theme.Client) *TranscriptUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	ui := &TranscriptUI{
		th:    th,
		theme: tc,
		bodySplit: split.SplitH{
			Ratio:    0,
			Bar:      unit.Dp(4),
			MinRatio: -0.75,
			MaxRatio: 0.75,
		},
	}

	return ui
}

func (ui *TranscriptUI) Layout(gtx layout.Context, ctx context.Context) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return layout.Dimensions{}
}
