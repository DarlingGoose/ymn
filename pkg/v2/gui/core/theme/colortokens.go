package theme

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

const (
	MinNormalTextContrast = 4.5
	PreferredTextContrast = 7.0 // Better for long reading
	MinLargeTextContrast  = 3.0
	MinUIContrast         = 3.0
)

type Theme struct {
	Name        string      `json:"name" yaml:"name"`
	LightColors ColorTokens `json:"lightColors" yaml:"lightColors"`
	DarkColors  ColorTokens `json:"darkColors" yaml:"darkColors"`

	//iconify name
	IconName string `json:"iconName" yaml:"iconName"`
}

type ColorTokens struct {
	// Core surfaces
	Background string `json:"background,omitempty" yaml:"background"`
	Surface    string `json:"surface,omitempty" yaml:"surface"`
	SurfaceAlt string `json:"surfaceAlt,omitempty" yaml:"surfaceAlt"`
	Border     string `json:"border,omitempty" yaml:"border"`
	Divider    string `json:"divider,omitempty" yaml:"divider"`

	// Text
	TextPrimary    string `json:"textPrimary,omitempty" yaml:"textPrimary"`
	TextSecondary  string `json:"textSecondary,omitempty" yaml:"textSecondary"`
	TextMuted      string `json:"textMuted,omitempty" yaml:"textMuted"`
	TextInverse    string `json:"textInverse,omitempty" yaml:"textInverse"`
	HighlightColor string `json:"highlightColor" yaml:"highlightColor"`

	// Brand/action
	Primary      string `json:"primary,omitempty" yaml:"primary"`
	PrimaryHover string `json:"primaryHover,omitempty" yaml:"primaryHover"`
	OnPrimary    string `json:"onPrimary,omitempty" yaml:"onPrimary"`

	Secondary      string `json:"secondary,omitempty" yaml:"secondary"`
	SecondaryHover string `json:"secondaryHover,omitempty" yaml:"secondaryHover"`
	OnSecondary    string `json:"onSecondary,omitempty" yaml:"onSecondary"`

	// Feedback
	Success string `json:"success,omitempty" yaml:"success"`
	Warning string `json:"warning,omitempty" yaml:"warning"`
	Danger  string `json:"danger,omitempty" yaml:"danger"`
	Info    string `json:"info,omitempty" yaml:"info"`

	// Interactive states
	FocusRing string `json:"focusRing,omitempty" yaml:"focusRing"`
	Selection string `json:"selection,omitempty" yaml:"selection"`
	Disabled  string `json:"disabled,omitempty" yaml:"disabled"`
}

func (c ColorTokens) MustValidate() error {
	values := map[string]string{
		"background":     c.Background,
		"surface":        c.Surface,
		"surfaceAlt":     c.SurfaceAlt,
		"border":         c.Border,
		"divider":        c.Divider,
		"textPrimary":    c.TextPrimary,
		"textSecondary":  c.TextSecondary,
		"textMuted":      c.TextMuted,
		"textInverse":    c.TextInverse,
		"primary":        c.Primary,
		"primaryHover":   c.PrimaryHover,
		"onPrimary":      c.OnPrimary,
		"secondary":      c.Secondary,
		"secondaryHover": c.SecondaryHover,
		"onSecondary":    c.OnSecondary,
		"success":        c.Success,
		"warning":        c.Warning,
		"danger":         c.Danger,
		"info":           c.Info,
		"focusRing":      c.FocusRing,
		"selection":      c.Selection,
		"disabled":       c.Disabled,
		"highlightColor": c.HighlightColor,
	}

	for name, value := range values {
		if _, err := utils.ParseHexNRGBA(value); err != nil {
			return fmt.Errorf("invalid color token %q=%q: %w", name, value, err)
		}
	}

	return nil
}
func (c ColorTokens) BackgroundNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Background)
}

func (c ColorTokens) SurfaceNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Surface)
}

func (c ColorTokens) SurfaceAltNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.SurfaceAlt)
}

func (c ColorTokens) BorderNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Border)
}

func (c ColorTokens) DividerNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Divider)
}

func (c ColorTokens) HighLightColor() color.NRGBA {
	return utils.HexNRGBA(c.HighlightColor)
}

func (c ColorTokens) TextPrimaryNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.TextPrimary)
}

func (c ColorTokens) TextSecondaryNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.TextSecondary)
}

func (c ColorTokens) TextMutedNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.TextMuted)
}

func (c ColorTokens) TextInverseNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.TextInverse)
}

func (c ColorTokens) PrimaryNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Primary)
}

func (c ColorTokens) PrimaryHoverNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.PrimaryHover)
}

func (c ColorTokens) OnPrimaryNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.OnPrimary)
}

func (c ColorTokens) SecondaryNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Secondary)
}

func (c ColorTokens) SecondaryHoverNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.SecondaryHover)
}

func (c ColorTokens) OnSecondaryNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.OnSecondary)
}

func (c ColorTokens) SuccessNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Success)
}

func (c ColorTokens) WarningNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Warning)
}

func (c ColorTokens) DangerNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Danger)
}

func (c ColorTokens) InfoNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Info)
}

func (c ColorTokens) FocusRingNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.FocusRing)
}

func (c ColorTokens) SelectionNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Selection)
}

func (c ColorTokens) DisabledNRGBA() color.NRGBA {
	return utils.HexNRGBA(c.Disabled)
}

type HighlightPart struct {
	Text      string
	Highlight bool
}

func SplitHighlightParts(text, query string, caseSensitive bool) []HighlightPart {
	if text == "" {
		return nil
	}

	if query == "" {
		return []HighlightPart{{Text: text}}
	}

	searchText := text
	searchQuery := query

	if !caseSensitive {
		searchText = strings.ToLower(text)
		searchQuery = strings.ToLower(query)
	}

	var parts []HighlightPart

	offset := 0
	for {
		idx := strings.Index(searchText[offset:], searchQuery)
		if idx < 0 {
			if offset < len(text) {
				parts = append(parts, HighlightPart{
					Text: text[offset:],
				})
			}
			break
		}

		start := offset + idx
		end := start + len(searchQuery)

		if start > offset {
			parts = append(parts, HighlightPart{
				Text: text[offset:start],
			})
		}

		parts = append(parts, HighlightPart{
			Text:      text[start:end],
			Highlight: true,
		})

		offset = end
	}

	return parts
}
