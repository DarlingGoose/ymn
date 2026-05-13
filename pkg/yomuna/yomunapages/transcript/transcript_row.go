package transcript

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/translation"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
	"github.com/DarlingGoose/ymn/pkg/yomuna/backend"
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
	th                             *material.Theme
	tc                             *theme.Client
	transcriptList                 widget.List
	transcriptRowClicks            map[string]*widget.Clickable
	transcriptRowTranslateClicks   map[string]*widget.Clickable
	transcriptRowRetranslateClicks map[string]*widget.Clickable

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
	invalidate    func()
}

func newTranscriptFollower(th *material.Theme, backend backend.Backend) transcriptFollower {
	transcriptList := widget.List{}
	transcriptList.Axis = layout.Vertical
	transcriptList.ScrollToEnd = true
	return transcriptFollower{
		backend:                        backend,
		th:                             th,
		compactTimestamps:              true,
		tc:                             theme.DefaultThemeClient,
		transcriptList:                 transcriptList,
		transcriptRowClicks:            make(map[string]*widget.Clickable),
		transcriptRowTranslateClicks:   make(map[string]*widget.Clickable),
		transcriptRowRetranslateClicks: make(map[string]*widget.Clickable),
		rowTranslations:                make(map[string]string),
		rowTranslationGenerating:       make(map[string]bool),
		rowTranslationShown:            map[string]bool{},

		rowMutex:               sync.RWMutex{},
		transcriptRows:         make([]transcriptRow, 0),
		maxTranscriptRows:      200,
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

func (t *transcriptFollower) WithInvalidate(invalidate func()) *transcriptFollower {
	t.invalidate = invalidate
	return t
}

func (t *transcriptFollower) WithSelectedRow(sr func(row transcriptRow)) {
	t.selectedRow = sr
}

func (t *transcriptFollower) MaxTranscriptRows() int {
	t.rowMutex.RLock()
	defer t.rowMutex.RUnlock()

	return clampTranscriptRowLimit(t.maxTranscriptRows)
}

func (t *transcriptFollower) SetMaxTranscriptRows(maxRows int) {
	t.rowMutex.Lock()
	defer t.rowMutex.Unlock()

	t.maxTranscriptRows = clampTranscriptRowLimit(maxRows)
	if len(t.transcriptRows) > t.maxTranscriptRows {
		t.transcriptRows = t.transcriptRows[len(t.transcriptRows)-t.maxTranscriptRows:]
		t.pruneTranscriptRowStateLocked()
	}
}

func clampTranscriptRowLimit(maxRows int) int {
	switch {
	case maxRows <= 0:
		return 200
	case maxRows < 25:
		return 25
	case maxRows > 5000:
		return 5000
	default:
		return maxRows
	}
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
	t.transcriptRowRetranslateClicks = map[string]*widget.Clickable{}
	//t.transcriptRowVoiceClicks = map[string]*widget.Clickable{}
	t.rowTranslationShown = map[string]bool{}

	t.selectedLineKey = ""
	t.selectedLineText = ""
}

func (t *transcriptFollower) AddRows(rows ...transcriptRow) {
	t.rowMutex.Lock()

	autoTranslateRows := make([]transcriptRow, 0, len(rows))
	for _, r := range rows {
		if r.Key == "" {
			r.Key = uuid.NewString()
		}
		if t.autoTranslate && !r.Info {
			t.rowTranslationShown[r.Key] = true
			autoTranslateRows = append(autoTranslateRows, r)
		}
		t.transcriptRows = append(t.transcriptRows, r)
	}

	var selected *transcriptRow
	if len(t.transcriptRows) > 0 {
		last := t.transcriptRows[len(t.transcriptRows)-1]
		if !last.Info {
			t.selectedLineKey = last.Key
			t.selectedLineText = last.Text // not last.Key
			row := last
			selected = &row
		}
	}

	if t.maxTranscriptRows <= 0 {
		t.rowMutex.Unlock()
		if selected != nil {
			t.selectedRow(*selected)
		}
		for _, row := range autoTranslateRows {
			t.showTranscriptRowTranslation(context.Background(), row)
		}
		t.invalidateUI()
		return
	}

	if len(t.transcriptRows) > t.maxTranscriptRows {
		t.transcriptRows = t.transcriptRows[len(t.transcriptRows)-t.maxTranscriptRows:]
	}

	t.pruneTranscriptRowStateLocked()
	t.rowMutex.Unlock()

	if selected != nil {
		t.selectedRow(*selected)
	}
	for _, row := range autoTranslateRows {
		t.showTranscriptRowTranslation(context.Background(), row)
	}
	t.invalidateUI()
}

func (t *transcriptFollower) invalidateUI() {
	if t.invalidate != nil {
		t.invalidate()
	}
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
	for key := range t.transcriptRowRetranslateClicks {
		if _, ok := valid[key]; !ok {
			delete(t.transcriptRowRetranslateClicks, key)
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
	rows := make([]transcriptRow, len(t.transcriptRows))
	copy(rows, t.transcriptRows)
	return rows
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
	for key, click := range t.transcriptRowRetranslateClicks {
		for click.Clicked(gtx) {
			t.forceTranscriptRowTranslation(context.Background(), key)
		}
	}
}
func (t *transcriptFollower) transcriptRowByKey(key string) (transcriptRow, bool) {
	for _, row := range t.GetRows() {
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
	t.rowMutex.RLock()
	selectedLineKey := t.selectedLineKey
	t.rowMutex.RUnlock()
	if selectedLineKey != "" {
		return selectedLineKey
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
	key := t.rowTranslationCacheKey(row)

	t.rowMutex.RLock()
	defer t.rowMutex.RUnlock()

	if !t.rowTranslationShown[row.Key] {

		return row.Text
	}
	if t.rowTranslationGenerating[key] {
		return "Translating..."
	}
	if text := strings.TrimSpace(t.rowTranslations[key]); text != "" {
		return text
	}
	return row.Text
}

func (t *transcriptFollower) isTranscriptRowTranslationShown(row transcriptRow) bool {
	t.rowMutex.RLock()
	defer t.rowMutex.RUnlock()
	return t.rowTranslationShown[row.Key]
}

func (t *transcriptFollower) isTranscriptRowTranslationGenerating(row transcriptRow) bool {
	key := t.rowTranslationCacheKey(row)
	if key == "" {
		return false
	}
	t.rowMutex.RLock()
	defer t.rowMutex.RUnlock()
	return t.rowTranslationGenerating[key]
}

func (t *transcriptFollower) rowTranslationCacheKey(row transcriptRow) string {
	source := utils.CleanInlineText(row.Text)
	t.rowMutex.RLock()
	targetLanguage := strings.TrimSpace(t.selectedTargetLanguage)
	activeGameName := strings.TrimSpace(t.activeGameName)
	t.rowMutex.RUnlock()
	if source == "" || targetLanguage == "" {
		return ""
	}
	return activeGameName + "\x00" + source + "\x00" + strings.ToLower(targetLanguage)
}

func (t *transcriptFollower) selectTranscriptRow(key string) {
	for _, row := range t.GetRows() {
		if row.Key != key {
			continue
		}
		if row.Info {
			return
		}
		t.rowMutex.Lock()
		t.selectedLineKey = row.Key
		t.selectedLineText = row.Text
		t.rowMutex.Unlock()
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

func (t *transcriptFollower) transcriptRowRetranslateClickable(key string) *widget.Clickable {
	if t.transcriptRowRetranslateClicks == nil {
		t.transcriptRowRetranslateClicks = make(map[string]*widget.Clickable)
	}
	if t.transcriptRowRetranslateClicks[key] == nil {
		t.transcriptRowRetranslateClicks[key] = new(widget.Clickable)
	}
	return t.transcriptRowRetranslateClicks[key]
}

func (t *transcriptFollower) showTranscriptRowTranslation(ctx context.Context, row transcriptRow) {
	if row.Info {
		return
	}
	key := t.rowTranslationCacheKey(row)
	if key == "" {
		return
	}

	t.rowMutex.Lock()
	if _, ok := t.rowTranslations[key]; ok {
		t.rowTranslationShown[row.Key] = true
		t.rowMutex.Unlock()
		t.invalidateUI()
		return
	}
	if t.rowTranslationGenerating[key] {
		t.rowTranslationShown[row.Key] = true
		t.rowMutex.Unlock()
		t.invalidateUI()
		return
	}
	t.rowTranslationShown[row.Key] = true
	t.rowTranslationGenerating[key] = true
	t.rowMutex.Unlock()

	t.invalidateUI()
	t.generateTranscriptRowTranslation(ctx, row, key)
}

func (t *transcriptFollower) forceTranscriptRowTranslation(ctx context.Context, rowKey string) {
	row, ok := t.transcriptRowByKey(rowKey)
	if !ok || row.Info {
		return
	}
	key := t.rowTranslationCacheKey(row)
	if key == "" {
		return
	}

	t.rowMutex.Lock()
	if t.rowTranslationGenerating[key] {
		t.rowTranslationShown[row.Key] = true
		t.rowMutex.Unlock()
		t.invalidateUI()
		return
	}
	delete(t.rowTranslations, key)
	t.rowTranslationShown[row.Key] = true
	t.rowTranslationGenerating[key] = true
	t.rowMutex.Unlock()

	t.invalidateUI()
	t.generateTranscriptRowTranslation(ctx, row, key, true)
}

func (t *transcriptFollower) generateTranscriptRowTranslation(ctx context.Context, row transcriptRow, key string, force ...bool) {
	if key == "" {
		return
	}
	source := utils.CleanInlineText(row.Text)
	if source == "" {
		return
	}
	t.rowMutex.RLock()
	gameName := t.activeGameName
	targetLanguage := t.selectedTargetLanguage
	cfg := t.translatorConfig
	forceGenerate := len(force) > 0 && force[0]
	t.rowMutex.RUnlock()
	go func() {
		var entry translation.Entry
		var err error
		if forceGenerate {
			entry, err = translation.Generate(ctx, cfg, gameName, source, targetLanguage)
		} else {
			var ok bool
			entry, ok, err = translation.Load(gameName, source, targetLanguage)
			if err != nil {
				slog.Error("failed loading transcript row translation", "err", err)
			}
			if !ok && err == nil {
				entry, err = translation.Generate(ctx, cfg, gameName, source, targetLanguage)
			}
		}
		result := rowTranslationResult{Key: key, RowKey: row.Key, Entry: entry, Err: err}

		t.rowMutex.Lock()

		delete(t.rowTranslationGenerating, result.Key)
		if result.Err != nil {
			slog.Error("failed generating transcript row translation", "err", result.Err)
			t.rowTranslationShown[result.RowKey] = false
			t.rowMutex.Unlock()
			t.invalidateUI()
			return
		}
		if translationText := strings.TrimSpace(result.Entry.Translation); translationText != "" {
			t.rowTranslations[result.Key] = translationText
			t.rowTranslationShown[result.RowKey] = true
		}
		t.rowMutex.Unlock()
		t.invalidateUI()

	}()
}
