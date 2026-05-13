package toggles

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

type Toggle struct {
	clickable widget.Clickable

	Flip    *tween.Flip
	TrackBG *tween.ColorTween
	ThumbBG *tween.ColorTween

	Checked  bool
	Disabled bool

	Label string

	Width       unit.Dp
	Height      unit.Dp
	ThumbSize   unit.Dp
	LabelGap    unit.Dp
	TrackRadius unit.Dp

	BorderWidth unit.Dp
	FocusWidth  unit.Dp

	theme *theme.Client
}

func NewToggle(label string, checked bool) *Toggle {
	tc := theme.DefaultThemeClient
	tokens := tc.GetCurrentColorToken()

	t := &Toggle{
		theme: tc,

		Flip: tween.NewFlip(
			160*time.Millisecond,
			tween.EaseOutCubic,
		),
		TrackBG: tween.NewColorTween(
			140*time.Millisecond,
			tween.EaseOutCubic,
			tokens.DisabledNRGBA(),
		),
		ThumbBG: tween.NewColorTween(
			140*time.Millisecond,
			tween.EaseOutCubic,
			tokens.SurfaceNRGBA(),
		),

		Checked: checked,
		Label:   label,

		Width:       unit.Dp(46),
		Height:      unit.Dp(26),
		ThumbSize:   unit.Dp(16),
		LabelGap:    unit.Dp(10),
		TrackRadius: unit.Dp(8),

		BorderWidth: unit.Dp(1),
		FocusWidth:  unit.Dp(2),
	}

	t.JumpTo(checked)

	return t
}

func (t *Toggle) WithThemeClient(tc *theme.Client) *Toggle {
	if t == nil {
		return t
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	t.theme = tc
	t.SyncThemeTweens(time.Now())

	return t
}

func (t *Toggle) WithLabel(label string) *Toggle {
	if t == nil {
		return t
	}

	t.Label = label
	return t
}

func (t *Toggle) WithDisabled(disabled bool) *Toggle {
	if t == nil {
		return t
	}

	t.Disabled = disabled
	return t
}

func (t *Toggle) SetChecked(checked bool) {
	if t == nil || t.Checked == checked {
		return
	}

	t.Checked = checked

	if checked {
		t.animateOn()
		return
	}

	t.animateOff()
}

func (t *Toggle) Toggle() {
	if t == nil || t.Disabled {
		return
	}

	t.SetChecked(!t.Checked)
}

func (t *Toggle) JumpTo(checked bool) {
	if t == nil {
		return
	}

	t.Checked = checked

	style := t.style()

	if t.Flip != nil {
		t.Flip.JumpExpanded(checked)
	}

	if t.TrackBG != nil {
		if checked {
			t.TrackBG.JumpTo(style.TrackOn)
		} else {
			t.TrackBG.JumpTo(style.TrackOff)
		}
	}

	if t.ThumbBG != nil {
		if checked {
			t.ThumbBG.JumpTo(style.ThumbOn)
		} else {
			t.ThumbBG.JumpTo(style.ThumbOff)
		}
	}
}

// Update handles clicks and returns true if Checked changed.
func (t *Toggle) Update(gtx layout.Context) bool {
	if t == nil || t.Disabled {
		return false
	}

	changed := false
	for t.clickable.Clicked(gtx) {
		t.Toggle()
		gtx.Execute(op.InvalidateCmd{})
		changed = true
	}

	return changed
}

// Changed is kept as a compatibility alias.
func (t *Toggle) Changed(gtx layout.Context) bool {
	return t.Update(gtx)
}

func (t *Toggle) SyncThemeTweens(now time.Time) {
	if t == nil {
		return
	}

	style := t.style()

	track := style.TrackOff
	thumb := style.ThumbOff
	if t.Checked {
		track = style.TrackOn
		thumb = style.ThumbOn
	}
	if t.Disabled {
		track = style.TrackDisabled
		thumb = style.ThumbDisabled
	}

	if t.TrackBG != nil {
		t.TrackBG.AnimateToAt(now, track)
	}
	if t.ThumbBG != nil {
		t.ThumbBG.AnimateToAt(now, thumb)
	}
}

func (t *Toggle) animateOn() {
	style := t.style()

	if t.Flip != nil {
		t.Flip.Expand()
	}
	if t.TrackBG != nil {
		t.TrackBG.AnimateTo(style.TrackOn)
	}
	if t.ThumbBG != nil {
		t.ThumbBG.AnimateTo(style.ThumbOn)
	}
}

func (t *Toggle) animateOff() {
	style := t.style()

	if t.Flip != nil {
		t.Flip.Collapse()
	}
	if t.TrackBG != nil {
		t.TrackBG.AnimateTo(style.TrackOff)
	}
	if t.ThumbBG != nil {
		t.ThumbBG.AnimateTo(style.ThumbOff)
	}
}

func (t *Toggle) Layout(gtx layout.Context) layout.Dimensions {
	if t == nil {
		return layout.Dimensions{}
	}

	now := time.Now()
	style := t.style()

	// Keep local toggle tweens following global theme color tween.
	t.SyncThemeTweens(now)

	progress := 0.0
	flipRunning := false
	if t.Flip != nil {
		progress, flipRunning = t.Flip.Value(now)
	}

	trackColor := style.TrackOff
	trackRunning := false
	if t.TrackBG != nil {
		trackColor, trackRunning = t.TrackBG.Value(now)
	}

	thumbColor := style.ThumbOff
	thumbRunning := false
	if t.ThumbBG != nil {
		thumbColor, thumbRunning = t.ThumbBG.Value(now)
	}

	if t.Disabled {
		trackColor = style.TrackDisabled
		thumbColor = style.ThumbDisabled
	}

	if flipRunning || trackRunning || thumbRunning {
		gtx.Execute(op.InvalidateCmd{})
	}

	if t.theme != nil && t.theme.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	return t.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return t.layoutSwitch(gtx, style, progress, trackColor, thumbColor)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if t.Label == "" {
					return layout.Dimensions{}
				}
				return layout.Spacer{Width: t.LabelGap}.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if t.Label == "" {
					return layout.Dimensions{}
				}

				col := style.Text
				if t.Disabled {
					col = style.TextMuted
				}

				return t.layoutLabel(gtx, style, col)
			}),
		)
	})
}

func (t *Toggle) layoutSwitch(
	gtx layout.Context,
	style toggleStyle,
	progress float64,
	trackColor color.NRGBA,
	thumbColor color.NRGBA,
) layout.Dimensions {
	width := gtx.Dp(t.Width)
	height := gtx.Dp(t.Height)
	thumb := gtx.Dp(t.ThumbSize)

	if width <= 0 {
		width = 46
	}
	if height <= 0 {
		height = 26
	}
	if thumb <= 0 {
		thumb = height - 6
	}

	size := image.Pt(width, height)

	borderColor := style.TrackBorder
	borderWidth := t.BorderWidth

	if t.Checked {
		borderColor = style.TrackOn
	}

	if t.Disabled {
		borderColor = style.TrackDisabled
	}

	if t.clickable.Hovered() && !t.Disabled {
		borderColor = style.Focus
		borderWidth = t.FocusWidth
	}

	return utils.SurfaceOutlined(
		gtx,
		color.NRGBA{}, // transparent background
		t.TrackRadius,
		utils.SurfaceBorder{
			Color: borderColor,
			Width: borderWidth,
		},
		func(gtx layout.Context) layout.Dimensions {
			t.layoutThumb(gtx, progress, thumbColor, width, height, thumb)
			return layout.Dimensions{Size: size}
		},
	)
}

func (t *Toggle) layoutThumb(
	gtx layout.Context,
	progress float64,
	thumbColor color.NRGBA,
	width int,
	height int,
	thumb int,
) {
	padding := (height - thumb) / 2
	if padding < 0 {
		padding = 0
	}

	minX := padding
	maxX := width - thumb - padding
	if maxX < minX {
		maxX = minX
	}

	x := tween.MapInt(progress, minX, maxX)
	y := padding

	thumbRect := image.Rect(x, y, x+thumb, y+thumb)

	paint.FillShape(
		gtx.Ops,
		thumbColor,
		clip.Ellipse(thumbRect).Op(gtx.Ops),
	)
}

func (t *Toggle) layoutLabel(
	gtx layout.Context,
	style toggleStyle,
	col color.NRGBA,
) layout.Dimensions {
	th := material.NewTheme()

	lbl := material.Body1(th, t.Label)
	lbl.Color = col
	lbl.Alignment = text.Middle

	theme.ApplyTypography(&lbl, style.Typo, theme.TextRoleLabel)

	return lbl.Layout(gtx)
}

type toggleStyle struct {
	Tokens *theme.ColorTokens
	Typo   theme.TypographyTokens

	TrackOff      color.NRGBA
	TrackOn       color.NRGBA
	TrackDisabled color.NRGBA
	TrackBorder   color.NRGBA

	ThumbOff      color.NRGBA
	ThumbOn       color.NRGBA
	ThumbDisabled color.NRGBA

	Text      color.NRGBA
	TextMuted color.NRGBA
	Focus     color.NRGBA
}

func (t *Toggle) style() toggleStyle {
	tc := t.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		t.theme = tc
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	return toggleStyle{
		Tokens: tokens,
		Typo:   typo,

		TrackOff:      color.NRGBA{}, // no fill
		TrackOn:       tokens.PrimaryNRGBA(),
		TrackDisabled: tokens.DisabledNRGBA(),
		TrackBorder:   tokens.BorderNRGBA(),

		ThumbOff:      tokens.TextMutedNRGBA(),
		ThumbOn:       tokens.PrimaryNRGBA(),
		ThumbDisabled: tokens.DisabledNRGBA(),

		Text:      tokens.TextPrimaryNRGBA(),
		TextMuted: tokens.TextMutedNRGBA(),
		Focus:     tokens.FocusRingNRGBA(),
	}
}
