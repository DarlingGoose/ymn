package media

import (
	"context"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/player"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

type VideoRenderer struct {
	Player   *InlineVideoPlayer
	Controls *Controls

	Theme *material.Theme
	tc    *theme.Client

	src player.Source
	err error
}

func NewVideoRenderer() *VideoRenderer {
	return &VideoRenderer{
		Player:   NewInlineVideoPlayer(),
		Controls: NewControls(),

		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,
	}
}

func (r *VideoRenderer) WithThemeClient(tc *theme.Client) *VideoRenderer {
	if r == nil {
		return r
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	r.tc = tc

	if r.Controls != nil {
		r.Controls.WithThemeClient(tc)
	}

	return r
}

func (r *VideoRenderer) WithMaterialTheme(th *material.Theme) *VideoRenderer {
	if r == nil {
		return r
	}
	if th != nil {
		r.Theme = th
	}

	if r.Controls != nil {
		r.Controls.WithMaterialTheme(th)
	}

	return r
}

func (r *VideoRenderer) Load(ctx context.Context, src player.Source) error {
	if r == nil {
		return nil
	}

	r.src = src

	if r.Player == nil {
		r.Player = NewInlineVideoPlayer()
	}

	r.err = r.Player.Load(src.Path)
	return r.err
}

func (r *VideoRenderer) Layout(gtx layout.Context) layout.Dimensions {
	if r == nil || r.Player == nil {
		return layout.Dimensions{}
	}

	if r.Theme == nil {
		r.Theme = material.NewTheme()
	}
	if r.tc == nil {
		r.tc = theme.DefaultThemeClient
	}
	if r.Controls == nil {
		r.Controls = NewControls().
			WithThemeClient(r.tc).
			WithMaterialTheme(r.Theme)
	}

	if r.tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	if err := r.Error(); err != nil {
		return r.layoutError(gtx, err.Error())
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return r.Player.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return r.Controls.Layout(gtx, r.Player)
		}),
	)
}

func (r *VideoRenderer) Close() error {
	if r == nil || r.Player == nil {
		return nil
	}

	r.src = player.Source{}
	r.err = nil

	return r.Player.Close()
}

func (r *VideoRenderer) State() player.State {
	if r == nil {
		return player.StateIdle
	}
	if r.err != nil {
		return player.StateError
	}
	if r.Player == nil {
		return player.StateIdle
	}
	return r.Player.State()
}

func (r *VideoRenderer) Error() error {
	if r == nil {
		return nil
	}
	if r.err != nil {
		return r.err
	}
	if r.Player != nil {
		return r.Player.Error()
	}
	return nil
}

func (r *VideoRenderer) layoutError(gtx layout.Context, msg string) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				r.Theme,
				r.tc,
				theme.TextRoleBody,
				theme.ThemeColorError,
				msg,
			)
		})
	})
}
