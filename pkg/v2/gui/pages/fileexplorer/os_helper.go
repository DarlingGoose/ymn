package fileexplorer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media/player"
)

func readDir(dir, query string, showHidden bool, sortBy SortBy, desc bool) ([]entry, error) {
	items, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(strings.TrimSpace(query))

	out := make([]entry, 0, len(items))

	for _, item := range items {
		name := item.Name()

		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}

		if query != "" && !strings.Contains(strings.ToLower(name), query) {
			continue
		}

		info, err := item.Info()
		if err != nil {
			continue
		}

		out = append(out, entry{
			Name:    name,
			Path:    filepath.Join(dir, name),
			IsDir:   item.IsDir(),
			Size:    fileSize(info),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		a := out[i]
		b := out[j]

		if sortBy != SortByKind && a.IsDir != b.IsDir {
			return a.IsDir
		}

		less := false

		switch sortBy {
		case SortBySize:
			if a.Size == b.Size {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			} else {
				less = a.Size < b.Size
			}

		case SortByModified:
			if a.ModTime.Equal(b.ModTime) {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			} else {
				less = a.ModTime.Before(b.ModTime)
			}

		case SortByKind:
			aKind := entryKind(a)
			bKind := entryKind(b)
			if aKind == bKind {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			} else {
				less = aKind < bKind
			}

		default:
			less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}

		if desc {
			return !less
		}

		return less
	})

	return out, nil
}

func entryKind(e entry) string {
	if e.IsDir {
		return "000-dir"
	}

	ext := strings.ToLower(filepath.Ext(e.Name))
	if ext == "" {
		return "zzz-file"
	}

	return ext
}

func fileSize(info fs.FileInfo) int64 {
	if info == nil || info.IsDir() {
		return 0
	}
	return info.Size()
}

func formatSize(size int64) string {
	const unit = 1024

	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div := int64(unit)
	exp := 0

	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}

func DefaultCommonPlaces(root string) []CommonPlace {
	home, _ := os.UserHomeDir()

	var places []CommonPlace

	add := func(label, path, icon string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}

		expanded := expandHome(path)
		abs, err := filepath.Abs(expanded)
		if err != nil {
			abs = expanded
		}

		if _, err := os.Stat(abs); err != nil {
			return
		}

		places = append(places, CommonPlace{
			Label: label,
			Path:  abs,
			Icon:  icon,
		})
	}

	add("Home", home, "lucide:home")
	add("Desktop", filepath.Join(home, "Desktop"), "lucide:monitor")
	add("Downloads", filepath.Join(home, "Downloads"), "lucide:download")
	add("Documents", filepath.Join(home, "Documents"), "lucide:file-text")
	add("Pictures", filepath.Join(home, "Pictures"), "lucide:image")
	add("Music", filepath.Join(home, "Music"), "lucide:music")
	add("Videos", filepath.Join(home, "Videos"), "lucide:film")

	if root != "" && !samePath(root, home) {
		add("Start", root, "lucide:folder-open")
	}

	add("Config", filepath.Join(home, ".config"), "lucide:settings")
	add("Root", string(filepath.Separator), "lucide:hard-drive")

	return dedupePlaces(places)
}

func dedupePlaces(in []CommonPlace) []CommonPlace {
	seen := make(map[string]struct{}, len(in))
	out := make([]CommonPlace, 0, len(in))

	for _, place := range in {
		key := filepath.Clean(place.Path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, place)
	}

	return out
}

func samePath(a, b string) bool {
	if a == "" || b == "" {
		return false
	}

	aa, err := filepath.Abs(expandHome(a))
	if err == nil {
		a = aa
	}

	bb, err := filepath.Abs(expandHome(b))
	if err == nil {
		b = bb
	}

	return filepath.Clean(a) == filepath.Clean(b)
}

func expandHome(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return path
	}

	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return path
	}

	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}

	return path
}

func supportsMediaPreview(path string) bool {
	switch media.DetectKind(path) {
	case player.KindImage, player.KindAudio, player.KindVideo:
		return true
	default:
		return false
	}
}
func isArchive(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".zip", ".tar", ".gz", ".tgz", ".bz2", ".xz", ".7z", ".rar":
		return true
	default:
		return false
	}
}
func detectFileCategory(path string, info fs.FileInfo) string {
	if info != nil && info.IsDir() {
		return "Directory"
	}

	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".sh", ".zsh", ".js", ".css":
		return "Text" //code todo
	case ".txt", ".log", ".md", ".csv", ".html":
		return "Text"
	case ".json":
		return "JSON"
	case ".yaml", ".yml":
		return "YAML"
	case ".toml":
		return "TOML"
	case ".xml":
		return "XML"
	case ".zip", ".tar", ".gz", ".tgz", ".bz2", ".xz", ".7z", ".rar":
		return "Archive"
	case ".exe", ".dll", ".msi":
		return "Windows executable"
	case ".so", ".dylib":
		return "Shared library"
	case ".appimage":
		return "AppImage"
	case ".desktop":
		return "Desktop entry"
	case ".pdf":
		return "PDF document"
	}

	if info != nil && info.Mode().IsRegular() && info.Mode()&0o111 != 0 {
		return "Executable"
	}

	if ext != "" {
		return strings.TrimPrefix(ext, ".") + " file"
	}

	return "File"
}

func isLikelyTextFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt", ".log", ".md", ".json", ".yaml", ".yml", ".toml", ".xml", ".csv", ".ini", ".conf", ".desktop":
		return true
	default:
		return false
	}
}
