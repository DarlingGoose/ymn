package toggles

import (
	"gioui.org/layout"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

type ThemeModeToggle struct {
	Toggle *Toggle

	theme *theme.Client
}

func NewThemeModeToggle(tc *theme.Client) *ThemeModeToggle {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	t := &ThemeModeToggle{
		theme: tc,
		Toggle: NewToggle("Dark mode", tc.GetCurrentMode() == theme.ModeDark).
			WithThemeClient(tc),
	}

	t.SyncFromTheme()

	return t
}

func (t *ThemeModeToggle) WithThemeClient(tc *theme.Client) *ThemeModeToggle {
	if t == nil {
		return t
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	t.theme = tc

	if t.Toggle == nil {
		t.Toggle = NewToggle("", tc.GetCurrentMode() == theme.ModeDark)
	}

	t.Toggle.WithThemeClient(tc)
	t.SyncFromTheme()

	return t
}

func (t *ThemeModeToggle) SyncFromTheme() {
	if t == nil || t.theme == nil || t.Toggle == nil {
		return
	}

	isDark := t.theme.GetCurrentMode() == theme.ModeDark
	t.Toggle.JumpTo(isDark)
	t.syncLabel()
}

func (t *ThemeModeToggle) Update(gtx layout.Context) bool {
	if t == nil || t.theme == nil || t.Toggle == nil {
		return false
	}

	t.syncLabel()

	if !t.Toggle.Update(gtx) {
		return false
	}

	nextMode := theme.ModeLight
	if t.Toggle.Checked {
		nextMode = theme.ModeDark
	}

	if err := t.theme.SetMode(nextMode); err != nil {
		t.SyncFromTheme()
		return false
	}

	t.syncLabel()

	return true
}

func (t *ThemeModeToggle) Layout(gtx layout.Context) layout.Dimensions {
	if t == nil || t.Toggle == nil {
		return layout.Dimensions{}
	}

	return t.Toggle.Layout(gtx)
}

func (t *ThemeModeToggle) syncLabel() {
	if t == nil || t.Toggle == nil || t.theme == nil {
		return
	}

	if t.theme.GetCurrentMode() == theme.ModeDark {
		t.Toggle.Label = "Dark mode"
		return
	}

	t.Toggle.Label = "Light mode"
}
