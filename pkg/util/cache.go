package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func VoiceCacheRootDir() string {
	return filepath.Join(ConfigBaseDir(), "voices")
}

func GameVoiceCacheDir(gameName string) string {
	return filepath.Join(VoiceCacheRootDir(), SanitizeName(gameName))
}

func VoiceCachePath(gameName, voice, ext string) (string, error) {
	gameName = strings.TrimSpace(gameName)
	voice = strings.TrimSpace(voice)
	if gameName == "" {
		return "", fmt.Errorf("game name is required")
	}
	if voice == "" {
		return "", fmt.Errorf("voice file is empty")
	}
	sum := sha256.Sum256([]byte(gameName + "\x00" + voice))
	name := hex.EncodeToString(sum[:]) + NormalizeAudioExt(FirstNonEmpty(ext, filepath.Ext(voice)))
	return filepath.Join(GameVoiceCacheDir(gameName), name), nil
}

func GameVoiceCacheSize(gameName string) (int64, error) {
	return DirectorySize(GameVoiceCacheDir(gameName))
}

func ClearGameVoiceCache(gameName string) error {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		return fmt.Errorf("game name is required")
	}
	return os.RemoveAll(GameVoiceCacheDir(gameName))
}

func DirectorySize(path string) (int64, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return 0, fmt.Errorf("path is required")
	}
	var total int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		total += info.Size()
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func NormalizeAudioExt(value string) string {
	value = strings.TrimSpace(value)
	if value != "" && !strings.Contains(value, string(os.PathSeparator)) && !strings.Contains(value, ".") {
		value = "." + value
	}
	ext := strings.ToLower(filepath.Ext(value))
	switch ext {
	case ".ogg", ".oga", ".wav", ".mp3", ".m4a", ".flac", ".opus":
		return ext
	default:
		return ".ogg"
	}
}
