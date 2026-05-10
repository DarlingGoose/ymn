package input

import (
	"errors"
	"image"
	"image/color"
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type Kind int

const (
	KindText Kind = iota
	KindPassword
	KindAPIKey
	KindSearch
	KindPath
	KindNumber
	KindInteger
)

type TextInput struct {
	Editor widget.Editor
	List   layout.List

	Label string
	Hint  string
	Kind  Kind

	// Behavior.
	Multiline bool
	ReadOnly  bool
	Disabled  bool
	MaxLen    int

	// Password/API-key behavior.
	RevealSecret bool
	CanCopy      bool
	CanClear     bool
	LeadingIcon  string
	TrailingIcon string

	// Optional text normalization after changes.
	// Useful for path cleanup, trimming search text, etc.
	Normalize func(text string) string

	// Search-specific callback.
	// Submit still goes through OnSubmit.
	OnSearch func(text string)

	// Validation.
	Rules            []Rule
	ValidateOnChange bool
	LastError        error
	Touched          bool

	// Optional callbacks.
	OnChange func(text string)
	OnSubmit func(text string)
	OnClear  func()
	OnCopy   func(text string)

	clearButton  widget.Clickable
	copyButton   widget.Clickable
	revealButton widget.Clickable

	Theme *material.Theme
	tc    *theme.Client

	Role      theme.TextRole
	LabelRole theme.TextRole
	ErrorRole theme.TextRole

	Width           unit.Dp
	Height          unit.Dp
	MultilineHeight unit.Dp
	Radius          unit.Dp
	Inset           unit.Dp
	Gap             unit.Dp
	IconSize        unit.Dp
}

func NewTextInput(label, hint string, rules ...Rule) *TextInput {
	in := &TextInput{
		Label: label,
		Hint:  hint,
		Rules: rules,

		CanClear: true,
		CanCopy:  false,

		ValidateOnChange: true,

		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,

		Role:      theme.TextRoleBody,
		LabelRole: theme.TextRoleLabel,
		ErrorRole: theme.TextRoleCaption,

		Height:          unit.Dp(44),
		MultilineHeight: unit.Dp(120),
		Radius:          unit.Dp(12),
		Inset:           unit.Dp(12),
		Gap:             unit.Dp(6),
		IconSize:        unit.Dp(18),

		List: layout.List{Axis: layout.Vertical},
	}
	in.syncEditorConfig()
	return in
}

func NewPasswordInput(label, hint string, rules ...Rule) *TextInput {
	in := NewTextInput(label, hint, rules...)
	in.Kind = KindPassword
	in.CanCopy = false
	in.syncEditorConfig()
	return in
}

func NewAPIKeyInput(label, hint string, rules ...Rule) *TextInput {
	in := NewTextInput(label, hint, rules...)
	in.Kind = KindAPIKey
	in.CanCopy = true
	in.syncEditorConfig()
	return in
}

func NewMultilineInput(label, hint string, rules ...Rule) *TextInput {
	in := NewTextInput(label, hint, rules...)
	in.Multiline = true
	in.syncEditorConfig()
	return in
}

func (in *TextInput) WithThemeClient(tc *theme.Client) *TextInput {
	if in == nil {
		return in
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	in.tc = tc
	return in
}

func (in *TextInput) WithRules(rules ...Rule) *TextInput {
	if in == nil {
		return in
	}
	in.Rules = append(in.Rules, rules...)
	return in
}

func (in *TextInput) Text() string {
	if in == nil {
		return ""
	}
	return in.Editor.Text()
}

func (in *TextInput) SetText(v string) {
	if in == nil {
		return
	}
	in.Editor.SetText(v)
	in.LastError = in.Validate()
}

func (in *TextInput) Clear() {
	if in == nil {
		return
	}
	in.Editor.SetText("")
	in.Touched = true
	in.LastError = in.Validate()

	if in.OnClear != nil {
		in.OnClear()
	}
	if in.OnChange != nil {
		in.OnChange("")
	}
}

func (in *TextInput) Validate() error {
	if in == nil {
		return nil
	}
	return Validate(in.Text(), in.Rules...)
}

func (in *TextInput) Valid() bool {
	return in == nil || in.Validate() == nil
}

func (in *TextInput) ErrorText() string {
	if in == nil || in.LastError == nil {
		return ""
	}

	var errs []error
	if errors.As(in.LastError, &errs) {
		parts := make([]string, 0, len(errs))
		for _, err := range errs {
			if err != nil {
				parts = append(parts, err.Error())
			}
		}
		return strings.Join(parts, ", ")
	}

	return in.LastError.Error()
}

func (in *TextInput) Layout(gtx layout.Context) layout.Dimensions {
	if in == nil {
		return layout.Dimensions{}
	}
	if in.Theme == nil {
		in.Theme = material.NewTheme()
	}
	if in.tc == nil {
		in.tc = theme.DefaultThemeClient
	}
	if in.tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	in.syncEditorConfig()
	in.update(gtx)

	if in.Disabled {
		gtx = gtx.Disabled()
	}

	children := []layout.FlexChild{}

	if strings.TrimSpace(in.Label) != "" {
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(
					gtx,
					in.Theme,
					in.tc,
					in.LabelRole,
					theme.ThemeColorTextSecondary,
					in.Label,
				)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		)
	}

	children = append(children, layout.Rigid(in.layoutBox))

	if errText := in.ErrorText(); errText != "" {
		children = append(children,
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return in.layoutError(gtx, errText)
			}),
		)
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, children...)
}

func (in *TextInput) update(gtx layout.Context) {
	changed := false

	for {
		ev, ok := in.Editor.Update(gtx)
		if !ok {
			break
		}

		switch ev := ev.(type) {
		case widget.ChangeEvent:
			changed = true
			in.Touched = true
		case widget.SubmitEvent:
			in.Touched = true
			if in.OnSubmit != nil {
				in.OnSubmit(ev.Text)
			}
		}
	}

	for in.clearButton.Clicked(gtx) {
		in.Clear()
		changed = false // Clear already calls OnChange.
	}

	for in.copyButton.Clicked(gtx) {
		in.copy(gtx)
	}

	for in.revealButton.Clicked(gtx) {
		in.RevealSecret = !in.RevealSecret
		in.syncEditorConfig()
		gtx.Execute(op.InvalidateCmd{})
	}

	if changed {
		if in.Normalize != nil {
			normalized := in.Normalize(in.Text())
			if normalized != in.Text() {
				in.Editor.SetText(normalized)
			}
		}

		if in.ValidateOnChange {
			in.LastError = in.Validate()
		}

		textValue := in.Text()

		if in.OnChange != nil {
			in.OnChange(textValue)
		}

		if in.Kind == KindSearch && in.OnSearch != nil {
			in.OnSearch(textValue)
		}
	}
}

func (in *TextInput) syncEditorConfig() {
	in.Editor.SingleLine = !in.Multiline
	in.Editor.ReadOnly = in.ReadOnly
	in.Editor.MaxLen = in.MaxLen
	in.Editor.Submit = !in.Multiline

	switch {
	case in.Kind == KindPassword && !in.RevealSecret:
		in.Editor.Mask = '•'
	case in.Kind == KindAPIKey && !in.RevealSecret:
		in.Editor.Mask = '•'
	default:
		in.Editor.Mask = 0
	}
}

func (in *TextInput) copy(gtx layout.Context) {
	value := in.Text()
	if value == "" {
		return
	}

	gtx.Execute(clipboard.WriteCmd{
		Type: "text/plain",
		Data: io.NopCloser(strings.NewReader(value)),
	})

	if in.OnCopy != nil {
		in.OnCopy(value)
	}
}

func (in *TextInput) layoutBox(gtx layout.Context) layout.Dimensions {
	style := in.style()

	height := gtx.Dp(in.Height)
	if in.Multiline {
		height = gtx.Dp(in.MultilineHeight)
	}
	if height <= 0 {
		height = 44
	}

	if in.Width > 0 {
		w := gtx.Dp(in.Width)
		gtx.Constraints.Min.X = w
		gtx.Constraints.Max.X = w
	}

	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	borderColor := style.Border
	if in.LastError != nil {
		borderColor = style.Danger
	}

	return utils.SurfaceOutlined(
		gtx,
		style.BG,
		in.Radius,
		utils.SurfaceBorder{
			Color: borderColor,
			Width: unit.Dp(1),
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Left:   in.Inset,
				Right:  in.Inset,
				Top:    unit.Dp(0),
				Bottom: unit.Dp(0),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				var children []layout.FlexChild

				if in.LeadingIcon != "" {
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return iconify.DefaultIconify.Layout(
								gtx,
								in.LeadingIcon,
								in.IconSize,
								style.Icon,
							)
						})
					}))
				}

				children = append(children,
					layout.Flexed(1, in.layoutEditor),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return in.layoutActions(gtx, style)
					}),
				)

				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx, children...)
			})
		},
	)
}

func (in *TextInput) layoutEditor(gtx layout.Context) layout.Dimensions {
	style := in.style()

	editor := material.Editor(in.Theme, &in.Editor, in.Hint)
	editor.Color = style.Text
	editor.HintColor = style.Muted
	editor.SelectionColor = style.Selection

	typography := in.tc.GetCurrentTypography()
	tmpLabel := material.Body1(in.Theme, "")
	theme.ApplyTypography(&tmpLabel, typography, in.Role)

	editor.Font = tmpLabel.Font
	editor.TextSize = tmpLabel.TextSize
	editor.LineHeight = tmpLabel.LineHeight

	if !in.Multiline {
		in.Editor.Alignment = text.Start
		return layout.Center.Layout(gtx, editor.Layout)
	}

	return layout.UniformInset(unit.Dp(2)).Layout(gtx, editor.Layout)
}

func (in *TextInput) layoutActions(gtx layout.Context, style inputStyle) layout.Dimensions {
	hasText := in.Text() != ""

	children := []layout.FlexChild{}

	if in.CanCopy {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return in.layoutIconButton(
				gtx,
				&in.copyButton,
				"lucide:copy",
				style.Icon,
				!hasText,
			)
		}))
	}

	if in.Kind == KindPassword || in.Kind == KindAPIKey {
		icon := "lucide:eye"
		if in.RevealSecret {
			icon = "lucide:eye-off"
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return in.layoutIconButton(
				gtx,
				&in.revealButton,
				icon,
				style.Icon,
				false,
			)
		}))
	}

	if in.CanClear {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return in.layoutIconButton(
				gtx,
				&in.clearButton,
				"lucide:x",
				style.Icon,
				!hasText || in.ReadOnly,
			)
		}))
	}

	if len(children) == 0 {
		return layout.Dimensions{}
	}

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx, children...)
}

func (in *TextInput) layoutIconButton(
	gtx layout.Context,
	clickable *widget.Clickable,
	iconName string,
	col color.NRGBA,
	disabled bool,
) layout.Dimensions {
	size := gtx.Dp(unit.Dp(30))
	if size <= 0 {
		size = 30
	}

	if disabled {
		col.A = 90
		gtx = gtx.Disabled()
	}

	gtx.Constraints.Min = image.Pt(size, size)
	gtx.Constraints.Max = image.Pt(size, size)

	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return iconify.DefaultIconify.Layout(gtx, iconName, in.IconSize, col)
		})
	})
}

func (in *TextInput) layoutError(gtx layout.Context, msg string) layout.Dimensions {
	style := in.style()

	lbl := material.Body1(in.Theme, msg)
	lbl.Color = style.Danger
	theme.ApplyTypography(&lbl, in.tc.GetCurrentTypography(), in.ErrorRole)
	return lbl.Layout(gtx)
}

func (in *TextInput) style() inputStyle {
	tc := in.tc
	if tc == nil {
		tc = theme.DefaultThemeClient
		in.tc = tc
	}

	tokens := tc.GetCurrentColorToken()

	return inputStyle{
		BG:        tokens.SurfaceNRGBA(),
		Border:    tokens.BorderNRGBA(),
		Text:      tokens.TextPrimaryNRGBA(),
		Muted:     tokens.TextMutedNRGBA(),
		Icon:      tokens.TextSecondaryNRGBA(),
		Danger:    tokens.DangerNRGBA(),
		Selection: tokens.SelectionNRGBA(),
	}
}
