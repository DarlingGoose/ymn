package media

import (
	"context"
	"fmt"
	"strings"
	"sync"

	audioplayer "github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/audioPlayer"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/backend/media/player"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

type RendererFactory func() player.Renderer

type TypeHandler struct {
	Kind       player.Kind
	Factory    RendererFactory
	Extensions map[string]struct{}
}

type Registry struct {
	mu       sync.RWMutex
	byKind   map[player.Kind]TypeHandler
	byExt    map[string]player.Kind
	fallback RendererFactory
}

func NewRegistry() *Registry {
	return &Registry{
		byKind: make(map[player.Kind]TypeHandler),
		byExt:  make(map[string]player.Kind),
	}
}

func (r *Registry) Register(kind player.Kind, factory RendererFactory, exts ...string) {
	if r == nil || kind == "" || factory == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	handler := TypeHandler{
		Kind:       kind,
		Factory:    factory,
		Extensions: make(map[string]struct{}, len(exts)),
	}

	for _, ext := range exts {
		ext = normalizeExt(ext)
		if ext == "" {
			continue
		}
		handler.Extensions[ext] = struct{}{}
		r.byExt[ext] = kind
	}

	r.byKind[kind] = handler
}

func (r *Registry) SetFallback(factory RendererFactory) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.fallback = factory
	r.mu.Unlock()
}

func (r *Registry) KindForExt(ext string) player.Kind {
	if r == nil {
		return player.KindUnknown
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if kind, ok := r.byExt[normalizeExt(ext)]; ok {
		return kind
	}
	return player.KindUnknown
}

func (r *Registry) NewRenderer(src player.Source) (player.Renderer, error) {
	if r == nil {
		return nil, fmt.Errorf("media registry is nil")
	}

	src = WithDetectedKind(src)

	r.mu.RLock()
	handler, ok := r.byKind[src.Kind]
	fallback := r.fallback
	r.mu.RUnlock()

	if ok && handler.Factory != nil {
		return handler.Factory(), nil
	}

	if fallback != nil {
		return fallback(), nil
	}

	return nil, fmt.Errorf("unsupported media kind %q for %s", src.Kind, src.Path)
}

func normalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

var DefaultRegistry = NewRegistry()

func init() {
	RegisterDefaults(nil)
}
func RegisterDefaults(tc *theme.Client) {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	DefaultRegistry.Register(
		player.KindImage,
		func() player.Renderer {
			return NewImageRenderer().
				WithThemeClient(tc)
		},
		".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp",
	)

	DefaultRegistry.Register(
		player.KindVideo,
		func() player.Renderer { return NewVideoRenderer() },
		".mp4", ".mkv", ".webm", ".mov", ".avi", ".wmv", ".m4v",
	)

	DefaultRegistry.Register(
		player.KindAudio,
		func() player.Renderer {
			return NewAudioRenderer(audioplayer.New()).WithThemeClient(tc)
		},
		".mp3", ".wav", ".ogg", ".flac", ".m4a", ".aac", ".opus",
	)

	DefaultRegistry.Register(
		player.KindDocument,
		func() player.Renderer { return NewUnsupportedRenderer("Document preview not implemented yet") },
		".pdf", ".txt", ".md", ".json", ".yaml", ".yml", ".toml", ".csv", ".log",
	)
}

func LoadRenderer(ctx context.Context, r *Registry, src player.Source) (player.Renderer, error) {
	if r == nil {
		r = DefaultRegistry
	}

	src = WithDetectedKind(src)

	renderer, err := r.NewRenderer(src)
	if err != nil {
		return nil, err
	}

	if err := renderer.Load(ctx, src); err != nil {
		_ = renderer.Close()
		return nil, err
	}

	return renderer, nil
}
