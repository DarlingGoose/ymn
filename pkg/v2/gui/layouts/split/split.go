package split

import (
	"image"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

type Split struct {
	// Axis controls the split direction.
	//
	// layout.Horizontal:
	//   first widget is left, second widget is right.
	//
	// layout.Vertical:
	//   first widget is top, second widget is bottom.
	Axis layout.Axis

	// Ratio keeps the current layout.
	// 0 is center.
	// -1 gives all space to the first widget.
	// 1 gives all space to the second widget.
	Ratio float32

	// Bar is the divider size.
	// For Horizontal, this is width.
	// For Vertical, this is height.
	Bar unit.Dp

	// Optional ratio limits.
	// Use values in range [-1, 1].
	MinRatio float32
	MaxRatio float32

	drag    bool
	dragID  pointer.ID
	dragPos float32
}

const defaultSplitBarSize = unit.Dp(10)

func (s *Split) Layout(gtx layout.Context, first, second layout.Widget) layout.Dimensions {
	switch s.Axis {
	case layout.Vertical:
		return s.layoutVertical(gtx, first, second)
	default:
		return s.layoutHorizontal(gtx, first, second)
	}
}

func (s *Split) ToggleAxis() {
	if s.Axis == layout.Horizontal {
		s.Axis = layout.Vertical
		return
	}

	s.Axis = layout.Horizontal
}
func (s *Split) layoutHorizontal(gtx layout.Context, left, right layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultSplitBarSize)
	}

	maxX := gtx.Constraints.Max.X
	maxY := gtx.Constraints.Max.Y

	proportion := (s.Ratio + 1) / 2

	leftSize := int(proportion*float32(maxX) - float32(bar)/2)
	if leftSize < 0 {
		leftSize = 0
	}

	rightOffset := leftSize + bar
	rightSize := maxX - rightOffset
	if rightSize < 0 {
		rightSize = 0
	}

	s.handleInput(
		gtx,
		image.Rect(leftSize, 0, rightOffset, maxY),
		float32(maxX),
		func(e pointer.Event) float32 {
			return e.Position.X
		},
		func() {
			pointer.CursorColResize.Add(gtx.Ops)
		},
	)

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(leftSize, maxY))
		left(gtx)
	}

	{
		off := op.Offset(image.Pt(rightOffset, 0)).Push(gtx.Ops)
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(rightSize, maxY))
		right(gtx)
		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}
func (s *Split) layoutVertical(gtx layout.Context, top, bottom layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultSplitBarSize)
	}

	maxX := gtx.Constraints.Max.X
	maxY := gtx.Constraints.Max.Y

	proportion := (s.Ratio + 1) / 2

	topSize := int(proportion*float32(maxY) - float32(bar)/2)
	if topSize < 0 {
		topSize = 0
	}

	bottomOffset := topSize + bar
	bottomSize := maxY - bottomOffset
	if bottomSize < 0 {
		bottomSize = 0
	}

	s.handleInput(
		gtx,
		image.Rect(0, topSize, maxX, bottomOffset),
		float32(maxY),
		func(e pointer.Event) float32 {
			return e.Position.Y
		},
		func() {
			pointer.CursorRowResize.Add(gtx.Ops)
		},
	)

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(maxX, topSize))
		top(gtx)
	}

	{
		off := op.Offset(image.Pt(0, bottomOffset)).Push(gtx.Ops)
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(maxX, bottomSize))
		bottom(gtx)
		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}
func (s *Split) handleInput(
	gtx layout.Context,
	barRect image.Rectangle,
	totalSize float32,
	pos func(pointer.Event) float32,
	cursor func(),
) {
	if totalSize <= 0 {
		return
	}

	area := clip.Rect(barRect).Push(gtx.Ops)
	defer area.Pop()

	event.Op(gtx.Ops, s)
	cursor()

	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: s,
			Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel,
		})
		if !ok {
			break
		}

		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}

		switch e.Kind {
		case pointer.Press:
			if s.drag {
				break
			}

			s.drag = true
			s.dragID = e.PointerID
			s.dragPos = pos(e)

		case pointer.Drag:
			if s.dragID != e.PointerID {
				break
			}

			newPos := pos(e)
			delta := newPos - s.dragPos
			s.dragPos = newPos

			s.Ratio += delta * 2 / totalSize
			s.recalcRatio()

			if e.Priority < pointer.Grabbed {
				gtx.Execute(pointer.GrabCmd{
					Tag: s,
					ID:  s.dragID,
				})
			}

		case pointer.Release, pointer.Cancel:
			if s.dragID == e.PointerID {
				s.drag = false
			}
		}
	}
}
func (s *Split) recalcRatio() {
	if s.Ratio < -1 {
		s.Ratio = -1
	}

	if s.Ratio > 1 {
		s.Ratio = 1
	}

	if s.MaxRatio >= -1 && s.MaxRatio <= 1 {
		if s.Ratio > s.MaxRatio {
			s.Ratio = s.MaxRatio
		}
	}

	if s.MinRatio >= -1 && s.MinRatio <= 1 {
		if s.Ratio < s.MinRatio {
			s.Ratio = s.MinRatio
		}
	}
}
