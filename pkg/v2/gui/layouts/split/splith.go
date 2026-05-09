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

type SplitH struct {
	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the top, 1 completely to the bottom.
	Ratio float32

	// Bar is the height for resizing the layout.
	Bar unit.Dp

	MinRatio float32
	MaxRatio float32

	drag   bool
	dragID pointer.ID
	dragY  float32
}

const defaultBarHeight = unit.Dp(10)

func (s *SplitH) Layout(gtx layout.Context, top, bottom layout.Widget) layout.Dimensions {
	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarHeight)
	}

	proportion := (s.Ratio + 1) / 2

	topSize := int(proportion*float32(gtx.Constraints.Max.Y) - float32(bar))
	if topSize < 0 {
		topSize = 0
	}

	bottomOffset := topSize + bar
	bottomSize := gtx.Constraints.Max.Y - bottomOffset
	if bottomSize < 0 {
		bottomSize = 0
	}

	{ // handle input
		barRect := image.Rect(
			0,
			topSize,
			gtx.Constraints.Max.X,
			bottomOffset,
		)

		area := clip.Rect(barRect).Push(gtx.Ops)

		event.Op(gtx.Ops, s)
		pointer.CursorRowResize.Add(gtx.Ops)

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

				s.dragID = e.PointerID
				s.dragY = e.Position.Y
				s.drag = true

			case pointer.Drag:
				if s.dragID != e.PointerID {
					break
				}

				deltaY := e.Position.Y - s.dragY
				s.dragY = e.Position.Y

				deltaRatio := deltaY * 2 / float32(gtx.Constraints.Max.Y)
				s.Ratio += deltaRatio
				s.recalcRatio()

				if e.Priority < pointer.Grabbed {
					gtx.Execute(pointer.GrabCmd{
						Tag: s,
						ID:  s.dragID,
					})
				}

			case pointer.Release, pointer.Cancel:
				s.drag = false
			}
		}

		area.Pop()
	}

	{
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(
			gtx.Constraints.Max.X,
			topSize,
		))
		top(gtx)
	}

	{
		off := op.Offset(image.Pt(0, bottomOffset)).Push(gtx.Ops)
		gtx := gtx
		gtx.Constraints = layout.Exact(image.Pt(
			gtx.Constraints.Max.X,
			bottomSize,
		))
		bottom(gtx)
		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (s *SplitH) recalcRatio() {
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
