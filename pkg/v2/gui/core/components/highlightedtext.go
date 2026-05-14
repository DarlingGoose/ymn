package components

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type HighlightedText struct {
	Theme *material.Theme

	Query         string
	CaseSensitive bool

	// Typography roles.
	TextRole      theme.TextRole
	HighlightRole theme.TextRole

	// Text color roles.
	TextColorRole          theme.TextColorRole
	HighlightTextColorRole theme.TextColorRole

	// Optional hard overrides.
	// If A == 0, theme color is used.
	TextColorOverride          color.NRGBA
	HighlightTextColorOverride color.NRGBA
	HighlightColorOverride     color.NRGBA

	// Background role for highlighted runs.
	// If HighlightColorOverride.A != 0, this is ignored.

	// Padding around highlighted runs.
	HighlightInset layout.Inset

	theme *theme.Client
}

func NewHighlightedText(th *material.Theme) *HighlightedText {
	if th == nil {
		th = material.NewTheme()
	}

	return &HighlightedText{
		Theme: th,
		theme: theme.DefaultThemeClient,

		TextRole:      theme.TextRoleBody,
		HighlightRole: theme.TextRoleBody,

		TextColorRole:          theme.ThemeColorTextPrimary,
		HighlightTextColorRole: theme.ThemeColorTextInverse,

		HighlightInset: layout.Inset{
			Left:   unit.Dp(2),
			Right:  unit.Dp(2),
			Top:    unit.Dp(1),
			Bottom: unit.Dp(1),
		},
	}
}

func (h *HighlightedText) WithThemeClient(tc *theme.Client) *HighlightedText {
	if h == nil {
		return h
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	h.theme = tc
	return h
}

func (h *HighlightedText) WithQuery(query string) *HighlightedText {
	if h == nil {
		return h
	}

	h.Query = query
	return h
}

func (h *HighlightedText) WithCaseSensitive(caseSensitive bool) *HighlightedText {
	if h == nil {
		return h
	}

	h.CaseSensitive = caseSensitive
	return h
}

func (h *HighlightedText) WithTextRole(role theme.TextRole) *HighlightedText {
	if h == nil {
		return h
	}

	h.TextRole = role
	return h
}

func (h *HighlightedText) WithHighlightRole(role theme.TextRole) *HighlightedText {
	if h == nil {
		return h
	}

	h.HighlightRole = role
	return h
}

func (h *HighlightedText) WithTextColorRole(role theme.TextColorRole) *HighlightedText {
	if h == nil {
		return h
	}

	h.TextColorRole = role
	return h
}

func (h *HighlightedText) WithHighlightTextColorRole(role theme.TextColorRole) *HighlightedText {
	if h == nil {
		return h
	}

	h.HighlightTextColorRole = role
	return h
}

func (h *HighlightedText) WithHighlightColor(col color.NRGBA) *HighlightedText {
	if h == nil {
		return h
	}

	h.HighlightColorOverride = col
	return h
}

func (h *HighlightedText) Layout(gtx layout.Context, text string) layout.Dimensions {
	if h == nil {
		return layout.Dimensions{}
	}

	if h.Theme == nil {
		h.Theme = material.NewTheme()
	}

	if h.theme == nil {
		h.theme = theme.DefaultThemeClient
	}

	if h.theme.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	parts := theme.SplitHighlightParts(text, h.Query, h.CaseSensitive)
	if len(parts) == 0 {
		return layout.Dimensions{}
	}

	children := make([]layout.FlexChild, 0, len(parts))
	for _, part := range parts {
		part := part

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if part.Highlight {
				return h.layoutHighlightedPart(gtx, part.Text)
			}

			return h.layoutTextPart(gtx, part.Text, false)
		}))
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx, children...)
}

func (h *HighlightedText) layoutTextPart(
	gtx layout.Context,
	textValue string,
	highlight bool,
) layout.Dimensions {
	tc := h.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	typography := tc.GetCurrentTypography()
	colors := tc.GetCurrentColorToken()

	textRole := h.TextRole
	colorRole := h.TextColorRole
	colorOverride := h.TextColorOverride

	if highlight {
		textRole = h.HighlightRole
		colorRole = h.HighlightTextColorRole
		colorOverride = h.HighlightTextColorOverride
	}

	lbl := material.Body1(h.Theme, textValue)
	theme.ApplyTypography(&lbl, typography, textRole)

	if colorOverride.A != 0 {
		lbl.Color = colorOverride
	} else {
		lbl.Color = theme.SelectTextColor(colors, colorRole)
	}

	lbl.MaxLines = 1

	return lbl.Layout(gtx)
}

func (h *HighlightedText) layoutHighlightedPart(
	gtx layout.Context,
	textValue string,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)

	dims := h.HighlightInset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return h.layoutTextPart(gtx, textValue, true)
	})

	call := macro.Stop()

	bg := h.highlightColor()

	paint.FillShape(
		gtx.Ops,
		bg,
		clip.UniformRRect(
			image.Rectangle{Max: dims.Size},
			gtx.Dp(unit.Dp(4)),
		).Op(gtx.Ops),
	)

	call.Add(gtx.Ops)

	return dims
}

func (h *HighlightedText) highlightColor() color.NRGBA {
	if h.HighlightColorOverride.A != 0 {
		return h.HighlightColorOverride
	}

	tc := h.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	tokens := tc.GetCurrentColorToken()
	if tokens == nil {
		return color.NRGBA{R: 255, G: 221, B: 120, A: 255}
	}

	// Good default: use primary as the highlight chip.
	return tokens.PrimaryNRGBA()
}
