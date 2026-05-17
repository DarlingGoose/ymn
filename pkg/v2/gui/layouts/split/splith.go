package split

import (
	"image"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
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

	// AnimationDuration controls how long pane visibility changes take.
	// If zero, a default duration is used. Negative values disable animation.
	AnimationDuration time.Duration

	drag           bool
	dragID         pointer.ID
	dragY          float32
	sizeTween      *tween.Tween
	sizeTarget     int
	sizeVisibility int
	sizeReady      bool
	lastTopSize    int
	lastBarSize    int
	lastMaxHeight  int
}

const defaultBarHeight = unit.Dp(10)
const defaultVisibilityAnimationDuration = 180 * time.Millisecond

func (s *SplitH) Layout(gtx layout.Context, top, bottom layout.Widget) layout.Dimensions {
	if s == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	s.normalizeVisibility()

	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarHeight)
	}

	max := gtx.Constraints.Max
	if max.X <= 0 || max.Y <= 0 {
		return layout.Dimensions{Size: max}
	}

	targetTopSize, targetBarSize := s.targetSizes(max.Y, bar)
	visibility := s.visibilityState()
	topSize, barSize, running := s.animatedSizes(targetTopSize, targetBarSize, max.Y, visibility)
	if running {
		gtx.Execute(op.InvalidateCmd{})
	}

	if !running && s.HideTop {
		s.lastTopSize = topSize
		s.lastBarSize = barSize
		s.lastMaxHeight = max.Y
		if bottom != nil {
			bottomGtx := gtx
			bottomGtx.Constraints = layout.Exact(gtx.Constraints.Max)
			bottom(bottomGtx)
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	if !running && s.HideBottom {
		s.lastTopSize = topSize
		s.lastBarSize = barSize
		s.lastMaxHeight = max.Y
		if top != nil {
			topGtx := gtx
			topGtx.Constraints = layout.Exact(gtx.Constraints.Max)
			top(topGtx)
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	bottomOffset := topSize + barSize
	if bottomOffset > max.Y {
		bottomOffset = max.Y
	}

	bottomSize := max.Y - bottomOffset
	if bottomSize < 0 {
		bottomSize = 0
	}

	if barSize > 0 && !s.HideTop && !s.HideBottom {
		s.layoutDragBar(gtx, topSize, bottomOffset)
	}

	if top != nil && topSize > 0 {
		topGtx := gtx
		topGtx.Constraints = layout.Exact(image.Pt(max.X, topSize))
		top(topGtx)
	}

	if bottom != nil && bottomSize > 0 {
		off := op.Offset(image.Pt(0, bottomOffset)).Push(gtx.Ops)
		bottomGtx := gtx
		bottomGtx.Constraints = layout.Exact(image.Pt(max.X, bottomSize))
		bottom(bottomGtx)
		off.Pop()
	}

	s.lastTopSize = topSize
	s.lastBarSize = barSize
	s.lastMaxHeight = max.Y

	return layout.Dimensions{Size: max}
}

func (s *SplitH) targetSizes(maxY, bar int) (topSize, barSize int) {
	if s == nil {
		return 0, 0
	}

	if s.HideTop {
		return 0, 0
	}

	if s.HideBottom {
		return maxY, 0
	}

	proportion := (s.Ratio + 1) / 2
	topSize = int(proportion*float32(maxY) - float32(bar)/2)

	if topSize < 0 {
		topSize = 0
	}

	if topSize > maxY-bar {
		topSize = maxY - bar
	}
	if topSize < 0 {
		topSize = 0
	}

	return topSize, bar
}

func (s *SplitH) animatedSizes(targetTopSize, targetBarSize, maxY, visibility int) (topSize, barSize int, running bool) {
	if s == nil {
		return targetTopSize, targetBarSize, false
	}

	if s.AnimationDuration < 0 {
		s.sizeReady = true
		s.sizeTarget = targetTopSize
		s.sizeVisibility = visibility
		return targetTopSize, targetBarSize, false
	}

	if s.sizeTween == nil {
		duration := s.AnimationDuration
		if duration == 0 {
			duration = defaultVisibilityAnimationDuration
		}
		s.sizeTween = tween.New(duration, tween.EaseInOutCubic)
	}

	startTopSize := s.lastTopSize
	if s.lastMaxHeight > 0 && s.lastMaxHeight != maxY {
		startTopSize = scaleInt(startTopSize, maxY, s.lastMaxHeight)
	}

	if !s.sizeReady {
		s.sizeTween.JumpTo(0, targetTopSize)
		s.sizeReady = true
		s.sizeTarget = targetTopSize
		s.sizeVisibility = visibility
		return targetTopSize, targetBarSize, false
	}

	if s.sizeTarget != targetTopSize || s.sizeVisibility != visibility {
		if s.sizeVisibility == visibility && !s.sizeTween.Running() {
			s.sizeTween.JumpTo(0, targetTopSize)
		} else {
			s.sizeTween.JumpTo(0, startTopSize)
			s.sizeTween.MoveTo(0, targetTopSize)
		}
		s.sizeTarget = targetTopSize
		s.sizeVisibility = visibility
	}

	pt, running := s.sizeTween.Tick(time.Now())
	topSize = pt.Y
	if topSize < 0 {
		topSize = 0
	}
	if topSize > maxY {
		topSize = maxY
	}

	barSize = targetBarSize
	if running && (s.lastBarSize > 0 || targetBarSize > 0) {
		progress := 0
		if targetTopSize > startTopSize {
			progress = topSize - startTopSize
			if progress < 0 {
				progress = -progress
			}
		} else {
			progress = startTopSize - topSize
			if progress < 0 {
				progress = -progress
			}
		}
		total := targetTopSize - startTopSize
		if total < 0 {
			total = -total
		}
		if total > 0 {
			barSize = s.lastBarSize + (targetBarSize-s.lastBarSize)*progress/total
		}
	}
	if barSize < 0 {
		barSize = 0
	}

	return topSize, barSize, running
}

func (s *SplitH) visibilityState() int {
	if s == nil {
		return 0
	}
	switch {
	case s.HideTop:
		return 1
	case s.HideBottom:
		return 2
	default:
		return 3
	}
}

func scaleInt(value, next, prev int) int {
	if prev <= 0 {
		return value
	}
	return value * next / prev
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
