package transcript

import (
	"testing"

	"github.com/DarlingGoose/vntext/pkg/engine"
)

func TestNormalizeHookGroupKeepsUsefulPrefix(t *testing.T) {
	hook := normalizeHookGroup(" Thread A@0 ")
	if hook != "@0" {
		t.Fatalf("expected hook group, got %q", hook)
	}
}

func TestHookMatchesFilterSupportsSavedGroupFilters(t *testing.T) {
	if !hookMatchesFilter("@13F548:KSH_dl.exe", "Thread A@13F548:KSH_dl.exe") {
		t.Fatal("expected saved hook group filter to match full hook")
	}

	if !hookMatchesFilter("Thread A@0", "Thread A@0") {
		t.Fatal("expected full hook filter to normalize and match its group")
	}

	if !hookMatchesFilter("Thread A@0", "Thread B@0") {
		t.Fatal("full hook filters should match another full hook with the same group")
	}
}

func TestTranscriptDisplayAddsHookLabelOnlyWhenEnabled(t *testing.T) {
	follower := newTranscriptFollower(nil, nil)
	row := transcriptRow{
		Key:  "line-1",
		Hook: "Thread A@0",
		Text: "こんにちは",
	}

	if got := follower.transcriptRowDisplayText(row); got != "こんにちは" {
		t.Fatalf("expected plain transcript text, got %q", got)
	}

	follower.SetShowHookLabels(true)
	if got := follower.transcriptRowDisplayText(row); got != "[@0] こんにちは" {
		t.Fatalf("expected hook-prefixed transcript text, got %q", got)
	}
}

func TestTranscriptRowFromEngineLineFallsBackToRawText(t *testing.T) {
	row := transcriptRowFromEngineLine(engine.Line{
		Raw: "Textractor: inserting hook: TextOutA",
	})

	if row.Text != "Textractor: inserting hook: TextOutA" {
		t.Fatalf("expected raw text fallback, got %q", row.Text)
	}
}

func TestTranscriptRowLooksLikeLanguageFiltersReplacementNoise(t *testing.T) {
	cases := []struct {
		name string
		row  transcriptRow
		want bool
	}{
		{name: "japanese", row: transcriptRow{Text: "こんにちは"}, want: true},
		{name: "speaker", row: transcriptRow{Speaker: "黒幕", Text: "「秘密だ」"}, want: true},
		{name: "replacement char", row: transcriptRow{Text: "�"}, want: false},
		{name: "empty square", row: transcriptRow{Text: "□"}, want: false},
		{name: "punctuation only", row: transcriptRow{Text: "!?..."}, want: false},
		{name: "info row", row: transcriptRow{Info: true, Text: "Following logs"}, want: true},
	}

	for _, tc := range cases {
		if got := transcriptRowLooksLikeLanguage(tc.row); got != tc.want {
			t.Fatalf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
