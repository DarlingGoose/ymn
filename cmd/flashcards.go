package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const (
	defaultAnkiModel = "Game Flashcard"
)

type Flashcard struct {
	ID                 string     `json:"id"`
	GameName           string     `json:"game_name"`
	Text               string     `json:"text"`
	Meaning            string     `json:"meaning"`
	Reading            string     `json:"reading,omitempty"`
	PronunciationText  string     `json:"pronunciation_text,omitempty"`
	PronunciationPitch string     `json:"pronunciation_pitch,omitempty"`
	AudioPath          string     `json:"audio_path,omitempty"`
	SourceLine         string     `json:"source_line,omitempty"`
	SourcePath         string     `json:"source_path,omitempty"`
	SelectionStart     int        `json:"selection_start,omitempty"`
	SelectionEnd       int        `json:"selection_end,omitempty"`
	AnkiDeck           string     `json:"anki_deck,omitempty"`
	AnkiModel          string     `json:"anki_model,omitempty"`
	AnkiNoteID         int64      `json:"anki_note_id,omitempty"`
	AnkiLastSyncAt     *time.Time `json:"anki_last_sync_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func flashcardDir() string {
	return filepath.Join(configBaseDir(), "flashcards")
}

func flashcardExportDir() string {
	return filepath.Join(configBaseDir(), "exports")
}

func flashcardPath(gameName string) string {
	name := sanitizeName(gameName)
	if name == "" {
		name = "default"
	}
	return filepath.Join(flashcardDir(), name+".json")
}

func ankiDeckName(gameName string) string {
	return strings.TrimSpace(gameName)
}

func loadFlashcards(gameName string) ([]Flashcard, error) {
	path := flashcardPath(gameName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read flashcards: %w", err)
	}

	var cards []Flashcard
	if err := json.Unmarshal(data, &cards); err != nil {
		return nil, fmt.Errorf("decode flashcards: %w", err)
	}
	return cards, nil
}

func saveFlashcards(gameName string, cards []Flashcard) error {
	if err := os.MkdirAll(flashcardDir(), 0o755); err != nil {
		return fmt.Errorf("create flashcard directory: %w", err)
	}

	data, err := json.MarshalIndent(cards, "", "  ")
	if err != nil {
		return fmt.Errorf("encode flashcards: %w", err)
	}

	if err := os.WriteFile(flashcardPath(gameName), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write flashcards: %w", err)
	}
	return nil
}

func addFlashcard(card Flashcard) error {
	card = normalizeFlashcard(card)
	if err := validateFlashcard(card); err != nil {
		return err
	}

	cards, err := loadFlashcards(card.GameName)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	card.ID = fmt.Sprintf("%s-%d", sanitizeName(card.GameName), now.UnixNano())
	card.AnkiDeck = ankiDeckName(card.GameName)
	card.AnkiModel = defaultAnkiModel
	card.CreatedAt = now
	card.UpdatedAt = now
	cards = append(cards, card)
	return saveFlashcards(card.GameName, cards)
}

func updateFlashcard(card Flashcard) error {
	card = normalizeFlashcard(card)
	if err := validateFlashcard(card); err != nil {
		return err
	}
	if strings.TrimSpace(card.ID) == "" {
		return fmt.Errorf("flashcard id cannot be empty")
	}

	cards, err := loadFlashcards(card.GameName)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for i := range cards {
		if cards[i].ID != card.ID {
			continue
		}
		card.CreatedAt = cards[i].CreatedAt
		if card.CreatedAt.IsZero() {
			card.CreatedAt = now
		}
		card.AnkiDeck = firstNonEmpty(card.AnkiDeck, cards[i].AnkiDeck, ankiDeckName(card.GameName))
		card.AnkiModel = firstNonEmpty(card.AnkiModel, cards[i].AnkiModel, defaultAnkiModel)
		card.AnkiNoteID = cards[i].AnkiNoteID
		card.AnkiLastSyncAt = cards[i].AnkiLastSyncAt
		card.UpdatedAt = now
		cards[i] = card
		return saveFlashcards(card.GameName, cards)
	}

	return fmt.Errorf("flashcard %q not found", card.ID)
}

func deleteFlashcard(gameName, cardID string) error {
	gameName = strings.TrimSpace(gameName)
	cardID = strings.TrimSpace(cardID)
	if gameName == "" {
		return fmt.Errorf("game name cannot be empty")
	}
	if cardID == "" {
		return fmt.Errorf("flashcard id cannot be empty")
	}

	cards, err := loadFlashcards(gameName)
	if err != nil {
		return err
	}

	filtered := cards[:0]
	removed := false
	for _, card := range cards {
		if card.ID == cardID {
			removed = true
			continue
		}
		filtered = append(filtered, card)
	}
	if !removed {
		return fmt.Errorf("flashcard %q not found", cardID)
	}
	return saveFlashcards(gameName, filtered)
}

func normalizeFlashcard(card Flashcard) Flashcard {
	card.Text = strings.TrimSpace(card.Text)
	card.Meaning = strings.TrimSpace(card.Meaning)
	card.GameName = strings.TrimSpace(card.GameName)
	card.Reading = strings.TrimSpace(card.Reading)
	card.PronunciationText = strings.TrimSpace(card.PronunciationText)
	card.PronunciationPitch = strings.TrimSpace(card.PronunciationPitch)
	card.AudioPath = strings.TrimSpace(card.AudioPath)
	card.SourceLine = strings.TrimSpace(card.SourceLine)
	card.SourcePath = strings.TrimSpace(card.SourcePath)
	if card.AnkiDeck == "" {
		card.AnkiDeck = ankiDeckName(card.GameName)
	}
	if card.AnkiModel == "" {
		card.AnkiModel = defaultAnkiModel
	}
	if card.SelectionStart < 0 {
		card.SelectionStart = 0
	}
	if card.SelectionEnd < card.SelectionStart {
		card.SelectionEnd = card.SelectionStart
	}
	return card
}

func validateFlashcard(card Flashcard) error {
	if card.Text == "" {
		return fmt.Errorf("flashcard text cannot be empty")
	}
	if card.Meaning == "" {
		return fmt.Errorf("flashcard meaning cannot be empty")
	}
	if card.GameName == "" {
		return fmt.Errorf("game name cannot be empty")
	}
	return nil
}

func flashcardFuriganaText(card Flashcard) string {
	text := strings.TrimSpace(card.Text)
	reading := strings.TrimSpace(card.Reading)
	if text == "" || reading == "" || text == reading || !containsKanji(text) {
		return ""
	}
	return reading
}

func flashcardFuriganaHTML(card Flashcard) string {
	text := strings.TrimSpace(card.Text)
	furigana := flashcardFuriganaText(card)
	if text == "" {
		return ""
	}
	if furigana == "" {
		return htmlEscapedSingleLine(text)
	}
	return "<ruby>" + htmlEscapedSingleLine(text) + "<rt>" + htmlEscapedSingleLine(furigana) + "</rt></ruby>"
}

func htmlEscapedSingleLine(text string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&#39;")
	return replacer.Replace(strings.TrimSpace(text))
}

func containsKanji(text string) bool {
	for _, r := range text {
		if unicode.In(r, unicode.Han) {
			return true
		}
	}
	return false
}
