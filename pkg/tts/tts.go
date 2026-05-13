package tts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	vntts "github.com/DarlingGoose/vntext/pkg/tts"
	"github.com/DarlingGoose/wgl/pkg/util"
)

type Reference struct {
	Speaker string
	Audio   string
	Text    string
}

func SpeakWithF5(ctx context.Context, gameName, text string, ref Reference) (string, error) {
	text = cleanTTSInput(text)
	if text == "" {
		return "", fmt.Errorf("text cannot be empty")
	}
	ref.Speaker = strings.TrimSpace(ref.Speaker)
	ref.Audio = strings.TrimSpace(ref.Audio)
	ref.Text = cleanTTSInput(ref.Text)
	if ref.Speaker == "" {
		return "", fmt.Errorf("select a TTS reference speaker first")
	}
	if ref.Audio == "" || !util.IsExistingFile(ref.Audio) {
		return "", fmt.Errorf("reference audio for %s is not available", ref.Speaker)
	}
	if ref.Text == "" {
		return "", fmt.Errorf("reference text for %s is not available", ref.Speaker)
	}

	dir := cacheDir(gameName)
	file := cacheFileName(gameName, text, ref)
	path := filepath.Join(dir, file)
	if util.IsExistingFile(path) {
		return path, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create tts cache dir: %w", err)
	}

	engine, err := vntts.NewF5(vntts.F5Config{})
	if err != nil {
		return "", err
	}
	defer engine.Close()

	speakerID := speakerID(ref.Speaker)
	if err := engine.LoadVoices(ctx, vntts.Speaker{
		ID:             speakerID,
		Name:           ref.Speaker,
		VoiceClipsPath: []string{ref.Audio},
		ReferenceText:  ref.Text,
		Language:       "Japanese",
	}); err != nil {
		return "", err
	}

	result, err := engine.Speak(ctx, text,
		vntts.WithSpeaker(speakerID),
		vntts.WithOutput(dir, file),
		vntts.WithRemoveSilence(false),
		vntts.WithDevice("vulkan"),
		vntts.WithLanguage("Japanese"),
		vntts.WithDeviceFallback("cpu"),
	)
	if err != nil {
		slog.Error("failed generating speach", "err", err, "dir", dir, "file", file)
		return "", err
	}
	if result == nil || strings.TrimSpace(result.AudioPath) == "" {
		return "", fmt.Errorf("tts did not return an audio file")
	}
	if !util.IsExistingFile(result.AudioPath) {
		return "", fmt.Errorf("tts output was not created: %s", result.AudioPath)
	}
	return result.AudioPath, nil
}

func cacheDir(gameName string) string {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		gameName = "unknown-game"
	}
	return filepath.Join(util.ConfigBaseDir(), "tts", safeFilename(gameName))
}

func cacheFileName(gameName, text string, ref Reference) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strings.TrimSpace(gameName),
		cleanTTSInput(text),
		strings.TrimSpace(ref.Speaker),
		strings.TrimSpace(ref.Audio),
		cleanTTSInput(ref.Text),
	}, "\x00")))
	return hex.EncodeToString(sum[:]) + ".wav"
}

func speakerID(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	if name == "" {
		return "speaker"
	}
	return safeFilename(name)
}

func safeFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "value"
	}
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-.")
	if out == "" {
		return "value"
	}
	return out
}

func cleanTTSInput(text string) string {
	text = strings.ReplaceAll(text, `\n`, " ")
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}
