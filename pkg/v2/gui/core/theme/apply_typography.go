package theme

import (
	"strings"

	"gioui.org/font"
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
