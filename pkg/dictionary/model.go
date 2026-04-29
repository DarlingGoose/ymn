package dictionary

import (
	"sync"

	"github.com/Seann-Moser/jpndict"
)

type Lookup struct {
	Key               string
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
