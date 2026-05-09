package tween

import (
	"image/color"
	"sync"
	"time"
)

type ColorTween struct {
	mu sync.Mutex

	curve    Curve
	duration time.Duration

	from color.NRGBA
	to   color.NRGBA

	value     color.NRGBA
	startedAt time.Time
	running   bool
}

func NewColorTween(duration time.Duration, curve Curve, initial color.NRGBA) *ColorTween {
	if curve.Fn == nil {
		curve = Linear
	}

	return &ColorTween{
		curve:    curve,
		duration: duration,
		from:     initial,
		to:       initial,
		value:    initial,
	}
}

func (c *ColorTween) AnimateTo(next color.NRGBA) {
	c.AnimateToAt(time.Now(), next)
}

func (c *ColorTween) AnimateToAt(now time.Time, next color.NRGBA) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.tickLocked(now)

	if c.to == next {
		return
	}

	c.from = c.value
	c.to = next
	c.startedAt = now
	c.running = true

	if c.duration <= 0 {
		c.value = next
		c.running = false
	}
}

func (c *ColorTween) JumpTo(next color.NRGBA) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.from = next
	c.to = next
	c.value = next
	c.running = false
}

func (c *ColorTween) Value(now time.Time) (color.NRGBA, bool) {
	if c == nil {
		return color.NRGBA{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.tickLocked(now)

	return c.value, c.running
}

func (c *ColorTween) Current() color.NRGBA {
	if c == nil {
		return color.NRGBA{}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.value
}

func (c *ColorTween) Target() color.NRGBA {
	if c == nil {
		return color.NRGBA{}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.to
}

func (c *ColorTween) Running() bool {
	if c == nil {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.running
}

func (c *ColorTween) SetDuration(duration time.Duration) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.duration = duration
}

func (c *ColorTween) SetCurve(curve Curve) {
	if c == nil {
		return
	}

	if curve.Fn == nil {
		curve = Linear
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.curve = curve
}

func (c *ColorTween) Stop(finish bool) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if finish {
		c.value = c.to
	} else {
		c.to = c.value
	}

	c.from = c.value
	c.running = false
}

func (c *ColorTween) tickLocked(now time.Time) {
	if !c.running {
		return
	}

	if c.duration <= 0 {
		c.value = c.to
		c.running = false
		return
	}

	progress := float64(now.Sub(c.startedAt)) / float64(c.duration)
	if progress >= 1 {
		c.value = c.to
		c.running = false
		return
	}

	if progress < 0 {
		progress = 0
	}

	c.value = Color(c.curve.At(progress), c.from, c.to)
}
