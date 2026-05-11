package components

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type PerformanceMonitor struct {
	Title string

	// Enabled controls both sampling and rendering.
	Enabled bool

	// AutoInvalidate keeps the widget updating while visible.
	AutoInvalidate bool

	// SampleEvery controls how often FPS/memory snapshots are refreshed.
	SampleEvery time.Duration

	Radius unit.Dp
	Inset  unit.Dp
	Gap    unit.Dp

	theme *theme.Client

	lastFrame     time.Time
	sampleStarted time.Time
	frames        int

	fps     float64
	frameMS float64
	minMS   float64
	maxMS   float64
	heap    uint64
	sys     uint64
	goros   int
	gc      uint32
}

func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		Title:          "Performance",
		Enabled:        true,
		AutoInvalidate: true,
		SampleEvery:    500 * time.Millisecond,
		Radius:         unit.Dp(14),
		Inset:          unit.Dp(12),
		Gap:            unit.Dp(8),
		theme:          theme.DefaultThemeClient,
	}
}

func (p *PerformanceMonitor) WithThemeClient(tc *theme.Client) *PerformanceMonitor {
	if p == nil {
		return p
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	p.theme = tc
	return p
}

func (p *PerformanceMonitor) Snapshot() PerformanceSnapshot {
	if p == nil {
		return PerformanceSnapshot{}
	}
	return PerformanceSnapshot{
		FPS:        p.fps,
		FrameMS:    p.frameMS,
		MinFrameMS: p.minMS,
		MaxFrameMS: p.maxMS,
		HeapAlloc:  p.heap,
		Sys:        p.sys,
		Goroutines: p.goros,
		NumGC:      p.gc,
	}
}

type PerformanceSnapshot struct {
	FPS        float64
	FrameMS    float64
	MinFrameMS float64
	MaxFrameMS float64
	HeapAlloc  uint64
	Sys        uint64
	Goroutines int
	NumGC      uint32
}

func (p *PerformanceMonitor) Layout(gtx layout.Context) layout.Dimensions {
	if p == nil || !p.Enabled {
		return layout.Dimensions{}
	}

	now := time.Now()
	p.update(now)

	if p.AutoInvalidate {
		gtx.Execute(op.InvalidateCmd{})
	}

	tc := p.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		p.theme = tc
	}
	if tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	return utils.SurfaceOutlined(
		gtx,
		tokens.SurfaceAltNRGBA(),
		p.Radius,
		utils.SurfaceBorder{Color: tokens.BorderNRGBA(), Width: unit.Dp(1)},
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(p.Inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.label(gtx, typo, theme.TextRoleLabel, tokens.TextPrimaryNRGBA(), p.Title)
					}),
					layout.Rigid(layout.Spacer{Height: p.Gap}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.metric(gtx, typo, tokens.TextSecondaryNRGBA(), tokens.PrimaryNRGBA(), "FPS", fmt.Sprintf("%.0f", p.fps), clamp01(p.fps/120.0))
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.metric(gtx, typo, tokens.TextSecondaryNRGBA(), tokens.InfoNRGBA(), "Frame", fmt.Sprintf("%.2f ms", p.frameMS), clamp01(p.frameMS/33.33))
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.metric(gtx, typo, tokens.TextSecondaryNRGBA(), tokens.WarningNRGBA(), "Heap", formatBytes(p.heap), clamp01(float64(p.heap)/(512*1024*1024)))
					}),
					layout.Rigid(layout.Spacer{Height: p.Gap}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						text := fmt.Sprintf("min %.2fms  max %.2fms  goroutines %d  GC %d", p.minMS, p.maxMS, p.goros, p.gc)
						return p.label(gtx, typo, theme.TextRoleCaption, tokens.TextMutedNRGBA(), text)
					}),
				)
			})
		},
	)
}

func (p *PerformanceMonitor) update(now time.Time) {
	if p.SampleEvery <= 0 {
		p.SampleEvery = 500 * time.Millisecond
	}

	if p.sampleStarted.IsZero() {
		p.sampleStarted = now
		p.lastFrame = now
		p.minMS = 0
		p.maxMS = 0
		return
	}

	if !p.lastFrame.IsZero() {
		dtMS := float64(now.Sub(p.lastFrame).Microseconds()) / 1000.0
		if dtMS > 0 {
			p.frameMS = dtMS
			if p.minMS == 0 || dtMS < p.minMS {
				p.minMS = dtMS
			}
			if dtMS > p.maxMS {
				p.maxMS = dtMS
			}
		}
	}
	p.lastFrame = now
	p.frames++

	elapsed := now.Sub(p.sampleStarted)
	if elapsed < p.SampleEvery {
		return
	}

	p.fps = float64(p.frames) / elapsed.Seconds()
	p.frames = 0
	p.sampleStarted = now

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	p.heap = m.HeapAlloc
	p.sys = m.Sys
	p.goros = runtime.NumGoroutine()
	p.gc = m.NumGC
}

func (p *PerformanceMonitor) metric(
	gtx layout.Context,
	typo theme.TypographyTokens,
	textColor color.NRGBA,
	barColor color.NRGBA,
	name string,
	value string,
	amount float64,
) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return p.label(gtx, typo, theme.TextRoleCaption, textColor, name)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.label(gtx, typo, theme.TextRoleCaption, textColor, value)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.bar(gtx, barColor, amount)
		}),
	)
}

func (p *PerformanceMonitor) label(
	gtx layout.Context,
	typo theme.TypographyTokens,
	role theme.TextRole,
	col color.NRGBA,
	value string,
) layout.Dimensions {
	lbl := material.Body1(material.NewTheme(), value)
	lbl.Color = col
	lbl.Alignment = text.Middle
	theme.ApplyTypography(&lbl, typo, role)
	return lbl.Layout(gtx)
}

func (p *PerformanceMonitor) bar(gtx layout.Context, fill color.NRGBA, amount float64) layout.Dimensions {
	w := gtx.Constraints.Max.X
	if w <= 0 {
		w = gtx.Dp(unit.Dp(120))
	}
	h := gtx.Dp(unit.Dp(5))
	if h <= 0 {
		h = 5
	}

	amount = clamp01(amount)
	track := image.Rect(0, 0, w, h)
	filled := image.Rect(0, 0, int(float64(w)*amount), h)

	trackColor := fill
	trackColor.A = 45

	paint.FillShape(gtx.Ops, trackColor, clip.UniformRRect(track, h/2).Op(gtx.Ops))
	if filled.Dx() > 0 {
		paint.FillShape(gtx.Ops, fill, clip.UniformRRect(filled, h/2).Op(gtx.Ops))
	}

	return layout.Dimensions{Size: image.Pt(w, h)}
}

func clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

func formatBytes(v uint64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%d B", v)
	}
	div, exp := uint64(unit), 0
	for n := v / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(v)/float64(div), "KMGTPE"[exp])
}
