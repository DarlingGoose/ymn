package examples

import (
	"gioui.org/layout"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/tabs"
)

type TabApp struct {
	Tabs *tabs.Layout
}

func NewTabApp() *TabApp {
	return &TabApp{
		Tabs: tabs.New(
			tabs.NewTabFunc("settings", "Settings", "lucide:settings", func(gtx layout.Context) layout.Dimensions {
				// settings page
				return layout.Dimensions{}
			}),
			tabs.NewTabFunc("themes", "Themes", "lucide:palette", func(gtx layout.Context) layout.Dimensions {
				// themes page
				return layout.Dimensions{}
			}),
			tabs.NewTabFunc("about", "About", "lucide:info", func(gtx layout.Context) layout.Dimensions {
				// about page
				return layout.Dimensions{}
			}),
		),
	}
}

func (a *TabApp) Layout(gtx layout.Context) layout.Dimensions {
	a.Tabs.Update(gtx)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Later your topbar/sidebar can call a.Tabs.Buttons()
			// and render each Button.Clickable.
			return layout.Dimensions{}
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return a.Tabs.Layout(gtx)
		}),
	)
}
