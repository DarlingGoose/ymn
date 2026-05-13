package tween

import (
	"math"
	"sync"
	"time"
)

type RotationMode int

const (
	RotationStopped RotationMode = iota
	RotationTweening
	RotationContinuous
)

type RotationTween struct {
	mu sync.Mutex

	curve    Curve
	duration time.Duration

	from float64
	to   float64

	value     float64
	startedAt time.Time
	running   bool

	mode RotationMode

	// Used by continuous mode.
	speedRadPerSecond float64
	lastTick          time.Time
}

func NewRotationTween(duration time.Duration, curve Curve, initialRadians float64) *RotationTween {
	if curve.Fn == nil {
		curve = Linear
	}

	return &RotationTween{
		curve:    curve,
		duration: duration,
		from:     normalizeRadians(initialRadians),
		to:       normalizeRadians(initialRadians),
		value:    normalizeRadians(initialRadians),
		mode:     RotationStopped,
	}
}

func NewRotationTweenDeg(duration time.Duration, curve Curve, initialDegrees float64) *RotationTween {
	return NewRotationTween(duration, curve, Deg(initialDegrees))
}

func Deg(v float64) float64 {
	return v * math.Pi / 180
}

func RadToDeg(v float64) float64 {
	return v * 180 / math.Pi
}

func (r *RotationTween) AnimateToRadians(next float64) {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.tickLocked(now)

	next = normalizeRadians(next)

	if almostEqualRadians(r.value, next) {
		r.value = next
		r.to = next
		r.running = false
		r.mode = RotationStopped
		return
	}

	r.from = r.value
	r.to = next
	r.startedAt = now
	r.running = true
	r.mode = RotationTweening

	if r.duration <= 0 {
		r.value = next
		r.running = false
		r.mode = RotationStopped
	}
}

func (r *RotationTween) AnimateToDegrees(next float64) {
	r.AnimateToRadians(Deg(next))
}

func (r *RotationTween) JumpToRadians(next float64) {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	next = normalizeRadians(next)

	r.from = next
	r.to = next
	r.value = next
	r.running = false
	r.mode = RotationStopped
	r.lastTick = time.Time{}
}

func (r *RotationTween) JumpToDegrees(next float64) {
	r.JumpToRadians(Deg(next))
}

func (r *RotationTween) Value(now time.Time) (float64, bool) {
	if r == nil {
		return 0, false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tickLocked(now)

	return r.value, r.running
}

func (r *RotationTween) Degrees(now time.Time) (float64, bool) {
	v, running := r.Value(now)
	return RadToDeg(v), running
}

func (r *RotationTween) Running() bool {
	if r == nil {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.running
}

func (r *RotationTween) Mode() RotationMode {
	if r == nil {
		return RotationStopped
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.mode
}

func (r *RotationTween) tickLocked(now time.Time) {
	if !r.running {
		return
	}

	switch r.mode {
	case RotationContinuous:
		r.tickContinuousLocked(now)

	case RotationTweening:
		r.tickTweenLocked(now)

	default:
		r.running = false
		r.mode = RotationStopped
	}
}

func (r *RotationTween) tickContinuousLocked(now time.Time) {
	if r.lastTick.IsZero() {
		r.lastTick = now
		return
	}

	dt := now.Sub(r.lastTick).Seconds()
	if dt <= 0 {
		return
	}

	r.value = normalizeRadians(r.value + r.speedRadPerSecond*dt)
	r.lastTick = now
}

func (r *RotationTween) tickTweenLocked(now time.Time) {
	if r.duration <= 0 {
		r.value = r.to
		r.running = false
		r.mode = RotationStopped
		return
	}

	progress := float64(now.Sub(r.startedAt)) / float64(r.duration)
	if progress >= 1 {
		r.value = r.to
		r.running = false
		r.mode = RotationStopped
		return
	}

	if progress < 0 {
		progress = 0
	}

	curved := r.curve.At(progress)
	r.value = normalizeRadians(lerpFloat(r.from, r.to, curved))
}

func normalizeRadians(v float64) float64 {
	v = math.Mod(v, math.Pi*2)
	if v < 0 {
		v += math.Pi * 2
	}
	return v
}

func almostEqualRadians(a, b float64) bool {
	const epsilon = 0.000001
	return math.Abs(normalizeRadians(a)-normalizeRadians(b)) < epsilon
}

func (r *RotationTween) StartContinuousRadiansPerSecond(speed float64) {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.tickLocked(now)

	r.speedRadPerSecond = speed
	r.lastTick = now
	r.running = true
	r.mode = RotationContinuous
}

func (r *RotationTween) StartContinuousDegreesPerSecond(speed float64) {
	r.StartContinuousRadiansPerSecond(Deg(speed))
}

func (r *RotationTween) Stop(reset bool) {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.tickLocked(now)

	r.running = false
	r.mode = RotationStopped
	r.lastTick = time.Time{}

	if reset {
		r.value = 0
		r.from = 0
		r.to = 0
	}
}

func (r *RotationTween) ToggleContinuousDegreesPerSecond(speed float64, resetOnStop bool) {
	if r == nil {
		return
	}

	r.mu.Lock()
	running := r.running && r.mode == RotationContinuous
	r.mu.Unlock()

	if running {
		r.Stop(resetOnStop)
		return
	}

	r.StartContinuousDegreesPerSecond(speed)
}
