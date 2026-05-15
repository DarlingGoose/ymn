package media

import (
	"context"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media/player"
)

type View struct {
	Registry *Registry

	Source   player.Source
	Renderer player.Renderer
	ImageFit widget.Fit

	LastError error
}

func NewView(registry *Registry) *View {
	if registry == nil {
		registry = DefaultRegistry
	}

	return &View{
		Registry: registry,
	}
}

func (v *View) WithImageFit(fit widget.Fit) *View {
	if v == nil {
		return v
	}
	v.ImageFit = fit
	v.applyImageFit()
	return v
}

func (v *View) LoadPath(ctx context.Context, path string) error {
	return v.Load(ctx, player.NewSource(path))
}

func (v *View) Load(ctx context.Context, src player.Source) error {
	if v == nil {
		return nil
	}

	src = WithDetectedKind(src)

	if v.Source.Path != "" && v.Source.Path != src.Path {
		_ = v.Close()
	}

	renderer, err := LoadRenderer(ctx, v.Registry, src)
	if err != nil {
		v.LastError = err
		return err
	}

	v.Source = src
	v.Renderer = renderer
	v.applyImageFit()
	v.LastError = nil

	return nil
}

func (v *View) applyImageFit() {
	if v == nil || v.ImageFit == 0 {
		return
	}
	if renderer, ok := v.Renderer.(*ImageRenderer); ok {
		renderer.View.Fit = v.ImageFit
	}
}

func (v *View) Close() error {
	if v == nil {
		return nil
	}

	var err error
	if v.Renderer != nil {
		err = v.Renderer.Close()
	}

	v.Source = player.Source{}
	v.Renderer = nil
	v.LastError = err
	return err
}

func (v *View) State() player.State {
	if v == nil || v.Renderer == nil {
		return player.StateIdle
	}
	return v.Renderer.State()
}

func (v *View) Error() error {
	if v == nil {
		return nil
	}
	if v.LastError != nil {
		return v.LastError
	}
	if v.Renderer != nil {
		return v.Renderer.Error()
	}
	return nil
}

func (v *View) Layout(gtx layout.Context) layout.Dimensions {
	if v == nil {
		return layout.Dimensions{}
	}

	if err := v.Error(); err != nil {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return material.Body2(material.NewTheme(), err.Error()).Layout(gtx)
		})
	}

	if v.Renderer == nil {
		return layout.Dimensions{}
	}

	return v.Renderer.Layout(gtx)
}
