package media

import (
	"context"
	"fmt"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media/player"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type UnsupportedRenderer struct {
	Message string

	Theme *material.Theme
	tc    *theme.Client

	src player.Source
	err error
}

func NewUnsupportedRenderer(message string) *UnsupportedRenderer {
	if message == "" {
		message = "Preview not supported"
	}

	return &UnsupportedRenderer{
		Message: message,

		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,
	}
}

func (r *UnsupportedRenderer) WithThemeClient(tc *theme.Client) *UnsupportedRenderer {
	if r == nil {
		return r
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	r.tc = tc
	return r
}

func (r *UnsupportedRenderer) WithMaterialTheme(th *material.Theme) *UnsupportedRenderer {
	if r == nil {
		return r
	}
	if th != nil {
		r.Theme = th
	}
	return r
}

func (r *UnsupportedRenderer) Load(ctx context.Context, src player.Source) error {
	if r == nil {
		return nil
	}

	r.src = src
	r.err = fmt.Errorf("%s: %s", r.Message, src.Name)

	return nil
}

func (r *UnsupportedRenderer) Layout(gtx layout.Context) layout.Dimensions {
	if r == nil {
		return layout.Dimensions{}
	}

	if r.Theme == nil {
		r.Theme = material.NewTheme()
	}
	if r.tc == nil {
		r.tc = theme.DefaultThemeClient
	}
	if r.tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	msg := r.Message
	if r.src.Name != "" {
		msg += ": " + r.src.Name
	}

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				r.Theme,
				r.tc,
				theme.TextRoleBody,
				theme.ThemeColorTextMuted,
				msg,
			)
		})
	})
}

func (r *UnsupportedRenderer) Close() error {
	if r == nil {
		return nil
	}

	r.src = player.Source{}
	r.err = nil

	return nil
}

func (r *UnsupportedRenderer) State() player.State {
	if r == nil {
		return player.StateIdle
	}
	return player.StateReady
}

func (r *UnsupportedRenderer) Error() error {
	// UnsupportedRenderer is a graceful fallback, so it renders a message
	// instead of surfacing an error to MediaView.
	return nil
}
