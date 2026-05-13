package dictionary

import (
	"sync"

	"github.com/DarlingGoose/jpndict"
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
	response          *jpndict.Response
}

var dictionaryState struct {
	once sync.Once
	dict jpndict.Dictonary
	err  error
}
