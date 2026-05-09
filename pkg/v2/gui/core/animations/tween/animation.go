package tween

import (
	"image"
	"sync/atomic"
	"time"

	"gioui.org/op"
)

type Animation struct {
	tween   *Tween
	running atomic.Bool
}

func NewAnimation(tween *Tween) *Animation {
	a := &Animation{tween: tween}

	if tween != nil && tween.Running() {
		a.running.Store(true)
	}

	return a
}

func (a *Animation) MoveTo(x, y int) {
	if a == nil || a.tween == nil {
		return
	}

	a.tween.MoveTo(x, y)
	a.running.Store(true)
}

func (a *Animation) JumpTo(x, y int) {
	if a == nil || a.tween == nil {
		return
	}

	a.tween.JumpTo(x, y)
	a.running.Store(false)
}

func (a *Animation) Tick(now time.Time) (image.Point, bool) {
	if a == nil || a.tween == nil {
		return image.Point{}, false
	}

	pt, running := a.tween.Tick(now)
	a.running.Store(running)

	return pt, running
}

func (a *Animation) Offset(ops *op.Ops, now time.Time) bool {
	pt, running := a.Tick(now)
	op.Offset(pt).Add(ops)
	return running
}

func (a *Animation) Layout(ops *op.Ops, now time.Time, fn func()) bool {
	if ops == nil {
		if fn != nil {
			fn()
		}
		return false
	}

	pt, running := a.Tick(now)

	stack := op.Offset(pt).Push(ops)
	if fn != nil {
		fn()
	}
	stack.Pop()

	return running
}
func (a *Animation) Active() bool {
	if a == nil {
		return false
	}

	return a.running.Load()
}
