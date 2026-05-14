package media

import (
	"path/filepath"
	"strings"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media/player"
)

var DefaultKinds = map[string]player.Kind{
	// Images
	".png":  player.KindImage,
	".jpg":  player.KindImage,
	".jpeg": player.KindImage,
	".gif":  player.KindImage,
	".webp": player.KindImage,
	".bmp":  player.KindImage,

	// Audio
	".mp3":  player.KindAudio,
	".wav":  player.KindAudio,
	".ogg":  player.KindAudio,
	".flac": player.KindAudio,
	".m4a":  player.KindAudio,
	".aac":  player.KindAudio,
	".opus": player.KindAudio,

	// Video
	".mp4":  player.KindVideo,
	".mkv":  player.KindVideo,
	".webm": player.KindVideo,
	".mov":  player.KindVideo,
	".avi":  player.KindVideo,
	".wmv":  player.KindVideo,
	".m4v":  player.KindVideo,

	// Documents
	".pdf":  player.KindDocument,
	".txt":  player.KindDocument,
	".md":   player.KindDocument,
	".json": player.KindDocument,
	".yaml": player.KindDocument,
	".yml":  player.KindDocument,
	".toml": player.KindDocument,
	".csv":  player.KindDocument,
	".log":  player.KindDocument,
}

func DetectKind(path string) player.Kind {
	ext := strings.ToLower(filepath.Ext(path))
	if kind, ok := DefaultKinds[ext]; ok {
		return kind
	}
	return player.KindUnknown
}

func WithDetectedKind(src player.Source) player.Source {
	if src.Ext == "" {
		src.Ext = strings.ToLower(filepath.Ext(src.Path))
	}
	if src.Kind == "" || src.Kind == player.KindUnknown {
		src.Kind = DetectKind(src.Path)
	}
	if src.Name == "" {
		src.Name = filepath.Base(src.Path)
	}
	return src
}
