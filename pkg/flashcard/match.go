package flashcard

import (
	"slices"
)

type Match struct {
	Key       string
	Word      string
	Card      Flashcard
	StartRune int
	EndRune   int
}

func FindMatches(text string, words []string) []Match {
	if len(words) == 0 || text == "" {
		return nil
	}

	textRunes := []rune(text)
	occupied := make([]bool, len(textRunes))
	matches := make([]Match, 0)
	for _, word := range words {
		if word == "" {
			continue
		}
		wordRunes := []rune(word)
		if len(wordRunes) == 0 || len(wordRunes) > len(textRunes) {
			continue
		}
		for start := 0; start <= len(textRunes)-len(wordRunes); start++ {
			end := start + len(wordRunes)
			if transcriptRunesOccupied(occupied, start, end) {
				continue
			}
			if !slices.Equal(textRunes[start:end], wordRunes) {
				continue
			}

			matches = append(matches, Match{
				Word:      word,
				StartRune: start,
				EndRune:   end,
			}) //func transcriptHighlightColor(base color.NRGBA) color.NRGBA {
			//	return color.NRGBA{
			//		R: base.R,
			//		G: base.G,
			//		B: base.B,
			//		A: 72,
			//	}
			for i := start; i < end; i++ {
				occupied[i] = true
			}
		}
	}
	return matches
}

func transcriptRunesOccupied(occupied []bool, start, end int) bool {
	for i := start; i < end; i++ {
		if occupied[i] {
			return true
		}
	}
	return false
}
