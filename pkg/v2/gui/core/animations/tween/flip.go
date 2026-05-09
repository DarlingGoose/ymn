package tween

import (
	"sync"
	"time"
)

type Flip struct {
	mu sync.Mutex

	curve    Curve
	duration time.Duration

	from float64
	to   float64

	value     float64
	startedAt time.Time
	running   bool
	expanded  bool
}

func NewFlip(duration time.Duration, curve Curve) *Flip {
	if curve.Fn == nil {
		curve = Linear
	}

	return &Flip{
		curve:    curve,
		duration: duration,
		value:    0,
		from:     0,
		to:       0,
	}
}

func (f *Flip) Toggle() {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	f.tickLocked(now)

	if f.expanded {
		f.startLocked(now, 0)
		f.expanded = false
		return
	}

	f.startLocked(now, 1)
	f.expanded = true
}

func (f *Flip) Expand() {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	f.tickLocked(now)

	f.expanded = true
	f.startLocked(now, 1)
}

func (f *Flip) Collapse() {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	f.tickLocked(now)

	f.expanded = false
	f.startLocked(now, 0)
}

func (f *Flip) SetExpanded(expanded bool) {
	if expanded {
		f.Expand()
	} else {
		f.Collapse()
	}
}

func (f *Flip) JumpExpanded(expanded bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.expanded = expanded
	f.running = false

	if expanded {
		f.value = 1
		f.from = 1
		f.to = 1
		return
	}

	f.value = 0
	f.from = 0
	f.to = 0
}

func (f *Flip) Value(now time.Time) (float64, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.tickLocked(now)

	return f.value, f.running
}

func (f *Flip) Running() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.running
}

func (f *Flip) Expanded() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.expanded
}

func (f *Flip) startLocked(now time.Time, target float64) {
	target = Clamp01(target)

	if f.duration <= 0 {
		f.value = target
		f.from = target
		f.to = target
		f.running = false
		return
	}

	f.from = f.value
	f.to = target
	f.startedAt = now
	f.running = true
}

func (f *Flip) tickLocked(now time.Time) {
	if !f.running {
		return
	}

	if f.duration <= 0 {
		f.value = f.to
		f.running = false
		return
	}

	progress := float64(now.Sub(f.startedAt)) / float64(f.duration)
	if progress >= 1 {
		f.value = f.to
		f.running = false
		return
	}

	if progress < 0 {
		progress = 0
	}

	curved := f.curve.At(progress)
	f.value = lerpFloat(f.from, f.to, curved)
	f.value = Clamp01(f.value)
}
