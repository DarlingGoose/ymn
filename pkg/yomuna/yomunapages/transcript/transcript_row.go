package transcript

import (
	"context"
	"strings"
	"sync"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/translation"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
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
}

func newTranscriptFollower(th *material.Theme) transcriptFollower {
	transcriptList := widget.List{}
	transcriptList.Axis = layout.Vertical
	transcriptList.ScrollToEnd = true
	return transcriptFollower{
		th:                           th,
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
		fontSize:               unit.Sp(16), //allow this to be dynamicly set
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

func (t *transcriptFollower) WithSelectedRow(sr func(row transcriptRow)) {
	t.selectedRow = sr
}
func (t *transcriptFollower) SetFoundSize(fontsize unit.Sp) {
}

func (t *transcriptFollower) SetGame(gameName string) {
	//clear last logs if diff than current
	// clear all maps as wel
}

func (t *transcriptFollower) Reset(gameName string) {
	//clear last logs if diff than current
	// clear all maps as wel
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

func (t *transcriptFollower) AddRows(rows ...transcriptRow) {
	//todo set selected to the newest row?
	t.rowMutex.Lock()
	defer t.rowMutex.Unlock()
	for _, r := range rows {
		if r.Key == "" {
			r.Key = uuid.NewString()
		}
		t.transcriptRows = append(t.transcriptRows, r)
	}
	if !t.transcriptRows[len(t.transcriptRows)-1].Info {
		t.selectedLineKey = t.transcriptRows[len(t.transcriptRows)-1].Key
		t.selectedLineText = t.transcriptRows[len(t.transcriptRows)-1].Key
	}
	if t.maxTranscriptRows <= 0 {
		return
	}
	if len(t.transcriptRowClicks) < t.maxTranscriptRows {
		return
	}

	//func (p *Page) pruneTranscriptRowClicks(rows []transcriptRow) {
	//	valid := make(map[string]struct{}, len(rows))
	//	validVoice := make(map[string]struct{}, len(rows))
	//	for _, row := range rows {
	//		if !row.Info {
	//			valid[row.Key] = struct{}{}
	//			validVoice[row.Key] = struct{}{}
	//		}
	//	}
	//	for key := range p.transcriptRowClicks {
	//		if _, ok := valid[key]; !ok {
	//			delete(p.transcriptRowClicks, key)
	//		}
	//	}
	//	for key := range p.transcriptRowTranslateClicks {
	//		if _, ok := valid[key]; !ok {
	//			delete(p.transcriptRowTranslateClicks, key)
	//			delete(p.rowTranslationShown, key)
	//		}
	//	}
	//	for key := range p.transcriptRowVoiceClicks {
	//		if _, ok := validVoice[key]; !ok {
	//			delete(p.transcriptRowVoiceClicks, key)
	//		}
	//	}
	//	if p.selectedLineKey != "" {
	//		if _, ok := valid[p.selectedLineKey]; !ok {
	//			p.selectedLineKey = ""
	//			p.selectedLineText = ""
	//		}
	//	}
	//}
	t.transcriptRows = t.transcriptRows[len(t.transcriptRows)-t.maxTranscriptRows:]
}
func (t *transcriptFollower) HandeEvents(gtx layout.Context) {
	for key, click := range t.transcriptRowClicks {
		for click.Clicked(gtx) {
			t.selectTranscriptRow(key)
		}
	}
	for key, click := range t.transcriptRowTranslateClicks {
		for click.Clicked(gtx) {
			t.toggleTranscriptRowTranslation(context.Background(), key)
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

func (t *transcriptFollower) generateTranscriptRowTranslation(ctx context.Context, w *app.Window, row transcriptRow, key string) {
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

		if w != nil {
			w.Invalidate()
		}
	}()
}
