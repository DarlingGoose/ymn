package anki

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/DarlingGoose/ymn/pkg/flashcard"
)

func TestInvokeRetriesEOF(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Fatalf("hijack response: %v", err)
			}
			_ = conn.Close()
			return
		}
		_ = json.NewEncoder(w).Encode(ankiResponse{Result: json.RawMessage(`true`)})
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		client:  &http.Client{Timeout: 2 * time.Second},
	}
	var result bool
	if err := client.invoke("version", nil, &result); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	if !result {
		t.Fatal("expected decoded result to be true")
	}
}

func TestAddNoteUsesYomunaModelFields(t *testing.T) {
	requests := make(chan ankiRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests <- req
		_ = json.NewEncoder(w).Encode(ankiResponse{Result: json.RawMessage(`123`)})
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		client:  &http.Client{Timeout: 2 * time.Second},
	}
	_, err := client.AddNote(flashcard.Flashcard{
		GameName:   "game",
		Text:       "しかも",
		Meaning:    "moreover",
		Reading:    "しかも",
		AnkiDeck:   "deck",
		AnkiModel:  flashcard.DefaultAnkiModel,
		SourceLine: "しかも、いい。",
	})
	if err != nil {
		t.Fatalf("add note: %v", err)
	}

	req := <-requests
	params := req.Params.(map[string]any)
	note := params["note"].(map[string]any)
	fields := note["fields"].(map[string]any)
	for _, name := range []string{"Text", "Reading", "Meaning", "Context", "Audio"} {
		if _, ok := fields[name]; !ok {
			t.Fatalf("expected field %q in %#v", name, fields)
		}
	}
	for _, name := range []string{"Front", "Back"} {
		if _, ok := fields[name]; ok {
			t.Fatalf("did not expect legacy field %q in %#v", name, fields)
		}
	}
}

func TestSyncFlashcardsSavesProgressOnLargeDeckFailure(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	gameName := "large deck"
	cards := make([]flashcard.Flashcard, ankiProgressSaveEach+2)
	for i := range cards {
		cards[i] = flashcard.Flashcard{
			GameName: gameName,
			Text:     "word-" + strconv.Itoa(i),
			Meaning:  "meaning",
		}
	}
	if err := flashcard.SaveFlashcards(gameName, cards); err != nil {
		t.Fatalf("save flashcards: %v", err)
	}

	var addNoteCalls int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ankiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		switch req.Action {
		case "createDeck":
			_ = json.NewEncoder(w).Encode(ankiResponse{Result: json.RawMessage(`true`)})
		case "addNote":
			addNoteCalls++
			if addNoteCalls > int64(ankiProgressSaveEach) {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Fatalf("hijack response: %v", err)
				}
				_ = conn.Close()
				return
			}
			_ = json.NewEncoder(w).Encode(ankiResponse{Result: json.RawMessage(strconv.FormatInt(addNoteCalls, 10))})
		default:
			t.Fatalf("unexpected action %q", req.Action)
		}
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		client:  &http.Client{Timeout: 2 * time.Second},
	}
	if _, err := client.SyncFlashcardsToAnki(gameName, server.URL, false); err == nil {
		t.Fatal("expected sync failure")
	}

	saved, err := flashcard.LoadFlashcards(gameName)
	if err != nil {
		t.Fatalf("load flashcards: %v", err)
	}
	for i := 0; i < ankiProgressSaveEach; i++ {
		if saved[i].AnkiNoteID == 0 {
			t.Fatalf("expected card %d progress to be saved", i)
		}
	}
	if saved[ankiProgressSaveEach].AnkiNoteID != 0 {
		t.Fatalf("did not expect failed card note id to be saved")
	}

	entries, err := os.ReadDir(filepath.Join(os.Getenv("XDG_CACHE_HOME"), "wgl", "flashcards"))
	if err != nil {
		t.Fatalf("read flashcard dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one flashcard file, got %d", len(entries))
	}
}
