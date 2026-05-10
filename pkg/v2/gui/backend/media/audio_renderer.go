package media

import (
	"context"
	"fmt"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/player"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

type AudioRenderer struct {
	Player   player.LoadablePlayable
	Controls *Controls

	Theme *material.Theme
	tc    *theme.Client

	src player.Source
	err error
}

func NewAudioRenderer(player player.LoadablePlayable) *AudioRenderer {
	return &AudioRenderer{
		Player:   player,
		Controls: NewControls(),

		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,
	}
}

func (r *AudioRenderer) WithThemeClient(tc *theme.Client) *AudioRenderer {
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

func (r *AudioRenderer) WithMaterialTheme(th *material.Theme) *AudioRenderer {
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

func (r *AudioRenderer) Load(ctx context.Context, src player.Source) error {
	if r == nil {
		return nil
	}

	r.src = src

	if r.Player == nil {
		r.err = fmt.Errorf("no audio player backend configured")
		return r.err
	}

	r.err = r.Player.Load(src.Path)
	return r.err
}

func (r *AudioRenderer) Layout(gtx layout.Context) layout.Dimensions {
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

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				r.Theme,
				r.tc,
				theme.TextRoleH4,
				theme.ThemeColorTextPrimary,
				"Audio Preview",
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return r.Controls.Layout(gtx, r.Player)
		}),
	)
}

func (r *AudioRenderer) Close() error {
	if r == nil || r.Player == nil {
		return nil
	}

	if closer, ok := r.Player.(interface{ Close() error }); ok {
		return closer.Close()
	}

	return r.Player.Stop()
}
func (r *AudioRenderer) State() player.State {
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

func (r *AudioRenderer) Error() error {
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
