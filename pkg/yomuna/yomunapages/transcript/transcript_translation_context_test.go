package transcript

import "testing"

func TestTranslationContextForRowUsesPreviousNonInfoRows(t *testing.T) {
	t.Helper()

	follower := newTranscriptFollower(nil, nil)
	follower.transcriptRows = []transcriptRow{
		{Key: "one", Speaker: "Alice", Text: "一行目"},
		{Key: "info", Info: true, Text: "Following logs"},
		{Key: "two", Text: "二行目"},
		{Key: "three", Speaker: "Bob", Text: "三行目"},
		{Key: "target", Text: "翻訳対象"},
	}

	got := follower.translationContextForRow("target", 2)
	want := "二行目\nBob: 三行目"
	if got != want {
		t.Fatalf("translationContextForRow() = %q, want %q", got, want)
	}
}

func TestTranslationContextForRowSkipsCurrentRow(t *testing.T) {
	follower := newTranscriptFollower(nil, nil)
	follower.transcriptRows = []transcriptRow{
		{Key: "target", Text: "翻訳対象"},
	}

	if got := follower.translationContextForRow("target", 4); got != "" {
		t.Fatalf("expected no context, got %q", got)
	}
}
