package grid

import (
	"gioui.org/layout"
)

type ScrollGrid struct {
	Grid Grid
	List layout.List
}

func NewScrollGrid() *ScrollGrid {
	return &ScrollGrid{
		List: layout.List{
			Axis: layout.Vertical,
		},
	}
}

func (g *ScrollGrid) Layout(gtx layout.Context, count int, widget Widget) layout.Dimensions {
	if g == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	if count <= 0 || widget == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	return g.Grid.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		cols := g.Grid.columnCount(gtx)
		rows := rowCount(count, cols)
		gap := g.Grid.gapDp(gtx)

		return g.List.Layout(gtx, rows, func(gtx layout.Context, row int) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return g.Grid.layoutRow(gtx, count, cols, row, widget)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if row >= rows-1 {
						return layout.Dimensions{}
					}

					return layout.Spacer{Height: gap}.Layout(gtx)
				}),
			)
		})
	})
}
