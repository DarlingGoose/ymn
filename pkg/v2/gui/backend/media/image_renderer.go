package media

import (
	"context"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media/player"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type ImageRenderer struct {
	View ImageView

	Theme *material.Theme
	tc    *theme.Client

	src player.Source
	err error
}

func NewImageRenderer() *ImageRenderer {
	return &ImageRenderer{
		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,
	}
}

func (r *ImageRenderer) WithThemeClient(tc *theme.Client) *ImageRenderer {
	if r == nil {
		return r
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	r.tc = tc
	return r
}

func (r *ImageRenderer) WithMaterialTheme(th *material.Theme) *ImageRenderer {
	if r == nil {
		return r
	}
	if th != nil {
		r.Theme = th
	}
	return r
}

func (r *ImageRenderer) Load(ctx context.Context, src player.Source) error {
	if r == nil {
		return nil
	}

	r.src = src
	r.err = r.View.Load(src.Path)
	return r.err
}

func (r *ImageRenderer) Layout(gtx layout.Context) layout.Dimensions {
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

	if r.View.Loading() {
		return r.layoutLoading(gtx)
	}

	if err := r.Error(); err != nil {
		return r.layoutError(gtx, err.Error())
	}

	return r.View.Draw(gtx)
}

func (r *ImageRenderer) Close() error {
	if r == nil {
		return nil
	}

	r.src = player.Source{}
	r.err = nil
	return nil
}

func (r *ImageRenderer) State() player.State {
	if r == nil {
		return player.StateIdle
	}

	if r.View.Loading() {
		return player.StateLoading
	}

	if r.Error() != nil {
		return player.StateError
	}

	if r.src.Path != "" {
		return player.StateReady
	}

	return player.StateIdle
}

func (r *ImageRenderer) Error() error {
	if r == nil {
		return nil
	}
	if r.err != nil {
		return r.err
	}
	return r.View.Err()
}

func (r *ImageRenderer) layoutLoading(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(16),
			Bottom: unit.Dp(16),
			Left:   unit.Dp(16),
			Right:  unit.Dp(16),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				r.Theme,
				r.tc,
				theme.TextRoleBody,
				theme.ThemeColorTextMuted,
				"Loading image preview...",
			)
		})
	})
}

func (r *ImageRenderer) layoutError(gtx layout.Context, msg string) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(16),
			Bottom: unit.Dp(16),
			Left:   unit.Dp(16),
			Right:  unit.Dp(16),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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
