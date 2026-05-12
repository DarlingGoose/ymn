package transcript

import (
	"context"
	"strings"
	"sync"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/translation"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
	"github.com/DarlingGoose/wgl/pkg/yomuna/backend"
	"github.com/google/uuid"
)

type transcriptRow struct {
	Key     string
	Hook    string
	Text    string
	Speaker string
	Raw     string
	Info    bool
	Time    string
}
type rowTranslationResult struct {
	Key    string
	RowKey string
	Entry  translation.Entry
	Err    error
}

type transcriptFollower struct {
	th                           *material.Theme
	tc                           *theme.Client
	transcriptList               widget.List
	transcriptRowClicks          map[string]*widget.Clickable
	transcriptRowTranslateClicks map[string]*widget.Clickable

	selectedLineKey  string
	selectedLineText string

	rowMutex          sync.RWMutex
	transcriptRows    []transcriptRow
	maxTranscriptRows int

	fontSize          unit.Sp
	radius            unit.Dp
	iconRadius        unit.Dp
	compactTimestamps bool

	rowTranslationShown      map[string]bool
	rowTranslations          map[string]string
	rowTranslationGenerating map[string]bool

	selectedTargetLanguage string

	activeGameName   string
	selectedRow      func(row transcriptRow)
	translatorConfig translation.Config

	backend backend.Backend

	autoTranslate bool
}

func newTranscriptFollower(th *material.Theme, backend backend.Backend) transcriptFollower {
	transcriptList := widget.List{}
	transcriptList.Axis = layout.Vertical
	transcriptList.ScrollToEnd = true
	return transcriptFollower{
		backend:                      backend,
		th:                           th,
		compactTimestamps:            true,
		tc:                           theme.DefaultThemeClient,
		transcriptList:               transcriptList,
		transcriptRowClicks:          make(map[string]*widget.Clickable),
		transcriptRowTranslateClicks: make(map[string]*widget.Clickable),
		rowTranslations:              make(map[string]string),
		rowTranslationGenerating:     make(map[string]bool),
		rowTranslationShown:          map[string]bool{},

		rowMutex:               sync.RWMutex{},
		transcriptRows:         make([]transcriptRow, 0),
		maxTranscriptRows:      200,         //todo add way to set this
		fontSize:               unit.Sp(22), //allow this to be dynamicly set
		radius:                 unit.Dp(12),
		iconRadius:             unit.Dp(8),
		selectedTargetLanguage: "english", //todo add way to set this
		activeGameName:         "",        //todo add way to set this
		selectedRow: func(row transcriptRow) {

		},
	}
}

func (t *transcriptFollower) WithTranslatorConfig(cfg translation.Config) *transcriptFollower {
	t.translatorConfig = cfg
	return t
}

func (t *transcriptFollower) WithAutoTranslate(at bool) *transcriptFollower {
	t.autoTranslate = at
	return t
}
func (t *transcriptFollower) WithSelectedRow(sr func(row transcriptRow)) {
	t.selectedRow = sr
}

func (t *transcriptFollower) SetGame(gameName string) {
	t.rowMutex.Lock()
	defer t.rowMutex.Unlock()

	if t.activeGameName == gameName {
		return
	}

	t.activeGameName = gameName
	t.resetLocked()
}

func (t *transcriptFollower) Reset(gameName string) {
	t.rowMutex.Lock()
	defer t.rowMutex.Unlock()

	t.activeGameName = gameName
	t.resetLocked()
}

func (t *transcriptFollower) resetLocked() {
	t.transcriptRows = nil
	t.transcriptRowClicks = map[string]*widget.Clickable{}
	t.transcriptRowTranslateClicks = map[string]*widget.Clickable{}
	//t.transcriptRowVoiceClicks = map[string]*widget.Clickable{}
	t.rowTranslationShown = map[string]bool{}

	t.selectedLineKey = ""
	t.selectedLineText = ""
}

func (t *transcriptFollower) AddRows(rows ...transcriptRow) {
	t.rowMutex.Lock()
	defer t.rowMutex.Unlock()

	for _, r := range rows {
		if r.Key == "" {
			r.Key = uuid.NewString()
		}
		if t.autoTranslate {
			//todo idk fix
			t.transcriptRowTranslateClickable(r.Key).Click()
			t.rowTranslationShown[r.Key] = true

		}
		t.transcriptRows = append(t.transcriptRows, r)
	}

	if len(t.transcriptRows) > 0 {
		last := t.transcriptRows[len(t.transcriptRows)-1]
		if !last.Info {
			t.selectedLineKey = last.Key
			t.selectedLineText = last.Text // not last.Key
		}
	}

	if t.maxTranscriptRows <= 0 {
		return
	}

	if len(t.transcriptRows) > t.maxTranscriptRows {
		t.transcriptRows = t.transcriptRows[len(t.transcriptRows)-t.maxTranscriptRows:]
	}

	t.pruneTranscriptRowStateLocked()
}

func (t *transcriptFollower) pruneTranscriptRowStateLocked() {
	valid := make(map[string]struct{}, len(t.transcriptRows))

	for _, row := range t.transcriptRows {
		if !row.Info {
			valid[row.Key] = struct{}{}
		}
	}

	for key := range t.transcriptRowClicks {
		if _, ok := valid[key]; !ok {
			delete(t.transcriptRowClicks, key)
		}
	}

	for key := range t.transcriptRowTranslateClicks {
		if _, ok := valid[key]; !ok {
			delete(t.transcriptRowTranslateClicks, key)
			delete(t.rowTranslationShown, key)
		}
	}

	//for key := range t.transcriptRowVoiceClicks {
	//	if _, ok := valid[key]; !ok {
	//		delete(t.transcriptRowVoiceClicks, key)
	//	}
	//}

	if t.selectedLineKey != "" {
		if _, ok := valid[t.selectedLineKey]; !ok {
			t.selectedLineKey = ""
			t.selectedLineText = ""
		}
	}
}

func (t *transcriptFollower) SetFoundSize(fontsize unit.Sp) {
}

func (t *transcriptFollower) SetCompactTimestamp(compact bool) {
	//clear last logs if diff than current
	// clear all maps as wel

}

func (t *transcriptFollower) WithTargetLanguage(targetLanguage string) {

}

func (t *transcriptFollower) WithThemeClient(tc *theme.Client) *transcriptFollower {
	t.tc = tc
	return t
}

func (t *transcriptFollower) GetRows() []transcriptRow {
	t.rowMutex.RLock()
	defer t.rowMutex.RUnlock()
	return t.transcriptRows
}

func (t *transcriptFollower) HandeEvents(gtx layout.Context) {
	for key, click := range t.transcriptRowClicks {
		for click.Clicked(gtx) {
			t.selectTranscriptRow(key)
		}
	}

	for key, click := range t.transcriptRowTranslateClicks {
		for click.Clicked(gtx) {
			t.toggleTranscriptRowTranslation(gtx, context.Background(), key)
		}
	}
}
func (t *transcriptFollower) transcriptRowByKey(key string) (transcriptRow, bool) {
	for _, row := range t.transcriptRows {
		if row.Key == key {
			return row, true
		}
	}
	return transcriptRow{}, false
}

func (t *transcriptFollower) transcriptRowClickable(key string) *widget.Clickable {
	if t.transcriptRowClicks == nil {
		t.transcriptRowClicks = make(map[string]*widget.Clickable)
	}
	if t.transcriptRowClicks[key] == nil {
		t.transcriptRowClicks[key] = new(widget.Clickable)
	}
	return t.transcriptRowClicks[key]
}

func (t *transcriptFollower) currentTranscriptRowKey() string {
	if t.selectedLineKey != "" {
		return t.selectedLineKey
	}
	rows := t.GetRows()
	for i := len(rows) - 1; i >= 0; i-- {
		if !rows[i].Info {
			return rows[i].Key
		}
	}
	return ""
}

func (t *transcriptFollower) transcriptRowDisplayText(row transcriptRow) string {
	if !t.isTranscriptRowTranslationShown(row) {

		return row.Text
	}
	key := t.rowTranslationCacheKey(row)
	if t.rowTranslationGenerating[key] {
		return "Translating..."
	}
	if text := strings.TrimSpace(t.rowTranslations[key]); text != "" {
		return text
	}
	return row.Text
}

func (t *transcriptFollower) isTranscriptRowTranslationShown(row transcriptRow) bool {
	return t.rowTranslationShown[row.Key]
}

func (t *transcriptFollower) rowTranslationCacheKey(row transcriptRow) string {
	source := utils.CleanInlineText(row.Text)
	targetLanguage := strings.TrimSpace(t.selectedTargetLanguage)
	if source == "" || targetLanguage == "" {
		return ""
	}
	return strings.TrimSpace(t.activeGameName) + "\x00" + source + "\x00" + strings.ToLower(targetLanguage)
}

func (t *transcriptFollower) selectTranscriptRow(key string) {
	for _, row := range t.GetRows() {
		if row.Key != key {
			continue
		}
		if row.Info {
			return
		}
		t.selectedLineKey = row.Key
		t.selectedLineText = row.Text
		t.selectedRow(row)
		return
	}
}

func (t *transcriptFollower) transcriptRowTranslateClickable(key string) *widget.Clickable {
	if t.transcriptRowTranslateClicks == nil {
		t.transcriptRowTranslateClicks = make(map[string]*widget.Clickable)
	}
	if t.transcriptRowTranslateClicks[key] == nil {
		t.transcriptRowTranslateClicks[key] = new(widget.Clickable)
	}
	return t.transcriptRowTranslateClicks[key]
}

func (t *transcriptFollower) generateTranscriptRowTranslation(gtx layout.Context, ctx context.Context, row transcriptRow, key string) {
	if key == "" || t.rowTranslationGenerating[key] {
		return
	}
	source := utils.CleanInlineText(row.Text)
	if source == "" {
		return
	}
	t.rowTranslationGenerating[key] = true
	gameName := t.activeGameName
	targetLanguage := t.selectedTargetLanguage
	cfg := t.translatorConfig
	go func() {
		entry, err := translation.Generate(ctx, cfg, gameName, source, targetLanguage)
		result := rowTranslationResult{Key: key, RowKey: row.Key, Entry: entry, Err: err}
		delete(t.rowTranslationGenerating, result.Key)
		t.rowTranslations[result.Key] = result.Entry.Translation
		t.rowTranslationShown[result.RowKey] = true

		gtx.Execute(op.InvalidateCmd{})

	}()
}
