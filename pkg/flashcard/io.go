package flashcard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DarlingGoose/wgl/pkg/util"
)

func flashcardDir() string {
	return filepath.Join(util.ConfigBaseDir(), "flashcards")
}

func FlashcardExportDir() string {
	return filepath.Join(util.ConfigBaseDir(), "exports")
}

func flashcardPath(gameName string) string {
	name := util.SanitizeName(gameName)
	if name == "" {
		name = "default"
	}
	return filepath.Join(flashcardDir(), name+".json")
}

func LoadFlashcards(gameName string) ([]Flashcard, error) {
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

func SaveFlashcards(gameName string, cards []Flashcard) error {
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

func RenameGameFlashcards(oldGameName, newGameName string) error {
	oldGameName = strings.TrimSpace(oldGameName)
	newGameName = strings.TrimSpace(newGameName)
	if oldGameName == "" || newGameName == "" || strings.EqualFold(oldGameName, newGameName) {
		return nil
	}

	oldPath := flashcardPath(oldGameName)
	newPath := flashcardPath(newGameName)
	if oldPath == newPath {
		return nil
	}

	oldCards, err := LoadFlashcards(oldGameName)
	if err != nil {
		return err
	}
	if len(oldCards) == 0 {
		return nil
	}

	newCards, err := LoadFlashcards(newGameName)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(newCards)+len(oldCards))
	for _, card := range newCards {
		seen[card.DuplicateKey()] = struct{}{}
	}
	for _, card := range oldCards {
		card.GameName = newGameName
		card.AnkiDeck = util.AnkiDeckName(newGameName)
		card.NormalizeFlashcard()
		if _, ok := seen[card.DuplicateKey()]; ok {
			continue
		}
		seen[card.DuplicateKey()] = struct{}{}
		newCards = append(newCards, card)
	}
	if err := SaveFlashcards(newGameName, newCards); err != nil {
		return err
	}
	_ = os.Remove(oldPath)
	return nil
}
