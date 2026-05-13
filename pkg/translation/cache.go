package translation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	jpndicttranslate "github.com/DarlingGoose/jpndict/translate"
	"github.com/DarlingGoose/ymn/pkg/util"
)

type Entry struct {
	GameName       string    `json:"game_name,omitempty"`
	SourceText     string    `json:"source_text"`
	TargetLanguage string    `json:"target_language"`
	Translation    string    `json:"translation"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Provider string

const (
	ProviderOllama           Provider = "ollama"
	ProviderOpenAI           Provider = "openai"
	ProviderGemini           Provider = "gemini"
	ProviderOpenAICompatible Provider = "openai_compatible"
)

type Config struct {
	Provider Provider
	APIKey   string
	BaseURL  string
	Model    string
}

var cacheMu sync.Mutex

func Load(gameName, sourceText, targetLanguage string) (Entry, bool, error) {
	key := cacheKey(gameName, sourceText, targetLanguage)
	if key == "" {
		return Entry{}, false, nil
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	entries, err := readCache()
	if err != nil {
		return Entry{}, false, err
	}
	entry, ok := entries[key]
	return entry, ok, nil
}

func Save(entry Entry) error {
	entry.GameName = cleanCacheText(entry.GameName)
	entry.SourceText = cleanCacheText(entry.SourceText)
	entry.TargetLanguage = cleanCacheText(entry.TargetLanguage)
	entry.Translation = strings.TrimSpace(entry.Translation)
	if entry.SourceText == "" {
		return fmt.Errorf("source sentence cannot be empty")
	}
	if entry.TargetLanguage == "" {
		return fmt.Errorf("target language cannot be empty")
	}
	if entry.Translation == "" {
		return fmt.Errorf("translation cannot be empty")
	}
	entry.UpdatedAt = time.Now()

	key := cacheKey(entry.GameName, entry.SourceText, entry.TargetLanguage)
	if key == "" {
		return fmt.Errorf("translation cache key cannot be empty")
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	entries, err := readCache()
	if err != nil {
		return err
	}
	entries[key] = entry
	return writeCache(entries)
}

func GenerateWithOllama(ctx context.Context, gameName, sourceText, targetLanguage string) (Entry, error) {
	return Generate(ctx, Config{Provider: ProviderOllama}, gameName, sourceText, targetLanguage)
}

func Generate(ctx context.Context, cfg Config, gameName, sourceText, targetLanguage string) (Entry, error) {
	sourceText = cleanCacheText(sourceText)
	targetLanguage = cleanCacheText(targetLanguage)
	if sourceText == "" {
		return Entry{}, fmt.Errorf("source sentence cannot be empty")
	}

	toLanguage, err := jpndictLanguage(targetLanguage)
	if err != nil {
		return Entry{}, err
	}

	client, err := newClient(cfg)
	if err != nil {
		return Entry{}, err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, &jpndicttranslate.Request{
		Text:               sourceText,
		ToLanguage:         toLanguage,
		GameTitle:          strings.TrimSpace(gameName),
		PreferLiteral:      false,
		SurroundingContext: "",
	})
	if err != nil {
		return Entry{}, err
	}
	if resp == nil || strings.TrimSpace(resp.Text) == "" {
		return Entry{}, jpndicttranslate.ErrNotFound
	}

	entry := Entry{
		GameName:       gameName,
		SourceText:     sourceText,
		TargetLanguage: targetLanguage,
		Translation:    strings.TrimSpace(resp.Text),
	}
	if err := Save(entry); err != nil {
		return Entry{}, err
	}
	return entry, nil
}

func newClient(cfg Config) (jpndicttranslate.Translate, error) {
	switch cfg.Provider {
	case "", ProviderOllama:
		return jpndicttranslate.NewOllamaClient(jpndicttranslate.OllamaConfig{
			BaseURL: strings.TrimSpace(cfg.BaseURL),
			Model:   strings.TrimSpace(cfg.Model),
		}), nil
	case ProviderOpenAI:
		return jpndicttranslate.NewOpenAIClient(jpndicttranslate.OpenAIConfig{
			APIKey:  strings.TrimSpace(cfg.APIKey),
			BaseURL: strings.TrimSpace(cfg.BaseURL),
			Model:   strings.TrimSpace(cfg.Model),
		}), nil
	case ProviderGemini:
		return jpndicttranslate.NewGeminiClient(jpndicttranslate.GeminiConfig{
			APIKey:  strings.TrimSpace(cfg.APIKey),
			BaseURL: strings.TrimSpace(cfg.BaseURL),
			Model:   strings.TrimSpace(cfg.Model),
		}), nil
	case ProviderOpenAICompatible:
		return jpndicttranslate.NewOpenAICompatibleClient(jpndicttranslate.OpenAICompatibleConfig{
			Name:    "openai-compatible",
			APIKey:  strings.TrimSpace(cfg.APIKey),
			BaseURL: strings.TrimSpace(cfg.BaseURL),
			Model:   strings.TrimSpace(cfg.Model),
		}), nil
	default:
		return nil, fmt.Errorf("unsupported translator provider %q", cfg.Provider)
	}
}

func jpndictLanguage(label string) (jpndicttranslate.Language, error) {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "english", "en":
		return jpndicttranslate.LanguageEnglish, nil
	case "japanese", "jpn", "ja":
		return jpndicttranslate.LanguageJapanese, nil
	default:
		return "", fmt.Errorf("jpndict translation generation currently supports English and Japanese, not %s", label)
	}
}

func readCache() (map[string]Entry, error) {
	path := cachePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]Entry), nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return make(map[string]Entry), nil
	}
	entries := make(map[string]Entry)
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func writeCache(entries map[string]Entry) error {
	path := cachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func cachePath() string {
	return filepath.Join(util.ConfigBaseDir(), "translations", "sentences.json")
}

func cacheKey(gameName, sourceText, targetLanguage string) string {
	sourceText = cleanCacheText(sourceText)
	targetLanguage = strings.ToLower(cleanCacheText(targetLanguage))
	if sourceText == "" || targetLanguage == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(cleanCacheText(gameName) + "\x00" + sourceText + "\x00" + targetLanguage))
	return hex.EncodeToString(sum[:])
}

func cleanCacheText(text string) string {
	text = strings.ReplaceAll(text, `\n`, " ")
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}
