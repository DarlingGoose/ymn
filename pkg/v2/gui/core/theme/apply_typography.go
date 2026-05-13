package theme

import (
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

type TextRole string

const (
	TextRoleDisplay    TextRole = "display"
	TextRoleH1         TextRole = "h1"
	TextRoleH2         TextRole = "h2"
	TextRoleH3         TextRole = "h3"
	TextRoleH4         TextRole = "h4"
	TextRoleBodyLarge  TextRole = "bodyLarge"
	TextRoleBody       TextRole = "body"
	TextRoleBodySmall  TextRole = "bodySmall"
	TextRoleLabel      TextRole = "label"
	TextRoleLabelSmall TextRole = "labelSmall"
	TextRoleCode       TextRole = "code"
	TextRoleCaption    TextRole = "caption"
)

type TextColorRole int

const (
	ThemeColorTextPrimary TextColorRole = iota
	ThemeColorTextSecondary
	ThemeColorTextMuted
	ThemeColorTextInverse
	ThemeColorOnPrimary
	ThemeColorWarning
	ThemeColorError
	ThemeColorPrimary
)

func ApplyTypography(
	lbl *material.LabelStyle,
	tokens TypographyTokens,
	role TextRole,
) {
	if lbl == nil {
		return
	}

	tokens.FillDefaults()

	style := typographyStyle(tokens, role)

	if style.Size > 0 {
		lbl.TextSize = style.Size
	}

	if style.LineHeight > 0 {
		lbl.LineHeight = style.LineHeight
	}

	if style.Weight > 0 {
		lbl.Font.Weight = fontWeight(style.Weight)
	}
	face := strings.TrimSpace(style.Font)
	switch strings.ToLower(face) {
	case "mono", "monospace", "code":
		lbl.Font.Typeface = font.Typeface("Go Mono")
	case "", "sans", "sans-serif", "default":
		// Keep Gio/material default.
	default:
		lbl.Font.Typeface = font.Typeface(face)
	}
}

func typographyStyle(tokens TypographyTokens, role TextRole) TextStyleToken {
	switch role {
	case TextRoleDisplay:
		return tokens.Display
	case TextRoleH1:
		return tokens.H1
	case TextRoleH2:
		return tokens.H2
	case TextRoleH3:
		return tokens.H3
	case TextRoleH4:
		return tokens.H4
	case TextRoleBodyLarge:
		return tokens.BodyLarge
	case TextRoleBodySmall:
		return tokens.BodySmall
	case TextRoleLabel:
		return tokens.Label
	case TextRoleLabelSmall:
		return tokens.LabelSmall
	case TextRoleCode:
		return tokens.Code
	case TextRoleCaption:
		return tokens.Caption
	case TextRoleBody:
		fallthrough
	default:
		return tokens.Body
	}
}

func fontWeight(weight int) font.Weight {
	switch {
	case weight <= 100:
		return font.Thin
	case weight <= 200:
		return font.ExtraLight
	case weight <= 300:
		return font.Light
	case weight <= 400:
		return font.Normal
	case weight <= 500:
		return font.Medium
	case weight <= 600:
		return font.SemiBold
	case weight <= 700:
		return font.Bold
	case weight <= 800:
		return font.ExtraBold
	default:
		return font.Black
	}
}

func SelectTextColor(tokens *ColorTokens, role TextColorRole) color.NRGBA {
	if tokens == nil {
		return color.NRGBA{A: 255}
	}

	switch role {
	case ThemeColorPrimary:
		return tokens.PrimaryNRGBA()
	case ThemeColorWarning:
		return tokens.WarningNRGBA()
	case ThemeColorError:
		return tokens.DangerNRGBA()
	case ThemeColorTextSecondary:
		return tokens.TextSecondaryNRGBA()
	case ThemeColorTextMuted:
		return tokens.TextMutedNRGBA()
	case ThemeColorTextInverse:
		return tokens.TextInverseNRGBA()
	case ThemeColorOnPrimary:
		return tokens.OnPrimaryNRGBA()
	case ThemeColorTextPrimary:
		fallthrough
	default:
		return tokens.TextPrimaryNRGBA()
	}
}

func ThemedLabel(
	gtx layout.Context,
	th *material.Theme,
	tc *Client,
	role TextRole,
	colorRole TextColorRole,
	text string,
) layout.Dimensions {
	if th == nil {
		th = material.NewTheme()
	}

	if tc == nil {
		tc = DefaultThemeClient
	}

	if tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	typography := tc.GetCurrentTypography()
	colors := tc.GetCurrentColorToken()

	lbl := material.Body1(th, text)
	ApplyTypography(&lbl, typography, role)
	lbl.Color = SelectTextColor(colors, colorRole)

	return lbl.Layout(gtx)
}
