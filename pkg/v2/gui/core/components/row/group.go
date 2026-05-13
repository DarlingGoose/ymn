package row

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

type Item struct {
	Row     *Row
	Control layout.Widget
}

type Group struct {
	Items []Item
	Gap   unit.Dp
}

func NewGroup(items ...Item) *Group {
	return &Group{
		Items: items,
		Gap:   unit.Dp(8),
	}
}

func Wrap(r *Row, control layout.Widget) Item {
	return Item{
		Row:     r,
		Control: control,
	}
}

func (g *Group) Layout(gtx layout.Context) layout.Dimensions {
	if g == nil {
		return layout.Dimensions{}
	}

	children := make([]layout.FlexChild, 0, len(g.Items)*2)

	for i := range g.Items {
		item := g.Items[i]
		if item.Row == nil || item.Control == nil {
			continue
		}

		if len(children) > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Height: g.Gap}.Layout))
		}

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return item.Row.Layout(gtx, item.Control)
		}))
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, children...)
}
