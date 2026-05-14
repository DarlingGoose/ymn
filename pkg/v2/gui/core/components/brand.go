package components

import (
	"context"
	"image"
	"image/color"
	"strings"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

type Brand struct {
	Title string

	// Icon is an Iconify name, for example "lucide:box", "mdi:duck", etc.
	Icon string

	// Image is optional. If set, it wins over Icon.
	Image image.Image

	IconSize unit.Dp
	Gap      unit.Dp
	Radius   unit.Dp
	Inset    unit.Dp

	ShowTitle bool

	Role theme.TextRole

	theme *theme.Client
}

func NewBrand(title string) *Brand {
	return &Brand{
		Title: title,

		IconSize: unit.Dp(28),
		Gap:      unit.Dp(10),
		Radius:   unit.Dp(10),
		Inset:    unit.Dp(0),

		ShowTitle: true,
		Role:      theme.TextRoleH3,

		theme: theme.DefaultThemeClient,
	}
}

func (b *Brand) WithThemeClient(tc *theme.Client) *Brand {
	if b == nil {
		return b
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	b.theme = tc
	return b
}

func (b *Brand) WithIcon(icon string) *Brand {
	if b == nil {
		return b
	}

	b.Icon = strings.TrimSpace(icon)
	return b
}

func (b *Brand) WithImage(img image.Image) *Brand {
	if b == nil {
		return b
	}

	b.Image = img
	return b
}

func (b *Brand) WithShowTitle(show bool) *Brand {
	if b == nil {
		return b
	}

	b.ShowTitle = show
	return b
}

func (b *Brand) Layout(gtx layout.Context) layout.Dimensions {
	if b == nil {
		return layout.Dimensions{}
	}

	tc := b.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		b.theme = tc
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	textColor := tokens.TextPrimaryNRGBA()
	muted := tokens.TextMutedNRGBA()
	bg := tokens.SurfaceAltNRGBA()

	return layout.UniformInset(b.Inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if !b.ShowTitle {
			return b.layoutMark(gtx, bg, textColor, muted, typo)
		}

		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return b.layoutMark(gtx, bg, textColor, muted, typo)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if strings.TrimSpace(b.Title) == "" {
					return layout.Dimensions{}
				}

				return layout.Spacer{Width: b.Gap}.Layout(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if strings.TrimSpace(b.Title) == "" {
					return layout.Dimensions{}
				}

				lbl := material.Body1(material.NewTheme(), b.Title)
				lbl.Color = textColor
				lbl.Alignment = text.Middle

				theme.ApplyTypography(&lbl, typo, b.Role)

				return lbl.Layout(gtx)
			}),
		)
	})
}

func (b *Brand) layoutMark(
	gtx layout.Context,
	bg color.NRGBA,
	textColor color.NRGBA,
	iconColor color.NRGBA,
	typo theme.TypographyTokens,
) layout.Dimensions {
	size := gtx.Dp(b.IconSize)
	if size <= 0 {
		size = 28
	}

	gtx.Constraints.Min.X = size
	gtx.Constraints.Max.X = size
	gtx.Constraints.Min.Y = size
	gtx.Constraints.Max.Y = size

	if b.Image != nil {
		return b.layoutImage(gtx, size)
	}

	if strings.TrimSpace(b.Icon) != "" {
		return b.layoutIcon(gtx, size, iconColor)
	}

	return utils.Surface(
		gtx,
		bg,
		b.Radius,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				acronym := utils.Acronym(b.Title)
				if acronym == "" {
					acronym = "?"
				}

				lbl := material.Body1(material.NewTheme(), acronym)
				lbl.Color = textColor
				lbl.Alignment = text.Middle

				theme.ApplyTypography(&lbl, typo, theme.TextRoleLabel)

				return lbl.Layout(gtx)
			})
		},
	)
}

func (b *Brand) layoutIcon(
	gtx layout.Context,
	sizePx int,
	col color.NRGBA,
) layout.Dimensions {
	ic, err := iconify.DefaultIconify.Icon(context.Background(), b.Icon)
	if err != nil || ic == nil {
		return b.layoutFallbackAcronym(gtx, sizePx)
	}

	size := b.IconSize
	if size <= 0 {
		size = unit.Dp(28)
	}

	iconGtx := gtx
	iconGtx.Constraints.Min.X = sizePx
	iconGtx.Constraints.Max.X = sizePx
	iconGtx.Constraints.Min.Y = sizePx
	iconGtx.Constraints.Max.Y = sizePx

	return ic.Layout(iconGtx, size, col)
}

func (b *Brand) layoutImage(gtx layout.Context, sizePx int) layout.Dimensions {
	if b.Image == nil {
		return layout.Dimensions{Size: image.Pt(sizePx, sizePx)}
	}

	rect := image.Rectangle{Max: image.Pt(sizePx, sizePx)}

	clipStack := clip.UniformRRect(rect, gtx.Dp(b.Radius)).Push(gtx.Ops)
	defer clipStack.Pop()

	op := paint.NewImageOp(b.Image)
	op.Add(gtx.Ops)

	paint.PaintOp{}.Add(gtx.Ops)

	return layout.Dimensions{Size: rect.Size()}
}

func (b *Brand) layoutFallbackAcronym(gtx layout.Context, sizePx int) layout.Dimensions {
	tc := b.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	return utils.Surface(
		gtx,
		tokens.SurfaceAltNRGBA(),
		b.Radius,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				acronym := utils.Acronym(b.Title)
				if acronym == "" {
					acronym = "?"
				}

				lbl := material.Body1(material.NewTheme(), acronym)
				lbl.Color = tokens.TextPrimaryNRGBA()
				lbl.Alignment = text.Middle

				theme.ApplyTypography(&lbl, typo, theme.TextRoleLabel)

				return lbl.Layout(gtx)
			})
		},
	)
}
