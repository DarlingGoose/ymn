package grid

import "gioui.org/layout"

type ItemWidget[T any] func(gtx layout.Context, item T, index int) layout.Dimensions

func LayoutSlice[T any](
	gtx layout.Context,
	g *Grid,
	items []T,
	widget ItemWidget[T],
) layout.Dimensions {
	if g == nil {
		g = &Grid{}
	}

	return g.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
		return widget(gtx, items[index], index)
	})
}

func LayoutScrollSlice[T any](
	gtx layout.Context,
	g *ScrollGrid,
	items []T,
	widget ItemWidget[T],
) layout.Dimensions {
	if g == nil {
		g = NewScrollGrid()
	}

	return g.Layout(gtx, len(items), func(gtx layout.Context, index int) layout.Dimensions {
		return widget(gtx, items[index], index)
	})
}
