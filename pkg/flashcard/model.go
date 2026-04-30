package flashcard

import (
	"fmt"
	"strings"
	"time"

	"github.com/Seann-Moser/wgl/pkg/util"
)

const (
	DefaultAnkiModel = "WGL Flashcard"
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

func (card *Flashcard) NormalizeFlashcard() {
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
		card.AnkiDeck = util.AnkiDeckName(card.GameName)
	}
	if card.AnkiModel == "" {
		card.AnkiModel = DefaultAnkiModel
	}
	if card.SelectionStart < 0 {
		card.SelectionStart = 0
	}
	if card.SelectionEnd < card.SelectionStart {
		card.SelectionEnd = card.SelectionStart
	}
}

func (card *Flashcard) Valid() error {
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

func (card *Flashcard) Furigana() string {
	text := strings.TrimSpace(card.Text)
	reading := strings.TrimSpace(card.Reading)
	if text == "" || reading == "" || text == reading || !util.ContainsKanji(text) {
		return ""
	}
	return reading
}

func (card *Flashcard) FlashcardFuriganaHTML() string {
	text := strings.TrimSpace(card.Text)
	furigana := card.Furigana()
	if text == "" {
		return ""
	}
	if furigana == "" {
		return util.HtmlEscapedSingleLine(text)
	}
	return "<ruby>" + util.HtmlEscapedSingleLine(text) +
		"<rt>" + util.HtmlEscapedSingleLine(furigana) + "</rt></ruby>"
}

func (card *Flashcard) AnkiReading() string {
	return strings.TrimSpace(card.Reading)
}

func (card *Flashcard) AnkiFuriganaHTML() string {
	return card.FlashcardFuriganaHTML()
}

func (card *Flashcard) DuplicateKey() string {
	return strings.ToLower(strings.Join([]string{
		strings.TrimSpace(card.Text),
		strings.TrimSpace(card.Meaning),
	}, "\n"))
}

func (card *Flashcard) DeckName() string {
	return util.FirstNonEmpty(card.AnkiDeck, card.AnkiDeck, util.AnkiDeckName(card.GameName))
}
