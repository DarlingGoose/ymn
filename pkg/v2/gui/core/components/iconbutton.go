package components

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

type TextCollapseMode int

const (
	TextCollapseNever TextCollapseMode = iota
	TextCollapseWhenNarrow
)

type IconPlacement int

const (
	IconLeading IconPlacement = iota
	IconTrailing
)

type IconButton struct {
	Clickable *widget.Clickable

	Icon        *iconify.SVGIcon
	LoadingIcon *iconify.SVGIcon

	Spinner *tween.RotationTween
	BG      *tween.ColorTween

	Text    string
	Acronym string

	Loading  bool
	Disabled bool

	IconPlacement IconPlacement

	TextCollapseMode  TextCollapseMode
	CollapseTextBelow unit.Dp

	Height   unit.Dp
	MinWidth unit.Dp
	Radius   unit.Dp
	PaddingX unit.Dp
	Gap      unit.Dp
	IconSize unit.Dp

	Role theme.TextRole

	theme *theme.Client

	FillWidth       bool
	CompactPaddingX unit.Dp
	IconYOffset     unit.Dp
	TextYOffset     unit.Dp
}

type iconButtonStyle struct {
	Tokens *theme.ColorTokens
	Typo   theme.TypographyTokens

	BG         color.NRGBA
	HoverBG    color.NRGBA
	ActiveBG   color.NRGBA
	DisabledBG color.NRGBA

	Text         color.NRGBA
	TextDisabled color.NRGBA

	Icon         color.NRGBA
	IconDisabled color.NRGBA
}

func NewIconButton(text string, clickable *widget.Clickable, icon *iconify.SVGIcon) *IconButton {
	if clickable == nil {
		clickable = &widget.Clickable{}
	}

	tc := theme.DefaultThemeClient
	tokens := tc.GetCurrentColorToken()

	return &IconButton{
		theme:     tc,
		Clickable: clickable,

		Icon:    icon,
		Text:    text,
		Acronym: utils.Acronym(text),

		Spinner: tween.NewRotationTweenDeg(
			160*time.Millisecond,
			tween.EaseOutCubic,
			0,
		),
		BG: tween.NewColorTween(
			120*time.Millisecond,
			tween.EaseOutCubic,
			tokens.PrimaryNRGBA(),
		),
		FillWidth:       true,
		CompactPaddingX: unit.Dp(0),
		IconPlacement:   IconLeading,

		TextCollapseMode:  TextCollapseWhenNarrow,
		CollapseTextBelow: unit.Dp(96),

		Height:      unit.Dp(42),
		MinWidth:    unit.Dp(42),
		Radius:      unit.Dp(12),
		PaddingX:    unit.Dp(14),
		Gap:         unit.Dp(8),
		TextYOffset: unit.Dp(4),
		IconSize:    unit.Dp(18),

		Role: theme.TextRoleLabel,
	}
}

func (b *IconButton) WithThemeClient(tc *theme.Client) *IconButton {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	b.theme = tc

	style := b.style()
	if b.BG != nil {
		b.BG.JumpTo(style.BG)
	}

	return b
}
func (b *IconButton) style() iconButtonStyle {
	tc := b.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		b.theme = tc
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	return iconButtonStyle{
		Tokens: tokens,
		Typo:   typo,

		BG:         tokens.PrimaryNRGBA(),
		HoverBG:    tokens.PrimaryHoverNRGBA(),
		ActiveBG:   tokens.SecondaryNRGBA(),
		DisabledBG: tokens.DisabledNRGBA(),

		Text:         tokens.OnPrimaryNRGBA(),
		TextDisabled: tokens.TextMutedNRGBA(),

		Icon:         tokens.OnPrimaryNRGBA(),
		IconDisabled: tokens.TextMutedNRGBA(),
	}
}

func (b *IconButton) targetBG(style iconButtonStyle) color.NRGBA {
	switch {
	case b.Disabled:
		return style.DisabledBG
	case b.Clickable != nil && b.Clickable.Pressed():
		return style.ActiveBG
	case b.Clickable != nil && b.Clickable.Hovered():
		return style.HoverBG
	default:
		return style.BG
	}
}
func (b *IconButton) WithRole(role theme.TextRole) *IconButton {
	b.Role = role
	return b
}
func (b *IconButton) SetLoading(loading bool) {
	if b == nil || b.Loading == loading {
		return
	}

	b.Loading = loading

	if b.Spinner == nil {
		return
	}

	if loading {
		b.Spinner.StartContinuousDegreesPerSecond(420)
		return
	}

	b.Spinner.Stop(true)
}
func (b *IconButton) layoutIconBox(
	gtx layout.Context,
	style iconButtonStyle,
	enabled bool,
	angle float64,
) layout.Dimensions {
	iconSize := gtx.Dp(b.IconSize)
	if iconSize <= 0 {
		iconSize = 18
	}

	boxH := iconSize + gtx.Dp(unit.Dp(4))
	boxW := iconSize

	gtx.Constraints.Min.X = boxW
	gtx.Constraints.Max.X = boxW
	gtx.Constraints.Min.Y = boxH
	gtx.Constraints.Max.Y = boxH

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		y := gtx.Dp(b.IconYOffset)
		if y != 0 {
			stack := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
			defer stack.Pop()
		}

		return b.layoutIcon(gtx, style, enabled, angle)
	})
}
func (b *IconButton) Clicked(gtx layout.Context) bool {
	if b == nil || b.Clickable == nil || b.Disabled || b.Loading {
		return false
	}

	for b.Clickable.Clicked(gtx) {
		return true
	}

	return false
}

func (b *IconButton) Layout(gtx layout.Context) layout.Dimensions {
	if b == nil {
		return layout.Dimensions{}
	}

	if b.Clickable == nil {
		b.Clickable = &widget.Clickable{}
	}

	if b.Disabled {
		style := b.style()
		bg := style.DisabledBG
		if b.BG != nil {
			b.BG.JumpTo(bg)
		}

		return b.layoutSurface(gtx, style, bg, false, 0)
	}

	return b.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		style := b.style()
		now := time.Now()

		targetBG := b.targetBG(style)

		if b.BG != nil {
			// Use AnimateToAt with the "same target" guard so this does not restart every frame.
			b.BG.AnimateToAt(now, targetBG)
		}

		bg := targetBG
		bgRunning := false
		if b.BG != nil {
			bg, bgRunning = b.BG.Value(now)
		}

		angle := 0.0
		spinRunning := false
		if b.Spinner != nil {
			angle, spinRunning = b.Spinner.Value(now)
		}

		if bgRunning || spinRunning {
			gtx.Execute(op.InvalidateCmd{})
		}

		if b.theme != nil && b.theme.ColorTweenRunning() {
			gtx.Execute(op.InvalidateCmd{})
		}

		return b.layoutSurface(gtx, style, bg, true, angle)
	})
}

func (b *IconButton) layoutSurface(
	gtx layout.Context,
	style iconButtonStyle,
	bg color.NRGBA,
	enabled bool,
	angle float64,
) layout.Dimensions {
	height := gtx.Dp(b.Height)
	minWidth := gtx.Dp(b.MinWidth)

	if height <= 0 {
		height = 40
	}

	hideText := b.shouldHideText(gtx)

	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	if b.FillWidth && gtx.Constraints.Max.X > 0 {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	} else if minWidth > 0 && gtx.Constraints.Min.X < minWidth {
		gtx.Constraints.Min.X = minWidth
	}

	paddingX := b.PaddingX
	if hideText {
		paddingX = b.CompactPaddingX
	}

	return utils.Surface(gtx, bg, b.Radius, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Left:  paddingX,
			Right: paddingX,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := b.children(gtx, style, enabled, angle)
				if len(children) == 0 {
					return layout.Dimensions{Size: image.Pt(gtx.Constraints.Min.X, height)}
				}

				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx, children...)
			})
		})
	})
}
func (b *IconButton) children(
	gtx layout.Context,
	style iconButtonStyle,
	enabled bool,
	angle float64,
) []layout.FlexChild {
	hideText := b.shouldHideText(gtx)

	hasRealIcon := b.Icon != nil || b.LoadingIcon != nil || b.Loading
	hasText := b.Text != ""

	iconChild := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return b.layoutIconBox(gtx, style, enabled, angle)
	})

	gapChild := layout.Rigid(layout.Spacer{Width: b.Gap}.Layout)

	textChild := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		y := gtx.Dp(b.TextYOffset)
		if y != 0 {
			stack := op.Offset(image.Pt(0, y)).Push(gtx.Ops)
			defer stack.Pop()
		}
		return b.layoutText(gtx, style, enabled, b.Text)
	})

	acronymChild := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return b.layoutText(gtx, style, enabled, b.effectiveAcronym())
	})

	if !hideText {
		switch {
		case hasRealIcon && hasText && b.IconPlacement == IconTrailing:
			return []layout.FlexChild{textChild, gapChild, iconChild}
		case hasRealIcon && hasText:
			return []layout.FlexChild{iconChild, gapChild, textChild}
		case hasRealIcon:
			return []layout.FlexChild{iconChild}
		case hasText:
			return []layout.FlexChild{textChild}
		default:
			return nil
		}
	}

	if hasRealIcon {
		return []layout.FlexChild{iconChild}
	}

	if hasText {
		return []layout.FlexChild{acronymChild}
	}

	return nil
}

func (b *IconButton) layoutIcon(
	gtx layout.Context,
	style iconButtonStyle,
	enabled bool,
	angle float64,
) layout.Dimensions {
	col := style.Icon
	if !enabled {
		col = style.IconDisabled
	}

	icon := b.Icon
	if b.Loading && b.LoadingIcon != nil {
		icon = b.LoadingIcon
	}

	if icon == nil {
		return layout.Dimensions{}
	}

	if b.Loading {
		return icon.LayoutRotated(gtx, b.IconSize, col, angle)
	}

	return icon.Layout(gtx, b.IconSize, col)
}

func (b *IconButton) layoutText(
	gtx layout.Context,
	style iconButtonStyle,
	enabled bool,
	value string,
) layout.Dimensions {
	if value == "" {
		return layout.Dimensions{}
	}

	col := style.Text
	if !enabled {
		col = style.TextDisabled
	}

	th := material.NewTheme()

	lbl := material.Body1(th, value)
	lbl.Color = col
	lbl.Alignment = text.Middle

	theme.ApplyTypography(&lbl, style.Typo, b.Role)

	return lbl.Layout(gtx)
}

func (b *IconButton) shouldHideText(gtx layout.Context) bool {
	if b == nil {
		return false
	}

	if b.TextCollapseMode == TextCollapseNever {
		return false
	}

	collapsePx := gtx.Dp(b.CollapseTextBelow)
	if collapsePx <= 0 {
		return false
	}

	return gtx.Constraints.Max.X > 0 && gtx.Constraints.Max.X <= collapsePx
}

func (b *IconButton) effectiveAcronym() string {
	if b == nil {
		return ""
	}

	if b.Acronym != "" {
		return b.Acronym
	}

	return utils.Acronym(b.Text)
}
