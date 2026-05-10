package audioplayer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/player"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
)

const beepOutputSampleRate = beep.SampleRate(44100)

var (
	beepSpeakerMu         sync.Mutex
	beepSpeakerReady      bool
	beepSpeakerSampleRate beep.SampleRate
)

type beepBackend struct {
	mu sync.Mutex

	cfg Config

	streamer beep.StreamSeekCloser
	format   beep.Format

	ctrl   *beep.Ctrl
	volume *effects.Volume
	done   chan struct{}

	queued bool

	state player.State
	path  string
	err   error

	closed bool

	muted      bool
	lastVolume float32
}

func newBeepBackend(cfg Config) *beepBackend {
	return &beepBackend{
		cfg:        cfg,
		state:      player.StateIdle,
		lastVolume: 1,
	}
}

func (p *beepBackend) Load(path string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.err = ErrClosed
		return p.err
	}

	if err := p.closeOpenFileLocked(); err != nil {
		p.err = err
		return err
	}

	path = strings.TrimSpace(path)
	if path == "" {
		p.err = fmt.Errorf("audio path is empty")
		return p.err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		p.err = err
		return err
	}

	if err := ensureBeepSpeaker(); err != nil {
		p.err = err
		return err
	}

	ctx := context.Background()
	cancel := func() {}
	if p.cfg.DecodeTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, p.cfg.DecodeTimeout)
	}
	defer cancel()

	streamer, format, err := decodeWithFFmpegPCM(ctx, abs)
	if err != nil {
		p.err = err
		p.state = player.StateError
		return err
	}

	p.streamer = streamer
	p.format = format
	p.path = abs
	p.state = player.StateReady
	p.err = nil
	p.done = make(chan struct{})
	p.queued = false

	p.rebuildChainLocked(true)

	return nil
}

func (p *beepBackend) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.err = ErrClosed
		return p.err
	}
	if p.streamer == nil || p.ctrl == nil || p.volume == nil {
		p.err = ErrNoFileOpen
		return p.err
	}

	// If Play is called after natural completion, rewind and queue again.
	if p.streamer.Len() > 0 && p.streamer.Position() >= p.streamer.Len() {
		if err := p.streamer.Seek(0); err != nil {
			p.err = fmt.Errorf("beep seek: %w", err)
			return p.err
		}

		p.done = make(chan struct{})
		p.queued = false
		p.rebuildChainLocked(true)
	}

	done := p.done

	speaker.Lock()
	p.ctrl.Paused = false
	speaker.Unlock()

	if !p.queued {
		p.queued = true

		speaker.Play(beep.Seq(
			p.volume,
			beep.Callback(func() {
				p.finish(done)
			}),
		))
	}

	p.state = player.StatePlaying
	p.err = nil
	return nil
}

func (p *beepBackend) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.err = ErrClosed
		return p.err
	}
	if p.streamer == nil || p.ctrl == nil {
		p.err = ErrNoFileOpen
		return p.err
	}

	speaker.Lock()
	p.ctrl.Paused = true
	speaker.Unlock()

	p.state = player.StatePaused
	p.err = nil
	return nil
}

func (p *beepBackend) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.err = ErrClosed
		return p.err
	}
	if p.streamer == nil {
		return nil
	}

	speaker.Lock()

	if p.ctrl != nil {
		p.ctrl.Paused = true
		p.ctrl.Streamer = nil
	}

	err := p.streamer.Seek(0)

	p.done = make(chan struct{})
	p.queued = false
	p.rebuildChainLocked(true)

	speaker.Unlock()

	if err != nil {
		p.err = fmt.Errorf("beep seek: %w", err)
		return p.err
	}

	p.state = player.StateStopped
	p.err = nil
	return nil
}

func (p *beepBackend) Seek(pos time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.err = ErrClosed
		return p.err
	}
	if p.streamer == nil {
		p.err = ErrNoFileOpen
		return p.err
	}

	if pos < 0 {
		pos = 0
	}

	samples := p.format.SampleRate.N(pos)
	if p.streamer.Len() > 0 && samples > p.streamer.Len() {
		samples = p.streamer.Len()
	}

	wasPaused := true
	if p.ctrl != nil {
		wasPaused = p.ctrl.Paused
	}

	speaker.Lock()

	if p.ctrl != nil {
		p.ctrl.Streamer = nil
	}

	err := p.streamer.Seek(samples)

	p.rebuildChainLocked(wasPaused)

	speaker.Unlock()

	if err != nil {
		p.err = fmt.Errorf("beep seek: %w", err)
		return p.err
	}

	// If the previous stream had completed, allow Play to queue again.
	if p.streamer.Len() == 0 || p.streamer.Position() < p.streamer.Len() {
		select {
		case <-p.done:
			p.done = make(chan struct{})
			p.queued = false
		default:
		}
	}

	p.err = nil
	return nil
}

func (p *beepBackend) Position() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || p.streamer == nil {
		return 0
	}

	speaker.Lock()
	pos := p.streamer.Position()
	speaker.Unlock()

	return p.format.SampleRate.D(pos)
}

func (p *beepBackend) Duration() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed || p.streamer == nil || p.streamer.Len() <= 0 {
		return 0
	}

	return p.format.SampleRate.D(p.streamer.Len())
}

func (p *beepBackend) SetVolume(volume float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.err = ErrClosed
		return p.err
	}

	volume = clamp01(volume)
	p.lastVolume = volume

	if p.volume != nil {
		speaker.Lock()
		p.volume.Volume = volumeToBeep(volume)
		p.volume.Silent = p.muted || volume == 0
		speaker.Unlock()
	}

	p.err = nil
	return nil
}

func (p *beepBackend) Volume() float32 {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.lastVolume
}

func (p *beepBackend) SetMuted(muted bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.err = ErrClosed
		return p.err
	}

	p.muted = muted

	if p.volume != nil {
		speaker.Lock()
		p.volume.Silent = muted || p.lastVolume == 0
		speaker.Unlock()
	}

	p.err = nil
	return nil
}

func (p *beepBackend) Muted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.muted
}

func (p *beepBackend) State() player.State {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.state
}

func (p *beepBackend) Error() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.err
}

func (p *beepBackend) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	err := p.closeOpenFileLocked()

	p.closed = true
	p.state = player.StateIdle
	p.err = err

	return err
}

func (p *beepBackend) Wait() error {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return ErrClosed
	}
	if p.streamer == nil {
		p.mu.Unlock()
		return ErrNoFileOpen
	}

	done := p.done
	p.mu.Unlock()

	if done == nil {
		return ErrNoFileOpen
	}

	<-done
	return nil
}

func (p *beepBackend) closeOpenFileLocked() error {
	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = true
		p.ctrl.Streamer = nil
		speaker.Unlock()
		p.ctrl = nil
	}

	p.volume = nil
	p.queued = false

	if p.done != nil {
		select {
		case <-p.done:
		default:
			close(p.done)
		}
		p.done = nil
	}

	var firstErr error

	if p.streamer != nil {
		if err := p.streamer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		p.streamer = nil
	}

	p.path = ""
	p.format = beep.Format{}

	if !p.closed {
		p.state = player.StateIdle
	}

	return firstErr
}

func (p *beepBackend) finish(done chan struct{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state != player.StateIdle && p.state != player.StateError {
		p.state = player.StateStopped
	}

	p.queued = false

	if done != nil {
		select {
		case <-done:
		default:
			close(done)
		}
	}
}

func (p *beepBackend) rebuildChainLocked(paused bool) {
	p.ctrl = &beep.Ctrl{
		Streamer: p.playbackStreamerLocked(),
		Paused:   paused,
	}

	p.volume = &effects.Volume{
		Streamer: p.ctrl,
		Base:     2,
		Volume:   volumeToBeep(p.lastVolume),
		Silent:   p.muted || p.lastVolume == 0,
	}
}

func (p *beepBackend) playbackStreamerLocked() beep.Streamer {
	if p.streamer == nil {
		return beep.Silence(-1)
	}

	if p.format.SampleRate == beepSpeakerSampleRate {
		return p.streamer
	}

	return beep.Resample(4, p.format.SampleRate, beepSpeakerSampleRate, p.streamer)
}

func ensureBeepSpeaker() error {
	beepSpeakerMu.Lock()
	defer beepSpeakerMu.Unlock()

	if beepSpeakerReady {
		return nil
	}

	if err := speaker.Init(beepOutputSampleRate, beepOutputSampleRate.N(time.Second/10)); err != nil {
		return fmt.Errorf("init speaker: %w", err)
	}

	beepSpeakerReady = true
	beepSpeakerSampleRate = beepOutputSampleRate

	return nil
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// effects.Volume uses a logarithmic-ish volume.
// 0.0 should be silent, 1.0 should be neutral.
func volumeToBeep(v float32) float64 {
	if v <= 0 {
		return -8
	}
	if v >= 1 {
		return 0
	}

	x := 0.0
	for v < 1 {
		v *= 2
		x--
		if x <= -8 {
			return -8
		}
	}

	return x
}
