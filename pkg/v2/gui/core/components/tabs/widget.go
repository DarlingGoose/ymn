package tabs

import "gioui.org/layout"

var tabLayoutIDCounter uint64

type Widget interface {
	Layout(gtx layout.Context) layout.Dimensions
}

type WidgetFunc func(gtx layout.Context) layout.Dimensions

func (fn WidgetFunc) Layout(gtx layout.Context) layout.Dimensions {
	if fn == nil {
		return layout.Dimensions{}
	}

	return fn(gtx)
}
