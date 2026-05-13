package animations

import (
	"image"
	"image/color"
	"math"
	"sync"

	"gioui.org/io/input"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type ProgressBar struct {
	rwMu     sync.RWMutex
	Max      float32
	Current  float32
	Width    int
	Height   int
	Color    color.NRGBA
	progress float32
}

func NewProgressBar(height, width int, max float32, color color.NRGBA) *ProgressBar {
	if height < 1 {
		height = 20
	}
	if width < 1 {
		width = 100
	}
	return &ProgressBar{
		Max:      max,
		Current:  0,
		Width:    height,
		Height:   width,
		Color:    color,
		progress: 0,
	}
}

func (b *ProgressBar) Draw(ops *op.Ops, source input.Source) {
	// Calculate how much of the progress bar to draw,
	// based on the current time.
	b.rwMu.Lock()
	defer b.rwMu.RUnlock()

	if b.progress < 1 {
		source.Execute(op.InvalidateCmd{})
	} else {
		b.progress = 1
	}

	width := float32(b.Width) * b.progress
	defer clip.Rect{Max: image.Pt(int(width), b.Height)}.Push(ops).Pop()
	paint.ColorOp{Color: b.Color}.Add(ops)
	paint.PaintOp{}.Add(ops)
}

func (b *ProgressBar) SetColor(color color.NRGBA) {
	b.rwMu.Lock()
	b.Color = color
	b.rwMu.Unlock()
}

func (b *ProgressBar) SetMax(max float32) {
	b.rwMu.Lock()
	if max <= 0 {
		b.rwMu.Unlock()
		return
	}
	if max < b.Current {
		b.Current = max
	}
	b.Max = max
	b.rwMu.Unlock()
	b.Calc()
}
func (b *ProgressBar) Add(amount float32) {
	b.rwMu.Lock()
	b.Current += amount
	b.rwMu.Unlock()
	b.Calc()
}

func (b *ProgressBar) Increment() {
	b.rwMu.Lock()
	if b.Current >= b.Max {
		b.rwMu.Unlock()
		return
	}
	b.Current++
	b.rwMu.Unlock()
	b.Calc()
}

func (b *ProgressBar) Calc() {
	if b.Max == 0 {
		return
	}
	b.rwMu.Lock()
	defer b.rwMu.RUnlock()
	b.progress = float32(math.Abs(float64(b.Current / b.Max)))
	if b.progress > 1 {
		b.progress = 1
	}
}
