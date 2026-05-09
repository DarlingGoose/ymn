package tween

import (
	"image"
	"sync"
	"time"
)

type Tween struct {
	mu sync.Mutex

	xOffset Offset
	yOffset Offset

	curve    Curve
	duration time.Duration

	startedAt time.Time
	running   bool
}

func New(duration time.Duration, curve Curve) *Tween {
	if curve.Fn == nil {
		curve = Linear
	}

	return &Tween{
		curve:    curve,
		duration: duration,
	}
}

func (t *Tween) Point() image.Point {
	return image.Pt(
		t.xOffset.Current(),
		t.yOffset.Current(),
	)
}

func (t *Tween) SetCurve(curve Curve) {
	if curve.Fn == nil {
		curve = Linear
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.curve = curve
}

func (t *Tween) SetDuration(duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.duration = duration
}

// MoveTo starts a new animation from the current offset to the given offset.
func (t *Tween) MoveTo(x, y int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	// If already running, first resolve the current animated value.
	t.tickLocked(now)

	t.xOffset.startingOffset.Store(t.xOffset.offset.Load())
	t.yOffset.startingOffset.Store(t.yOffset.offset.Load())

	t.xOffset.endingOffset.Store(int64(x))
	t.yOffset.endingOffset.Store(int64(y))

	t.startedAt = now
	t.running = true

	if t.duration <= 0 {
		t.xOffset.offset.Store(int64(x))
		t.yOffset.offset.Store(int64(y))
		t.running = false
	}
}

// JumpTo immediately sets the current offset without animation.
func (t *Tween) JumpTo(x, y int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.xOffset.offset.Store(int64(x))
	t.yOffset.offset.Store(int64(y))

	t.xOffset.startingOffset.Store(int64(x))
	t.yOffset.startingOffset.Store(int64(y))

	t.xOffset.endingOffset.Store(int64(x))
	t.yOffset.endingOffset.Store(int64(y))

	t.running = false
}

// Tick advances the tween.
// The bool return value tells you whether it is still animating.
func (t *Tween) Tick(now time.Time) (image.Point, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.tickLocked(now)

	return t.Point(), t.running
}

func (t *Tween) Running() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.running
}

func (t *Tween) tickLocked(now time.Time) {
	if !t.running {
		return
	}

	if t.duration <= 0 {
		t.finishLocked()
		return
	}

	progress := float64(now.Sub(t.startedAt)) / float64(t.duration)
	if progress >= 1 {
		t.finishLocked()
		return
	}

	if progress < 0 {
		progress = 0
	}

	curved := t.curve.At(progress)

	x := lerpInt64(
		t.xOffset.startingOffset.Load(),
		t.xOffset.endingOffset.Load(),
		curved,
	)

	y := lerpInt64(
		t.yOffset.startingOffset.Load(),
		t.yOffset.endingOffset.Load(),
		curved,
	)

	t.xOffset.offset.Store(x)
	t.yOffset.offset.Store(y)
}

func (t *Tween) finishLocked() {
	t.xOffset.offset.Store(t.xOffset.endingOffset.Load())
	t.yOffset.offset.Store(t.yOffset.endingOffset.Load())
	t.running = false
}
