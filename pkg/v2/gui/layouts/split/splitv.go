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

type SplitV struct {
	// Ratio keeps the current layout.
	// 0 is center, -1 completely to the left, 1 completely to the right.
	Ratio float32

	// Bar is the width for resizing the layout.
	Bar unit.Dp

	MinRatio float32
	MaxRatio float32

	// HideLeft hides the left pane.
	// If both HideLeft and HideRight are true, HideRight is ignored so one pane remains visible.
	HideLeft bool

	// HideRight hides the right pane.
	// If both HideLeft and HideRight are true, HideRight is ignored so one pane remains visible.
	HideRight bool

	drag   bool
	dragID pointer.ID
	dragX  float32
}

const defaultBarWidth = unit.Dp(10)

func (s *SplitV) Layout(gtx layout.Context, left, right layout.Widget) layout.Dimensions {
	if s == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	s.normalizeVisibility()

	if s.HideLeft {
		if right != nil {
			gtx.Constraints = layout.Exact(gtx.Constraints.Max)
			right(gtx)
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	if s.HideRight {
		if left != nil {
			gtx.Constraints = layout.Exact(gtx.Constraints.Max)
			left(gtx)
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	max := gtx.Constraints.Max
	if max.X <= 0 || max.Y <= 0 {
		return layout.Dimensions{Size: max}
	}

	proportion := (s.Ratio + 1) / 2
	leftSize := int(proportion*float32(max.X) - float32(bar)/2)

	if leftSize < 0 {
		leftSize = 0
	}

	rightOffset := leftSize + bar
	if rightOffset > max.X {
		rightOffset = max.X
	}

	rightSize := max.X - rightOffset
	if rightSize < 0 {
		rightSize = 0
	}

	s.layoutDragBar(gtx, leftSize, rightOffset)

	if left != nil {
		leftGtx := gtx
		leftGtx.Constraints = layout.Exact(image.Pt(leftSize, max.Y))
		left(leftGtx)
	}

	if right != nil {
		off := op.Offset(image.Pt(rightOffset, 0)).Push(gtx.Ops)
		rightGtx := gtx
		rightGtx.Constraints = layout.Exact(image.Pt(rightSize, max.Y))
		right(rightGtx)
		off.Pop()
	}

	return layout.Dimensions{Size: max}
}

func (s *SplitV) layoutDragBar(gtx layout.Context, leftSize, rightOffset int) {
	max := gtx.Constraints.Max

	// Important: Y max should be max.Y, not max.X.
	barRect := image.Rect(leftSize, 0, rightOffset, max.Y)
	area := clip.Rect(barRect).Push(gtx.Ops)

	event.Op(gtx.Ops, s)
	pointer.CursorColResize.Add(gtx.Ops)

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
			s.dragX = e.Position.X
			s.drag = true

		case pointer.Drag:
			if s.dragID != e.PointerID {
				break
			}

			deltaX := e.Position.X - s.dragX
			s.dragX = e.Position.X

			if max.X > 0 {
				deltaRatio := deltaX * 2 / float32(max.X)
				s.Ratio += deltaRatio
				s.recalcRatio()
			}

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

func (s *SplitV) SetHidden(leftHidden, rightHidden bool) {
	if s == nil {
		return
	}

	s.HideLeft = leftHidden
	s.HideRight = rightHidden
	s.normalizeVisibility()
}

func (s *SplitV) ShowBoth() {
	if s == nil {
		return
	}

	s.HideLeft = false
	s.HideRight = false
}

func (s *SplitV) ShowLeftOnly() {
	if s == nil {
		return
	}

	s.HideLeft = false
	s.HideRight = true
}

func (s *SplitV) ShowRightOnly() {
	if s == nil {
		return
	}

	s.HideLeft = true
	s.HideRight = false
}

func (s *SplitV) ToggleLeft() {
	if s == nil {
		return
	}

	s.HideLeft = !s.HideLeft
	s.normalizeVisibility()
}

func (s *SplitV) ToggleRight() {
	if s == nil {
		return
	}

	s.HideRight = !s.HideRight
	s.normalizeVisibility()
}

func (s *SplitV) LeftVisible() bool {
	if s == nil {
		return false
	}
	s.normalizeVisibility()
	return !s.HideLeft
}

func (s *SplitV) RightVisible() bool {
	if s == nil {
		return false
	}
	s.normalizeVisibility()
	return !s.HideRight
}

func (s *SplitV) normalizeVisibility() {
	if s == nil {
		return
	}

	// One side should always remain visible.
	// If both are hidden, prefer showing left.
	if s.HideLeft && s.HideRight {
		s.HideLeft = false
	}
}

func (s *SplitV) recalcRatio() {
	if s == nil {
		return
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
