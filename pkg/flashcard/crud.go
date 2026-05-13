package flashcard

import (
	"fmt"
	"strings"
	"time"

	"github.com/DarlingGoose/wgl/pkg/util"
)

func AddFlashcard(card Flashcard) error {
	card.NormalizeFlashcard()
	if err := card.Valid(); err != nil {
		return err
	}

	cards, err := LoadFlashcards(card.GameName)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	card.ID = fmt.Sprintf("%s-%d",
		util.SanitizeName(card.GameName),
		now.UnixNano())
	card.AnkiDeck = util.AnkiDeckName(card.GameName)
	card.AnkiModel = DefaultAnkiModel
	card.CreatedAt = now
	card.UpdatedAt = now
	cards = append(cards, card)
	return SaveFlashcards(card.GameName, cards)
}

func AddFlashcards(gameName string, newCards []Flashcard) (added int, skipped int, err error) {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		return 0, 0, fmt.Errorf("game name cannot be empty")
	}
	if len(newCards) == 0 {
		return 0, 0, fmt.Errorf("no flashcards to add")
	}

	cards, err := LoadFlashcards(gameName)
	if err != nil {
		return 0, 0, err
	}

	seen := make(map[string]struct{}, len(cards)+len(newCards))
	for _, card := range cards {
		seen[card.DuplicateKey()] = struct{}{}
	}

	now := time.Now().UTC()
	for i, card := range newCards {
		card.NormalizeFlashcard()

		card.GameName = gameName
		if err := card.Valid(); err != nil {
			return added, skipped, err
		}
		key := card.DuplicateKey()
		if _, ok := seen[key]; ok {
			skipped++
			continue
		}
		card.ID = fmt.Sprintf("%s-%d-%d", util.SanitizeName(card.GameName), now.UnixNano(), i)
		card.AnkiDeck = util.AnkiDeckName(card.GameName)
		card.AnkiModel = DefaultAnkiModel
		card.CreatedAt = now
		card.UpdatedAt = now
		cards = append(cards, card)
		seen[key] = struct{}{}
		added++
	}
	if added == 0 {
		return added, skipped, nil
	}
	return added, skipped, SaveFlashcards(gameName, cards)
}

func UpdateFlashcard(card Flashcard) error {
	card.NormalizeFlashcard()
	if err := card.Valid(); err != nil {
		return err
	}
	if strings.TrimSpace(card.ID) == "" {
		return fmt.Errorf("flashcard id cannot be empty")
	}

	cards, err := LoadFlashcards(card.GameName)
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
		card.AnkiDeck = util.FirstNonEmpty(card.DeckName(), cards[i].AnkiDeck, util.AnkiDeckName(card.GameName))
		card.AnkiModel = util.FirstNonEmpty(card.AnkiModel, cards[i].AnkiModel, DefaultAnkiModel)
		card.AnkiNoteID = cards[i].AnkiNoteID
		card.AnkiLastSyncAt = cards[i].AnkiLastSyncAt
		card.UpdatedAt = now
		cards[i] = card
		return SaveFlashcards(card.GameName, cards)
	}

	return fmt.Errorf("flashcard %q not found", card.ID)
}

func DeleteFlashcard(gameName, cardID string) error {
	gameName = strings.TrimSpace(gameName)
	cardID = strings.TrimSpace(cardID)
	if gameName == "" {
		return fmt.Errorf("game name cannot be empty")
	}
	if cardID == "" {
		return fmt.Errorf("flashcard id cannot be empty")
	}

	cards, err := LoadFlashcards(gameName)
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
	return SaveFlashcards(gameName, filtered)
}
