package player

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"gioui.org/layout"
)

type Kind string

const (
	KindUnknown  Kind = "unknown"
	KindImage    Kind = "image"
	KindAudio    Kind = "audio"
	KindVideo    Kind = "video"
	KindDocument Kind = "document"
)

type State string

const (
	StateIdle    State = "idle"
	StateLoading State = "loading"
	StateReady   State = "ready"
	StatePlaying State = "playing"
	StatePaused  State = "paused"
	StateStopped State = "stopped"
	StateError   State = "error"
)

type Source struct {
	Path string
	Kind Kind
	MIME string
	Ext  string
	Name string
}

func NewSource(path string) Source {
	ext := strings.ToLower(filepath.Ext(path))
	return Source{
		Path: path,
		Ext:  ext,
		Name: filepath.Base(path),
	}
}

type Metadata struct {
	Width    int
	Height   int
	Duration time.Duration
	MIME     string
	Extra    map[string]string
}

type Renderer interface {
	Load(ctx context.Context, src Source) error
	Layout(gtx layout.Context) layout.Dimensions
	Close() error
	State() State
	Error() error
}

type Playable interface {
	Play() error
	Pause() error
	Stop() error
	Seek(pos time.Duration) error
	SetVolume(v float32) error

	Position() time.Duration
	Duration() time.Duration
	State() State
	Error() error
}

type LoadablePlayable interface {
	Playable
	Load(path string) error
}
