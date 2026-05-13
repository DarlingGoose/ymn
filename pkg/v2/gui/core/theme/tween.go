package theme

import (
	"fmt"
	"image/color"
	"time"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

type ColorTokenTweens struct {
	Background *tween.ColorTween
	Surface    *tween.ColorTween
	SurfaceAlt *tween.ColorTween
	Border     *tween.ColorTween
	Divider    *tween.ColorTween

	TextPrimary   *tween.ColorTween
	TextSecondary *tween.ColorTween
	TextMuted     *tween.ColorTween
	TextInverse   *tween.ColorTween

	Primary      *tween.ColorTween
	PrimaryHover *tween.ColorTween
	OnPrimary    *tween.ColorTween

	Secondary      *tween.ColorTween
	SecondaryHover *tween.ColorTween
	OnSecondary    *tween.ColorTween

	Success *tween.ColorTween
	Warning *tween.ColorTween
	Danger  *tween.ColorTween
	Info    *tween.ColorTween

	FocusRing *tween.ColorTween
	Selection *tween.ColorTween
	Disabled  *tween.ColorTween
}

func NewColorTokenTweens(duration time.Duration, curve tween.Curve, initial ColorTokens) *ColorTokenTweens {
	newTween := func(hex string) *tween.ColorTween {
		return tween.NewColorTween(duration, curve, utils.HexNRGBA(hex))
	}

	return &ColorTokenTweens{
		Background: newTween(initial.Background),
		Surface:    newTween(initial.Surface),
		SurfaceAlt: newTween(initial.SurfaceAlt),
		Border:     newTween(initial.Border),
		Divider:    newTween(initial.Divider),

		TextPrimary:   newTween(initial.TextPrimary),
		TextSecondary: newTween(initial.TextSecondary),
		TextMuted:     newTween(initial.TextMuted),
		TextInverse:   newTween(initial.TextInverse),

		Primary:      newTween(initial.Primary),
		PrimaryHover: newTween(initial.PrimaryHover),
		OnPrimary:    newTween(initial.OnPrimary),

		Secondary:      newTween(initial.Secondary),
		SecondaryHover: newTween(initial.SecondaryHover),
		OnSecondary:    newTween(initial.OnSecondary),

		Success: newTween(initial.Success),
		Warning: newTween(initial.Warning),
		Danger:  newTween(initial.Danger),
		Info:    newTween(initial.Info),

		FocusRing: newTween(initial.FocusRing),
		Selection: newTween(initial.Selection),
		Disabled:  newTween(initial.Disabled),
	}
}

func (t *ColorTokenTweens) AnimateToAt(now time.Time, next ColorTokens) {
	if t == nil {
		return
	}

	t.Background.AnimateToAt(now, utils.HexNRGBA(next.Background))
	t.Surface.AnimateToAt(now, utils.HexNRGBA(next.Surface))
	t.SurfaceAlt.AnimateToAt(now, utils.HexNRGBA(next.SurfaceAlt))
	t.Border.AnimateToAt(now, utils.HexNRGBA(next.Border))
	t.Divider.AnimateToAt(now, utils.HexNRGBA(next.Divider))

	t.TextPrimary.AnimateToAt(now, utils.HexNRGBA(next.TextPrimary))
	t.TextSecondary.AnimateToAt(now, utils.HexNRGBA(next.TextSecondary))
	t.TextMuted.AnimateToAt(now, utils.HexNRGBA(next.TextMuted))
	t.TextInverse.AnimateToAt(now, utils.HexNRGBA(next.TextInverse))

	t.Primary.AnimateToAt(now, utils.HexNRGBA(next.Primary))
	t.PrimaryHover.AnimateToAt(now, utils.HexNRGBA(next.PrimaryHover))
	t.OnPrimary.AnimateToAt(now, utils.HexNRGBA(next.OnPrimary))

	t.Secondary.AnimateToAt(now, utils.HexNRGBA(next.Secondary))
	t.SecondaryHover.AnimateToAt(now, utils.HexNRGBA(next.SecondaryHover))
	t.OnSecondary.AnimateToAt(now, utils.HexNRGBA(next.OnSecondary))

	t.Success.AnimateToAt(now, utils.HexNRGBA(next.Success))
	t.Warning.AnimateToAt(now, utils.HexNRGBA(next.Warning))
	t.Danger.AnimateToAt(now, utils.HexNRGBA(next.Danger))
	t.Info.AnimateToAt(now, utils.HexNRGBA(next.Info))

	t.FocusRing.AnimateToAt(now, utils.HexNRGBA(next.FocusRing))
	t.Selection.AnimateToAt(now, utils.HexNRGBA(next.Selection))
	t.Disabled.AnimateToAt(now, utils.HexNRGBA(next.Disabled))
}

func (t *ColorTokenTweens) Value(now time.Time) (ColorTokens, bool) {
	if t == nil {
		return ColorTokens{}, false
	}

	var running bool

	value := func(ct *tween.ColorTween) string {
		c, r := ct.Value(now)
		if r {
			running = true
		}
		return nrgbaHex(c)
	}

	return ColorTokens{
		Background: value(t.Background),
		Surface:    value(t.Surface),
		SurfaceAlt: value(t.SurfaceAlt),
		Border:     value(t.Border),
		Divider:    value(t.Divider),

		TextPrimary:   value(t.TextPrimary),
		TextSecondary: value(t.TextSecondary),
		TextMuted:     value(t.TextMuted),
		TextInverse:   value(t.TextInverse),

		Primary:      value(t.Primary),
		PrimaryHover: value(t.PrimaryHover),
		OnPrimary:    value(t.OnPrimary),

		Secondary:      value(t.Secondary),
		SecondaryHover: value(t.SecondaryHover),
		OnSecondary:    value(t.OnSecondary),

		Success: value(t.Success),
		Warning: value(t.Warning),
		Danger:  value(t.Danger),
		Info:    value(t.Info),

		FocusRing: value(t.FocusRing),
		Selection: value(t.Selection),
		Disabled:  value(t.Disabled),
	}, running
}

func (t *ColorTokenTweens) JumpTo(next ColorTokens) {
	if t == nil {
		return
	}

	t.Background.JumpTo(utils.HexNRGBA(next.Background))
	t.Surface.JumpTo(utils.HexNRGBA(next.Surface))
	t.SurfaceAlt.JumpTo(utils.HexNRGBA(next.SurfaceAlt))
	t.Border.JumpTo(utils.HexNRGBA(next.Border))
	t.Divider.JumpTo(utils.HexNRGBA(next.Divider))

	t.TextPrimary.JumpTo(utils.HexNRGBA(next.TextPrimary))
	t.TextSecondary.JumpTo(utils.HexNRGBA(next.TextSecondary))
	t.TextMuted.JumpTo(utils.HexNRGBA(next.TextMuted))
	t.TextInverse.JumpTo(utils.HexNRGBA(next.TextInverse))

	t.Primary.JumpTo(utils.HexNRGBA(next.Primary))
	t.PrimaryHover.JumpTo(utils.HexNRGBA(next.PrimaryHover))
	t.OnPrimary.JumpTo(utils.HexNRGBA(next.OnPrimary))

	t.Secondary.JumpTo(utils.HexNRGBA(next.Secondary))
	t.SecondaryHover.JumpTo(utils.HexNRGBA(next.SecondaryHover))
	t.OnSecondary.JumpTo(utils.HexNRGBA(next.OnSecondary))

	t.Success.JumpTo(utils.HexNRGBA(next.Success))
	t.Warning.JumpTo(utils.HexNRGBA(next.Warning))
	t.Danger.JumpTo(utils.HexNRGBA(next.Danger))
	t.Info.JumpTo(utils.HexNRGBA(next.Info))

	t.FocusRing.JumpTo(utils.HexNRGBA(next.FocusRing))
	t.Selection.JumpTo(utils.HexNRGBA(next.Selection))
	t.Disabled.JumpTo(utils.HexNRGBA(next.Disabled))
}

func nrgbaHex(c color.NRGBA) string {
	if c.A == 255 {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}

	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}
