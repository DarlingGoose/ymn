package transcript

import (
	"fmt"
	"image/color"
	"strings"
	"sync"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/jpndict"
	"github.com/DarlingGoose/wgl/pkg/japanese"
	"github.com/DarlingGoose/wgl/pkg/translation"
	"github.com/DarlingGoose/wgl/pkg/util"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
	"github.com/DarlingGoose/wgl/pkg/yomuna/backend"
)

const (
	focusedFuriganaHidden = "hidden"
	focusedFuriganaAbove  = "above"
	focusedFuriganaBelow  = "below"
)

type SentenceAnalysis struct {
	th                     *material.Theme
	tc                     *theme.Client
	backend                backend.Backend
	selectedTargetLanguage string
	translatorConfig       translation.Config

	focusedTokenClicks  map[string]*widget.Clickable
	furiganaHiddenClick widget.Clickable
	furiganaAboveClick  widget.Clickable
	lookupFontDownClick widget.Clickable
	lookupFontUpClick   widget.Clickable
	lookupAudioClicks   map[string]*widget.Clickable

	autoTranslate bool

	focusedFuriganaMode    string
	focusedFuriganaDefault string

	selectedFocusedTokenKey  string
	selectedFocusedTokenWord string
	selectedFocusedTokenNote string
	focusedLookupPendingKey  string

	lookupMu         sync.RWMutex
	lookupGeneration int
	lookupResults    []*jpndict.Response
	lookupError      string
	lookupPending    bool
	lookupQuery      string
	lookupTokenKey   string
	lookupList       widget.List
	lookupFontSize   unit.Sp
	lookupAudio      map[string]*lookupAudioState
	invalidate       func()

	sentenceFontSize  unit.Sp
	furigiganFontSize unit.Sp
	structureList     widget.List

	line *transcriptRow
}

func NewSentenceAnalysis(th *material.Theme, backend backend.Backend) *SentenceAnalysis {
	structureList := widget.List{}
	structureList.Axis = layout.Vertical
	lookupList := widget.List{}
	lookupList.Axis = layout.Vertical
	return &SentenceAnalysis{
		tc:                     theme.DefaultThemeClient,
		backend:                backend,
		th:                     th,
		selectedTargetLanguage: "english",
		focusedFuriganaMode:    focusedFuriganaAbove,
		focusedFuriganaDefault: focusedFuriganaAbove,
		sentenceFontSize:       unit.Sp(24),
		furigiganFontSize:      unit.Sp(12),
		structureList:          structureList,
		lookupList:             lookupList,
		lookupFontSize:         unit.Sp(14),
		focusedTokenClicks:     make(map[string]*widget.Clickable),
		lookupAudioClicks:      make(map[string]*widget.Clickable),
		lookupAudio:            make(map[string]*lookupAudioState),
	}
}

type lookupAudioState struct {
	Query   string
	Pending bool
	Cached  bool
	Error   string
	Resp    *jpndict.Response
}

func (t *SentenceAnalysis) WithTranslatorConfig(cfg translation.Config) *SentenceAnalysis {
	t.translatorConfig = cfg
	return t
}

func (t *SentenceAnalysis) WithAutoTranslate(at bool) *SentenceAnalysis {
	t.autoTranslate = at
	return t
}

func (t *SentenceAnalysis) WithThemeClient(tc *theme.Client) *SentenceAnalysis {
	if t == nil {
		return t
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	t.tc = tc
	return t
}

func (t *SentenceAnalysis) WithInvalidate(invalidate func()) *SentenceAnalysis {
	if t == nil {
		return t
	}
	t.invalidate = invalidate
	return t
}

func (t *SentenceAnalysis) SetSentence(line *transcriptRow) {
	if t == nil {
		return
	}
	if line == nil || line.Info || strings.TrimSpace(line.Text) == "" {
		t.Reset()
		return
	}
	row := *line
	row.Text = utils.CleanInlineText(row.Text)
	if t.line == nil || t.line.Key != row.Key || t.line.Text != row.Text {
		t.clearFocusedLookup()
	}
	t.line = &row
}
func (t *SentenceAnalysis) Reset() {
	if t == nil {
		return
	}
	t.line = nil
	t.clearFocusedLookup()
}
func (t *SentenceAnalysis) HandeEvents(gtx layout.Context) {
	for t.furiganaHiddenClick.Clicked(gtx) {
		t.focusedFuriganaMode = focusedFuriganaHidden
	}
	for t.furiganaAboveClick.Clicked(gtx) {
		t.focusedFuriganaMode = focusedFuriganaAbove
	}
	for t.lookupFontDownClick.Clicked(gtx) {
		t.adjustLookupFontSize(-1)
	}
	for t.lookupFontUpClick.Clicked(gtx) {
		t.adjustLookupFontSize(1)
	}
	for key, click := range t.focusedTokenClicks {
		for click.Clicked(gtx) {
			t.selectFocusedToken(key)
		}
	}
	for key, click := range t.lookupAudioClicks {
		for click.Clicked(gtx) {
			t.playLookupAudio(key)
		}
	}
}

func (t *SentenceAnalysis) adjustLookupFontSize(delta unit.Sp) {
	next := t.lookupFontSize + delta
	next = clampLookupFontSize(next)
	if next == t.lookupFontSize {
		return
	}
	t.lookupFontSize = next
	t.invalidateUI()
}

func structureFlashcardWord(token japanese.Token) string {
	switch token.POSMajor() {
	case "動詞", "形容詞":
		return util.FirstNonEmpty(strings.TrimSpace(token.BaseForm), strings.TrimSpace(token.Surface))
	default:
		return strings.TrimSpace(token.Surface)
	}
}
func (t *SentenceAnalysis) selectFocusedToken(key string) {
	if key != "" && t.selectedFocusedTokenKey == key {
		t.clearFocusedLookup()
		t.invalidateUI()
		return
	}
	text := t.structureSourceText()
	analysis, err := japanese.AnalyzeSentence(text)
	//analysis, errText := t.currentStructureAnalysis()
	if err != nil {
		//p.showError("Dictionary Lookup Failed", errText)
		return
	}
	for _, token := range analysis.Tokens {
		if structureTokenKey(token) != key {
			continue
		}
		word := structureFlashcardWord(token)
		if word == "" {
			word = strings.TrimSpace(token.Surface)
		}
		t.selectedFocusedTokenKey = key
		t.selectedFocusedTokenWord = word
		t.selectedFocusedTokenNote = ""
		t.focusedLookupPendingKey = ""
		t.startFocusedTokenLookup(key, word)
		return
	}
}

func (t *SentenceAnalysis) clearFocusedLookup() {
	t.lookupMu.Lock()
	defer t.lookupMu.Unlock()
	t.selectedFocusedTokenKey = ""
	t.selectedFocusedTokenWord = ""
	t.selectedFocusedTokenNote = ""
	t.focusedLookupPendingKey = ""
	t.lookupGeneration++
	t.lookupResults = nil
	t.lookupError = ""
	t.lookupPending = false
	t.lookupQuery = ""
	t.lookupTokenKey = ""
	t.lookupAudioClicks = make(map[string]*widget.Clickable)
	t.lookupAudio = make(map[string]*lookupAudioState)
}

func (t *SentenceAnalysis) startFocusedTokenLookup(key, word string) {
	word = strings.TrimSpace(word)
	if word == "" {
		return
	}
	t.lookupMu.Lock()
	t.lookupGeneration++
	generation := t.lookupGeneration
	t.lookupPending = true
	t.lookupError = ""
	t.lookupResults = nil
	t.lookupQuery = word
	t.lookupTokenKey = key
	t.focusedLookupPendingKey = key
	t.lookupMu.Unlock()
	t.invalidateUI()

	go func() {
		var (
			results []*jpndict.Response
			err     error
		)
		if t.backend == nil {
			err = fmt.Errorf("dictionary backend is not available")
		} else {
			results, err = t.backend.SearchAllTerm(jpndict.Search{Text: word})
		}

		t.lookupMu.Lock()
		if generation != t.lookupGeneration {
			t.lookupMu.Unlock()
			return
		}
		t.lookupPending = false
		t.focusedLookupPendingKey = ""
		if err != nil {
			t.lookupError = err.Error()
		} else if len(results) == 0 {
			t.lookupError = "No dictionary entry found for " + word
		} else {
			t.lookupResults = results
		}
		t.lookupMu.Unlock()
		t.invalidateUI()
	}()
}

func (t *SentenceAnalysis) lookupSnapshot() (query, tokenKey, errText string, pending bool, results []*jpndict.Response) {
	t.lookupMu.RLock()
	defer t.lookupMu.RUnlock()
	results = append([]*jpndict.Response(nil), t.lookupResults...)
	return t.lookupQuery, t.lookupTokenKey, t.lookupError, t.lookupPending, results
}

func (t *SentenceAnalysis) lookupAudioClickable(key string) *widget.Clickable {
	if t.lookupAudioClicks == nil {
		t.lookupAudioClicks = make(map[string]*widget.Clickable)
	}
	if t.lookupAudioClicks[key] == nil {
		t.lookupAudioClicks[key] = new(widget.Clickable)
	}
	return t.lookupAudioClicks[key]
}

func (t *SentenceAnalysis) registerLookupAudio(key, query string, resp *jpndict.Response) {
	key = strings.TrimSpace(key)
	query = strings.TrimSpace(query)
	if key == "" || query == "" {
		return
	}
	t.lookupMu.Lock()
	defer t.lookupMu.Unlock()
	if t.lookupAudio == nil {
		t.lookupAudio = make(map[string]*lookupAudioState)
	}
	state := t.lookupAudio[key]
	if state == nil {
		state = &lookupAudioState{}
		t.lookupAudio[key] = state
	}
	state.Query = query
	if resp != nil && resp.HasAudio() {
		state.Cached = true
		state.Resp = resp
	}
}

func (t *SentenceAnalysis) lookupAudioSnapshot(key string) (pending, cached bool, errText string) {
	t.lookupMu.RLock()
	defer t.lookupMu.RUnlock()
	state := t.lookupAudio[key]
	if state == nil {
		return false, false, ""
	}
	return state.Pending, state.Cached, state.Error
}

func (t *SentenceAnalysis) playLookupAudio(key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}

	t.lookupMu.Lock()
	state := t.lookupAudio[key]
	if state == nil || strings.TrimSpace(state.Query) == "" {
		t.lookupMu.Unlock()
		return
	}
	if state.Pending {
		t.lookupMu.Unlock()
		return
	}
	if state.Resp != nil && state.Resp.HasAudio() {
		resp := state.Resp
		state.Error = ""
		t.lookupMu.Unlock()
		go func() {
			if _, err := resp.PlayAudio(false); err != nil {
				t.setLookupAudioError(key, err)
			}
		}()
		return
	}

	query := state.Query
	state.Pending = true
	state.Error = ""
	t.lookupMu.Unlock()
	t.invalidateUI()

	go func() {
		var (
			resp *jpndict.Response
			err  error
		)
		if t.backend == nil {
			err = fmt.Errorf("dictionary backend is not available")
		} else {
			resp, err = t.backend.SearchTerm(jpndict.Search{Text: query, WithAudio: true})
		}
		if err == nil && (resp == nil || !resp.HasAudio()) {
			err = fmt.Errorf("no cached audio found for %s", query)
		}
		if err == nil {
			_, err = resp.PlayAudio(false)
		}

		t.lookupMu.Lock()
		state := t.lookupAudio[key]
		if state != nil {
			state.Pending = false
			if err != nil {
				state.Error = err.Error()
			} else {
				state.Error = ""
				state.Cached = true
				state.Resp = resp
			}
		}
		t.lookupMu.Unlock()
		t.invalidateUI()
	}()
}

func (t *SentenceAnalysis) setLookupAudioError(key string, err error) {
	if err == nil {
		return
	}
	t.lookupMu.Lock()
	if state := t.lookupAudio[key]; state != nil {
		state.Error = err.Error()
		state.Pending = false
	}
	t.lookupMu.Unlock()
	t.invalidateUI()
}

func (t *SentenceAnalysis) invalidateUI() {
	if t != nil && t.invalidate != nil {
		t.invalidate()
	}
}

func (t *SentenceAnalysis) SetLookupFontSize(size unit.Sp) {
	if t == nil {
		return
	}
	t.lookupFontSize = clampLookupFontSize(size)
}

func (t *SentenceAnalysis) LookupFontSize() unit.Sp {
	if t == nil {
		return unit.Sp(14)
	}
	return t.lookupFontSize
}

func clampLookupFontSize(size unit.Sp) unit.Sp {
	if size < unit.Sp(11) {
		return unit.Sp(11)
	}
	if size > unit.Sp(24) {
		return unit.Sp(24)
	}
	return size
}

func (t *SentenceAnalysis) currentAnalysis() (japanese.Analysis, string) {
	text := t.structureSourceText()
	if strings.TrimSpace(text) == "" {
		return japanese.Analysis{}, ""
	}
	var (
		analysis japanese.Analysis
		err      error
	)
	if t.backend != nil {
		analysis, err = t.backend.AnalyzeSentence(text)
	} else {
		analysis, err = japanese.AnalyzeSentence(text)
	}
	if err != nil {
		return japanese.Analysis{}, err.Error()
	}
	return analysis, ""
}

func (t *SentenceAnalysis) structureSourceText() string {
	if t.line == nil {
		return ""
	}
	return t.line.Text
}
func (t *SentenceAnalysis) focusedSentenceTokenWidth(gtx layout.Context, token japanese.Token) int {
	surfaceRunes := len([]rune(utils.CleanInlineText(token.Surface)))
	readingRunes := len([]rune(focusedTokenReading(token)))
	runes := surfaceRunes
	if t.focusedFuriganaMode == focusedFuriganaAbove && readingRunes > runes {
		runes = readingRunes
	}
	if runes <= 0 {
		runes = 1
	}
	size := float32(t.sentenceFontSize)
	// This is intentionally conservative: Gio's label layout, token padding,
	// furigana, and the inter-token gap all need to fit before the wrap point.
	return gtx.Dp(unit.Dp(float32(runes)*size*0.95 + 28))
}

func (t *SentenceAnalysis) focusedTokenClickable(key string) *widget.Clickable {
	if t.focusedTokenClicks == nil {
		t.focusedTokenClicks = make(map[string]*widget.Clickable)
	}
	if t.focusedTokenClicks[key] == nil {
		t.focusedTokenClicks[key] = new(widget.Clickable)
	}
	return t.focusedTokenClicks[key]
}

func (t *SentenceAnalysis) pruneFocusedTokenClicks(tokens []japanese.Token) {
	valid := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		valid[structureTokenKey(token)] = struct{}{}
	}
	for key := range t.focusedTokenClicks {
		if _, ok := valid[key]; !ok {
			delete(t.focusedTokenClicks, key)
		}
	}
	if t.selectedFocusedTokenKey != "" {
		if _, ok := valid[t.selectedFocusedTokenKey]; !ok {
			t.selectedFocusedTokenKey = ""
			t.selectedFocusedTokenWord = ""
			t.selectedFocusedTokenNote = ""
			t.focusedLookupPendingKey = ""
			t.clearFocusedLookup()
		}
	}
}
func (t *SentenceAnalysis) focusedTokenSurfaceSlotHeight(gtx layout.Context) unit.Dp {
	size := float32(t.sentenceFontSize)
	return unit.Dp(size + 12)
}

func (t *SentenceAnalysis) focusedTokenReadingFontSize() unit.Sp {
	size := t.sentenceFontSize * unit.Sp(0.5)
	if size < unit.Sp(10) {
		return unit.Sp(10)
	}
	if size > unit.Sp(18) {
		return unit.Sp(18)
	}
	return size
}

func (t *SentenceAnalysis) focusedTokenReadingSlotHeight() unit.Dp {
	return unit.Dp(float32(t.focusedTokenReadingFontSize()) + 8)
}

func (t *SentenceAnalysis) layoutFocusedTokenSlot(gtx layout.Context, height unit.Dp, w layout.Widget) layout.Dimensions {
	slotHeight := gtx.Dp(height)
	if slotHeight <= 0 {
		return w(gtx)
	}
	local := gtx
	local.Constraints.Min.Y = slotHeight
	local.Constraints.Max.Y = slotHeight
	dims := layout.Center.Layout(local, w)
	dims.Size.Y = slotHeight
	return dims
}

// todo fix this
func focusedTokenColor(theme *theme.ColorTokens, token japanese.Token, selected, inFlashcards, dictionaryReady bool) color.NRGBA {
	primary := theme.PrimaryNRGBA()
	surfaceAlt := theme.SurfaceAltNRGBA()

	if selected {
		return color.NRGBA{R: primary.R, G: primary.G, B: primary.B, A: 88}
	}
	if inFlashcards {
		return color.NRGBA{R: primary.R, G: primary.G, B: primary.B, A: 54}
	}
	if dictionaryReady {
		return color.NRGBA{R: surfaceAlt.R, G: surfaceAlt.G, B: surfaceAlt.B, A: 210}
	}
	switch token.POSMajor() {
	//case "名詞":
	//	return color.NRGBA{R: theme.Color.Secondary.R, G: theme.Color.Secondary.G, B: theme.Color.Secondary.B, A: 44}
	//case "動詞":
	//	return color.NRGBA{R: theme.Color.Primary.R, G: theme.Color.Primary.G, B: theme.Color.Primary.B, A: 42}
	//case "形容詞", "副詞":
	//	return color.NRGBA{R: theme.Color.Warning.R, G: theme.Color.Warning.G, B: theme.Color.Warning.B, A: 42}
	//case "助詞", "助動詞":
	//	return color.NRGBA{R: theme.Color.Tertiary.R, G: theme.Color.Tertiary.G, B: theme.Color.Tertiary.B, A: 32}
	default:
		return color.NRGBA{R: surfaceAlt.R, G: surfaceAlt.G, B: surfaceAlt.B, A: 180}
	}
}

func focusedTokenReading(token japanese.Token) string {
	if !util.ContainsKanji(token.Surface) {
		return ""
	}
	reading := utils.CleanInlineText(token.Reading)
	if reading == "" || reading == utils.CleanInlineText(token.Surface) {
		return ""
	}
	return katakanaToHiragana(reading)
}
func katakanaToHiragana(text string) string {
	var b strings.Builder
	for _, r := range text {
		if r >= 'ァ' && r <= 'ヶ' {
			r -= 0x60
		}
		b.WriteRune(r)
	}
	return b.String()
}

//func (p *Page) layoutFocusedSentenceCard(gtx layout.Context) layout.Dimensions {
//	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutCardHeader(gtx, "Focused Sentence", "Select text, click a saved word, or use the latest transcript line")
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusedSentenceText(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusedFuriganaControls(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusedTokenActions(gtx)
//			}),
//			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			//	analysis, errText := p.currentStructureAnalysis()
//			//	if errText == "" && len(analysis.Tokens) == 0 {
//			//		return layout.Dimensions{}
//			//	}
//			//	return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			//		return p.layoutFocusedSentenceChips(gtx, analysis, errText)
//			//	})
//			//}),
//			//layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			//	return p.layoutFocusedLookupBar(gtx)
//			//}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusedTranslationSection(gtx)
//			}),
//			//layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//			//layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			//	gtx.Constraints.Min = gtx.Constraints.Max
//			//	return p.layoutSentenceStructurePanel(gtx, true)
//			//}),
//		)
//	})
//}
