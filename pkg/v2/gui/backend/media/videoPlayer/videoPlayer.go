package videoplayer

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"os/exec"
	"sync"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

type Player struct {
	Path   string
	Width  int
	Height int
	FPS    int

	PlayAudio bool
	Volume    float32

	mu sync.RWMutex

	running bool
	paused  bool
	seekTo  time.Duration
	playFrom time.Duration
	startedAt time.Time

	frameA *image.RGBA
	frameB *image.RGBA
	front  *image.RGBA
	back   *image.RGBA

	imageOp paint.ImageOp
	hasOp   bool

	cancel context.CancelFunc

	ffmpeg *exec.Cmd
	ffplay *exec.Cmd

	Invalidate func()
}

func New(path string, width, height, fps int, invalidate func()) *Player {
	if fps <= 0 {
		fps = 30
	}

	a := image.NewRGBA(image.Rect(0, 0, width, height))
	b := image.NewRGBA(image.Rect(0, 0, width, height))

	return &Player{
		Path:       path,
		Width:      width,
		Height:     height,
		FPS:        fps,
		Volume:     0.8,
		frameA:     a,
		frameB:     b,
		front:      a,
		back:       b,
		Invalidate: invalidate,
	}
}

func (p *Player) Start(ctx context.Context) {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	start := p.seekTo
	p.cancel = cancel
	p.running = true
	p.paused = false
	p.playFrom = start
	p.startedAt = time.Now()
	p.mu.Unlock()

	go p.videoLoop(ctx)

	if p.PlayAudio {
		go p.startAudio(ctx, start)
	}
}

func (p *Player) Stop() {
	p.mu.Lock()
	p.seekTo = 0
	p.playFrom = 0
	p.startedAt = time.Time{}
	p.hasOp = false
	if p.cancel != nil {
		p.cancel()
	}

	if p.ffmpeg != nil && p.ffmpeg.Process != nil {
		_ = p.ffmpeg.Process.Kill()
	}

	if p.ffplay != nil && p.ffplay.Process != nil {
		_ = p.ffplay.Process.Kill()
	}

	p.running = false
	p.paused = false
	p.cancel = nil
	p.ffmpeg = nil
	p.ffplay = nil
	p.mu.Unlock()
}

func (p *Player) Pause() {
	p.mu.Lock()
	if p.running {
		p.seekTo = p.positionLocked(time.Now())
		p.playFrom = p.seekTo
	}
	p.paused = true
	p.running = false
	p.startedAt = time.Time{}
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	if p.ffmpeg != nil && p.ffmpeg.Process != nil {
		_ = p.ffmpeg.Process.Kill()
		p.ffmpeg = nil
	}
	if p.ffplay != nil && p.ffplay.Process != nil {
		_ = p.ffplay.Process.Kill()
		p.ffplay = nil
	}
	p.mu.Unlock()
}

func (p *Player) Resume(ctx context.Context) {
	p.mu.Lock()
	running := p.running
	p.mu.Unlock()

	if running {
		return
	}

	p.Start(ctx)
}

func (p *Player) Seek(ctx context.Context, d time.Duration) {
	p.mu.Lock()
	wasPlaying := p.running && !p.paused
	p.stopProcessesLocked()
	p.seekTo = d
	p.playFrom = d
	p.startedAt = time.Time{}
	p.running = false
	p.paused = !wasPlaying
	p.mu.Unlock()

	if wasPlaying {
		p.Start(ctx)
	}
}

func (p *Player) SetPosition(d time.Duration) {
	p.mu.Lock()
	p.seekTo = d
	p.mu.Unlock()
}

func (p *Player) Position() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.positionLocked(time.Now())
}

func (p *Player) Running() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

func (p *Player) SetVolume(v float32) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	p.mu.Lock()
	pos := p.positionLocked(time.Now())
	playing := p.running && !p.paused
	p.Volume = v
	p.mu.Unlock()

	if playing && p.PlayAudio {
		p.restartAudio(context.Background(), pos)
	}
}

func (p *Player) videoLoop(ctx context.Context) {
	for {
		seek := p.currentSeek()

		err := p.runFFmpeg(ctx, seek)
		if err != nil {
			p.setStopped()
			return
		}

		select {
		case <-ctx.Done():
			p.setStopped()
			return
		default:
			p.setStopped()
			return
		}
	}
}

func (p *Player) runFFmpeg(ctx context.Context, seek time.Duration) error {
	args := []string{
		"-hide_banner",
		"-loglevel", "error",
	}

	if seek > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", seek.Seconds()))
	}

	args = append(args,
		"-re",
		"-i", p.Path,
		"-an",
		"-vf",
		fmt.Sprintf(
			"fps=%d,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,format=rgba",
			p.FPS,
			p.Width,
			p.Height,
			p.Width,
			p.Height,
		),
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"pipe:1",
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	p.mu.Lock()
	p.ffmpeg = cmd
	p.mu.Unlock()

	frameSize := p.Width * p.Height * 4
	frameDelay := time.Second / time.Duration(p.FPS)
	ticker := time.NewTicker(frameDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return ctx.Err()
		default:
		}

		p.mu.RLock()
		paused := p.paused
		p.mu.RUnlock()

		if paused {
			time.Sleep(20 * time.Millisecond)
			continue
		}

		p.mu.Lock()
		back := p.back
		p.mu.Unlock()

		_, err := io.ReadFull(stdout, back.Pix[:frameSize])
		if err != nil {
			_ = cmd.Wait()
			return err
		}

		p.mu.Lock()

		p.front, p.back = p.back, p.front

		// This is the practical Gio GPU upload path.
		// Gio will upload this image as a texture internally when painted.
		p.imageOp = paint.NewImageOp(p.front)
		p.hasOp = true

		p.mu.Unlock()

		if p.Invalidate != nil {
			p.Invalidate()
		}

		<-ticker.C
	}
}

func (p *Player) startAudio(ctx context.Context, seek time.Duration) {
	args := []string{
		"-nodisp",
		"-autoexit",
		"-loglevel", "error",
		"-volume", fmt.Sprintf("%d", int(p.Volume*100)),
	}

	if seek > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", seek.Seconds()))
	}

	args = append(args, p.Path)

	cmd := exec.CommandContext(ctx, "ffplay", args...)

	p.mu.Lock()
	p.ffplay = cmd
	p.mu.Unlock()

	_ = cmd.Run()
}

func (p *Player) restartAudio(ctx context.Context, seek time.Duration) {
	p.mu.Lock()
	if p.ffplay != nil && p.ffplay.Process != nil {
		_ = p.ffplay.Process.Kill()
		p.ffplay = nil
	}
	p.mu.Unlock()

	go p.startAudio(ctx, seek)
}

func (p *Player) Layout(gtx layout.Context) layout.Dimensions {
	max := gtx.Constraints.Max

	if max.X <= 0 {
		max.X = p.Width
	}
	if max.Y <= 0 {
		max.Y = p.Height
	}

	paint.FillShape(
		gtx.Ops,
		color.NRGBA{R: 14, G: 14, B: 18, A: 255},
		clip.Rect{Max: max}.Op(),
	)

	p.mu.RLock()
	img := p.imageOp
	ok := p.hasOp
	p.mu.RUnlock()

	if !ok {
		return layout.Dimensions{Size: max}
	}

	return widget.Image{
		Src:      img,
		Fit:      widget.Contain,
		Position: layout.Center,
		Scale:    1.0 / gtx.Metric.PxPerDp,
	}.Layout(gtx)
}

func (p *Player) currentSeek() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playFrom
}

func (p *Player) setStopped() {
	p.mu.Lock()
	p.seekTo = p.positionLocked(time.Now())
	p.playFrom = p.seekTo
	p.running = false
	p.startedAt = time.Time{}
	p.stopProcessesLocked()
	p.mu.Unlock()
}

func (p *Player) positionLocked(now time.Time) time.Duration {
	pos := p.seekTo
	if p.running && !p.startedAt.IsZero() {
		pos = p.playFrom + now.Sub(p.startedAt)
	}
	if pos < 0 {
		return 0
	}
	return pos
}

func (p *Player) stopProcessesLocked() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	if p.ffmpeg != nil && p.ffmpeg.Process != nil {
		_ = p.ffmpeg.Process.Kill()
	}
	if p.ffplay != nil && p.ffplay.Process != nil {
		_ = p.ffplay.Process.Kill()
	}
	p.ffmpeg = nil
	p.ffplay = nil
}
