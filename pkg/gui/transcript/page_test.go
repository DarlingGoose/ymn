package transcript

import (
	"strings"
	"testing"

	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
	"github.com/DarlingGoose/wgl/pkg/dictionary"
	flashcards "github.com/DarlingGoose/wgl/pkg/flashcard"
)

func TestTranscriptRowsKeepEscapedNewlinesInSpeakerLine(t *testing.T) {
	p := New(barethemes.New(barethemes.ModeDark, barethemes.PaletteMoonlitLibrary, true))
	p.SetRawTranscript(`[2026-05-04T12:30:19-07:00][speaker:キリア][voice:v_kiri0001]: 「無論だ中佐。情報将校ごときがコトを解決できると\n思わぬことだ。最後は武力がものを言う」`)

	rows := p.transcriptRows()
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d: %#v", len(rows), rows)
	}
	if rows[0].Speaker != "キリア" {
		t.Fatalf("expected speaker キリア, got %q", rows[0].Speaker)
	}
	if rows[0].Voice != "v_kiri0001" {
		t.Fatalf("expected voice v_kiri0001, got %q", rows[0].Voice)
	}
	if got := rows[0].Text; got != "「無論だ中佐。情報将校ごときがコトを解決できると 思わぬことだ。最後は武力がものを言う」" {
		t.Fatalf("unexpected row text: %q", got)
	}
}

func TestSetRawTranscriptSelectsLatestLineWhenNewLineArrives(t *testing.T) {
	p := New(barethemes.New(barethemes.ModeDark, barethemes.PaletteMoonlitLibrary, true))
	p.SetRawTranscript(strings.Join([]string{
		`[2026-05-04T12:30:19-07:00][speaker:キリア][voice:v_kiri0001]: 最初の行`,
		`[2026-05-04T12:30:20-07:00][speaker:キリア][voice:v_kiri0002]: 二番目の行`,
	}, "\n"))

	rows := p.transcriptRows()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d: %#v", len(rows), rows)
	}
	p.selectTranscriptRow(rows[0].Key)
	if got := p.structureSourceText(); got != "最初の行" {
		t.Fatalf("expected manual selection before update, got %q", got)
	}

	p.SetRawTranscript(strings.Join([]string{
		`[2026-05-04T12:30:19-07:00][speaker:キリア][voice:v_kiri0001]: 最初の行`,
		`[2026-05-04T12:30:20-07:00][speaker:キリア][voice:v_kiri0002]: 二番目の行`,
		`[2026-05-04T12:30:21-07:00][speaker:キリア][voice:v_kiri0003]: 最新の行`,
	}, "\n"))

	if got := p.structureSourceText(); got != "最新の行" {
		t.Fatalf("expected newest line to be focused, got %q", got)
	}
	if got := p.selectedLineText; got != "最新の行" {
		t.Fatalf("expected newest line to be selected, got %q", got)
	}
}

func TestLookupFlashcardExistsMatchesAddedWord(t *testing.T) {
	p := New(barethemes.New(barethemes.ModeDark, barethemes.PaletteMoonlitLibrary, true))
	p.activeGameName = "test game"
	p.flashcards = []flashcards.Flashcard{{
		GameName: "test game",
		Text:     "言葉",
		Meaning:  "word",
		Reading:  "ことば",
	}}

	lookup := dictionary.Lookup{
		Query:   "言葉",
		Meaning: "word",
		Reading: "ことば",
	}
	if !p.lookupFlashcardExists(lookup) {
		t.Fatal("expected lookup to match existing flashcard")
	}
}

func TestLookupFlashcardExistsMatchesFocusedTokenBaseForm(t *testing.T) {
	p := New(barethemes.New(barethemes.ModeDark, barethemes.PaletteMoonlitLibrary, true))
	p.activeGameName = "test game"
	p.SetRawTranscript(`勉強した`)

	analysis, errText := p.currentStructureAnalysis()
	if errText != "" {
		t.Fatalf("unexpected analysis error: %s", errText)
	}
	var selected bool
	for _, token := range analysis.Tokens {
		if token.Surface == "し" && token.BaseForm == "する" {
			p.selectedFocusedTokenKey = structureTokenKey(token)
			p.selectedFocusedTokenWord = token.BaseForm
			selected = true
			break
		}
	}
	if !selected {
		t.Fatalf("expected conjugated し token in analysis: %#v", analysis.Tokens)
	}
	p.flashcards = []flashcards.Flashcard{{
		GameName: "test game",
		Text:     "し",
		Meaning:  "do",
	}}

	lookup := dictionary.Lookup{
		Query:   "する",
		Meaning: "do",
	}
	if !p.lookupFlashcardExists(lookup) {
		t.Fatal("expected base-form lookup to match existing conjugated token flashcard")
	}
}
