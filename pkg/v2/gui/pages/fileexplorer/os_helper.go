package fileexplorer

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
			Name:        name,
			Path:        filepath.Join(dir, name),
			IsDir:       item.IsDir(),
			Size:        fileSize(info),
			ModTime:     info.ModTime(),
			CreatedTime: createdTime(info),
			Mode:        info.Mode(),
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

		case SortByCreated:
			if a.CreatedTime.Equal(b.CreatedTime) {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			} else {
				less = a.CreatedTime.Before(b.CreatedTime)
			}

		case SortByKind:
			aKind := entryKind(a)
			bKind := entryKind(b)
			if aKind == bKind {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			} else {
				less = aKind < bKind
			}

		case SortByExtension:
			aExt := entryExtension(a)
			bExt := entryExtension(b)
			if aExt == bExt {
				less = strings.ToLower(a.Name) < strings.ToLower(b.Name)
			} else {
				less = aExt < bExt
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

func entryExtension(e entry) string {
	if e.IsDir {
		return "000-dir"
	}
	ext := strings.ToLower(filepath.Ext(e.Name))
	if ext == "" {
		return "zzz"
	}
	return strings.TrimPrefix(ext, ".")
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

func isZipArchive(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".zip")
}

func extractZipArchive(srcPath string) (string, error) {
	srcPath = strings.TrimSpace(srcPath)
	if srcPath == "" {
		return "", fmt.Errorf("archive path is required")
	}
	if !isZipArchive(srcPath) {
		return "", fmt.Errorf("only zip archives can be extracted")
	}

	absSrc, err := filepath.Abs(srcPath)
	if err != nil {
		return "", fmt.Errorf("resolve archive path: %w", err)
	}

	destDir, err := uniqueExtractDir(defaultZipExtractDir(absSrc))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", fmt.Errorf("create extract dir: %w", err)
	}

	reader, err := zip.OpenReader(absSrc)
	if err != nil {
		return "", fmt.Errorf("open zip archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if err := extractZipEntry(file, destDir); err != nil {
			return "", err
		}
	}

	return destDir, nil
}

func defaultZipExtractDir(srcPath string) string {
	base := filepath.Base(srcPath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" || base == "." {
		base = "archive"
	}
	return filepath.Join(filepath.Dir(srcPath), base)
}

func uniqueExtractDir(base string) (string, error) {
	base = strings.TrimSpace(base)
	if base == "" {
		return "", fmt.Errorf("extract dir is required")
	}

	if _, err := os.Stat(base); os.IsNotExist(err) {
		return base, nil
	} else if err != nil {
		return "", fmt.Errorf("check extract dir: %w", err)
	}

	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s %d", base, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("check extract dir: %w", err)
		}
	}

	return "", fmt.Errorf("could not find available extract dir for %s", base)
}

func extractZipEntry(file *zip.File, destDir string) error {
	cleanName := path.Clean(strings.ReplaceAll(file.Name, "\\", "/"))
	if cleanName == "." || cleanName == ".." || path.IsAbs(cleanName) || strings.HasPrefix(cleanName, "../") {
		return fmt.Errorf("zip entry escapes destination: %s", file.Name)
	}

	destPath := filepath.Join(append([]string{destDir}, strings.Split(cleanName, "/")...)...)
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(destPath, file.Mode()); err != nil {
			return fmt.Errorf("create zip directory: %w", err)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("create zip parent directory: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("open zip entry: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("create zip entry: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("extract zip entry: %w", err)
	}

	return nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}
	return t.Format("2006-01-02 15:04:05")
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
