package dropdowns

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

type ThemeDropdown struct {
	Dropdown *Dropdown

	theme *theme.Client

	lastThemeName string
	OnError       func(error)
}

func NewThemeDropdown(tc *theme.Client) *ThemeDropdown {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	d := &ThemeDropdown{
		theme: tc,
	}

	d.Dropdown = NewDropdown(d.themeItems()).
		WithThemeClient(tc).
		WithRole(theme.TextRoleLabel)

	d.syncSelection()

	d.Dropdown.SelectItemEvent(func(item DropdownItem, valid bool) {
		if !valid {
			return
		}

		if strings.EqualFold(item.Value, d.currentThemeName()) {
			return
		}

		if err := tc.SelectTheme(item.Value); err != nil {
			if d.OnError != nil {
				d.OnError(err)
			}
			return
		}

		d.lastThemeName = item.Value
	})

	return d
}

func (d *ThemeDropdown) WithThemeClient(tc *theme.Client) *ThemeDropdown {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	d.theme = tc

	if d.Dropdown == nil {
		d.Dropdown = NewDropdown(d.themeItems())
	} else {
		d.Dropdown.SetItems(d.themeItems())
	}

	d.Dropdown.WithThemeClient(tc)
	d.syncSelection()

	return d
}

func (d *ThemeDropdown) Layout(gtx layout.Context, overlay *overlay.Overlay) layout.Dimensions {
	if d == nil {
		return layout.Dimensions{}
	}

	if d.Dropdown == nil {
		d.Dropdown = NewDropdown(d.themeItems()).WithThemeClient(d.theme)
	}

	// Keep selected item synced if another component changed the theme.
	current := d.currentThemeName()
	if !strings.EqualFold(current, d.lastThemeName) {
		d.syncSelection()
	}

	if d.theme != nil && d.theme.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	return d.Dropdown.Layout(gtx, overlay)
}

func (d *ThemeDropdown) syncSelection() {
	if d == nil || d.Dropdown == nil {
		return
	}

	current := d.currentThemeName()
	d.Dropdown.SelectItem(current)
	d.lastThemeName = current
}

func (d *ThemeDropdown) currentThemeName() string {
	if d == nil || d.theme == nil {
		return ""
	}

	return d.theme.GetCurrentThemeName()
}

func (d *ThemeDropdown) themeItems() []DropdownItem {
	tc := d.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	themes := tc.GetThemes()
	items := make([]DropdownItem, 0, len(themes))

	for _, t := range themes {
		if t == nil || strings.TrimSpace(t.Name) == "" {
			continue
		}

		items = append(items, DropdownItem{
			Label: t.Name,
			Value: t.Name,
		})
	}

	return items
}

func makeItemsClickable(n int) []widget.Clickable {
	if n <= 0 {
		return nil
	}

	return make([]widget.Clickable, n)
}
