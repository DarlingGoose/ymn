package audioplayer

import (
	"errors"
	"time"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/player"
)

var (
	ErrNoFileOpen = errors.New("no audio file is open")
	ErrClosed     = errors.New("audio player is closed")
)

type Config struct {
	// DecodeTimeout is used by ffmpeg decode.
	// Zero means no explicit timeout.
	DecodeTimeout time.Duration
}

type BeepPlayer struct {
	*beepBackend
}

var _ player.LoadablePlayable = (*BeepPlayer)(nil)

func NewBeepPlayer() (*BeepPlayer, error) {
	return NewBeepPlayerWithConfig(Config{})
}

func NewBeepPlayerWithConfig(cfg Config) (*BeepPlayer, error) {
	return &BeepPlayer{
		beepBackend: newBeepBackend(cfg),
	}, nil
}

// New is a small convenience for registry factories.
func New() player.LoadablePlayable {
	p, err := NewBeepPlayer()
	if err != nil {
		return nil
	}
	return p
}
