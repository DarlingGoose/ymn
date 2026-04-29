package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultAnkiConnectURL = "http://127.0.0.1:8765"

type ankiRequest struct {
	Action  string `json:"action"`
	Version int    `json:"version"`
	Params  any    `json:"params,omitempty"`
}

type ankiResponse struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

type ankiClient struct {
	baseURL string
	client  *http.Client
}

type ankiSyncResult struct {
	DeckName string
	Total    int
	Created  int
	Updated  int
}

func newAnkiClient(baseURL string) *ankiClient {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultAnkiConnectURL
	}
	return &ankiClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ankiClient) invoke(action string, params any, out any) error {
	payload, err := json.Marshal(ankiRequest{
		Action:  action,
		Version: 6,
		Params:  params,
	})
	if err != nil {
		return fmt.Errorf("encode anki request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create anki request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call anki at %s: %w", c.baseURL, err)
	}
	defer resp.Body.Close()

	var parsed ankiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("decode anki response: %w", err)
	}
	if parsed.Error != "" {
		return fmt.Errorf("anki action %s failed: %s", action, parsed.Error)
	}
	if out == nil {
		return nil
	}
	if len(parsed.Result) == 0 || string(parsed.Result) == "null" {
		return nil
	}
	if err := json.Unmarshal(parsed.Result, out); err != nil {
		return fmt.Errorf("decode anki result: %w", err)
	}
	return nil
}

func (c *ankiClient) createDeck(name string) error {
	var result any
	return c.invoke("createDeck", map[string]string{"deck": name}, &result)
}

func (c *ankiClient) addNote(card Flashcard) (int64, error) {
	var result int64
	audioTag, err := c.storeFlashcardAudio(card)
	if err != nil {
		return 0, err
	}
	err = c.invoke("addNote", map[string]any{
		"note": map[string]any{
			"deckName":  card.AnkiDeck,
			"modelName": card.AnkiModel,
			"fields": map[string]string{
				"Front":   html.EscapeString(card.Text),
				"Back":    flashcardBackHTML(card, audioTag),
				"Meaning": card.Meaning,
				"Reading": card.Reading,
				"Text":    card.Text,
				"Audio":   audioTag,
			},
			"options": map[string]any{
				"allowDuplicate": false,
				"duplicateScope": "deck",
			},
			"tags": []string{"wgl", sanitizeName(card.GameName)},
		},
	}, &result)
	if err != nil {
		return 0, err
	}
	if result == 0 {
		return 0, fmt.Errorf("anki did not return a note id")
	}
	return result, nil
}

func (c *ankiClient) updateNote(card Flashcard) error {
	audioTag, err := c.storeFlashcardAudio(card)
	if err != nil {
		return err
	}
	return c.invoke("updateNoteFields", map[string]any{
		"note": map[string]any{
			"id": card.AnkiNoteID,
			"fields": map[string]string{
				"Front": html.EscapeString(card.Text),
				"Back":  flashcardBackHTML(card, audioTag),
			},
		},
	}, nil)
}

func (c *ankiClient) storeMediaFile(filename string, data []byte) error {
	var result string
	return c.invoke("storeMediaFile", map[string]any{
		"filename": filename,
		"data":     base64.StdEncoding.EncodeToString(data),
	}, &result)
}

func (c *ankiClient) storeFlashcardAudio(card Flashcard) (string, error) {
	if !isExistingFile(card.AudioPath) {
		return "", nil
	}

	data, err := os.ReadFile(card.AudioPath)
	if err != nil {
		return "", fmt.Errorf("read audio for %q: %w", card.Text, err)
	}

	filename := ankiAudioFilename(card)
	if err := c.storeMediaFile(filename, data); err != nil {
		return "", fmt.Errorf("store audio for %q: %w", card.Text, err)
	}
	return "[sound:" + filename + "]", nil
}

func (c *ankiClient) sync() error {
	var result any
	return c.invoke("sync", nil, &result)
}

func (c *ankiClient) deleteNotes(noteIDs []int64) error {
	if len(noteIDs) == 0 {
		return nil
	}
	var result any
	return c.invoke("deleteNotes", map[string]any{
		"notes": noteIDs,
	}, &result)
}

func deleteFlashcardFromAnki(card Flashcard, baseURL string, pushSync bool) error {
	if card.AnkiNoteID <= 0 {
		return nil
	}

	client := newAnkiClient(baseURL)
	if err := client.deleteNotes([]int64{card.AnkiNoteID}); err != nil {
		return err
	}
	if pushSync {
		if err := client.sync(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ankiClient) ensureModel() error {
	return c.invoke("createModel", map[string]any{
		"modelName": defaultAnkiModel,
		"inOrderFields": []string{
			"Text",
			"Reading",
			"Meaning",
			"Context",
			"Audio",
		},
		"css": `
.card {
  font-family: sans-serif;
  font-size: 28px;
  text-align: center;
  color: black;
  background-color: white;
}

.front {
  font-size: 42px;
  font-weight: 600;
}

.reading {
  font-size: 24px;
  color: #666;
  margin-top: 0.5rem;
}

.meaning {
  font-size: 32px;
}

.context {
  font-size: 20px;
  color: #777;
  margin-top: 1rem;
}
`,
		"cardTemplates": []map[string]string{
			{
				"Name": "Text → Meaning",
				"Front": `
<div class="front">{{Text}}</div>
<div class="reading">{{Reading}}</div>
`,
				"Back": `
{{FrontSide}}
<hr>
<div class="meaning">{{Meaning}}</div>
<div class="context">{{Context}}</div>
<br><br>
{{Audio}}
`,
			},
			{
				"Name": "Meaning → Text",
				"Front": `
<div class="front meaning">{{Meaning}}</div>
`,
				"Back": `
{{FrontSide}}
<hr>
<div class="front">{{Text}}</div>
<div class="reading">{{Reading}}</div>
<div class="context">{{Context}}</div>
<br><br>
{{Audio}}
`,
			},
		},
	}, nil)
}

func syncFlashcardsToAnki(gameName, baseURL string, pushSync bool) (ankiSyncResult, error) {
	cards, err := loadFlashcards(gameName)
	if err != nil {
		return ankiSyncResult{}, err
	}
	if len(cards) == 0 {
		return ankiSyncResult{}, fmt.Errorf("no flashcards saved for %s", gameName)
	}

	client := newAnkiClient(baseURL)
	_ = client.ensureModel()

	deckName := ankiDeckName(gameName)
	if err := client.createDeck(deckName); err != nil {
		return ankiSyncResult{}, err
	}

	result := ankiSyncResult{
		DeckName: deckName,
		Total:    len(cards),
	}
	now := time.Now().UTC()

	for i := range cards {
		cards[i] = normalizeFlashcard(cards[i])
		cards[i].AnkiDeck = deckName
		cards[i].AnkiModel = defaultAnkiModel

		if cards[i].AnkiNoteID > 0 {
			if err := client.updateNote(cards[i]); err != nil {
				if strings.Contains(err.Error(), "Note was not found") {
					noteID, err := client.addNote(cards[i])
					if err != nil {
						return ankiSyncResult{}, fmt.Errorf("sync card %q: %w", cards[i].Text, err)
					}
					cards[i].AnkiNoteID = noteID
					result.Created++
				} else {
					return ankiSyncResult{}, fmt.Errorf("sync card %q: %w", cards[i].Text, err)

				}
			} else {
				result.Updated++
			}

		} else {
			noteID, err := client.addNote(cards[i])
			if err != nil {
				return ankiSyncResult{}, fmt.Errorf("sync card %q: %w", cards[i].Text, err)
			}
			cards[i].AnkiNoteID = noteID
			result.Created++
		}
		cards[i].UpdatedAt = now
		cards[i].AnkiLastSyncAt = &now
	}

	if err := saveFlashcards(gameName, cards); err != nil {
		return ankiSyncResult{}, err
	}
	if pushSync {
		if err := client.sync(); err != nil {
			return ankiSyncResult{}, err
		}
	}

	return result, nil
}

func flashcardBackHTML(card Flashcard, audioTag string) string {
	var parts []string
	parts = append(parts, escapedTextHTML(card.Meaning))
	if furigana := flashcardFuriganaHTML(card); strings.TrimSpace(furigana) != "" {
		parts = append(parts, "<small>Furigana: "+furigana+"</small>")
	}
	if strings.TrimSpace(card.Reading) != "" {
		parts = append(parts, "<small>Reading: "+html.EscapeString(card.Reading)+"</small>")
	}
	if strings.TrimSpace(card.PronunciationText) != "" {
		line := "<small>Pronunciation: " + html.EscapeString(card.PronunciationText)
		if strings.TrimSpace(card.PronunciationPitch) != "" {
			line += " (" + html.EscapeString(card.PronunciationPitch) + ")"
		}
		line += "</small>"
		parts = append(parts, line)
	}
	if strings.TrimSpace(audioTag) != "" {
		parts = append(parts, audioTag)
	}
	if strings.TrimSpace(card.SourceLine) != "" {
		parts = append(parts, "<small>"+html.EscapeString(card.SourceLine)+"</small>")
	}
	return strings.Join(parts, "<br><br>")
}

func escapedTextHTML(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	escaped := make([]string, 0, len(lines))
	for _, line := range lines {
		escaped = append(escaped, html.EscapeString(strings.TrimSpace(line)))
	}
	return strings.Join(escaped, "<br>")
}

func ankiAudioFilename(card Flashcard) string {
	ext := filepath.Ext(card.AudioPath)
	if ext == "" {
		ext = ".mp3"
	}
	return fmt.Sprintf("wgl-%s-audio%s", sanitizeName(card.ID), ext)
}

func exportFlashcardsToTSV(gameName string, cards []Flashcard) (string, error) {
	if err := os.MkdirAll(flashcardExportDir(), 0o755); err != nil {
		return "", fmt.Errorf("create export directory: %w", err)
	}

	path := filepath.Join(flashcardExportDir(), sanitizeName(gameName)+"-anki.tsv")
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create export file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	rows := [][]string{
		{"Front", "Back", "Source", "Deck"},
	}
	for _, card := range cards {
		rows = append(rows, []string{
			card.Text,
			card.Meaning,
			card.SourceLine,
			ankiDeckName(gameName),
		})
	}
	if err := writer.WriteAll(rows); err != nil {
		return "", fmt.Errorf("write export file: %w", err)
	}
	return path, nil
}
