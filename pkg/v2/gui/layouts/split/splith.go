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

	// HideTop hides the top pane.
	// If both HideTop and HideBottom are true, HideBottom is ignored so one pane remains visible.
	HideTop bool

	// HideBottom hides the bottom pane.
	// If both HideTop and HideBottom are true, HideBottom is ignored so one pane remains visible.
	HideBottom bool

	drag   bool
	dragID pointer.ID
	dragY  float32
}

const defaultBarHeight = unit.Dp(10)

func (s *SplitH) Layout(gtx layout.Context, top, bottom layout.Widget) layout.Dimensions {
	if s == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	s.normalizeVisibility()

	if s.HideTop {
		if bottom != nil {
			bottomGtx := gtx
			bottomGtx.Constraints = layout.Exact(gtx.Constraints.Max)
			bottom(bottomGtx)
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	if s.HideBottom {
		if top != nil {
			topGtx := gtx
			topGtx.Constraints = layout.Exact(gtx.Constraints.Max)
			top(topGtx)
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarHeight)
	}

	max := gtx.Constraints.Max
	if max.X <= 0 || max.Y <= 0 {
		return layout.Dimensions{Size: max}
	}

	proportion := (s.Ratio + 1) / 2
	topSize := int(proportion*float32(max.Y) - float32(bar)/2)

	if topSize < 0 {
		topSize = 0
	}

	bottomOffset := topSize + bar
	if bottomOffset > max.Y {
		bottomOffset = max.Y
	}

	bottomSize := max.Y - bottomOffset
	if bottomSize < 0 {
		bottomSize = 0
	}

	s.layoutDragBar(gtx, topSize, bottomOffset)

	if top != nil {
		topGtx := gtx
		topGtx.Constraints = layout.Exact(image.Pt(max.X, topSize))
		top(topGtx)
	}

	if bottom != nil {
		off := op.Offset(image.Pt(0, bottomOffset)).Push(gtx.Ops)
		bottomGtx := gtx
		bottomGtx.Constraints = layout.Exact(image.Pt(max.X, bottomSize))
		bottom(bottomGtx)
		off.Pop()
	}

	return layout.Dimensions{Size: max}
}

func (s *SplitH) layoutDragBar(gtx layout.Context, topSize, bottomOffset int) {
	max := gtx.Constraints.Max

	barRect := image.Rect(
		0,
		topSize,
		max.X,
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

			if max.Y > 0 {
				deltaRatio := deltaY * 2 / float32(max.Y)
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

func (s *SplitH) SetHidden(topHidden, bottomHidden bool) {
	if s == nil {
		return
	}

	s.HideTop = topHidden
	s.HideBottom = bottomHidden
	s.normalizeVisibility()
}

func (s *SplitH) ShowBoth() {
	if s == nil {
		return
	}

	s.HideTop = false
	s.HideBottom = false
}

func (s *SplitH) ShowTopOnly() {
	if s == nil {
		return
	}

	s.HideTop = false
	s.HideBottom = true
}

func (s *SplitH) ShowBottomOnly() {
	if s == nil {
		return
	}

	s.HideTop = true
	s.HideBottom = false
}

func (s *SplitH) ToggleTop() {
	if s == nil {
		return
	}

	s.HideTop = !s.HideTop
	s.normalizeVisibility()
}

func (s *SplitH) ToggleBottom() {
	if s == nil {
		return
	}

	s.HideBottom = !s.HideBottom
	s.normalizeVisibility()
}

func (s *SplitH) TopVisible() bool {
	if s == nil {
		return false
	}

	s.normalizeVisibility()
	return !s.HideTop
}

func (s *SplitH) BottomVisible() bool {
	if s == nil {
		return false
	}

	s.normalizeVisibility()
	return !s.HideBottom
}

func (s *SplitH) normalizeVisibility() {
	if s == nil {
		return
	}

	// One side should always remain visible.
	// If both are hidden, prefer showing top.
	if s.HideTop && s.HideBottom {
		s.HideTop = false
	}
}

func (s *SplitH) recalcRatio() {
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
