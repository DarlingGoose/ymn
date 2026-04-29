package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	jpndict "github.com/Seann-Moser/jpndict"
)

type dictionaryLookup struct {
	Query             string
	Headword          string
	Reading           string
	Meaning           string
	PronunciationText string
	Pitch             string
	AudioPath         string
}

var dictionaryState struct {
	once sync.Once
	dict jpndict.Dictonary
	err  error
}

func dictionaryCacheDir() string {
	return filepath.Join(configBaseDir(), "jpndict")
}

func loadDictionary() (jpndict.Dictonary, error) {
	dictionaryState.once.Do(func() {
		dict, err := jpndict.NewJiTenDex(dictionaryCacheDir(), "", true)
		if err != nil {
			dictionaryState.err = err
			return
		}
		if err := dict.Download(); err != nil {
			dictionaryState.err = err
			return
		}
		dictionaryState.dict = dict
	})
	if dictionaryState.err != nil {
		return nil, dictionaryState.err
	}
	return dictionaryState.dict, nil
}

func lookupDictionaryWord(word string) (*dictionaryLookup, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	dict, err := loadDictionary()
	if err != nil {
		return nil, err
	}

	resp, err := dict.Search(word)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Entry == nil {
		return nil, fmt.Errorf("no dictionary entry found for %q", word)
	}

	lookup := &dictionaryLookup{
		Query:    word,
		Headword: strings.TrimSpace(resp.Entry.Headword),
		Reading:  strings.TrimSpace(resp.Entry.Reading),
		Meaning:  summarizeDictionaryEntry(resp.Entry),
	}

	if lookup.Meaning == "" {
		lookup.Meaning = strings.TrimSpace(resp.Text)
	}

	if resp.Entry.Pronunciation != nil {
		lookup.PronunciationText = strings.TrimSpace(resp.Entry.Pronunciation.Text)
		lookup.Pitch = strings.TrimSpace(resp.Entry.Pronunciation.Pitch)
		if audioPath := strings.TrimSpace(resp.Entry.Pronunciation.Audio); audioPath != "" && isExistingFile(audioPath) {
			lookup.AudioPath = audioPath
		}
	}

	if lookup.Meaning == "" {
		return nil, fmt.Errorf("dictionary entry for %q did not contain a usable meaning", word)
	}
	return lookup, nil
}

func summarizeDictionaryEntry(entry *jpndict.Entry) string {
	if entry == nil {
		return ""
	}

	lines := make([]string, 0, len(entry.Senses))
	for i, sense := range entry.Senses {
		gloss := strings.TrimSpace(strings.Join(sense.Glosses, "; "))
		if gloss == "" {
			continue
		}

		line := gloss
		if len(sense.PartsOfSpeech) > 0 {
			line = "[" + strings.Join(sense.PartsOfSpeech, ", ") + "] " + line
		}
		if len(sense.Notes) > 0 {
			line += " (" + strings.Join(sense.Notes, "; ") + ")"
		}
		if len(entry.Senses) > 1 {
			line = fmt.Sprintf("%d. %s", i+1, line)
		}
		lines = append(lines, line)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
