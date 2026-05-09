package grid

import (
	"image"

	"gioui.org/layout"
	"gioui.org/unit"
)

const (
	defaultMinCellWidth = unit.Dp(160)
	defaultGap          = unit.Dp(8)
)

type Widget func(gtx layout.Context, index int) layout.Dimensions

type Grid struct {
	// Columns forces a fixed column count when > 0.
	// If Columns <= 0, the grid calculates columns from MinCellWidth.
	Columns int

	// MinColumns and MaxColumns clamp responsive column count.
	// Use 0 to disable either limit.
	MinColumns int
	MaxColumns int

	// MinCellWidth is used when Columns <= 0.
	MinCellWidth unit.Dp

	// Gap is the spacing between cells horizontally and vertically.
	Gap unit.Dp

	// Inset wraps the whole grid.
	Inset layout.Inset

	// StretchLastRow controls whether incomplete last rows stretch items
	// across all available columns.
	//
	// false:
	//   last row items keep normal column width.
	//
	// true:
	//   last row items expand to fill the row.
	StretchLastRow bool
}

func (g *Grid) gapDp(gtx layout.Context) unit.Dp {
	if g.Gap < 0 {
		return 0
	}

	if g.Gap == 0 {
		return defaultGap
	}

	return g.Gap
}

func (g *Grid) gapPx(gtx layout.Context) int {
	return gtx.Dp(g.gapDp(gtx))
}
func (g *Grid) Layout(gtx layout.Context, count int, widget Widget) layout.Dimensions {
	if count <= 0 || widget == nil {
		return g.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
		})
	}

	return g.Inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		cols := g.columnCount(gtx)
		rows := rowCount(count, cols)
		gap := g.gapDp(gtx)

		children := make([]layout.FlexChild, 0, rows*2)

		for row := 0; row < rows; row++ {
			row := row

			if row > 0 && gap > 0 {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: gap}.Layout(gtx)
				}))
			}

			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return g.layoutRow(gtx, count, cols, row, widget)
			}))
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx, children...)
	})
}

func (g *Grid) layoutRow(
	gtx layout.Context,
	count int,
	cols int,
	row int,
	widget Widget,
) layout.Dimensions {
	start := row * cols
	end := start + cols
	if end > count {
		end = count
	}

	itemsInRow := end - start
	if itemsInRow <= 0 {
		return layout.Dimensions{}
	}

	renderCols := cols
	if g.StretchLastRow && itemsInRow < cols {
		renderCols = itemsInRow
	}

	gapPx := g.gapPx(gtx)
	gapDp := g.gapDp(gtx)

	cellWidth := cellWidth(gtx.Constraints.Max.X, renderCols, gapPx)

	children := make([]layout.FlexChild, 0, itemsInRow*2)

	for col := 0; col < itemsInRow; col++ {
		index := start + col

		if col > 0 && gapDp > 0 {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Spacer{Width: gapDp}.Layout(gtx)
			}))
		}

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = cellWidth
			gtx.Constraints.Max.X = cellWidth
			gtx.Constraints.Min.Y = 0

			return widget(gtx, index)
		}))
	}

	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx, children...)
}

func (g *Grid) columnCount(gtx layout.Context) int {
	maxX := gtx.Constraints.Max.X
	if maxX <= 0 {
		return 1
	}

	if g.Columns > 0 {
		return clampColumns(g.Columns, g.MinColumns, g.MaxColumns)
	}

	minCellWidth := gtx.Dp(g.MinCellWidth)
	if minCellWidth <= 1 {
		minCellWidth = gtx.Dp(defaultMinCellWidth)
	}

	gap := g.gapPx(gtx)

	cols := maxX / minCellWidth
	if cols < 1 {
		cols = 1
	}

	for cols > 1 {
		totalGap := gap * (cols - 1)
		if cols*minCellWidth+totalGap <= maxX {
			break
		}
		cols--
	}

	return clampColumns(cols, g.MinColumns, g.MaxColumns)
}

func rowCount(count int, cols int) int {
	if cols <= 0 {
		cols = 1
	}

	rows := count / cols
	if count%cols != 0 {
		rows++
	}

	return rows
}

func cellWidth(maxX int, cols int, gap int) int {
	if cols <= 1 {
		return maxX
	}

	totalGap := gap * (cols - 1)
	available := maxX - totalGap
	if available <= 0 {
		return 0
	}

	return available / cols
}

func clampColumns(cols int, minCols int, maxCols int) int {
	if cols < 1 {
		cols = 1
	}

	if minCols > 0 && cols < minCols {
		cols = minCols
	}

	if maxCols > 0 && cols > maxCols {
		cols = maxCols
	}

	if cols < 1 {
		return 1
	}

	return cols
}
