package input

import (
	"path/filepath"
	"strings"

	"gioui.org/widget/material"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

func NewSearchInput(hint string, rules ...Rule) *TextInput {
	in := NewTextInput("", hint, rules...)
	in.Kind = KindSearch
	in.LeadingIcon = "lucide:search"
	in.CanClear = true
	in.CanCopy = false
	in.ValidateOnChange = true
	in.Editor.Submit = true
	in.syncEditorConfig()
	return in
}

func NewLabeledSearchInput(label, hint string, rules ...Rule) *TextInput {
	in := NewSearchInput(hint, rules...)
	in.Label = label
	return in
}

func NewPathInput(label, hint string, rules ...Rule) *TextInput {
	in := NewTextInput(label, hint, rules...)
	in.Kind = KindPath
	in.LeadingIcon = "lucide:folder"
	in.CanClear = true
	in.CanCopy = true
	in.ValidateOnChange = true
	in.Role = theme.TextRoleCode
	in.syncEditorConfig()
	return in
}

func NewFilePathInput(label, hint string, rules ...Rule) *TextInput {
	in := NewPathInput(label, hint, append([]Rule{
		FileExists(),
	}, rules...)...)
	in.LeadingIcon = "lucide:file"
	return in
}

func NewDirPathInput(label, hint string, rules ...Rule) *TextInput {
	in := NewPathInput(label, hint, append([]Rule{
		DirExists(),
	}, rules...)...)
	in.LeadingIcon = "lucide:folder"
	return in
}

func NewExecutablePathInput(label, hint string, rules ...Rule) *TextInput {
	in := NewPathInput(label, hint, append([]Rule{
		FileExists(),
		ExecutableFile(),
	}, rules...)...)
	in.LeadingIcon = "lucide:file-cog"
	return in
}

func NewCleanPathInput(label, hint string, rules ...Rule) *TextInput {
	in := NewPathInput(label, hint, rules...)
	in.Normalize = func(text string) string {
		text = strings.TrimSpace(text)
		if text == "" {
			return ""
		}
		return filepath.Clean(text)
	}
	return in
}

func NewNumberInput(label, hint string, rules ...Rule) *NumberInput {
	base := NewTextInput(label, hint, rules...)
	base.Kind = KindNumber
	base.LeadingIcon = "lucide:hash"
	base.CanClear = true
	base.CanCopy = false
	base.ValidateOnChange = true
	base.syncEditorConfig()

	n := &NumberInput{
		TextInput:     base,
		AllowDecimal:  true,
		AllowNegative: true,
		Step:          1,
	}

	base.Rules = append([]Rule{
		n.numberRule(),
	}, base.Rules...)

	return n
}

func NewIntegerInput(label, hint string, rules ...Rule) *NumberInput {
	n := NewNumberInput(label, hint, rules...)
	n.Kind = KindInteger
	n.AllowDecimal = false
	n.Step = 1
	n.LeadingIcon = "lucide:hash"
	return n
}

func NewPortInput(label, hint string, rules ...Rule) *NumberInput {
	n := NewIntegerInput(label, hint, append([]Rule{
		NumberRange(1, 65535),
	}, rules...)...)
	n.LeadingIcon = "lucide:ethernet-port"
	return n
}

func NewPercentInput(label, hint string, rules ...Rule) *NumberInput {
	n := NewNumberInput(label, hint, append([]Rule{
		NumberRange(0, 100),
	}, rules...)...)
	n.AllowNegative = false
	n.Step = 1
	n.LeadingIcon = "lucide:percent"
	return n
}

func (in *TextInput) WithMaterialTheme(th *material.Theme) *TextInput {
	if in == nil {
		return in
	}
	if th != nil {
		in.Theme = th
	}
	return in
}
