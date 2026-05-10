package media

import (
	"context"
	"fmt"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	player2 "github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/player"
	videoplayer "github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/videoPlayer"
)

type InlineVideoPlayer struct {
	Path string

	player   *videoplayer.Player
	width    int
	height   int
	duration time.Duration
	position time.Duration
	volume   float32
	state    player2.State
	err      error
	started  bool
}

func NewInlineVideoPlayer() *InlineVideoPlayer {
	return &InlineVideoPlayer{
		state:  player2.StateIdle,
		volume: 0.8,
	}
}

func (p *InlineVideoPlayer) Load(path string) error {
	p.Stop()

	p.Path = path
	p.position = 0
	p.duration = 0
	p.err = nil

	if duration, err := probeDuration(context.Background(), path); err == nil {
		p.duration = duration
	}

	width, height, err := probeVideoSize(context.Background(), path)
	if err != nil {
		width, height = 1280, 720
	}
	p.width = width
	p.height = height

	p.player = videoplayer.New(path, width, height, 30, nil)
	p.player.PlayAudio = true
	p.player.SetVolume(p.volume)
	p.player.SetPosition(0)
	p.state = player2.StatePaused
	return nil
}

func (p *InlineVideoPlayer) Play() error {
	if p.player == nil {
		return fmt.Errorf("no media loaded")
	}
	if p.state == player2.StatePlaying {
		return nil
	}

	if !p.started {
		p.player.Start(context.Background())
		p.started = true
	} else {
		p.player.Resume(context.Background())
	}

	p.state = player2.StatePlaying
	p.err = nil
	return nil
}

func (p *InlineVideoPlayer) Pause() error {
	if p.player == nil {
		return nil
	}
	p.player.Pause()
	p.state = player2.StatePaused
	return nil
}

func (p *InlineVideoPlayer) Stop() error {
	if p.player != nil {
		p.player.Stop()
		p.player.SetPosition(0)
	}
	p.started = false
	p.position = 0
	p.state = player2.StateStopped
	p.err = nil
	return nil
}

func (p *InlineVideoPlayer) Close() error {
	err := p.Stop()
	p.player = nil
	p.Path = ""
	p.width = 0
	p.height = 0
	p.duration = 0
	p.position = 0
	p.state = player2.StateIdle
	return err
}

func (p *InlineVideoPlayer) Seek(pos time.Duration) error {
	if pos < 0 {
		pos = 0
	}
	if p.duration > 0 && pos > p.duration {
		pos = p.duration
	}

	p.position = pos
	if p.player == nil {
		return nil
	}

	if !p.started {
		p.player.SetPosition(pos)
		return nil
	}

	p.player.Seek(context.Background(), pos)
	if p.state == player2.StatePaused {
		p.player.Pause()
	}
	return nil
}

func (p *InlineVideoPlayer) SetVolume(v float32) error {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.volume = v
	if p.player != nil {
		p.player.SetVolume(v)
	}
	return nil
}

func (p *InlineVideoPlayer) Position() time.Duration {
	if p.state == player2.StatePlaying && p.player != nil {
		if !p.player.Running() {
			p.position = p.player.Position()
			if p.duration > 0 && p.position >= p.duration {
				p.position = p.duration
			}
			p.state = player2.StateStopped
			p.started = false
			return p.position
		}
		p.position = p.player.Position()
	}
	return p.position
}

func (p *InlineVideoPlayer) Duration() time.Duration {
	return p.duration
}

func (p *InlineVideoPlayer) State() player2.State {
	return p.state
}

func (p *InlineVideoPlayer) Error() error {
	return p.err
}

func (p *InlineVideoPlayer) Layout(gtx layout.Context) layout.Dimensions {
	if p.player == nil {
		return layout.Dimensions{}
	}
	_ = p.Position()
	if p.state == player2.StatePlaying {
		gtx.Execute(op.InvalidateCmd{})
	}
	return p.player.Layout(gtx)
}
