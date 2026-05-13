package theme

import (
	"strings"

	"gioui.org/unit"
)

type TypographyTokens struct {
	Name    string         `json:"name" yaml:"name"`
	Display TextStyleToken `json:"display" yaml:"display"`
	H1      TextStyleToken `json:"h1" yaml:"h1"`
	H2      TextStyleToken `json:"h2" yaml:"h2"`
	H3      TextStyleToken `json:"h3" yaml:"h3"`
	H4      TextStyleToken `json:"h4" yaml:"h4"`

	BodyLarge TextStyleToken `json:"bodyLarge" yaml:"bodyLarge"`
	Body      TextStyleToken `json:"body" yaml:"body"`
	BodySmall TextStyleToken `json:"bodySmall" yaml:"bodySmall"`

	Label      TextStyleToken `json:"label" yaml:"label"`
	LabelSmall TextStyleToken `json:"labelSmall" yaml:"labelSmall"`

	Code    TextStyleToken `json:"code" yaml:"code"`
	Caption TextStyleToken `json:"caption" yaml:"caption"`
}

type TextStyleToken struct {
	Size       unit.Sp `json:"size" yaml:"size"`
	LineHeight unit.Sp `json:"lineHeight" yaml:"lineHeight"`

	// Gio text.Font.Weight values usually map nicely to 400, 500, 600, 700.
	Weight int `json:"weight" yaml:"weight"`

	// Optional. Useful if you later support different font faces.
	Font string `json:"font,omitempty" yaml:"font,omitempty"`

	// Letter spacing in Sp. Keep this 0 for most body text.
	LetterSpacing unit.Sp `json:"letterSpacing,omitempty" yaml:"letterSpacing,omitempty"`
}

var DefaultTypography = TypographyTokens{
	Name: "default",
	Display: TextStyleToken{
		Size:       unit.Sp(36),
		LineHeight: unit.Sp(44),
		Weight:     700,
	},
	H1: TextStyleToken{
		Size:       unit.Sp(30),
		LineHeight: unit.Sp(38),
		Weight:     700,
	},
	H2: TextStyleToken{
		Size:       unit.Sp(24),
		LineHeight: unit.Sp(32),
		Weight:     650,
	},
	H3: TextStyleToken{
		Size:       unit.Sp(20),
		LineHeight: unit.Sp(28),
		Weight:     600,
	},
	H4: TextStyleToken{
		Size:       unit.Sp(18),
		LineHeight: unit.Sp(26),
		Weight:     600,
	},

	BodyLarge: TextStyleToken{
		Size:       unit.Sp(17),
		LineHeight: unit.Sp(26),
		Weight:     400,
	},
	Body: TextStyleToken{
		Size:       unit.Sp(15),
		LineHeight: unit.Sp(23),
		Weight:     400,
	},
	BodySmall: TextStyleToken{
		Size:       unit.Sp(13),
		LineHeight: unit.Sp(20),
		Weight:     400,
	},

	Label: TextStyleToken{
		Size:       unit.Sp(14),
		LineHeight: unit.Sp(18),
		Weight:     600,
	},
	LabelSmall: TextStyleToken{
		Size:       unit.Sp(11),
		LineHeight: unit.Sp(16),
		Weight:     600,
	},

	Code: TextStyleToken{
		Size:       unit.Sp(13),
		LineHeight: unit.Sp(20),
		Weight:     400,
		Font:       "mono",
	},
	Caption: TextStyleToken{
		Size:       unit.Sp(12),
		LineHeight: unit.Sp(17),
		Weight:     400,
	},
}

func (t *TypographyTokens) FillDefaults() {
	if t == nil {
		return
	}

	if strings.TrimSpace(t.Name) == "" {
		t.Name = DefaultTypography.Name
	}

	fillTextStyle(&t.Display, DefaultTypography.Display)
	fillTextStyle(&t.H1, DefaultTypography.H1)
	fillTextStyle(&t.H2, DefaultTypography.H2)
	fillTextStyle(&t.H3, DefaultTypography.H3)
	fillTextStyle(&t.H4, DefaultTypography.H4)

	fillTextStyle(&t.BodyLarge, DefaultTypography.BodyLarge)
	fillTextStyle(&t.Body, DefaultTypography.Body)
	fillTextStyle(&t.BodySmall, DefaultTypography.BodySmall)

	fillTextStyle(&t.Label, DefaultTypography.Label)
	fillTextStyle(&t.LabelSmall, DefaultTypography.LabelSmall)

	fillTextStyle(&t.Code, DefaultTypography.Code)
	fillTextStyle(&t.Caption, DefaultTypography.Caption)
}

func fillTextStyle(dst *TextStyleToken, fallback TextStyleToken) {
	if dst.Size == 0 {
		dst.Size = fallback.Size
	}

	if dst.LineHeight == 0 {
		dst.LineHeight = fallback.LineHeight
	}

	if dst.Weight == 0 {
		dst.Weight = fallback.Weight
	}

	if strings.TrimSpace(dst.Font) == "" {
		dst.Font = fallback.Font
	}

	if dst.LetterSpacing == 0 {
		dst.LetterSpacing = fallback.LetterSpacing
	}
}

func appendOrReplaceTypography(existing []*TypographyTokens, next ...*TypographyTokens) []*TypographyTokens {
	byName := make(map[string]int, len(existing))
	out := make([]*TypographyTokens, 0, len(existing)+len(next))

	for _, t := range existing {
		if t == nil || strings.TrimSpace(t.Name) == "" {
			continue
		}

		cp := *t
		cp.FillDefaults()

		key := strings.ToLower(strings.TrimSpace(cp.Name))
		byName[key] = len(out)
		out = append(out, &cp)
	}

	for _, t := range next {
		if t == nil || strings.TrimSpace(t.Name) == "" {
			continue
		}

		cp := *t
		cp.FillDefaults()

		key := strings.ToLower(strings.TrimSpace(cp.Name))
		if idx, ok := byName[key]; ok {
			out[idx] = &cp
			continue
		}

		byName[key] = len(out)
		out = append(out, &cp)
	}

	return out
}

func safeFileName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(name))

	lastDash := false

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == ' ':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}
