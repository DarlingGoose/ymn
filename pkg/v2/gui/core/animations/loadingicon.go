package animations

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
)

type LoadingIcon struct {
	Icon    *iconify.SVGIcon
	Spinner *tween.RotationTween
}

func NewLoadingIcon(icon *iconify.SVGIcon) *LoadingIcon {
	return &LoadingIcon{
		Icon: icon,
		Spinner: tween.NewRotationTweenDeg(
			200*time.Millisecond,
			tween.EaseOutCubic,
			0,
		),
	}
}

func (l *LoadingIcon) Start() {
	l.Spinner.StartContinuousDegreesPerSecond(360)
}

func (l *LoadingIcon) Stop() {
	l.Spinner.Stop(true) // true = reset to 0
}

func (l *LoadingIcon) Toggle() {
	l.Spinner.ToggleContinuousDegreesPerSecond(360, true)
}

func (l *LoadingIcon) Layout(gtx layout.Context, size unit.Dp, col color.NRGBA) layout.Dimensions {
	angle, running := l.Spinner.Value(time.Now())
	if running {
		gtx.Execute(op.InvalidateCmd{})
	}

	return l.Icon.LayoutRotated(gtx, size, col, angle)
}
