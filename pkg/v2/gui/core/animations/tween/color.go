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
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.tickLocked(now)

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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.from = next
	c.to = next
	c.value = next
	c.running = false
}

func (c *ColorTween) Value(now time.Time) (color.NRGBA, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tickLocked(now)

	return c.value, c.running
}

func (c *ColorTween) Running() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.running
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
