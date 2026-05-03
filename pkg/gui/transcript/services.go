package transcript

import "context"

type SentenceTranslator interface {
	TranslateSentence(ctx context.Context, sentence string) (string, error)
}

type MockSentenceTranslator struct {
	Response string
	Err      error
}

func (m MockSentenceTranslator) TranslateSentence(_ context.Context, _ string) (string, error) {
	return m.Response, m.Err
}
