package dictionary

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DarlingGoose/jpndict"
	"github.com/DarlingGoose/wgl/pkg/util"
)

func init() {
	_, _ = loadDictionary()
}

func LookupWord(word string) (*Lookup, error) {
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

	lookup := dictionaryLookupFromResponse(word, resp)
	if lookup.Meaning == "" {
		return nil, fmt.Errorf("dictionary entry for %q did not contain a usable meaning", word)
	}
	return &lookup, nil
}

func LookupWords(word string) ([]Lookup, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return nil, fmt.Errorf("word cannot be empty")
	}

	dict, err := loadDictionary()
	if err != nil {
		return nil, err
	}

	responses, err := dict.SearchAll(word)
	if err != nil {
		return nil, err
	}
	if len(responses) == 0 {
		return nil, fmt.Errorf("no dictionary entry found for %q", word)
	}
	lookups := make([]Lookup, 0, len(responses))
	seen := make(map[string]struct{}, len(responses))
	for _, resp := range responses {
		lookup := dictionaryLookupFromResponse(word, resp)
		if lookup.Meaning == "" {
			continue
		}
		key := util.FirstNonEmpty(lookup.Query, lookup.Key, lookup.Headword)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		lookups = append(lookups, lookup)
	}
	if len(lookups) == 0 {
		return nil, fmt.Errorf("dictionary entry for %q did not contain a usable meaning", word)
	}
	return lookups, nil
}

func PlayLookupAudio(lookup Lookup) error {
	if lookup.response != nil {
		if _, err := lookup.response.PlayAudio(false); err != nil {
			return err
		}
		return nil
	}

	return PlayAudioForText(util.FirstNonEmpty(lookup.Query, lookup.Key, lookup.Headword, lookup.Reading))
}

func PlayAudioForText(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("word cannot be empty")
	}

	dict, err := loadDictionary()
	if err != nil {
		return err
	}

	resp, err := dict.Search(text)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("no dictionary entry found for %q", text)
	}
	_, err = resp.PlayAudio(false)
	return err
}

func SummarizeDictionaryEntry(entry *jpndict.Entry) string {
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

func dictionaryCacheDir() string {
	return filepath.Join(util.ConfigBaseDir(), "jpndict")
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

func dictionaryLookupFromResponse(query string, resp *jpndict.Response) Lookup {
	lookup := Lookup{
		Key:      strings.TrimSpace(util.FirstNonEmpty(resp.Key, query)),
		Query:    strings.TrimSpace(query),
		response: resp,
	}

	if resp.Entry != nil {
		lookup.Query = resp.Query
		lookup.Headword = strings.TrimSpace(resp.Entry.Headword)
		lookup.Reading = strings.TrimSpace(resp.Entry.Reading)
		lookup.Meaning = SummarizeDictionaryEntry(resp.Entry)
		if resp.Entry.Pronunciation != nil {
			lookup.PronunciationText = strings.TrimSpace(resp.Entry.Pronunciation.Text)
			lookup.Pitch = strings.TrimSpace(resp.Entry.Pronunciation.Pitch)
			if audioPath := strings.TrimSpace(resp.Entry.Pronunciation.Audio); audioPath != "" && util.IsExistingFile(audioPath) {
				lookup.AudioPath = audioPath
			}
		}
	}
	if lookup.Meaning == "" {
		lookup.Meaning = strings.TrimSpace(resp.Text)
	}
	return lookup
}
