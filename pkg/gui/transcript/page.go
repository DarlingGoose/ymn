package transcript

//
//import (
//	"context"
//	"fmt"
//	"image"
//	"image/color"
//	"log/slog"
//	"os"
//	"path/filepath"
//	"regexp"
//	"sort"
//	"strings"
//	"time"
//
//	"gioui.org/font"
//
//	"gioui.org/app"
//	"gioui.org/io/pointer"
//	"gioui.org/layout"
//	"gioui.org/op"
//	"gioui.org/op/clip"
//	"gioui.org/op/paint"
//	"gioui.org/unit"
//	"gioui.org/widget"
//	"gioui.org/widget/material"
//	bareui "github.com/DarlingGoose/bare/pkg/ui"
//	"github.com/DarlingGoose/bare/pkg/ui/icons"
//	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
//	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
//	"github.com/DarlingGoose/jpndict/audioplayer"
//	"github.com/DarlingGoose/vntext/pkg/engine"
//	"github.com/DarlingGoose/vntext/pkg/engine/auto"
//	vngame "github.com/DarlingGoose/vntext/pkg/game"
//	//"github.com/DarlingGoose/vntext/pkg/runner"
//	vnutil "github.com/DarlingGoose/vntext/pkg/util"
//	"github.com/DarlingGoose/wgl/pkg/anki"
//	"github.com/DarlingGoose/wgl/pkg/dictionary"
//	flashcards "github.com/DarlingGoose/wgl/pkg/flashcard"
//	"github.com/DarlingGoose/wgl/pkg/gui"
//	guitoast "github.com/DarlingGoose/wgl/pkg/gui/toast"
//	"github.com/DarlingGoose/wgl/pkg/japanese"
//	"github.com/DarlingGoose/wgl/pkg/translation"
//	wgltts "github.com/DarlingGoose/wgl/pkg/tts"
//	"github.com/DarlingGoose/wgl/pkg/util"
//)
//
//const (
//	compactWidth          = 1080
//	transcriptStackWidth  = 1240
//	transcriptMediumWidth = 1480
//
//	composerFocusFlashcards        = "flashcards"
//	composerFocusSentenceStructure = "sentence_structure"
//
//	focusedFuriganaHidden = "hidden"
//	focusedFuriganaAbove  = "above"
//	focusedFuriganaBelow  = "below"
//)
//
//var _ gui.EvenHandler = &Page{}
//
//var ansiRE = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|[@-Z\\-_])`)
//
//type Page struct {
//	theme   barethemes.Theme
//	iconify *icons.Iconify
//
//	transcriptView       widget.Selectable
//	focusedSentenceView  widget.Selectable
//	transcriptFocusSplit widget.Float
//	transcriptList       widget.List
//	structureList        widget.List
//	lookupResultsList    widget.List
//	wordEditor           widget.Editor
//	meaningEditor        widget.Editor
//	translationEditor    widget.Editor
//	targetLanguageDrop   bareui.Dropdown
//	ttsSpeakerDrop       bareui.Dropdown
//	hideReadingInAnki    widget.Bool
//	searchWordButton     widget.Clickable
//	playAudioButton      widget.Clickable
//	addAllLookupButton   widget.Clickable
//	launchGameButton     widget.Clickable
//	syncAnkiButton       widget.Clickable
//	clearButton          widget.Clickable
//	deleteLogButton      widget.Clickable
//
//	playSentenceButton         widget.Clickable
//	translateSentenceButton    widget.Clickable
//	saveSentenceButton         widget.Clickable
//	focusedLookupButton        widget.Clickable
//	transcriptPopupAudioButton widget.Clickable
//	transcriptPopupCloseButton widget.Clickable
//	popupDismissClicks         [4]widget.Clickable
//	composerToggleButton       widget.Clickable
//	composerFlashcardsTab      widget.Clickable
//	composerSentenceTab        widget.Clickable
//	translationToggleButton    widget.Clickable
//	saveTranslationButton      widget.Clickable
//	generateTranslationButton  widget.Clickable
//	autoTranslateMissing       widget.Bool
//	furiganaHiddenButton       widget.Clickable
//	furiganaAboveButton        widget.Clickable
//	furiganaBelowButton        widget.Clickable
//	focusedTokenAddButton      widget.Clickable
//	focusedTokenAudioButton    widget.Clickable
//
//	transcriptHighlightClicks    map[string]*widget.Clickable
//	transcriptHighlightBounds    map[string]image.Rectangle
//	lookupResultAddClicks        map[string]*widget.Clickable
//	lookupResultPlayClicks       map[string]*widget.Clickable
//	structureTokenAddClicks      map[string]*widget.Clickable
//	structureTokenPlayClicks     map[string]*widget.Clickable
//	transcriptRowClicks          map[string]*widget.Clickable
//	transcriptRowTranslateClicks map[string]*widget.Clickable
//	transcriptRowVoiceClicks     map[string]*widget.Clickable
//	ttsSpeakerClicks             map[string]*widget.Clickable
//	focusedTokenClicks           map[string]*widget.Clickable
//	targetLanguageOptions        []gui.DropdownOption
//
//	activeGameName         string
//	logPath                string
//	ankiURL                string
//	pushSync               bool
//	statusText             string
//	currentConfig          *vngame.Game
//	runnerStatus           *runner.ProcessStatus
//	selectedTextSizeName   string
//	selectedRecentLines    string
//	transcriptTextSize     unit.Sp
//	focusedTextSize        unit.Sp
//	translateDetailSize    unit.Sp
//	recentLineLimit        int
//	autoPlayHighlightAudio bool
//	colorizeHighlights     bool
//	speakerOnlyRows        bool
//	compactTimestamps      bool
//	selectedTTSSpeaker     string
//
//	flashcards                []flashcards.Flashcard
//	lookupResult              *dictionary.Lookup
//	lookupResults             []dictionary.Lookup
//	displayTranscript         string
//	lastSyncedText            string
//	lastFocusedText           string
//	structureCacheKey         string
//	structureCache            japanese.Analysis
//	structureCacheErr         string
//	highlightCacheKey         string
//	highlightCache            []flashcards.Match
//	popupFlashcard            *flashcards.Flashcard
//	popupAnchor               image.Rectangle
//	popupBounds               image.Rectangle
//	popupMatchKey             string
//	popupWord                 string
//	selectedLineKey           string
//	selectedLineText          string
//	selectedFocusedTokenKey   string
//	selectedFocusedTokenWord  string
//	selectedFocusedTokenNote  string
//	focusedLookupPendingKey   string
//	translationCollapsed      bool
//	focusedFuriganaMode       string
//	focusedFuriganaDefault    string
//	selectedTargetLanguage    string
//	defaultTargetLanguage     string
//	targetLanguageUserSet     bool
//	translationLoadedKey      string
//	translationGeneratingKey  string
//	autoTranslationAttemptKey string
//	translationResultCh       chan translationResult
//	rowTranslationResultCh    chan rowTranslationResult
//	audioResultCh             chan audioPlaybackResult
//	flashcardAddResultCh      chan flashcardAddResult
//	focusedLookupResultCh     chan focusedTokenLookupResult
//	translatorConfig          translation.Config
//	composerFocus             string
//	composerMinimized         bool
//	composerLastUsed          time.Time
//	lastAutoWord              string
//	hideReadingSet            bool
//	rowTranslations           map[string]string
//	rowTranslationShown       map[string]bool
//	rowTranslationGenerating  map[string]bool
//	lookupAddPending          map[string]bool
//
//	OnError     func(title, body string)
//	OnNotify    func(title, body string, kin	//		rowTranslations:              make(map[string]string),
//	//		rowTranslationShown:          make(map[string]bool),
//	//		rowTranslationGenerating:     make(map[string]bool),d guitoast.NotificationType)
//	OnDeleteLog func(config *vngame.Game) error
//}
//
//type translationResult struct {
//	Key   string
//	Entry translation.Entry
//	Err   error
//}
//
//type rowTranslationResult struct {
//	Key    string
//	RowKey string
//	Entry  translation.Entry
//	Err    error
//}
//
//type audioPlaybackResult struct {
//	Title string
//	Err   error
//}
//
//type flashcardAddResult struct {
//	Key     string
//	Card    flashcards.Flashcard
//	Cards   []flashcards.Flashcard
//	Added   int
//	Skipped int
//	Err     error
//}
//
//type focusedTokenLookupResult struct {
//	Key     string
//	Word    string
//	Lookups []dictionary.Lookup
//	Err     error
//}
//
//type transcriptRow struct {
//	Key        string
//	Time       string
//	Speaker    string
//	Voice      string
//	Text       string
//	Info       bool
//	VocabWords []string
//}
//
//type ttsReference struct {
//	Speaker string
//	Voice   string
//	Text    string
//}
//
//func New(theme barethemes.Theme) *Page {
//	p := &Page{
//		theme:                        theme,
//		pushSync:                     true,
//		statusText:                   "Start the game to show live transcript text here.",
//		selectedTextSizeName:         "Medium",
//		selectedRecentLines:          "All Lines",
//		transcriptTextSize:           unit.Sp(16),
//		focusedTextSize:              unit.Sp(26),
//		translateDetailSize:          unit.Sp(15),
//		focusedFuriganaMode:          focusedFuriganaHidden,
//		focusedFuriganaDefault:       focusedFuriganaHidden,
//		selectedTargetLanguage:       util.DefaultTranslationLanguage,
//		defaultTargetLanguage:        util.DefaultTranslationLanguage,
//		translationResultCh:          make(chan translationResult, 1),
//		rowTranslationResultCh:       make(chan rowTranslationResult, 8),
//		audioResultCh:                make(chan audioPlaybackResult, 8),
//		flashcardAddResultCh:         make(chan flashcardAddResult, 8),
//		focusedLookupResultCh:        make(chan focusedTokenLookupResult, 8),
//		translatorConfig:             translation.Config{Provider: translation.ProviderOllama},
//		composerFocus:                composerFocusFlashcards,
//		composerMinimized:            true,
//		composerLastUsed:             time.Now(),
//		transcriptHighlightClicks:    make(map[string]*widget.Clickable),
//		transcriptHighlightBounds:    make(map[string]image.Rectangle),
//		lookupResultAddClicks:        make(map[string]*widget.Clickable),
//		lookupResultPlayClicks:       make(map[string]*widget.Clickable),
//		structureTokenAddClicks:      make(map[string]*widget.Clickable),
//		structureTokenPlayClicks:     make(map[string]*widget.Clickable),
//		transcriptRowClicks:          make(map[string]*widget.Clickable),
//		transcriptRowTranslateClicks: make(map[string]*widget.Clickable),
//		transcriptRowVoiceClicks:     make(map[string]*widget.Clickable),
//		ttsSpeakerClicks:             make(map[string]*widget.Clickable),
//		focusedTokenClicks:           make(map[string]*widget.Clickable),
//		rowTranslations:              make(map[string]string),
//		rowTranslationShown:          make(map[string]bool),
//		rowTranslationGenerating:     make(map[string]bool),
//		lookupAddPending:             make(map[string]bool),
//	}
//	p.transcriptFocusSplit.Value = 0.5
//	p.wordEditor.SingleLine = true
//	p.meaningEditor.SingleLine = false
//	p.translationEditor.SingleLine = false
//	gui.NewDropDownLayout(&p.targetLanguageDrop, "mdi:translate")
//	p.targetLanguageDrop.Width = unit.Dp(190)
//	p.targetLanguageDrop.OffsetY = unit.Dp(42)
//	gui.NewDropDownLayout(&p.ttsSpeakerDrop, "mdi:account-voice")
//	p.ttsSpeakerDrop.Width = unit.Dp(220)
//	p.ttsSpeakerDrop.OffsetY = unit.Dp(42)
//	p.targetLanguageOptions = gui.NewTranslationLanguageOptions()
//	p.transcriptList.Axis = layout.Vertical
//	p.transcriptList.ScrollToEnd = true
//	p.structureList.Axis = layout.Vertical
//	p.lookupResultsList.Axis = layout.Vertical
//	return p
//}
//
//func (p *Page) WithTheme(theme barethemes.Theme) *Page {
//	p.theme = theme
//	return p
//}
//
//func (p *Page) WithIcon(icon *icons.Iconify) *Page {
//	p.iconify = icon
//	return p
//}
//
//func (p *Page) SetContext(activeGameName, logPath, ankiURL string, cfg *vngame.Game) *Page {
//	p.activeGameName = strings.TrimSpace(activeGameName)
//	p.logPath = strings.TrimSpace(logPath)
//	p.ankiURL = strings.TrimSpace(ankiURL)
//	p.currentConfig = cfg
//	return p
//}
//
//func (p *Page) SetPushSync(pushSync bool) *Page {
//	p.pushSync = pushSync
//	return p
//}
//
//func (p *Page) SetRunningState(running bool, pid int) *Page {
//	if !running {
//		p.runnerStatus = nil
//		return p
//	}
//	p.runnerStatus = &runner.ProcessStatus{
//		PID:    pid,
//		Status: runner.StatusRunning,
//	}
//	return p
//}
//
//func (p *Page) SetTranscriptOptions(textSize unit.Sp, textSizeName string, recentLineLimit int, recentLinesName string) *Page {
//	p.transcriptTextSize = textSize
//	p.selectedTextSizeName = strings.TrimSpace(textSizeName)
//	p.recentLineLimit = recentLineLimit
//	p.selectedRecentLines = strings.TrimSpace(recentLinesName)
//	return p
//}
//
//func (p *Page) SetTranscriptDisplayOptions(speakerOnlyRows, compactTimestamps bool) *Page {
//	p.speakerOnlyRows = speakerOnlyRows
//	p.compactTimestamps = compactTimestamps
//	return p
//}
//
//func (p *Page) SetTranslateTextOptions(focusedSize, detailSize unit.Sp) *Page {
//	if focusedSize > 0 {
//		p.focusedTextSize = focusedSize
//	}
//	if detailSize > 0 {
//		p.translateDetailSize = detailSize
//	}
//	return p
//}
//
//func (p *Page) SetTranslatorConfig(cfg translation.Config) *Page {
//	p.translatorConfig = cfg
//	return p
//}
//
//func (p *Page) SetDefaultTargetLanguage(language string) *Page {
//	language = util.ResolveTranslationLanguage(language)
//	if strings.TrimSpace(language) == "" {
//		language = util.DefaultTranslationLanguage
//	}
//	if p.selectedTargetLanguage == "" || !p.targetLanguageUserSet || p.selectedTargetLanguage == p.defaultTargetLanguage {
//		p.selectedTargetLanguage = language
//		p.translationLoadedKey = ""
//		p.autoTranslationAttemptKey = ""
//	}
//	p.defaultTargetLanguage = language
//	return p
//}
//
//func (p *Page) SetFocusedFuriganaDefault(mode string) *Page {
//	mode = normalizeFocusedFuriganaMode(mode)
//	oldDefault := p.focusedFuriganaDefault
//	if oldDefault == "" || p.focusedFuriganaMode == "" || p.focusedFuriganaMode == oldDefault {
//		p.focusedFuriganaMode = mode
//	}
//	p.focusedFuriganaDefault = mode
//	return p
//}
//
//func (p *Page) SetAutoPlayHighlightAudio(enabled bool) *Page {
//	p.autoPlayHighlightAudio = enabled
//	return p
//}
//
//func (p *Page) SetColorizeHighlights(enabled bool) *Page {
//	p.colorizeHighlights = enabled
//	return p
//}
//
//func (p *Page) SetStatus(status string) *Page {
//	p.statusText = strings.TrimSpace(status)
//	return p
//}
//
//func (p *Page) SetRawTranscript(raw string) *Page {
//	next := limitTranscriptLines(sanitizeTranscriptForDisplay(raw), p.recentLineLimit)
//	if next != p.displayTranscript {
//		p.displayTranscript = next
//		p.invalidateHighlights()
//		p.selectLatestTranscriptRow()
//	}
//	return p
//}
//
//func (p *Page) ClearTranscript() {
//	p.displayTranscript = ""
//	p.lookupResult = nil
//	p.lookupResults = nil
//	p.focusedLookupPendingKey = ""
//	p.invalidateHighlights()
//	p.DismissPopup()
//	p.selectedLineKey = ""
//	p.selectedLineText = ""
//	p.statusText = "Transcript view cleared; waiting for new dialogue."
//	p.lastSyncedText = ""
//}
//
//func (p *Page) SetFlashcards(cards []flashcards.Flashcard) *Page {
//	p.flashcards = append([]flashcards.Flashcard(nil), cards...)
//	sort.Slice(p.flashcards, func(i, j int) bool {
//		return p.flashcards[i].UpdatedAt.After(p.flashcards[j].UpdatedAt)
//	})
//	p.invalidateHighlights()
//	return p
//}
//
//func (p *Page) Cards() []flashcards.Flashcard {
//	return append([]flashcards.Flashcard(nil), p.flashcards...)
//}
//
//func (p *Page) ReloadFlashcards() error {
//	if strings.TrimSpace(p.activeGameName) == "" {
//		p.flashcards = nil
//		return nil
//	}
//	cards, err := flashcards.LoadFlashcards(p.activeGameName)
//	if err != nil {
//		return err
//	}
//	p.SetFlashcards(cards)
//	return nil
//}
//
//func (p *Page) PopupFlashcard() *flashcards.Flashcard {
//	return p.popupFlashcard
//}
//
//func (p *Page) DismissPopup() {
//	p.popupFlashcard = nil
//	p.popupBounds = image.Rectangle{}
//	p.popupMatchKey = ""
//	p.popupWord = ""
//}
//
//func (p *Page) HandleEvents(gtx layout.Context, ctx context.Context, w *app.Window) {
//	p.drainTranslationResults()
//	p.drainRowTranslationResults()
//	p.drainAudioResults()
//	p.drainFlashcardAddResults()
//	p.drainFocusedTokenLookupResults()
//	p.maybeAutoGenerateTranslation(ctx, w)
//	if p.hideReadingInAnki.Update(gtx) {
//		p.hideReadingSet = true
//	}
//	p.targetLanguageDrop.Update(gtx)
//	p.ttsSpeakerDrop.Update(gtx)
//	p.syncHideReadingDefault()
//	for p.launchGameButton.Clicked(gtx) {
//		p.launchCurrentGameInBackground()
//	}
//
//	for p.syncAnkiButton.Clicked(gtx) {
//		if err := p.syncCurrentGameToAnki(); err != nil {
//			p.showError("Anki Sync Failed", err.Error())
//		} else {
//			p.showNotification("Anki Sync Complete", "Transcript flashcards synced to Anki.", guitoast.NotificationTypeSuccess)
//		}
//	}
//	for p.clearButton.Clicked(gtx) {
//		p.ClearTranscript()
//	}
//	for p.deleteLogButton.Clicked(gtx) {
//		p.deleteCurrentLog()
//	}
//	for p.playSentenceButton.Clicked(gtx) {
//		p.playCurrentLookupAudio(ctx, w)
//	}
//	for p.translateSentenceButton.Clicked(gtx) {
//		p.composerFocus = composerFocusSentenceStructure
//		p.composerMinimized = false
//		p.composerLastUsed = time.Now()
//	}
//	for p.saveSentenceButton.Clicked(gtx) {
//		p.composerFocus = composerFocusFlashcards
//		p.composerMinimized = false
//		p.composerLastUsed = time.Now()
//	}
//	for p.focusedLookupButton.Clicked(gtx) {
//		p.lookupCurrentWord()
//	}
//	for p.focusedTokenAddButton.Clicked(gtx) {
//		p.addFocusedTokenFlashcard()
//	}
//	for p.focusedTokenAudioButton.Clicked(gtx) {
//		p.playFocusedTokenAudio(ctx, w)
//	}
//	for p.translationToggleButton.Clicked(gtx) {
//		p.translationCollapsed = !p.translationCollapsed
//	}
//	for p.saveTranslationButton.Clicked(gtx) {
//		p.saveCurrentTranslation()
//	}
//	for p.generateTranslationButton.Clicked(gtx) {
//		p.generateCurrentTranslation(ctx, w)
//	}
//	for p.furiganaHiddenButton.Clicked(gtx) {
//		p.focusedFuriganaMode = focusedFuriganaHidden
//	}
//	for p.furiganaAboveButton.Clicked(gtx) {
//		p.focusedFuriganaMode = focusedFuriganaAbove
//	}
//	for p.furiganaBelowButton.Clicked(gtx) {
//		p.focusedFuriganaMode = focusedFuriganaBelow
//	}
//	for i := range p.targetLanguageOptions {
//		opt := &p.targetLanguageOptions[i]
//		for opt.Clickable.Clicked(gtx) {
//			p.selectedTargetLanguage = opt.Label
//			p.targetLanguageUserSet = true
//			p.translationLoadedKey = ""
//			p.autoTranslationAttemptKey = ""
//			p.targetLanguageDrop.Close()
//		}
//	}
//	for speaker, click := range p.ttsSpeakerClicks {
//		for click.Clicked(gtx) {
//			p.selectedTTSSpeaker = speaker
//			p.ttsSpeakerDrop.Close()
//		}
//	}
//	for p.searchWordButton.Clicked(gtx) {
//		p.lookupCurrentWord()
//	}
//	for p.playAudioButton.Clicked(gtx) {
//		p.playCurrentLookupAudio(ctx, w)
//	}
//	for p.addAllLookupButton.Clicked(gtx) {
//		p.addAllLookupFlashcards()
//	}
//	for key, click := range p.lookupResultAddClicks {
//		for click.Clicked(gtx) {
//			p.addLookupFlashcardByKey(key, w)
//		}
//	}
//	for key, click := range p.lookupResultPlayClicks {
//		for click.Clicked(gtx) {
//			p.playLookupAudioByKey(ctx, w, key)
//		}
//	}
//	for key, click := range p.structureTokenAddClicks {
//		for click.Clicked(gtx) {
//			p.addStructureTokenFlashcard(key)
//		}
//	}
//	for key, click := range p.structureTokenPlayClicks {
//		for click.Clicked(gtx) {
//			p.playStructureTokenAudio(ctx, w, key)
//		}
//	}
//	for key, click := range p.transcriptHighlightClicks {
//		for click.Clicked(gtx) {
//			p.openTranscriptHighlightPopup(key)
//		}
//	}
//	for key, click := range p.transcriptRowClicks {
//		for click.Clicked(gtx) {
//			p.selectTranscriptRow(key)
//		}
//	}
//	for key, click := range p.transcriptRowTranslateClicks {
//		for click.Clicked(gtx) {
//			p.toggleTranscriptRowTranslation(ctx, w, key)
//		}
//	}
//	for key, click := range p.transcriptRowVoiceClicks {
//		for click.Clicked(gtx) {
//			p.playTranscriptRowVoice(ctx, w, key)
//		}
//	}
//	for key, click := range p.focusedTokenClicks {
//		for click.Clicked(gtx) {
//			p.selectFocusedToken(key, w)
//		}
//	}
//	for p.transcriptPopupAudioButton.Clicked(gtx) {
//		if p.popupFlashcard == nil {
//			p.showError("Audio Playback Failed", "No flashcard is selected.")
//			continue
//		}
//		card := *p.popupFlashcard
//		p.startAudioPlayback(w, "Audio Playback Failed", func() error {
//			return p.playFlashcardAudio(ctx, card)
//		})
//	}
//	for p.transcriptPopupCloseButton.Clicked(gtx) {
//		p.DismissPopup()
//	}
//	for p.composerToggleButton.Clicked(gtx) {
//		p.composerMinimized = !p.composerMinimized
//		p.composerLastUsed = time.Now()
//	}
//	for i := range p.popupDismissClicks {
//		for p.popupDismissClicks[i].Clicked(gtx) {
//			p.DismissPopup()
//		}
//	}
//}
//
//func (p *Page) startAudioPlayback(w *app.Window, title string, play func() error) {
//	if play == nil {
//		return
//	}
//	if strings.TrimSpace(title) == "" {
//		title = "Audio Playback Failed"
//	}
//	go func() {
//		if err := play(); err != nil {
//			result := audioPlaybackResult{Title: title, Err: err}
//			select {
//			case p.audioResultCh <- result:
//			default:
//				slog.Warn("audio playback error dropped", "title", title, "error", err)
//			}
//			if w != nil {
//				w.Invalidate()
//			}
//		}
//	}()
//}
//
//func (p *Page) drainAudioResults() {
//	for {
//		select {
//		case result := <-p.audioResultCh:
//			if result.Err != nil {
//				p.showError(util.FirstNonEmpty(result.Title, "Audio Playback Failed"), result.Err.Error())
//			}
//		default:
//			return
//		}
//	}
//}
//
//func (p *Page) drainFlashcardAddResults() {
//	for {
//		select {
//		case result := <-p.flashcardAddResultCh:
//			delete(p.lookupAddPending, result.Key)
//			if result.Err != nil {
//				p.showError("Create Flashcard Failed", result.Err.Error())
//				continue
//			}
//			if result.Cards != nil {
//				p.SetFlashcards(result.Cards)
//			}
//			if result.Added > 0 {
//				p.showNotification("Flashcard Created", result.Card.Text+" was added.", guitoast.NotificationTypeSuccess)
//				continue
//			}
//			if result.Skipped > 0 {
//				p.showNotification("Flashcard Exists", result.Card.Text+" is already in your flashcards.", guitoast.NotificationTypeInfo)
//			}
//		default:
//			return
//		}
//	}
//}
//
//func (p *Page) drainFocusedTokenLookupResults() {
//	for {
//		select {
//		case result := <-p.focusedLookupResultCh:
//			if result.Key != p.focusedLookupPendingKey {
//				continue
//			}
//			p.focusedLookupPendingKey = ""
//			if result.Key != p.selectedFocusedTokenKey || result.Word != p.selectedFocusedTokenWord {
//				continue
//			}
//			if result.Err != nil {
//				p.meaningEditor.SetText("")
//				p.showError("Dictionary Lookup Failed", result.Err.Error())
//				continue
//			}
//			if len(result.Lookups) == 0 {
//				p.meaningEditor.SetText("")
//				p.showError("Dictionary Lookup Failed", "No dictionary matches were found for "+result.Word+".")
//				continue
//			}
//			p.lookupResults = result.Lookups
//			p.lookupResult = &result.Lookups[0]
//			p.meaningEditor.SetText(result.Lookups[0].Meaning)
//		default:
//			return
//		}
//	}
//}
//
//func (p *Page) LayoutPage(gtx layout.Context) layout.Dimensions {
//	if p.iconify == nil {
//		p.iconify = icons.NewIconify()
//	}
//	p.syncTranscriptEditor()
//	return p.layoutTranscriptPanel(gtx)
//}
//
//func (p *Page) LayoutPopupContent(gtx layout.Context) layout.Dimensions {
//	if p.popupFlashcard == nil {
//		return layout.Dimensions{}
//	}
//	card := *p.popupFlashcard
//	audioButton := bareui.Button{
//		Clickable: &p.transcriptPopupAudioButton,
//		Text:      "Play Audio",
//		Prefix:    "mdi:play-circle-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.H6(p.theme.Gio(), card.Text)
//			lbl.Color = p.theme.Color.Text
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body1(p.theme.Gio(), card.Meaning)
//			lbl.Color = p.theme.Color.Text
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			meta := p.flashcardMetaText(card)
//			if meta == "" {
//				return layout.Dimensions{}
//			}
//			return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(p.theme.Gio(), meta)
//				lbl.Color = p.theme.Color.TextMuted
//				return lbl.Layout(gtx)
//			})
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if !util.IsExistingFile(card.AudioPath) && !p.hasTTSReference() {
//				return layout.Dimensions{}
//			}
//			return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return audioButton.Layout(gtx, p.theme, p.iconify)
//			})
//		}),
//	)
//}
//
//func (p *Page) layoutTranscriptPanel(gtx layout.Context) layout.Dimensions {
//	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.Stack{}.Layout(gtx,
//			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
//				return layout.Inset{
//					Top:    unit.Dp(18),
//					Left:   unit.Dp(20),
//					Right:  unit.Dp(20),
//					Bottom: unit.Dp(18),
//				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return p.layoutTranscriptTopbar(gtx)
//						}),
//						layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//							return p.layoutTranscriptWorkspace(gtx)
//						}),
//					)
//				})
//			}),
//			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
//				return layout.Dimensions{}
//			}),
//		)
//	})
//}
//
//func (p *Page) layoutTranscriptTopbar(gtx layout.Context) layout.Dimensions {
//	launchButton := bareui.Button{
//		Clickable: &p.launchGameButton,
//		Text:      p.transcriptLaunchButtonLabel(),
//		Prefix:    p.transcriptLaunchButtonIcon(),
//		Variant:   p.transcriptLaunchButtonVariant(),
//	}
//	syncButton := bareui.Button{
//		Clickable: &p.syncAnkiButton,
//		Text:      "Sync Anki",
//		Prefix:    "mdi:cloud-sync-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	clearButton := bareui.Button{
//		Clickable: &p.clearButton,
//		Text:      "mdi:broom",
//		Icon:      true,
//		Variant:   bareui.ButtonGhost,
//	}
//	deleteLogButton := bareui.Button{
//		Clickable: &p.deleteLogButton,
//		Text:      "mdi:file-remove-outline",
//		Icon:      true,
//		Variant:   bareui.ButtonGhost,
//	}
//	statusText := "IDLE"
//	statusLive := false
//	if p.runnerStatus != nil {
//		statusText = "LIVE"
//		statusLive = true
//	}
//	if p.isCompactLayout(gtx) {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return p.layoutStatusPill(gtx, statusText, statusLive)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						lbl := material.Body1(p.theme.Gio(), util.FirstNonEmpty(p.activeGameName, "No game selected"))
//						lbl.Color = p.theme.Color.Text
//						return lbl.Layout(gtx)
//					}),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						lbl := material.Body2(p.theme.Gio(), p.logSizeText())
//						lbl.Color = p.theme.Color.TextMuted
//						return lbl.Layout(gtx)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return clearButton.Layout(gtx, p.theme, p.iconify)
//					}),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						if p.currentConfig == nil {
//							return deleteLogButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//						}
//						return deleteLogButton.Layout(gtx, p.theme, p.iconify)
//					}),
//				)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						return p.layoutTTSSpeakerDropdown(gtx)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						if p.runnerStatus != nil {
//							return launchButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//						}
//						return launchButton.Layout(gtx, p.theme, p.iconify)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						return syncButton.Layout(gtx, p.theme, p.iconify)
//					}),
//				)
//			}),
//			//layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			//	lbl := material.Body1(p.theme.Gio(), p.statusText)
//			//	lbl.Color = p.statusColor()
//			//	return lbl.Layout(gtx)
//			//}),
//		)
//	}
//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutStatusPill(gtx, statusText, statusLive)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body1(p.theme.Gio(), p.transcriptRunningStatusText())
//			lbl.Color = p.theme.Color.TextMuted
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutTTSSpeakerDropdown(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if p.runnerStatus != nil {
//				return launchButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//			}
//			return launchButton.Layout(gtx, p.theme, p.iconify)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return syncButton.Layout(gtx, p.theme, p.iconify)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(4))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body2(p.theme.Gio(), p.logSizeText())
//			lbl.Color = p.theme.Color.TextMuted
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return clearButton.Layout(gtx, p.theme, p.iconify)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if p.currentConfig == nil {
//				return deleteLogButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//			}
//			return deleteLogButton.Layout(gtx, p.theme, p.iconify)
//		}),
//	)
//}
//
//func (p *Page) layoutStatusPill(gtx layout.Context, text string, live bool) layout.Dimensions {
//	bg := p.theme.Color.SurfaceAlt
//	fg := p.theme.Color.TextMuted
//
//	if live {
//		fg = p.theme.Color.Primary
//		bg = color.NRGBA{R: fg.R, G: fg.G, B: fg.B, A: 42}
//	}
//
//	return RoundedSurfaceWrap(
//		gtx,
//		bg,
//		unit.Dp(p.theme.Radius.MD),
//		func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{
//				Top:    unit.Dp(7),
//				Bottom: unit.Dp(7),
//				Left:   unit.Dp(10),
//				Right:  unit.Dp(10),
//			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body2(p.theme.Gio(), text)
//				lbl.Color = fg
//				return lbl.Layout(gtx)
//			})
//		},
//	)
//}
//
//func (p *Page) layoutTTSSpeakerDropdown(gtx layout.Context) layout.Dimensions {
//	refs := p.ttsReferences()
//	p.pruneTTSSpeakerClicks(refs)
//	if len(refs) == 0 {
//		btn := bareui.Button{
//			Text:    "TTS Voice",
//			Prefix:  "mdi:account-voice",
//			Variant: bareui.ButtonSecondary,
//		}
//		return btn.Layout(gtx.Disabled(), p.theme, p.iconify)
//	}
//	if strings.TrimSpace(p.selectedTTSSpeaker) == "" || !ttsReferenceExists(refs, p.selectedTTSSpeaker) {
//		p.selectedTTSSpeaker = refs[0].Speaker
//	}
//	label := "TTS: " + p.selectedTTSSpeaker
//	return p.ttsSpeakerDrop.Layout(gtx, p.theme, p.iconify, label, func(gtx layout.Context) layout.Dimensions {
//		children := make([]layout.FlexChild, 0, len(refs))
//		for _, ref := range refs {
//			ref := ref
//			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				variant := bareui.ButtonSecondary
//				if ref.Speaker == p.selectedTTSSpeaker {
//					variant = bareui.ButtonPrimary
//				}
//				btn := bareui.Button{
//					Clickable: p.ttsSpeakerClickable(ref.Speaker),
//					Text:      ref.Speaker,
//					Prefix:    "mdi:account-voice",
//					Variant:   variant,
//				}
//				return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//					return btn.Layout(gtx, p.theme, p.iconify)
//				})
//			}))
//		}
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
//	})
//}
//
//func RoundedSurfaceWrap(
//	gtx layout.Context,
//	bg color.NRGBA,
//	radius unit.Dp,
//	w layout.Widget,
//) layout.Dimensions {
//	macro := op.Record(gtx.Ops)
//
//	dims := w(gtx)
//
//	call := macro.Stop()
//
//	rr := clip.RRect{
//		Rect: image.Rectangle{
//			Max: dims.Size,
//		},
//		NE: int(gtx.Dp(radius)),
//		NW: int(gtx.Dp(radius)),
//		SE: int(gtx.Dp(radius)),
//		SW: int(gtx.Dp(radius)),
//	}
//
//	paint.FillShape(gtx.Ops, bg, rr.Op(gtx.Ops))
//	call.Add(gtx.Ops)
//
//	return dims
//}
//
//func (p *Page) layoutTranscriptWorkspace(gtx layout.Context) layout.Dimensions {
//	if p.runnerStatus == nil || p.shouldStackTranscriptPage(gtx) {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//				gtx.Constraints.Min = gtx.Constraints.Max
//				return p.layoutTranscriptBodyPanel(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			//	if p.runnerStatus == nil {
//			//		return layout.Dimensions{}
//			//	}
//			//	return p.layoutFlashcardComposer(gtx)
//			//}),
//		)
//	}
//	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			gtx.Constraints.Min = gtx.Constraints.Max
//			return p.layoutTranscriptBodyPanel(gtx)
//		}),
//		//layout.Rigid(bareutils.SpacerW(unit.Dp(14))),
//		//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//		//	width := p.transcriptComposerWidth(gtx)
//		//	gtx.Constraints.Min.X = width
//		//	gtx.Constraints.Max.X = width
//		//	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
//		//	return p.layoutContextRail(gtx)
//		//}),
//	)
//}
//
//func (p *Page) layoutTranscriptBodyPanel(gtx layout.Context) layout.Dimensions {
//	gtx.Constraints.Min = gtx.Constraints.Max
//	liveRatio := p.transcriptFocusRatio()
//	focusedRatio := 1 - liveRatio
//
//	if p.runnerStatus == nil {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Flexed(liveRatio, func(gtx layout.Context) layout.Dimensions {
//				gtx.Constraints.Min = gtx.Constraints.Max
//				return p.layoutLiveTranscriptCard(gtx)
//			}),
//		)
//	}
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//		layout.Flexed(liveRatio, func(gtx layout.Context) layout.Dimensions {
//			gtx.Constraints.Min = gtx.Constraints.Max
//			return p.layoutLiveTranscriptCard(gtx)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutTranscriptFocusResizeHandle(gtx)
//		}),
//		layout.Flexed(focusedRatio, func(gtx layout.Context) layout.Dimensions {
//			gtx.Constraints.Min = gtx.Constraints.Max
//			return p.layoutFocusedSentenceCard(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutBottomActions(gtx)
//		}),
//	)
//}
//
//func (p *Page) layoutTranscriptFocusResizeHandle(gtx layout.Context) layout.Dimensions {
//	return layout.Inset{
//		Top:    unit.Dp(8),
//		Bottom: unit.Dp(8),
//	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return bareutils.RoundedSurface(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{
//				Top:    unit.Dp(6),
//				Bottom: unit.Dp(6),
//				Left:   unit.Dp(12),
//				Right:  unit.Dp(12),
//			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						lbl := material.Body2(p.theme.Gio(), "Live")
//						lbl.Color = p.theme.Color.TextMuted
//						return lbl.Layout(gtx)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						pointer.CursorRowResize.Add(gtx.Ops)
//						gtx.Constraints.Min.X = gtx.Constraints.Max.X
//						slider := material.Slider(p.theme.Gio(), &p.transcriptFocusSplit)
//						slider.Color = p.theme.Color.Primary
//						return slider.Layout(gtx)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						lbl := material.Body2(p.theme.Gio(), "Focused")
//						lbl.Color = p.theme.Color.TextMuted
//						return lbl.Layout(gtx)
//					}),
//				)
//			})
//		})
//	})
//}
//
//func (p *Page) transcriptFocusRatio() float32 {
//	if p.transcriptFocusSplit.Value < 0.25 {
//		p.transcriptFocusSplit.Value = 0.25
//	}
//	if p.transcriptFocusSplit.Value > 0.75 {
//		p.transcriptFocusSplit.Value = 0.75
//	}
//	return p.transcriptFocusSplit.Value
//}
//
//func (p *Page) layoutLiveTranscriptCard(gtx layout.Context) layout.Dimensions {
//	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutCardHeader(gtx, "Live Transcript", "Scanning mode: saved words are highlighted inline")
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//				gtx.Constraints.Min = gtx.Constraints.Max
//				if p.runnerStatus == nil {
//					return p.layoutTranscriptIdleState(gtx)
//				}
//				return p.layoutTranscriptEditor(gtx)
//			}),
//		)
//	})
//}
//
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
//
//func (p *Page) layoutFocusedSentenceText(gtx layout.Context) layout.Dimensions {
//	text := p.structureSourceText()
//	if text == "" {
//		text = "Start the game to inspect the latest sentence."
//	}
//	text = cleanInlineText(text)
//	p.syncFocusedSentenceView(text)
//	if p.focusedFuriganaMode != focusedFuriganaHidden {
//		return p.layoutFocusedSentenceWithFurigana(gtx, text)
//	}
//
//	lbl := material.H6(p.theme.Gio(), text)
//	lbl.Color = p.theme.Color.Text
//	lbl.TextSize = p.focusedSentenceTextSize(gtx)
//	lbl.State = &p.focusedSentenceView
//	return lbl.Layout(gtx)
//}
//
//func (p *Page) layoutFocusedSentenceWithFurigana(gtx layout.Context, sentence string) layout.Dimensions {
//	analysis, err := japanese.AnalyzeSentence(sentence)
//	if err != nil || len(analysis.Tokens) == 0 {
//		lbl := material.H6(p.theme.Gio(), sentence)
//		lbl.Color = p.theme.Color.Text
//		lbl.TextSize = p.focusedSentenceTextSize(gtx)
//		return lbl.Layout(gtx)
//	}
//	lines := p.focusedSentenceTokenLines(gtx, analysis.Tokens)
//	children := make([]layout.FlexChild, 0, len(lines))
//	for i, line := range lines {
//		line := line
//		if i > 0 {
//			children = append(children, layout.Rigid(bareutils.SpacerH(unit.Dp(5))))
//		}
//		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lineChildren := make([]layout.FlexChild, 0, len(line))
//			for _, token := range line {
//				token := token
//				lineChildren = append(lineChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutFocusedFuriganaToken(gtx, token)
//				}))
//			}
//			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, lineChildren...)
//		}))
//	}
//	p.pruneFocusedTokenClicks(analysis.Tokens)
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
//}
//
//func (p *Page) focusedSentenceTokenLines(gtx layout.Context, tokens []japanese.Token) [][]japanese.Token {
//	maxWidth := gtx.Constraints.Max.X
//	if maxWidth <= 0 {
//		return [][]japanese.Token{tokens}
//	}
//	lines := make([][]japanese.Token, 0, 2)
//	line := make([]japanese.Token, 0, len(tokens))
//	lineWidth := 0
//	for _, token := range tokens {
//		tokenWidth := p.focusedSentenceTokenWidth(gtx, token)
//		if len(line) > 0 && lineWidth+tokenWidth > maxWidth {
//			lines = append(lines, line)
//			line = make([]japanese.Token, 0, len(tokens))
//			lineWidth = 0
//		}
//		line = append(line, token)
//		lineWidth += tokenWidth
//	}
//	if len(line) > 0 {
//		lines = append(lines, line)
//	}
//	return lines
//}
//
//func (p *Page) focusedSentenceTokenWidth(gtx layout.Context, token japanese.Token) int {
//	surfaceRunes := len([]rune(cleanInlineText(token.Surface)))
//	readingRunes := len([]rune(focusedTokenReading(token)))
//	runes := surfaceRunes
//	if readingRunes > runes {
//		runes = readingRunes
//	}
//	if runes <= 0 {
//		runes = 1
//	}
//	size := float32(p.focusedSentenceTextSize(gtx))
//	return gtx.Dp(unit.Dp(float32(runes)*size*0.72 + 16))
//}
//
//func (p *Page) layoutFocusedFuriganaToken(gtx layout.Context, token japanese.Token) layout.Dimensions {
//	key := structureTokenKey(token)
//	click := p.focusedTokenClickable(key)
//	reading := focusedTokenReading(token)
//	surface := cleanInlineText(token.Surface)
//	if surface == "" {
//		return layout.Dimensions{}
//	}
//	_, inFlashcards := p.structureTokenFlashcard(token)
//	dictionaryReady := focusedTokenDictionaryReady(token)
//	bg := focusedTokenColor(p.theme, token, p.selectedFocusedTokenKey == key, inFlashcards, dictionaryReady)
//	children := make([]layout.FlexChild, 0, 4)
//	if p.focusedFuriganaMode == focusedFuriganaAbove {
//		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutFocusedTokenSlot(gtx, unit.Dp(18), func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusedTokenReading(gtx, reading)
//			})
//		}), layout.Rigid(bareutils.SpacerH(unit.Dp(2))))
//	}
//	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//		return p.layoutFocusedTokenSlot(gtx, p.focusedTokenSurfaceSlotHeight(gtx), func(gtx layout.Context) layout.Dimensions {
//			lbl := material.H6(p.theme.Gio(), surface)
//			lbl.Color = p.theme.Color.Text
//			lbl.TextSize = p.focusedSentenceTextSize(gtx)
//			return lbl.Layout(gtx)
//		})
//	}))
//	if p.focusedFuriganaMode == focusedFuriganaBelow {
//		children = append(children, layout.Rigid(bareutils.SpacerH(unit.Dp(2))), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutFocusedTokenSlot(gtx, unit.Dp(18), func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusedTokenReading(gtx, reading)
//			})
//		}))
//	}
//	if p.focusedFuriganaMode == focusedFuriganaAbove {
//		children = append(children, layout.Rigid(bareutils.SpacerH(unit.Dp(3))), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutFocusedTokenSlot(gtx, unit.Dp(14), func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusedTokenMarker(gtx, inFlashcards, dictionaryReady)
//			})
//		}))
//	}
//	return layout.Inset{Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			pointer.CursorPointer.Add(gtx.Ops)
//			return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
//				return layout.Inset{
//					Top:    unit.Dp(5),
//					Bottom: unit.Dp(5),
//					Left:   unit.Dp(6),
//					Right:  unit.Dp(6),
//				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
//				})
//			})
//		})
//	})
//}
//
//func (p *Page) layoutFocusedTokenSlot(gtx layout.Context, height unit.Dp, w layout.Widget) layout.Dimensions {
//	slotHeight := gtx.Dp(height)
//	if slotHeight <= 0 {
//		return w(gtx)
//	}
//	local := gtx
//	local.Constraints.Min.Y = slotHeight
//	local.Constraints.Max.Y = slotHeight
//	dims := layout.Center.Layout(local, w)
//	dims.Size.Y = slotHeight
//	return dims
//}
//
//func (p *Page) focusedTokenSurfaceSlotHeight(gtx layout.Context) unit.Dp {
//	size := float32(p.focusedSentenceTextSize(gtx))
//	return unit.Dp(size + 12)
//}
//
//func (p *Page) layoutFocusedTokenMarker(gtx layout.Context, inFlashcards, dictionaryReady bool) layout.Dimensions {
//	text := " "
//	fg := p.theme.Color.TextMuted
//	if inFlashcards {
//		text = "✓"
//		fg = p.theme.Color.Success
//	} else if dictionaryReady {
//		text = "·"
//		fg = p.theme.Color.Secondary
//	}
//	lbl := material.Body2(p.theme.Gio(), text)
//	lbl.Color = fg
//	lbl.TextSize = unit.Sp(12)
//	return lbl.Layout(gtx)
//}
//
//func (p *Page) layoutFocusedTokenReading(gtx layout.Context, reading string) layout.Dimensions {
//	if reading == "" {
//		reading = " "
//	}
//	lbl := material.Body2(p.theme.Gio(), reading)
//	lbl.Color = color.NRGBA{R: 255, G: 137, B: 103, A: 255}
//	lbl.TextSize = p.translateDetailTextSize()
//	return lbl.Layout(gtx)
//}
//
//func (p *Page) layoutFocusedFuriganaControls(gtx layout.Context) layout.Dimensions {
//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body2(p.theme.Gio(), "Furigana")
//			lbl.Color = p.theme.Color.TextMuted
//			lbl.TextSize = p.translateDetailTextSize()
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutFuriganaModeButton(gtx, &p.furiganaHiddenButton, focusedFuriganaHidden, "Hide")
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(6))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutFuriganaModeButton(gtx, &p.furiganaAboveButton, focusedFuriganaAbove, "Above")
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(6))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutFuriganaModeButton(gtx, &p.furiganaBelowButton, focusedFuriganaBelow, "Below")
//		}),
//	)
//}
//
//func (p *Page) layoutFocusedTokenActions(gtx layout.Context) layout.Dimensions {
//	word := cleanInlineText(util.FirstNonEmpty(
//		p.selectedFocusedTokenWord,
//		p.selectedTranscriptText(),
//	))
//
//	existingCard, hasExistingCard := p.focusedSelectedTokenFlashcard(word)
//
//	meaning := "Click a word block above to inspect it."
//	sourceLabel := "Select word"
//	sourceLive := false
//
//	if note := cleanInlineText(p.selectedFocusedTokenNote); note != "" {
//		meaning = note
//		sourceLabel = "Selection"
//		sourceLive = true
//	} else if p.lookupResult != nil && cleanInlineText(p.lookupResult.Meaning) != "" {
//		meaning = cleanInlineText(p.lookupResult.Meaning)
//		sourceLabel = "Lookup"
//		sourceLive = true
//	} else if word != "" {
//		if hasExistingCard && cleanInlineText(existingCard.Meaning) != "" {
//			meaning = cleanInlineText(existingCard.Meaning)
//			sourceLabel = "Saved"
//			sourceLive = true
//		} else {
//			sourceLabel = "No match"
//		}
//	}
//
//	canAdd := p.lookupResult != nil && !hasExistingCard
//
//	audioPath := ""
//	if p.lookupResult != nil {
//		audioPath = strings.TrimSpace(p.lookupResult.AudioPath)
//	}
//	if audioPath == "" && hasExistingCard {
//		audioPath = strings.TrimSpace(existingCard.AudioPath)
//	}
//	canPlayAudio := audioPath != "" || p.hasTTSReference()
//
//	addButton := bareui.Button{
//		Clickable: &p.focusedTokenAddButton,
//		Text:      "Add",
//		Prefix:    "mdi:plus-circle-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//
//	audioButton := bareui.Button{
//		Clickable: &p.focusedTokenAudioButton,
//		Text:      "Audio",
//		Prefix:    "mdi:volume-high",
//		Variant:   bareui.ButtonSecondary,
//	}
//
//	return bareutils.RoundedSurface(
//		gtx,
//		p.theme.Color.Surface,
//		unit.Dp(p.theme.Radius.LG),
//		func(gtx layout.Context) layout.Dimensions {
//			return layout.UniformInset(unit.Dp(p.theme.Space.SM)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return bareutils.RoundedSurface(
//					gtx,
//					p.theme.Color.SurfaceAlt,
//					unit.Dp(p.theme.Radius.MD),
//					func(gtx layout.Context) layout.Dimensions {
//						return layout.Inset{
//							Top:    unit.Dp(p.theme.Space.SM),
//							Bottom: unit.Dp(p.theme.Space.SM),
//							Left:   unit.Dp(p.theme.Space.MD),
//							Right:  unit.Dp(p.theme.Space.MD),
//						}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//							return layout.Flex{
//								Axis:      layout.Horizontal,
//								Alignment: layout.Middle,
//							}.Layout(gtx,
//								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//									return p.layoutFocusedTokenSummary(gtx, word, meaning, sourceLabel, sourceLive)
//								}),
//
//								layout.Rigid(bareutils.SpacerW(unit.Dp(p.theme.Space.MD))),
//
//								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//									return p.layoutFocusedTokenButtons(gtx, addButton, audioButton, canAdd, canPlayAudio)
//								}),
//							)
//						})
//					},
//				)
//			})
//		},
//	)
//}
//
//func (p *Page) layoutFocusedTokenSummary(
//	gtx layout.Context,
//	word string,
//	meaning string,
//	sourceLabel string,
//	sourceLive bool,
//) layout.Dimensions {
//	return layout.Flex{
//		Axis: layout.Vertical,
//	}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis:      layout.Horizontal,
//				Alignment: layout.Middle,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					title := word
//					if title == "" {
//						title = "Focused token"
//					}
//
//					lbl := material.Body1(p.theme.Gio(), title)
//					lbl.Color = p.theme.Color.Text
//					lbl.TextSize = unit.Sp(float32(p.translateDetailTextSize()) + 1)
//					return lbl.Layout(gtx)
//				}),
//
//				layout.Rigid(bareutils.SpacerW(unit.Dp(p.theme.Space.SM))),
//
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//
//					return p.layoutStatusPill(gtx, sourceLabel, sourceLive)
//				}),
//			)
//		}),
//
//		layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
//
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body2(p.theme.Gio(), meaning)
//			lbl.Color = p.theme.Color.TextMuted
//			lbl.TextSize = p.translateDetailTextSize()
//			return lbl.Layout(gtx)
//		}),
//	)
//}
//
//func (p *Page) layoutFocusedTokenButtons(
//	gtx layout.Context,
//	addButton bareui.Button,
//	audioButton bareui.Button,
//	canAdd bool,
//	canPlayAudio bool,
//) layout.Dimensions {
//	return layout.Flex{
//		Axis:      layout.Horizontal,
//		Alignment: layout.Middle,
//	}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if !canAdd {
//				return addButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//			}
//			return addButton.Layout(gtx, p.theme, p.iconify)
//		}),
//
//		layout.Rigid(bareutils.SpacerW(unit.Dp(p.theme.Space.SM))),
//
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if !canPlayAudio {
//				return audioButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//			}
//			return audioButton.Layout(gtx, p.theme, p.iconify)
//		}),
//	)
//}
//
//func (p *Page) layoutMiniStatusPill(gtx layout.Context, text string, live bool) layout.Dimensions {
//	bg := p.theme.Color.Surface
//	fg := p.theme.Color.TextMuted
//
//	if live {
//		fg = p.theme.Color.Primary
//		bg = color.NRGBA{
//			R: uint8((uint16(fg.R) + uint16(p.theme.Color.Surface.R)*5) / 6),
//			G: uint8((uint16(fg.G) + uint16(p.theme.Color.Surface.G)*5) / 6),
//			B: uint8((uint16(fg.B) + uint16(p.theme.Color.Surface.B)*5) / 6),
//			A: 255,
//		}
//	}
//
//	return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
//		return layout.Inset{
//			Top:    unit.Dp(3),
//			Bottom: unit.Dp(3),
//			Left:   unit.Dp(7),
//			Right:  unit.Dp(7),
//		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body2(p.theme.Gio(), text)
//			lbl.Color = fg
//			lbl.TextSize = unit.Sp(11)
//			return lbl.Layout(gtx)
//		})
//	})
//}
//func (p *Page) layoutFuriganaModeButton(gtx layout.Context, click *widget.Clickable, mode, label string) layout.Dimensions {
//	active := p.focusedFuriganaMode == mode
//	bg := p.theme.Color.Surface
//	fg := p.theme.Color.TextMuted
//	if active {
//		bg = p.theme.Color.Primary
//		fg = bareutils.ReadableOn(bg)
//	} else if click.Hovered() {
//		bg = p.theme.Color.SurfaceAlt
//		fg = p.theme.Color.Text
//	}
//	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		pointer.CursorPointer.Add(gtx.Ops)
//		return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{
//				Top:    unit.Dp(6),
//				Bottom: unit.Dp(6),
//				Left:   unit.Dp(9),
//				Right:  unit.Dp(9),
//			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body2(p.theme.Gio(), label)
//				lbl.Color = fg
//				lbl.TextSize = p.translateDetailTextSize()
//				return lbl.Layout(gtx)
//			})
//		})
//	})
//}
//
//func (p *Page) layoutFocusedSentenceChips(gtx layout.Context, analysis japanese.Analysis, errText string) layout.Dimensions {
//	if errText != "" {
//		lbl := material.Body1(p.theme.Gio(), errText)
//		lbl.Color = p.theme.Color.Warning
//		lbl.TextSize = p.translateDetailTextSize()
//		return lbl.Layout(gtx)
//	}
//	children := make([]layout.FlexChild, 0, min(4, len(analysis.Tokens)))
//	for _, token := range focusTokens(analysis.Tokens, 4) {
//		token := token
//		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{Right: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return p.layoutFocusChip(gtx, token)
//			})
//		}))
//	}
//	if len(children) == 0 {
//		return layout.Dimensions{}
//	}
//	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
//}
//
//func (p *Page) layoutFocusChip(gtx layout.Context, token japanese.Token) layout.Dimensions {
//	bg := barethemes.Mix(
//		p.theme.Color.Primary,
//		p.theme.Color.SurfaceAlt,
//		0.18,
//	)
//
//	border := barethemes.Mix(
//		p.theme.Color.Primary,
//		p.theme.Color.TextMuted,
//		0.22,
//	)
//
//	spacingX := unit.Dp(12)
//	spacingY := unit.Dp(8)
//	radius := unit.Dp(p.theme.Radius.SM)
//
//	return layout.Inset{
//		Top:    unit.Dp(2),
//		Bottom: unit.Dp(2),
//		Left:   unit.Dp(2),
//		Right:  unit.Dp(2),
//	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		macro := op.Record(gtx.Ops)
//
//		dims := layout.Inset{
//			Top:    spacingY,
//			Bottom: spacingY,
//			Left:   spacingX,
//			Right:  spacingX,
//		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{
//				Axis:      layout.Vertical,
//				Alignment: layout.Middle,
//			}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body2(
//						p.theme.Gio(),
//						posMajorLabel(token.POSMajor()),
//					)
//					lbl.Color = p.theme.Color.TextMuted
//					lbl.TextSize = p.translateDetailTextSize()
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(
//						p.theme.Gio(),
//						structureFlashcardWord(token),
//					)
//					lbl.Color = p.theme.Color.Text
//					lbl.TextSize = p.translateDetailTextSize()
//					lbl.Font.Weight = font.Bold
//					return lbl.Layout(gtx)
//				}),
//			)
//		})
//
//		call := macro.Stop()
//
//		rr := clip.UniformRRect(
//			image.Rectangle{Max: dims.Size},
//			gtx.Dp(radius),
//		)
//
//		paint.FillShape(gtx.Ops, bg, rr.Op(gtx.Ops))
//
//		stroke := clip.Stroke{
//			Path:  rr.Path(gtx.Ops),
//			Width: float32(gtx.Dp(unit.Dp(1))),
//		}.Op()
//
//		paint.FillShape(gtx.Ops, border, stroke)
//
//		call.Add(gtx.Ops)
//
//		return dims
//	})
//}
//func (p *Page) layoutFocusedLookupBar(gtx layout.Context) layout.Dimensions {
//	selected := p.selectedTranscriptText()
//	if selected == "" {
//		selected = "Highlight a word in the focused sentence to look it up."
//	}
//	lookupButton := bareui.Button{
//		Clickable: &p.focusedLookupButton,
//		Text:      "Lookup Selection",
//		Prefix:    "mdi:book-search-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body2(p.theme.Gio(), selected)
//			lbl.Color = p.theme.Color.TextMuted
//			lbl.TextSize = p.translateDetailTextSize()
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if p.selectedTranscriptText() == "" {
//				return lookupButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//			}
//			return lookupButton.Layout(gtx, p.theme, p.iconify)
//		}),
//	)
//}
//
//func (p *Page) layoutFocusedTranslationSection(gtx layout.Context) layout.Dimensions {
//	p.syncTranslationEditor()
//
//	return bareutils.RoundedSurface(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			children := []layout.FlexChild{
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutTranslationHeader(gtx)
//				}),
//			}
//			if !p.translationCollapsed {
//				children = append(children,
//					layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						editor := material.Editor(p.theme.Gio(), &p.translationEditor, "Type or edit the translation here")
//						editor.Color = p.theme.Color.Text
//						editor.HintColor = p.theme.Color.TextMuted
//						editor.TextSize = p.translateDetailTextSize()
//						maxHeight := min(gtx.Constraints.Max.Y, gtx.Dp(unit.Dp(96)))
//						gtx.Constraints.Min.Y = min(maxHeight, gtx.Dp(unit.Dp(58)))
//						gtx.Constraints.Max.Y = maxHeight
//						return editor.Layout(gtx)
//					}),
//				)
//			}
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
//		})
//	})
//}
//
//func (p *Page) layoutTranslationHeader(gtx layout.Context) layout.Dimensions {
//	chevron := "mdi:chevron-down"
//	if p.translationCollapsed {
//		chevron = "mdi:chevron-right"
//	}
//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.translationToggleButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				pointer.CursorPointer.Add(gtx.Ops)
//				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//					if p.iconify == nil {
//						lbl := material.Body1(p.theme.Gio(), "+")
//						lbl.Color = p.theme.Color.Text
//						return lbl.Layout(gtx)
//					}
//					return p.iconify.Layout(gtx, chevron, unit.Dp(18), p.theme.Color.Text)
//				})
//			})
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body1(p.theme.Gio(), "Live Translation")
//			lbl.Color = p.theme.Color.Text
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body2(p.theme.Gio(), "to")
//			lbl.Color = p.theme.Color.TextMuted
//			lbl.TextSize = p.translateDetailTextSize()
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			return p.targetLanguageDrop.Layout(gtx, p.theme, p.iconify, p.selectedTargetLanguage, func(gtx layout.Context) layout.Dimensions {
//				return gui.LayoutOptionMenu(gtx, p.targetLanguageOptions, p.selectedTargetLanguage, p.theme, p.iconify)
//			})
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			saveButton := bareui.Button{
//				Clickable: &p.saveTranslationButton,
//				Text:      "Save",
//				Prefix:    "mdi:content-save-outline",
//				Variant:   bareui.ButtonSecondary,
//			}
//			if strings.TrimSpace(p.translationEditor.Text()) == "" || strings.TrimSpace(p.structureSourceText()) == "" {
//				return saveButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//			}
//			return saveButton.Layout(gtx, p.theme, p.iconify)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body2(p.theme.Gio(), "Auto")
//					lbl.Color = p.theme.Color.TextMuted
//					lbl.TextSize = p.translateDetailTextSize()
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerW(unit.Dp(4))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					autoSwitch := material.Switch(p.theme.Gio(), &p.autoTranslateMissing, "Auto translate missing sentences")
//					autoSwitch.Color.Enabled = p.theme.Color.Primary
//					autoSwitch.Color.Disabled = p.theme.Color.Border
//					return autoSwitch.Layout(gtx)
//				}),
//			)
//		}),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			text := "Generate"
//			if p.translationGeneratingKey != "" {
//				text = "Generating"
//			}
//			generateButton := bareui.Button{
//				Clickable: &p.generateTranslationButton,
//				Text:      text,
//				Prefix:    "mdi:creation-outline",
//				Variant:   bareui.ButtonPrimary,
//			}
//			if p.translationGeneratingKey != "" || strings.TrimSpace(p.structureSourceText()) == "" {
//				return generateButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//			}
//			return generateButton.Layout(gtx, p.theme, p.iconify)
//		}),
//	)
//}
//
//func (p *Page) layoutTranscriptCard(gtx layout.Context, bg color.NRGBA, child layout.Widget) layout.Dimensions {
//	return bareutils.Panel(gtx, bg, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, child)
//	})
//}
//
//func (p *Page) layoutCardHeader(gtx layout.Context, title, hint string) layout.Dimensions {
//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			lbl := material.H6(p.theme.Gio(), title)
//			lbl.Color = p.theme.Color.Text
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if p.isCompactLayout(gtx) {
//				return layout.Dimensions{}
//			}
//			lbl := material.Body2(p.theme.Gio(), hint)
//			lbl.Color = p.theme.Color.TextMuted
//			lbl.TextSize = p.translateDetailTextSize()
//			return lbl.Layout(gtx)
//		}),
//	)
//}
//
//func (p *Page) focusedSentenceTextSize(gtx layout.Context) unit.Sp {
//	if p.focusedTextSize > 0 {
//		return p.focusedTextSize
//	}
//	if p.isCompactLayout(gtx) {
//		return unit.Sp(20)
//	}
//	return unit.Sp(26)
//}
//
//func (p *Page) translateDetailTextSize() unit.Sp {
//	if p.translateDetailSize > 0 {
//		return p.translateDetailSize
//	}
//	return unit.Sp(15)
//}
//
//func (p *Page) layoutBottomActions(gtx layout.Context) layout.Dimensions {
//	playButton := bareui.Button{
//		Clickable: &p.playSentenceButton,
//		Text:      "Play Audio",
//		Prefix:    "mdi:volume-high",
//		Variant:   bareui.ButtonSecondary,
//	}
//	structureButton := bareui.Button{
//		Clickable: &p.translateSentenceButton,
//		Text:      "Sentence Structure",
//		Prefix:    "mdi:translate",
//		Variant:   bareui.ButtonSecondary,
//	}
//	saveButton := bareui.Button{
//		Clickable: &p.saveSentenceButton,
//		Text:      "Save Sentence",
//		Prefix:    "mdi:heart-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return playButton.Layout(gtx, p.theme, p.iconify) }),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return structureButton.Layout(gtx, p.theme, p.iconify) }),
//		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return saveButton.Layout(gtx, p.theme, p.iconify) }),
//	)
//}
//
//func (p *Page) layoutContextRail(gtx layout.Context) layout.Dimensions {
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//		layout.Flexed(5, func(gtx layout.Context) layout.Dimensions {
//			gtx.Constraints.Min = gtx.Constraints.Max
//			return p.layoutWordDetailsCard(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//		layout.Flexed(4, func(gtx layout.Context) layout.Dimensions {
//			gtx.Constraints.Min = gtx.Constraints.Max
//			return p.layoutChoicesCard(gtx)
//		}),
//	)
//}
//
//func (p *Page) layoutWordDetailsCard(gtx layout.Context) layout.Dimensions {
//	card, hasCard := p.contextFlashcard()
//	title := "Vocabulary"
//	reading := "Select a saved word or run lookup"
//	meaning := "Word details appear here while the transcript stays readable."
//	meta := "No card selected"
//	if hasCard {
//		title = util.FirstNonEmpty(strings.TrimSpace(card.Text), strings.TrimSpace(p.popupWord), "Vocabulary")
//		reading = util.FirstNonEmpty(strings.TrimSpace(card.Reading), strings.TrimSpace(card.PronunciationText), "No reading saved")
//		meaning = util.FirstNonEmpty(strings.TrimSpace(card.Meaning), strings.TrimSpace(card.SourceLine), meaning)
//		meta = util.FirstNonEmpty(p.flashcardMetaText(card), "Saved flashcard")
//	} else if p.lookupResult != nil {
//		title = util.FirstNonEmpty(p.lookupResult.Query, p.lookupResult.Headword, p.lookupResult.Key)
//		reading = util.FirstNonEmpty(p.lookupResult.Reading, p.lookupResult.PronunciationText, "No reading found")
//		meaning = util.FirstNonEmpty(p.lookupResult.Meaning, meaning)
//		meta = "Dictionary lookup"
//	}
//	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
//		playButton := bareui.Button{Clickable: &p.playAudioButton, Text: "mdi:volume-high", Icon: true, Variant: bareui.ButtonSecondary}
//		//lookupButton := bareui.Button{Clickable: &p.searchWordButton, Text: "Lookup", Prefix: "mdi:book-search-outline", Variant: bareui.ButtonSecondary}
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//			//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			//			lbl := material.H6(p.theme.Gio(), "Word Details")
//			//			lbl.Color = p.theme.Color.Text
//			//			return lbl.Layout(gtx)
//			//		}),
//			//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			//			return p.layoutStatusPill(gtx, contextVocabPillText(hasCard), hasCard)
//			//		}),
//			//	)
//			//}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						lbl := material.H5(p.theme.Gio(), title)
//						lbl.Color = p.theme.Color.Text
//						lbl.TextSize = unit.Sp(32)
//						return lbl.Layout(gtx)
//					}),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return playButton.Layout(gtx, p.theme, p.iconify)
//					}),
//				)
//			}),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(p.theme.Gio(), reading)
//				lbl.Color = p.theme.Color.TextMuted
//				lbl.TextSize = p.translateDetailTextSize()
//				return lbl.Layout(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(p.theme.Gio(), meaning)
//				lbl.Color = p.theme.Color.Text
//				lbl.TextSize = p.translateDetailTextSize()
//				return lbl.Layout(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body2(p.theme.Gio(), meta)
//				lbl.Color = p.theme.Color.TextMuted
//				lbl.TextSize = p.translateDetailTextSize()
//				return lbl.Layout(gtx)
//			}),
//			//layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
//			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			//	return lookupButton.Layout(gtx, p.theme, p.iconify)
//			//}),
//		)
//	})
//}
//
//func (p *Page) layoutChoicesCard(gtx layout.Context) layout.Dimensions {
//	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutCardHeader(gtx, "Dialog Choices", "Structure")
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutChoiceRow(gtx, "1", "Inspect sentence structure")
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutChoiceRow(gtx, "2", "Create a flashcard from selected text")
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return p.layoutChoiceRow(gtx, "3", "Review saved transcript vocabulary")
//			}),
//			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body2(p.theme.Gio(), "Highlight text in Focused Sentence, then use Lookup Selection or Word Details.")
//				lbl.Color = p.theme.Color.TextMuted
//				lbl.TextSize = p.translateDetailTextSize()
//				return lbl.Layout(gtx)
//			}),
//		)
//	})
//}
//
//func (p *Page) layoutChoiceRow(gtx layout.Context, number, text string) layout.Dimensions {
//	return bareutils.RoundedSurface(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.Inset{
//			Top:    unit.Dp(10),
//			Bottom: unit.Dp(10),
//			Left:   unit.Dp(12),
//			Right:  unit.Dp(12),
//		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutStatusPill(gtx, number, false)
//				}),
//				layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(p.theme.Gio(), text)
//					lbl.Color = p.theme.Color.TextMuted
//					lbl.TextSize = p.translateDetailTextSize()
//					return lbl.Layout(gtx)
//				}),
//			)
//		})
//	})
//}
//

//
//func (p *Page) layoutTranscriptRow(gtx layout.Context, row transcriptRow) layout.Dimensions {
//	if row.Info {
//		return p.layoutTranscriptInfoRow(gtx, row)
//	}
//	click := p.transcriptRowClickable(row.Key)
//	selected := row.Key == p.currentTranscriptRowKey()
//	bg := p.theme.Color.Surface
//	fg := p.theme.Color.Text
//	timeColor := p.theme.Color.TextMuted
//	if selected {
//		bg = barethemes.Mix(p.theme.Color.Primary, p.theme.Color.SurfaceAlt, 0.22)
//		timeColor = p.theme.Color.Primary
//	}
//	if click.Hovered() && !selected {
//		bg = p.theme.Color.SurfaceAlt
//	}
//	displayText := p.transcriptRowDisplayText(row)
//	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		pointer.CursorPointer.Add(gtx.Ops)
//		return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{
//				Top:    unit.Dp(12),
//				Bottom: unit.Dp(12),
//				Left:   unit.Dp(14),
//				Right:  unit.Dp(12),
//			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						gtx.Constraints.Min.X = gtx.Dp(p.transcriptTimestampWidth())
//						lbl := material.Body2(p.theme.Gio(), p.transcriptTimestampText(row.Time))
//						lbl.Color = timeColor
//						return lbl.Layout(gtx)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(14))),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						if strings.TrimSpace(row.Speaker) == "" {
//							return layout.Dimensions{}
//						}
//						return p.layoutTranscriptSpeaker(gtx, row.Speaker, selected)
//					}),
//					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//						lbl := material.Body1(p.theme.Gio(), displayText)
//						lbl.Color = fg
//						if p.isTranscriptRowTranslationShown(row) {
//							lbl.Color = p.theme.Color.Primary
//						}
//						lbl.TextSize = p.transcriptTextSize
//						return lbl.Layout(gtx)
//					}),
//					layout.Rigid(bareutils.S					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return p.layoutTranscriptVoiceIcon(gtx, row)
//					}),pacerW(unit.Dp(10))),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return p.layoutRowVocabIndicators(gtx, row.VocabWords)
//					}),
//					layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return p.layoutTranscriptTranslateIcon(gtx, row)
//					}),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return p.layoutTranscriptVoiceIcon(gtx, row)
//					}),
//					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//						return p.layoutRowIcon(gtx, "mdi:heart-outline", true)
//					}),
//				)
//			})
//		})
//	})
//}
//
//func (p *Page) layoutTranscriptInfoRow(gtx layout.Context, row transcriptRow) layout.Dimensions {
//	return bareutils.RoundedSurface(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.Inset{
//			Top:    unit.Dp(9),
//			Bottom: unit.Dp(9),
//			Left:   unit.Dp(14),
//			Right:  unit.Dp(12),
//		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					gtx.Constraints.Min.X = gtx.Dp(p.transcriptTimestampWidth())
//					lbl := material.Body2(p.theme.Gio(), p.transcriptTimestampText(row.Time))
//					lbl.Color = p.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerW(unit.Dp(14))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutRowIcon(gtx, "mdi:information-outline", true)
//				}),
//				layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body2(p.theme.Gio(), row.Text)
//					lbl.Color = p.theme.Color.TextMuted
//					lbl.TextSize = p.translateDetailTextSize()
//					return lbl.Layout(gtx)
//				}),
//			)
//		})
//	})
//}
//
//func (p *Page) transcriptTimestampWidth() unit.Dp {
//	if p.compactTimestamps {
//		return unit.Dp(54)
//	}
//	return unit.Dp(78)
//}
//
//func (p *Page) transcriptTimestampText(timestamp string) string {
//	timestamp = strings.TrimSpace(timestamp)
//	if timestamp == "" {
//		return ""
//	}compactTimestamps
//	if !p.compactTimestamps {
//		return timestamp
//	}
//	fields := strings.Fields(timestamp)
//	if len(fields) > 0 {
//		return fields[len(fields)-1]
//	}
//	return timestamp
//}
//
//func (p *Page) layoutTranscriptSpeaker(gtx layout.Context, speaker string, selected bool) layout.Dimensions {
//	speaker = strings.TrimSpace(speaker)
//	if speaker == "" {
//		return layout.Dimensions{}
//	}
//	fg := p.theme.Color.Primary
//	bg := color.NRGBA{R: fg.R, G: fg.G, B: fg.B, A: 34}
//	if selected {
//		bg = color.NRGBA{R: fg.R, G: fg.G, B: fg.B, A: 54}
//	}
//	return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return RoundedSurfaceWrap(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{
//				Top:    unit.Dp(4),
//				Bottom: unit.Dp(4),
//				Left:   unit.Dp(8),
//				Right:  unit.Dp(8),
//			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body2(p.theme.Gio(), speaker)
//				lbl.Color = fg
//				return lbl.Layout(gtx)
//			})
//		})
//	})
//}
//
//func (p *Page) layoutRowVocabIndicators(gtx layout.Context, words []string) layout.Dimensions {
//	if len(words) == 0 {
//		return layout.Dimensions{}
//	}
//	visible := words
//	if len(visible) > 2 {
//		visible = visible[:2]
//	}
//	children := make([]layout.FlexChild, 0, len(visible)+1)
//	for _, word := range visible {
//		word := word
//		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return p.layoutVocabChip(gtx, word)
//			})
//		}))
//	}
//	if extra := len(words) - len(visible); extra > 0 {
//		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return p.layoutVocabChip(gtx, fmt.Sprintf("+%d", extra))
//			})
//		}))
//	}
//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
//}
//
//func (p *Page) layoutVocabChip(gtx layout.Context, text string) layout.Dimensions {
//	bg := color.NRGBA{R: p.theme.Color.Primary.R, G: p.theme.Color.Primary.G, B: p.theme.Color.Primary.B, A: 34}
//	return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
//		return layout.Inset{
//			Top:    unit.Dp(4),
//			Bottom: unit.Dp(4),
//			Left:   unit.Dp(7),
//			Right:  unit.Dp(7),
//		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body2(p.theme.Gio(), text)
//			lbl.Color = p.theme.Color.Primary
//			return lbl.Layout(gtx)
//		})
//	})
//}
//
//func (p *Page) layoutRowIcon(gtx layout.Context, icon string, enabled bool) layout.Dimensions {
//	if p.iconify == nil {
//		return layout.Dimensions{}
//	}
//	clr := p.theme.Color.TextMuted
//	if !enabled {
//		clr = color.NRGBA{R: clr.R, G: clr.G, B: clr.B, A: 80}
//	}
//	return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return bareutils.RoundedSurface(gtx, color.NRGBA{}, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
//			return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return p.iconify.Layout(gtx, icon, unit.Dp(16), clr)
//			})
//		})
//	})
//}
//
//func (p *Page) layoutTranscriptVoiceIcon(gtx layout.Context, row transcriptRow) layout.Dimensions {
//	if strings.TrimSpace(row.Voice) == "" && !p.hasTTSReference() {
//		return p.layoutRowIcon(gtx, "mdi:volume-off", false)
//	}
//	click := p.transcriptRowVoiceClickable(row.Key)
//	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		pointer.CursorPointer.Add(gtx.Ops)
//		icon := "mdi:volume-high"
//		if strings.TrimSpace(row.Voice) == "" {
//			icon = "mdi:account-voice"
//		}
//		return p.layoutRowIcon(gtx, icon, true)
//	})
//}
//
//func (p *Page) layoutTranscriptTranslateIcon(gtx layout.Context, row transcriptRow) layout.Dimensions {
//	enabled := strings.TrimSpace(row.Text) != "" && strings.TrimSpace(p.selectedTargetLanguage) != ""
//	if !enabled {
//		return p.layoutRowIcon(gtx, "mdi:translate", false)
//	}
//	click := p.transcriptRowTranslateClickable(row.Key)
//	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		pointer.CursorPointer.Add(gtx.Ops)
//		icon := "mdi:translate"
//		if p.isTranscriptRowTranslationShown(row) {
//			icon = "mdi:translate-off"
//		}
//		return p.layoutRowIcon(gtx, icon, true)
//	})
//}
//
//func (p *Page) layoutTranscriptLabel(gtx layout.Context, clr color.NRGBA, state *widget.Selectable) layout.Dimensions {
//	label := material.Body1(p.theme.Gio(), p.displayTranscript)
//	label.Color = clr
//	label.TextSize = p.transcriptTextSize
//	label.State = state
//	return label.Layout(gtx)
//}
//
//func (p *Page) layoutTranscriptPopup(gtx layout.Context) layout.Dimensions {
//	if p.popupFlashcard == nil {
//		p.popupBounds = image.Rectangle{}
//		return layout.Dimensions{}
//	}
//
//	card := *p.popupFlashcard
//	popupWidth := gtx.Dp(unit.Dp(280))
//	if popupWidth > gtx.Constraints.Max.X {
//		popupWidth = gtx.Constraints.Max.X
//	}
//	if popupWidth <= 0 {
//		return layout.Dimensions{}
//	}
//	popupHeightGuess := gtx.Dp(p.popupHeightGuess(card))
//
//	x := p.popupAnchor.Min.X
//	if x+popupWidth > gtx.Constraints.Max.X {
//		x = gtx.Constraints.Max.X - popupWidth
//	}
//	if x < 0 {
//		x = 0
//	}
//
//	y := p.popupAnchor.Min.Y - popupHeightGuess - gtx.Dp(unit.Dp(10))
//	if y < 0 {
//		y = p.popupAnchor.Max.Y + gtx.Dp(unit.Dp(10))
//	}
//	if y+popupHeightGuess > gtx.Constraints.Max.Y {
//		y = max(0, gtx.Constraints.Max.Y-popupHeightGuess)
//	}
//	p.popupBounds = image.Rect(x, y, x+popupWidth, y+popupHeightGuess)
//
//	p.layoutTranscriptPopupDismissRegions(gtx, p.popupBounds)
//
//	offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
//	local := gtx
//	local.Constraints.Min = image.Point{}
//	local.Constraints.Max = image.Pt(popupWidth, popupHeightGuess)
//	dims := p.layoutTranscriptPopupCard(local, card)
//	offset.Pop()
//	p.popupBounds = image.Rect(x, y, x+dims.Size.X, y+dims.Size.Y)
//	return layout.Dimensions{}
//}
//
//func (p *Page) layoutTranscriptPopupCard(gtx layout.Context, card flashcards.Flashcard) layout.Dimensions {
//	titleText := util.FirstNonEmpty(strings.TrimSpace(card.Text), strings.TrimSpace(p.popupWord), strings.TrimSpace(card.Reading), "Vocabulary")
//	bodyText := util.FirstNonEmpty(strings.TrimSpace(card.Meaning), strings.TrimSpace(card.SourceLine), "No saved meaning for this word yet.")
//	audioButton := bareui.Button{
//		Clickable: &p.transcriptPopupAudioButton,
//		Text:      "Play Audio",
//		Prefix:    "mdi:play-circle-outline",
//		Variant:   bareui.ButtonSecondary,
//	}
//	closeButton := bareui.Button{
//		Clickable: &p.transcriptPopupCloseButton,
//		Text:      "mdi:close",
//		Icon:      true,
//		Prefix:    "mdi:close",
//		Variant:   bareui.ButtonGhost,
//	}
//	borderColor := transcriptPopupBorderColor(p.theme.Color.Primary)
//	return layout.Stack{}.Layout(gtx,
//		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
//			return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//				return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//									lbl := material.Body1(p.theme.Gio(), "Vocabulary")
//									lbl.Color = p.theme.Color.TextMuted
//									return lbl.Layout(gtx)
//								}),
//								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//									return closeButton.Layout(gtx, p.theme, p.iconify)
//								}),
//							)
//						}),
//						layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							lbl := material.H6(p.theme.Gio(), titleText)
//							lbl.Color = p.theme.Color.Text
//							return lbl.Layout(gtx)
//						}),
//						layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							lbl := material.Body1(p.theme.Gio(), bodyText)
//							lbl.Color = p.theme.Color.Text
//							return lbl.Layout(gtx)
//						}),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							meta := p.flashcardMetaText(card)
//							if meta == "" {
//								return layout.Dimensions{}
//							}
//							return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//								lbl := material.Body1(p.theme.Gio(), meta)
//								lbl.Color = p.theme.Color.TextMuted
//								return lbl.Layout(gtx)
//							})
//						}),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							if !util.IsExistingFile(card.AudioPath) && !p.hasTTSReference() {
//								return layout.Dimensions{}
//							}
//							return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//								return audioButton.Layout(gtx, p.theme, p.iconify)
//							})
//						}),
//					)
//				})
//			})
//		}),
//		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
//			border := clip.Stroke{
//				Path:  clip.RRect{Rect: image.Rectangle{Max: gtx.Constraints.Max}, NW: gtx.Dp(unit.Dp(p.theme.Radius.MD)), NE: gtx.Dp(unit.Dp(p.theme.Radius.MD)), SW: gtx.Dp(unit.Dp(p.theme.Radius.MD)), SE: gtx.Dp(unit.Dp(p.theme.Radius.MD))}.Path(gtx.Ops),
//				Width: float32(gtx.Dp(unit.Dp(1))),
//			}.Op()
//			paint.FillShape(gtx.Ops, borderColor, border)
//			return layout.Dimensions{}
//		}),
//	)
//}
//
//func (p *Page) layoutTranscriptPopupDismissRegions(gtx layout.Context, popup image.Rectangle) {
//	regions := [4]image.Rectangle{
//		image.Rect(0, 0, gtx.Constraints.Max.X, popup.Min.Y),
//		image.Rect(0, popup.Min.Y, popup.Min.X, popup.Max.Y),
//		image.Rect(popup.Max.X, popup.Min.Y, gtx.Constraints.Max.X, popup.Max.Y),
//		image.Rect(0, popup.Max.Y, gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
//	}
//	for i, region := range regions {
//		if region.Empty() {
//			continue
//		}
//		offset := op.Offset(region.Min).Push(gtx.Ops)
//		local := gtx
//		local.Constraints.Min = region.Size()
//		local.Constraints.Max = region.Size()
//		p.popupDismissClicks[i].Layout(local, func(gtx layout.Context) layout.Dimensions {
//			return layout.Dimensions{Size: region.Size()}
//		})
//		offset.Pop()
//	}
//}
//
//func (p *Page) popupHeightGuess(card flashcards.Flashcard) unit.Dp {
//	bodyText := util.FirstNonEmpty(strings.TrimSpace(card.Meaning), strings.TrimSpace(card.SourceLine), "No saved meaning for this word yet.")
//	height := 92
//	height += min(3, 1+strings.Count(bodyText, "\n")) * 18
//	if meta := p.flashcardMetaText(card); meta != "" {
//		height += min(3, 1+strings.Count(meta, "\n")) * 14
//	}
//	if util.IsExistingFile(card.AudioPath) {
//		height += 34
//	}
//	if height < 112 {
//		height = 112
//	}
//	if height > 168 {
//		height = 168
//	}
//	return unit.Dp(height)
//}
//
//func (p *Page) shouldCollapseFlashcardComposer() bool {
//	if p.selectedTranscriptText() != "" {
//		return false
//	}
//	if strings.TrimSpace(p.wordEditor.Text()) != "" {
//		return false
//	}
//	if strings.TrimSpace(p.meaningEditor.Text()) != "" {
//		return false
//	}
//	return len(p.lookupResults) == 0
//}
//
//func (p *Page) resetFlashcardComposer() {
//	p.wordEditor.SetText("")
//	p.meaningEditor.SetText("")
//	p.hideReadingInAnki.Value = false
//	p.lastAutoWord = ""
//	p.hideReadingSet = false
//	p.lookupResult = nil
//	p.lookupResults = nil
//	p.focusedLookupPendingKey = ""
//	p.composerMinimized = true
//	p.composerLastUsed = time.Now()
//}
//
//func (p *Page) syncComposerMinimized() {
//	if p.composerHasActiveContent() {
//		p.composerMinimized = false
//		p.composerLastUsed = time.Now()
//		return
//	}
//	if p.shouldCollapseFlashcardComposer() && time.Since(p.composerLastUsed) > 4*time.Second {
//		p.composerMinimized = true
//	}
//}
//
//func (p *Page) composerHasActiveContent() bool {
//	if p.selectedTranscriptText() != "" {
//		return true
//	}
//	if strings.TrimSpace(p.wordEditor.Text()) != "" {
//		return true
//	}
//	if strings.TrimSpace(p.meaningEditor.Text()) != "" {
//		return true
//	}
//	return len(p.lookupResults) > 0
//}
//
//func (p *Page) syncHideReadingDefault() {
//	word := strings.TrimSpace(p.wordEditor.Text())
//	if word == p.lastAutoWord {
//		return
//	}
//	p.lastAutoWord = word
//	if p.hideRead//func (p *Page) layoutTranscriptIdleState(gtx layout.Context) layout.Dimensions {
////	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
////		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
////			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
////				lbl := material.H6(p.theme.Gio(), "Transcript Hidden")
////				lbl.Color = p.theme.Color.Text
////				return lbl.Layout(gtx)
////			}),
////			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
////			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
////				lbl := material.Body1(p.theme.Gio(), "Start the game to show live transcript text here.")
////				lbl.Color = p.theme.Color.TextMuted
////				return lbl.Layout(gtx)
////			}),
////			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
////			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
////				lbl := material.Body1(p.theme.Gio(), "The flashcard composer stays on this page next to the transcript.")
////				lbl.Color = p.theme.Color.TextMuted
////				return lbl.Layout(gtx)
////			}),
////		)
////	})
////}ingSet {
//		return
//	}
//	p.hideReadingInAnki.Value = util.ContainsKanji(word)
//}
//
//func (p *Page) layoutTranscriptIdleState(gtx layout.Context) layout.Dimensions {
//	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.H6(p.theme.Gio(), "Transcript Hidden")
//				lbl.Color = p.theme.Color.Text
//				return lbl.Layout(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(p.theme.Gio(), "Start the game to show live transcript text here.")
//				lbl.Color = p.theme.Color.TextMuted
//				return lbl.Layout(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(p.theme.Gio(), "The flashcard composer stays on this page next to the transcript.")
//				lbl.Color = p.theme.Color.TextMuted
//				return lbl.Layout(gtx)
//			}),
//		)
//	})
//}
//
//func (p *Page) layoutFlashcardComposer(gtx layout.Context) layout.Dimensions {
//	p.syncComposerMinimized()
//	if p.composerMinimized {
//		return p.layoutFlashcardComposerMini(gtx)
//	}
//	if p.shouldCollapseFlashcardComposer() {
//		return p.layoutFlashcardComposerHint(gtx)
//	}
//
//	minimizeButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-down", Icon: true, Prefix: "mdi:chevron-down", Variant: bareui.ButtonGhost}
//
//	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutComposerHeader(gtx, &minimizeButton)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					switch p.composerFocus {
//					case composerFocusSentenceStructure:
//						return p.layoutSentenceStructurePanel(gtx, false)
//					default:
//						return p.layoutFlashcardComposerForm(gtx)
//					}
//				}),
//			)
//		})
//	})
//}
//
//func (p *Page) layoutFlashcardComposerDocked(gtx layout.Context) layout.Dimensions {
//	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutComposerHeader(gtx, nil)
//				}),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					gtx.Constraints.Min = gtx.Constraints.Max
//					switch p.composerFocus {
//					case composerFocusSentenceStructure:
//						return p.layoutSentenceStructurePanel(gtx, true)
//					default:
//						return p.layoutFlashcardComposerForm(gtx)
//					}
//				}),
//			)
//		})
//	})
//}
//
//func (p *Page) layoutComposerHeader(gtx layout.Context, action *bareui.Button) layout.Dimensions {
//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//			return layout.Dimensions{}
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return p.layoutComposerFocusTabs(gtx)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if action == nil {
//				return layout.Dimensions{}
//			}
//			return layout.Inset{Left: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return action.Layout(gtx, p.theme, p.iconify)
//			})
//		}),
//	)
//}
//
//func (p *Page) layoutFlashcardComposerForm(gtx layout.Context) layout.Dimensions {
//	word := material.Editor(p.theme.Gio(), &p.wordEditor, "Word or phrase")
//	word.Color = p.theme.Color.Text
//	word.HintColor = p.theme.Color.TextMuted
//	meaning := material.Editor(p.theme.Gio(), &p.meaningEditor, "Meaning")
//	meaning.Color = p.theme.Color.Text
//	meaning.HintColor = p.theme.Color.TextMuted
//	//hideReadingCheck := material.CheckBox(p.theme.Gio(), &p.hideReadingInAnki, "Hide reading/furigana in Anki for this card")
//	//hideReadingCheck.Color = p.theme.Color.Text
//
//	searchButton := bareui.Button{Clickable: &p.searchWordButton, Text: "Lookup", Prefix: "mdi:book-search-outline", Variant: bareui.ButtonSecondary}
//	playButton := bareui.Button{Clickable: &p.playAudioButton, Text: "mdi:play-circle-outline", Icon: true, Prefix: "mdi:play-circle-outline", Variant: bareui.ButtonSecondary}
//	addAllButton := bareui.Button{Clickable: &p.addAllLookupButton, Text: "Add All Matches", Prefix: "mdi:playlist-plus", Variant: bareui.ButtonSecondary}
//
//	selected := p.selectedTranscriptText()
//	if selected == "" {
//		selected = "Select focused sentence or transcript text to prefill the flashcard word."
//	}
//
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.H6(p.theme.Gio(), "New Flashcard")
//			lbl.Color = p.theme.Color.Text
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body1(p.theme.Gio(), selected)
//			lbl.Color = p.theme.Color.TextMuted
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//		layout.Rigid(word.Layout),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			minHeight := unit.Dp(120)
//			if p.isCompactLayout(gtx) {
//				minHeight = unit.Dp(102)
//			}
//			gtx.Constraints.Min.Y = gtx.Dp(minHeight)
//			return meaning.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return searchButton.Layout(gtx, p.theme, p.iconify)
//				}),
//				layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return playButton.Layout(gtx, p.theme, p.iconify)
//				}),
//			)
//		}),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if len(p.lookupResults) <= 1 {
//				return layout.Dimensions{}
//			}
//			return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return addAllButton.Layout(gtx, p.theme, p.iconify)
//			})
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			if len(p.lookupResults) == 0 {
//				return layout.Dimensions{}
//			}
//			return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return p.layoutLookupResults(gtx)
//			})
//		}),
//	)
//}
//
//func (p *Page) layoutSentenceStructurePanel(gtx layout.Context, fillHeight bool) layout.Dimensions {
//	analysis, errText := p.currentStructureAnalysis()
//	if fillHeight {
//		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
//	} else if p.isCompactLayout(gtx) {
//		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(unit.Dp(320)))
//	} else {
//		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(unit.Dp(380)))
//	}
//	if gtx.Constraints.Max.Y <= 0 {
//		gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(260))
//	}
//
//	items := 1 + len(analysis.Tokens)
//	if len(analysis.Particles) > 0 {
//		items++
//	}
//	return material.List(p.theme.Gio(), &p.structureList).Layout(gtx, items, func(gtx layout.Context, index int) layout.Dimensions {
//		switch {
//		case index == 0:
//			return p.layoutStructureSummary(gtx, analysis, errText)
//		case len(analysis.Particles) > 0 && index == 1:
//			return layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return p.layoutParticleSummary(gtx, analysis.Particles)
//			})
//		default:
//			tokenIndex := index - 1
//			if len(analysis.Particles) > 0 {
//				tokenIndex--
//			}
//			if tokenIndex < 0 || tokenIndex >= len(analysis.Tokens) {
//				return layout.Dimensions{}
//			}
//			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				return p.layoutStructureToken(gtx, analysis.Tokens[tokenIndex])
//			})
//		}
//	})
//}
//
//func (p *Page) layoutStructureSummary(gtx layout.Context, analysis japanese.Analysis, errText string) layout.Dimensions {
//	return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				lbl := material.H6(p.theme.Gio(), "Sentence Structure")
//				lbl.Color = p.theme.Color.Text
//				return lbl.Layout(gtx)
//			}),
//			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				text := strings.TrimSpace(analysis.Sentence)
//				if text == "" {
//					text = "Select transcript text, or enter a flashcard word, to inspect sentence structure."
//				}
//				if errText != "" {
//					text = errText
//				}
//				lbl := material.Body1(p.theme.Gio(), text)
//				lbl.Color = p.theme.Color.TextMuted
//				return lbl.Layout(gtx)
//			}),
//		)
//	})
//}
//
//func (p *Page) layoutParticleSummary(gtx layout.Context, particles []japanese.Token) layout.Dimensions {
//	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			children := []layout.FlexChild{
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(p.theme.Gio(), "Particles")
//					lbl.Color = p.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//			}
//			for _, particle := range particles {
//				particle := particle
//				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//						lbl := material.Body1(p.theme.Gio(), particle.Surface+" - "+particleRole(particle.Surface))
//						lbl.Color = p.theme.Color.TextMuted
//						return lbl.Layout(gtx)
//					})
//				}))
//			}
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
//		})
//	})
//}
//
//func (p *Page) layoutStructureToken(gtx layout.Context, token japanese.Token) layout.Dimensions {
//	existingCard, hasExistingCard := p.structureTokenFlashcard(token)
//	addButton := bareui.Button{
//		Clickable: p.structureTokenAddClickable(structureTokenKey(token)),
//		Text:      "mdi:plus-circle-outline",
//		Icon:      true,
//		Variant:   bareui.ButtonPrimary,
//	}
//	playButton := bareui.Button{
//		Clickable: p.structureTokenPlayClickable(structureTokenKey(token)),
//		Text:      "mdi:play-circle-outline",
//		Icon:      true,
//		Variant:   bareui.ButtonSecondary,
//	}
//	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//							lbl := material.H6(p.theme.Gio(), token.Surface)
//							lbl.Color = p.theme.Color.Text
//							return lbl.Layout(gtx)
//						}),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							lbl := material.Body1(p.theme.Gio(), posMajorLabel(token.POSMajor()))
//							lbl.Color = p.theme.Color.Primary
//							return lbl.Layout(gtx)
//						}),
//						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//							if hasExistingCard {
//								if strings.TrimSpace(existingCard.AudioPath) == "" && !p.hasTTSReference() {
//									return layout.Dimensions{}
//								}
//								return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//									return playButton.Layout(gtx, p.theme, p.iconify)
//								})
//							}
//							if !canCreateStructureFlashcard(token) {
//								return layout.Dimensions{}
//							}
//							return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//								return addButton.Layout(gtx, p.theme, p.iconify)
//							})
//						}),
//					)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(p.theme.Gio(), tokenDetailText(token))
//					lbl.Color = p.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					if !hasExistingCard || strings.TrimSpace(existingCard.Meaning) == "" {
//						return layout.Dimensions{}
//					}
//					return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//						lbl := material.Body1(p.theme.Gio(), existingCard.Meaning)
//						lbl.Color = p.theme.Color.Text
//						return lbl.Layout(gtx)
//					})
//				}),
//			)
//		})
//	})
//}
//
//func (p *Page) currentStructureAnalysis() (japanese.Analysis, string) {
//	text := p.structureSourceText()
//	if text == "" {
//		p.structureCacheKey = ""
//		p.structureCache = japanese.Analysis{}
//		p.structureCacheErr = ""
//		return japanese.Analysis{}, ""
//	}
//	if text == p.structureCacheKey {
//		return p.structureCache, p.structureCacheErr
//	}
//	analysis, err := japanese.AnalyzeSentence(text)
//	p.structureCacheKey = text
//	p.structureCache = analysis
//	p.structureCacheErr = ""
//	if err != nil {
//		p.structureCache = japanese.Analysis{Sentence: text}
//		p.structureCacheErr = err.Error()
//	}
//	return p.structureCache, p.structureCacheErr
//}
//
//func (p *Page) addStructureTokenFlashcard(key string) {
//	if strings.TrimSpace(p.activeGameName) == "" {
//		p.showError("Create Flashcard Failed", "Select a game before creating flashcards.")
//		return
//	}
//	analysis, errText := p.currentStructureAnalysis()
//	if errText != "" {
//		p.showError("Create Flashcard Failed", errText)
//		return
//	}
//	for _, token := range analysis.Tokens {
//		if structureTokenKey(token) != key {
//			continue
//		}
//		word := structureFlashcardWord(token)
//		if word == "" {
//			p.showError("Create Flashcard Failed", "This structure component cannot be turned into a flashcard.")
//			return
//		}
//		lookups, err := dictionary.LookupWords(word)
//		if err != nil {
//			p.showError("Create Flashcard Failed", err.Error())
//			return
//		}
//		if len(lookups) == 0 {
//			p.showError("Create Flashcard Failed", "No dictionary matches were found for "+word+".")
//			return
//		}
//		card := p.flashcardFromLookup(lookups[0])
//		card.SourceLine = analysis.Sentence
//		if err := flashcards.AddFlashcard(card); err != nil {
//			p.showError("Create Flashcard Failed", err.Error())
//			return
//		}
//		_ = p.ReloadFlashcards()
//		p.showNotification("Flashcard Created", word+" was added from sentence structure.", guitoast.NotificationTypeSuccess)
//		return
//	}
//}
//
//func (p *Page) structureTokenAddClickable(key string) *widget.Clickable {
//	if p.structureTokenAddClicks == nil {
//		p.structureTokenAddClicks = make(map[string]*widget.Clickable)
//	}
//	if p.structureTokenAddClicks[key] == nil {
//		p.structureTokenAddClicks[key] = new(widget.Clickable)
//	}
//	return p.structureTokenAddClicks[key]
//}
//
//func (p *Page) playStructureTokenAudio(ctx context.Context, w *app.Window, key string) {
//	tokenCard, ok := p.structureTokenFlashcardByKey(key)
//	if !ok {
//		return
//	}
//	if strings.TrimSpace(tokenCard.AudioPath) == "" {
//		word := util.FirstNonEmpty(tokenCard.Text, tokenCard.Reading)
//		p.playTTSForText(ctx, w, word)
//		return
//	}
//	p.startAudioPlayback(w, "Audio Playback Failed", func() error {
//		return p.playFlashcardAudio(ctx, tokenCard)
//	})
//}
//
//func (p *Page) structureTokenPlayClickable(key string) *widget.Clickable {
//	if p.structureTokenPlayClicks == nil {
//		p.structureTokenPlayClicks = make(map[string]*widget.Clickable)
//	}
//	if p.structureTokenPlayClicks[key] == nil {
//		p.structureTokenPlayClicks[key] = new(widget.Clickable)
//	}
//	return p.structureTokenPlayClicks[key]
//}
//
//func (p *Page) focusedTokenClickable(key string) *widget.Clickable {
//	if p.focusedTokenClicks == nil {
//		p.focusedTokenClicks = make(map[string]*widget.Clickable)
//	}
//	if p.focusedTokenClicks[key] == nil {
//		p.focusedTokenClicks[key] = new(widget.Clickable)
//	}
//	return p.focusedTokenClicks[key]
//}
//
//func (p *Page) pruneFocusedTokenClicks(tokens []japanese.Token) {
//	valid := make(map[string]struct{}, len(tokens))
//	for _, token := range tokens {
//		valid[structureTokenKey(token)] = struct{}{}
//	}
//	for key := range p.focusedTokenClicks {
//		if _, ok := valid[key]; !ok {
//			delete(p.focusedTokenClicks, key)
//		}
//	}
//	if p.selectedFocusedTokenKey != "" {
//		if _, ok := valid[p.selectedFocusedTokenKey]; !ok {
//			p.selectedFocusedTokenKey = ""
//			p.selectedFocusedTokenWord = ""
//			p.selectedFocusedTokenNote = ""
//			p.focusedLookupPendingKey = ""
//		}
//	}
//}
//
//func (p *Page) selectFocusedToken(key string, w *app.Window) {
//	analysis, errText := p.currentStructureAnalysis()
//	if errText != "" {
//		p.showError("Dictionary Lookup Failed", errText)
//		return
//	}
//	for _, token := range analysis.Tokens {
//		if structureTokenKey(token) != key {
//			continue
//		}
//		word := structureFlashcardWord(token)
//		if word == "" {
//			word = strings.TrimSpace(token.Surface)
//		}
//		p.selectedFocusedTokenKey = key
//		p.selectedFocusedTokenWord = word
//		p.selectedFocusedTokenNote = ""
//		p.focusedLookupPendingKey = ""
//		p.wordEditor.SetText(word)
//		p.meaningEditor.SetText("")
//		p.lookupResult = nil
//		p.lookupResults = nil
//		if isParticleToken(token) {
//			note := particleRole(token.Surface)
//			p.selectedFocusedTokenNote = note
//			p.meaningEditor.SetText(note)
//			return
//		}
//		p.startFocusedTokenLookup(key, word, w)
//		return
//	}
//}
//
//func (p *Page) startFocusedTokenLookup(key, word string, w *app.Window) {
//	key = strings.TrimSpace(key)
//	word = strings.TrimSpace(word)
//	if key == "" || word == "" {
//		return
//	}
//	p.focusedLookupPendingKey = key
//	p.meaningEditor.SetText("Looking up...")
//	go func() {
//		lookups, err := dictionary.LookupWords(word)
//		result := focusedTokenLookupResult{
//			Key:     key,
//			Word:    word,
//			Lookups: lookups,
//			Err:     err,
//		}
//		select {
//		case p.focusedLookupResultCh <- result:
//		default:
//			slog.Warn("focused token lookup result dropped", "key", key)
//		}
//		if w != nil {
//			w.Invalidate()
//		}
//	}()
//}
//
//func (p *Page) addFocusedTokenFlashcard() {
//	if p.lookupResult == nil {
//		p.showError("Create Flashcard Failed", "Click a word block before adding a flashcard.")
//		return
//	}
//	if _, ok := p.focusedSelectedTokenFlashcard(p.selectedFocusedTokenWord); ok {
//		p.showNotification("Flashcard Exists", p.selectedFocusedTokenWord+" is already in your flashcards.", guitoast.NotificationTypeInfo)
//		return
//	}
//	card := p.flashcardFromLookup(*p.lookupResult)
//	if err := flashcards.AddFlashcard(card); err != nil {
//		p.showError("Create Flashcard Failed", err.Error())
//		return
//	}
//	_ = p.ReloadFlashcards()
//	p.showNotification("Flashcard Created", card.Text+" was added.", guitoast.NotificationTypeSuccess)
//}
//
//func (p *Page) playFocusedTokenAudio(ctx context.Context, w *app.Window) {
//	if p.lookupResult != nil && strings.TrimSpace(p.lookupResult.AudioPath) != "" {
//		lookup := *p.lookupResult
//		p.startAudioPlayback(w, "Audio Playback Failed", func() error {
//			return dictionary.PlayLookupAudio(lookup)
//		})
//		return
//	}
//	card, ok := p.focusedSelectedTokenFlashcard(p.selectedFocusedTokenWord)
//	if !ok || strings.TrimSpace(card.AudioPath) == "" {
//		p.playTTSForText(ctx, w, p.selectedFocusedTokenWord)
//		return
//	}
//	p.startAudioPlayback(w, "Audio Playback Failed", func() error {
//		return p.playFlashcardAudio(ctx, card)
//	})
//}
//
//func (p *Page) focusedSelectedTokenFlashcard(word string) (flashcards.Flashcard, bool) {
//	if p.selectedFocusedTokenKey != "" {
//		if card, ok := p.structureTokenFlashcardByKey(p.selectedFocusedTokenKey); ok {
//			return card, true
//		}
//	}
//	return p.flashcardForWordExact(word)
//}
//
//func (p *Page) structureTokenFlashcardByKey(key string) (flashcards.Flashcard, bool) {
//	analysis, _ := p.currentStructureAnalysis()
//	for _, token := range analysis.Tokens {
//		if structureTokenKey(token) != key {
//			continue
//		}
//		return p.structureTokenFlashcard(token)
//	}
//	return flashcards.Flashcard{}, false
//}
//
//func (p *Page) structureTokenFlashcard(token japanese.Token) (flashcards.Flashcard, bool) {
//	candidates := structureTokenFlashcardCandidates(token)
//	if len(candidates) == 0 {
//		return flashcards.Flashcard{}, false
//	}
//	for _, card := range p.flashcards {
//		cardWords := []string{card.Text, card.Reading, card.PronunciationText}
//		for _, cardWord := range cardWords {
//			cardWord = normalizeStructureMatchText(cardWord)
//			if cardWord == "" {
//				continue
//			}
//			for _, candidate := range candidates {
//				if cardWord == candidate {
//					return card, true
//				}
//			}
//		}
//	}
//	return flashcards.Flashcard{}, false
//}
//
//func (p *Page) structureSourceText() string {
//	selected := p.selectedTranscriptText()
//	if selected != "" {
//		if sentence := japanese.ExtractSenNtence(p.displayTranscript, selected); sentence != "" {
//			return cleanTranscriptFocusText(sentence)
//		}
//		return cleanTranscriptFocusText(selected)
//	}
//	if p.selectedLineKey != "" {
//		if selected := p.transcriptFocusTextForKey(p.selectedLineKey); selected != "" {
//			return selected
//		}
//	}
//	word := normalizeSelectionText(p.wordEditor.Text())
//	if word != "" {
//		if sentence := findFlashcardSourceLine(p.displayTranscript, word); sentence != "" {
//			return cleanTranscriptFocusText(sentence)
//		}
//		return cleanInlineText(word)
//	}
//	if latest := p.transcriptFocusTextForKey(""); latest != "" {
//		return latest
//	}
//	return ""
//}
//
//func (p *Page) translationCacheKey() string {
//	source := p.structureSourceText()
//	if strings.TrimSpace(source) == "" || strings.TrimSpace(p.selectedTargetLanguage) == "" {
//		return ""
//	}
//	return strings.TrimSpace(p.activeGameName) + "\x00" + cleanInlineText(source) + "\x00" + strings.ToLower(strings.TrimSpace(p.selectedTargetLanguage))
//}
//
//func (p *Page) syncTranslationEditor() {
//	key := p.translationCacheKey()
//	if key == p.translationLoadedKey {
//		return
//	}
//	p.translationLoadedKey = key
//	if key == "" {
//		p.translationEditor.SetText("")
//		return
//	}
//	entry, ok, err := translation.Load(p.activeGameName, p.structureSourceText(), p.selectedTargetLanguage)
//	if err != nil {
//		p.showError("Translation Cache Failed", err.Error())
//		return
//	}
//	if !ok {
//		p.translationEditor.SetText("")
//		return
//	}
//	p.translationEditor.SetText(entry.Translation)
//}
//
//func (p *Page) maybeAutoGenerateTranslation(ctx context.Context, w *app.Window) {
//	if !p.autoTranslateMissing.Value || p.translationGeneratingKey != "" {
//		return
//	}
//	key := p.translationCacheKey()
//	if key == "" || key == p.autoTranslationAttemptKey {
//		return
//	}
//	entry, ok, err := translation.Load(p.activeGameName, p.structureSourceText(), p.selectedTargetLanguage)
//	if err != nil {
//		p.autoTranslationAttemptKey = key
//		p.showError("Translation Cache Failed", err.Error())
//		return
//	}
//	if ok {
//		p.translationLoadedKey = key
//		p.translationEditor.SetText(entry.Translation)
//		return
//	}
//	p.autoTranslationAttemptKey = key
//	p.generateCurrentTranslation(ctx, w)
//}
//
//func (p *Page) saveCurrentTranslation() {
//	source := p.structureSourceText()
//	if strings.TrimSpace(source) == "" {
//		p.showError("Save Translation Failed", "There is no focused sentence to save.")
//		return
//	}
//	entry := translation.Entry{
//		GameName:       p.activeGameName,
//		SourceText:     source,
//		TargetLanguage: p.selectedTargetLanguage,
//		Translation:    p.translationEditor.Text(),
//	}
//	if err := translation.Save(entry); err != nil {
//		p.showError("Save Translation Failed", err.Error())
//		return
//	}
//	p.translationLoadedKey = p.translationCacheKey()
//	p.autoTranslationAttemptKey = p.translationLoadedKey
//	p.showNotification("Translation Saved", "Saved translation for "+p.selectedTargetLanguage+".", guitoast.NotificationTypeSuccess)
//}
//
//func (p *Page) generateCurrentTranslation(ctx context.Context, w *app.Window) {
//	source := p.structureSourceText()
//	if strings.TrimSpace(source) == "" {
//		p.showError("Generate Translation Failed", "There is no focused sentence to translate.")
//		return
//	}
//	key := p.translationCacheKey()
//	if key == "" || p.translationGeneratingKey != "" {
//		return
//	}
//	p.translationGeneratingKey = key
//	p.translationEditor.SetText("Generating translation...")
//
//	gameName := p.activeGameName
//	targetLanguage := p.selectedTargetLanguage
//	cfg := p.translatorConfig
//	go func() {
//		entry, err := translation.Generate(ctx, cfg, gameName, source, targetLanguage)
//		result := translationResult{Key: key, Entry: entry, Err: err}
//		select {
//		case p.translationResultCh <- result:
//		case <-ctx.Done():
//		}
//		if w != nil {
//			w.Invalidate()
//		}
//	}()
//}
//
//func (p *Page) drainTranslationResults() {
//	for {
//		select {
//		case result := <-p.translationResultCh:
//			if result.Key != p.translationGeneratingKey {
//				continue
//			}
//			p.translationGeneratingKey = ""
//			if result.Key != p.translationCacheKey() {
//				continue
//			}
//			if result.Err != nil {
//				p.translationEditor.SetText("")
//				p.showError("Generate Translation Failed", result.Err.Error())
//				continue
//			}
//			p.translationLoadedKey = result.Key
//			p.autoTranslationAttemptKey = result.Key
//			p.translationEditor.SetText(result.Entry.Translation)
//			p.showNotification("Translation Generated", "Generated and cached translation for "+result.Entry.TargetLanguage+".", guitoast.NotificationTypeSuccess)
//		default:
//			return
//		}
//	}
//}
//
//func (p *Page) toggleTranscriptRowTranslation(ctx context.Context, w *app.Window, rowKey string) {
//	row, ok := p.transcriptRowByKey(rowKey)
//	if !ok || row.Info {
//		return
//	}
//	if p.rowTranslationShown[row.Key] {
//		p.rowTranslationShown[row.Key] = false
//		return
//	}
//	key := p.rowTranslationCacheKey(row)
//	if key == "" {
//		p.showError("Translate Row Failed", "Select a target language before translating a row.")
//		return
//	}
//	if _, ok := p.rowTranslations[key]; ok {
//		p.rowTranslationShown[row.Key] = true
//		return
//	}
//	entry, ok, err := translation.Load(p.activeGameName, row.Text, p.selectedTargetLanguage)
//	if err != nil {
//		p.showError("Translate Row Failed", err.Error())
//		return
//	}
//	if ok {
//		p.rowTranslations[key] = entry.Translation
//		p.rowTranslationShown[row.Key] = true
//		return
//	}
//	p.rowTranslationShown[row.Key] = true
//	p.generateTranscriptRowTranslation(ctx, w, row, key)
//}
//
//func (p *Page) generateTranscriptRowTranslation(ctx context.Context, w *app.Window, row transcriptRow, key string) {
//	if key == "" || p.rowTranslationGenerating[key] {
//		return
//	}
//	source := cleanInlineText(row.Text)
//	if source == "" {
//		return
//	}
//	p.rowTranslationGenerating[key] = true
//	gameName := p.activeGameName
//	targetLanguage := p.selectedTargetLanguage
//	cfg := p.translatorConfig
//	go func() {
//		entry, err := translation.Generate(ctx, cfg, gameName, source, targetLanguage)
//		result := rowTranslationResult{Key: key, RowKey: row.Key, Entry: entry, Err: err}
//		select {
//		case p.rowTranslationResultCh <- result:
//		case <-ctx.Done():
//		}
//		if w != nil {
//			w.Invalidate()
//		}
//	}()
//}
//
//func (p *Page) drainRowTranslationResults() {
//	for {
//		select {
//		case result := <-p.rowTranslationResultCh:
//			delete(p.rowTranslationGenerating, result.Key)
//			if result.Err != nil {
//				p.rowTranslationShown[result.RowKey] = false
//				p.showError("Translate Row Failed", result.Err.Error())
//				continue
//			}
//			p.rowTranslations[result.Key] = result.Entry.Translation
//			p.rowTranslationShown[result.RowKey] = true
//			p.showNotification("Translation Generated", "Generated row translation for "+result.Entry.TargetLanguage+".", guitoast.NotificationTypeSuccess)
//		default:
//			return
//		}
//	}
//}
//
//func (p *Page) transcriptRowByKey(key string) (transcriptRow, bool) {
//	for _, row := range p.transcriptRows() {
//		if row.Key == key {
//			return row, true
//		}
//	}
//	return transcriptRow{}, false
//}
//
//func (p *Page) transcriptRowDisplayText(row transcriptRow) string {
//	if !p.isTranscriptRowTranslationShown(row) {
//		return row.Text
//	}
//	key := p.rowTranslationCacheKey(row)
//	if p.rowTranslationGenerating[key] {
//		return "Translating..."
//	}
//	if text := strings.TrimSpace(p.rowTranslations[key]); text != "" {
//		return text
//	}
//	return row.Text
//}
//
//func (p *Page) isTranscriptRowTranslationShown(row transcriptRow) bool {
//	return p.rowTranslationShown[row.Key]
//}
//
//func (p *Page) rowTranslationCacheKey(row transcriptRow) string {
//	source := cleanInlineText(row.Text)
//	targetLanguage := strings.TrimSpace(p.selectedTargetLanguage)
//	if source == "" || targetLanguage == "" {
//		return ""
//	}
//	return strings.TrimSpace(p.activeGameName) + "\x00" + source + "\x00" + strings.ToLower(targetLanguage)
//}
//
//func (p *Page) transcriptFocusTextForKey(key string) string {
//	rows := p.transcriptRows()
//	if len(rows) == 0 {
//		return ""
//	}
//	if strings.TrimSpace(key) != "" {
//		for _, row := range rows {
//			if row.Key == key {
//				if row.Info {
//					return ""
//				}
//				return cleanInlineText(row.Text)
//			}
//		}
//		return ""
//	}
//	for i := len(rows) - 1; i >= 0; i-- {
//		if !rows[i].Info {
//			return cleanInlineText(rows[i].Text)
//		}
//	}
//	return ""
//}
//
//func (p *Page) transcriptRows() []transcriptRow {
//	lines := strings.Split(strings.TrimSpace(p.displayTranscript), "\n")
//	rows := make([]transcriptRow, 0, len(lines))
//	var previousTimestamp string = unknownTimestamp
//	for i, line := range lines {
//		text := cleanInlineText(line)
//		if text == "" {
//			continue
//		}
//		timestamp, body, speaker, voice, info := splitTranscriptRow(text)
//		if strings.HasPrefix(timestamp, "--") {
//			timestamp = previousTimestamp
//		} else {
//			previousTimestamp = timestamp
//		}
//		speaker = cleanInlineText(speaker)
//		voice = strings.TrimSpace(voice)
//		body = cleanInlineText(body)
//		if body == "" {
//			continue
//		}
//		if p.speakerOnlyRows && (info || strings.TrimSpace(speaker) == "") {
//			continue
//		}
//		key := fmt.Sprintf("%d:%s", i, text)
//		if !info && len(rows) > 0 && !rows[len(rows)-1].Info && timestamp != unknownTimestamp && rows[len(rows)-1].Time == timestamp && rows[len(rows)-1].Speaker == speaker && rows[len(rows)-1].Voice == voice {
//			rows[len(rows)-1].Text = strings.TrimSpace(rows[len(rows)-1].Text + "\n" + body)
//			rows[len(rows)-1].VocabWords = p.vocabWordsInText(rows[len(rows)-1].Text)
//			continue
//		}
//		row := transcriptRow{Key: key, Time: timestamp, Speaker: speaker, Voice: voice, Text: body, Info: info}
//		if !info {
//			row.VocabWords = p.vocabWordsInText(body)
//		}
//		rows = append(rows, row)
//	}
//	p.pruneTranscriptRowClicks(rows)
//	return rows
//}
//
//func (p *Page) ttsReferences() []ttsReference {
//	rows := p.transcriptRows()
//	refsBySpeaker := map[string]ttsReference{}
//	order := make([]string, 0)
//	for _, row := range rows {
//		speaker := strings.TrimSpace(row.Speaker)
//		if speaker == "" || strings.TrimSpace(row.Voice) == "" || strings.TrimSpace(row.Text) == "" {
//			continue
//		}
//		if _, ok := refsBySpeaker[speaker]; !ok {
//			order = append(order, speaker)
//		}
//		refsBySpeaker[speaker] = ttsReference{
//			Speaker: speaker,
//			Voice:   row.Voice,
//			Text:    row.Text,
//		}
//	}
//	sort.Strings(order)
//	refs := make([]ttsReference, 0, len(order))
//	for _, speaker := range order {
//		refs = append(refs, refsBySpeaker[speaker])
//	}
//	return refs
//}
//
//func (p *Page) selectedTTSReference() (ttsReference, bool) {
//	refs := p.ttsReferences()
//	if len(refs) == 0 {
//		return ttsReference{}, false
//	}
//	selected := strings.TrimSpace(p.selectedTTSSpeaker)
//	for _, ref := range refs {
//		if ref.Speaker == selected {
//			return ref, true
//		}
//	}
//	return refs[0], true
//}
//
//func (p *Page) hasTTSReference() bool {
//	_, ok := p.selectedTTSReference()
//	return ok
//}
//
//func ttsReferenceExists(refs []ttsReference, speaker string) bool {
//	for _, ref := range refs {
//		if ref.Speaker == speaker {
//			return true
//		}
//	}
//	return false
//}
//
//func (p *Page) vocabWordsInText(text string) []string {
//	text = cleanInlineText(text)
//	if text == "" || len(p.flashcards) == 0 {
//		return nil
//	}
//	seen := make(map[string]struct{}, len(p.flashcards))
//	words := make([]string, 0, len(p.flashcards))
//	for _, card := range p.flashcards {
//		word := cleanInlineText(card.Text)
//		if word == "" {
//			continue
//		}
//		if _, ok := seen[word]; ok {
//			continue
//		}
//		seen[word] = struct{}{}
//		words = append(words, word)
//	}
//	sort.SliceStable(words, func(i, j int) bool {
//		return len([]rune(words[i])) > len([]rune(words[j]))
//	})
//	matches := flashcards.FindMatches(text, words)
//	if len(matches) == 0 {
//		return nil
//	}
//	out := make([]string, 0, len(matches))
//	for _, match := range matches {
//		if _, ok := seen[match.Word]; !ok {
//			continue
//		}
//		out = append(out, match.Word)
//		delete(seen, match.Word)
//	}
//	return out
//}
//
//func (p *Page) currentTranscriptRowKey() string {
//	if p.selectedLineKey != "" {
//		return p.selectedLineKey
//	}
//	rows := p.transcriptRows()
//	for i := len(rows) - 1; i >= 0; i-- {
//		if !rows[i].Info {
//			return rows[i].Key
//		}
//	}
//	return ""
//}
//
//func (p *Page) selectTranscriptRow(key string) {
//	for _, row := range p.transcriptRows() {
//		if row.Key != key {
//			continue
//		}
//		if row.Info {
//			return
//		}
//		p.selectedLineKey = row.Key
//		p.selectedLineText = row.Text
//		p.wordEditor.SetText("")
//		p.meaningEditor.SetText("")
//		p.lookupResult = nil
//		p.lookupResults = nil
//		p.focusedLookupPendingKey = ""
//		p.DismissPopup()
//		return
//	}
//}
//
//func (p *Page) selectLatestTranscriptRow() {
//	rows := p.transcriptRows()
//	for i := len(rows) - 1; i >= 0; i-- {
//		row := rows[i]
//		if row.Info {
//			continue
//		}
//		if row.Key == p.selectedLineKey && row.Text == p.selectedLineText {
//			return
//		}
//		p.selectedLineKey = row.Key
//		p.selectedLineText = row.Text
//		p.selectedFocusedTokenKey = ""
//		p.selectedFocusedTokenWord = ""
//		p.selectedFocusedTokenNote = ""
//		p.focusedLookupPendingKey = ""
//		p.wordEditor.SetText("")
//		p.meaningEditor.SetText("")
//		p.lookupResult = nil
//		p.lookupResults = nil
//		p.focusedLookupPendingKey = ""
//		p.DismissPopup()
//		return
//	}
//	p.selectedLineKey = ""
//	p.selectedLineText = ""
//}
//
//func (p *Page) playTranscriptRowVoice(ctx context.Context, w *app.Window, key string) {
//	for _, row := range p.transcriptRows() {
//		if row.Key != key {
//			continue
//		}
//		if strings.TrimSpace(row.Voice) == "" {
//			p.playTTSForText(ctx, w, row.Text)
//			return
//		}
//		if p.currentConfig == nil {
//			p.showError("Voice Playback Failed", "Game config is not loaded.")
//			return
//		}
//		voice := row.Voice
//		cfg := *p.currentConfig
//		p.startAudioPlayback(w, "Voice Playback Failed", func() error {
//			path, err := cachedTranscriptVoicePathForConfig(&cfg, voice)
//			if err != nil {
//				return err
//			}
//			player, err := audioplayer.NewPlayer(audioplayer.Config{Backend: audioplayer.BackendAuto})
//			if err != nil {
//				return err
//			}
//			return audioplayer.PlayAudioFile(player, path, true)
//		})
//		return
//	}
//}
//
//func (p *Page) cachedTranscriptVoicePath(voice string) (string, error) {
//	return cachedTranscriptVoicePathForConfig(p.currentConfig, voice)
//}
//
//func cachedTranscriptVoicePathForConfig(cfg *vngame.Game, voice string) (string, error) {
//	voice = strings.TrimSpace(voice)
//	if voice == "" {
//		return "", fmt.Errorf("voice file is empty")
//	}
//	if cfg == nil {
//		return "", fmt.Errorf("game config is not loaded")
//	}
//	initialExt := filepath.Ext(voice)
//	cachePath, err := util.VoiceCachePath(cfg.Name, voice, initialExt)
//	if err != nil {
//		return "", err
//	}
//	if initialExt != "" && util.IsExistingFile(cachePath) {
//		return cachePath, nil
//	}
//	inputPath := util.FirstNonEmpty(cfg.GamePath, cfg.Executable, cfg.WorkingDir)
//	eng, err := auto.SelectEngine(inputPath)
//	if err != nil {
//		return "", err
//	}
//	fileInfo, err := eng.GetFile(cfg, voice)
//	if err != nil {
//		return "", err
//	}
//	if fileInfo == nil {
//		return "", fmt.Errorf("voice file %q was not returned", voice)
//	}
//	cachePath, err = util.VoiceCachePath(cfg.Name, voice, engineFileCacheExt(fileInfo))
//	if err != nil {
//		return "", err
//	}
//	slog.Info("cache path", "path", cachePath)
//	if util.IsExistingFile(cachePath) {
//		return cachePath, nil
//	}
//	data := fileInfo.Data
//	if len(data) == 0 {
//		return "", fmt.Errorf("voice file %q is empty", voice)
//	}
//	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
//		return "", fmt.Errorf("create voice cache dir: %w", err)
//	}
//	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
//		return "", fmt.Errorf("write voice cache: %w", err)
//	}
//	return cachePath, nil
//}
//
//func (p *Page) transcriptRowClickable(key string) *widget.Clickable {
//	if p.transcriptRowClicks == nil {
//		p.transcriptRowClicks = make(map[string]*widget.Clickable)
//	}
//	if p.transcriptRowClicks[key] == nil {
//		p.transcriptRowClicks[key] = new(widget.Clickable)
//	}
//	return p.transcriptRowClicks[key]
//}
//
//func (p *Page) transcriptRowTranslateClickable(key string) *widget.Clickable {
//	if p.transcriptRowTranslateClicks == nil {
//		p.transcriptRowTranslateClicks = make(map[string]*widget.Clickable)
//	}
//	if p.transcriptRowTranslateClicks[key] == nil {
//		p.transcriptRowTranslateClicks[key] = new(widget.Clickable)
//	}
//	return p.transcriptRowTranslateClicks[key]
//}
//
//func (p *Page) transcriptRowVoiceClickable(key string) *widget.Clickable {
//	if p.transcriptRowVoiceClicks == nil {
//		p.transcriptRowVoiceClicks = make(map[string]*widget.Clickable)
//	}
//	if p.transcriptRowVoiceClicks[key] == nil {
//		p.transcriptRowVoiceClicks[key] = new(widget.Clickable)
//	}
//	return p.transcriptRowVoiceClicks[key]
//}
//
//func (p *Page) ttsSpeakerClickable(speaker string) *widget.Clickable {
//	if p.ttsSpeakerClicks == nil {
//		p.ttsSpeakerClicks = make(map[string]*widget.Clickable)
//	}
//	if p.ttsSpeakerClicks[speaker] == nil {
//		p.ttsSpeakerClicks[speaker] = new(widget.Clickable)
//	}
//	return p.ttsSpeakerClicks[speaker]
//}
//
//func (p *Page) pruneTTSSpeakerClicks(refs []ttsReference) {
//	valid := make(map[string]struct{}, len(refs))
//	for _, ref := range refs {
//		valid[ref.Speaker] = struct{}{}
//	}
//	for speaker := range p.ttsSpeakerClicks {
//		if _, ok := valid[speaker]; !ok {
//			delete(p.ttsSpeakerClicks, speaker)
//		}
//	}
//}
//
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
//
//const unknownTimestamp = "----:-- --:--:--"
//
//func splitTranscriptTimestamp(line string) (string, string) {
//	timestamp, body, _, _, _ := splitTranscriptRow(line)
//	return timestamp, body
//}
//
//func splitTranscriptRow(line string) (string, string, string, string, bool) {
//	line = strings.TrimSpace(line)
//	data, err := ParseLogLine(line)
//	if err != nil {
//		return unknownTimestamp, line, "", "", false
//	}
//	if data.RawTime == "" {
//		return unknownTimestamp, line, "", "", false
//	}
//	timestamp := data.Time.Format("2006/01 15:04:05")
//	speaker := cleanInlineText(data.Speaker)
//	voice := strings.TrimSpace(data.Voice)
//	switch strings.ToLower(strings.TrimSpace(data.Speaker)) {
//	case "system":
//		return timestamp, cleanInlineText(data.Text), "", "", true
//	case "new session":
//		return timestamp, "New session", "", "", true
//	}
//	if cleanInlineText(data.Text) == "" {
//		return timestamp, "New session", "", "", true
//	}
//	return timestamp, data.Text, speaker, voice, false
//}
//
//func (p *Page) contextFlashcard() (flashcards.Flashcard, bool) {
//	if p.popupFlashcard != nil {
//		return *p.popupFlashcard, true
//	}
//	word := strings.TrimSpace(p.wordEditor.Text())
//	if word == "" {
//		word = strings.TrimSpace(p.popupWord)
//	}
//	return p.contextFlashcardForWord(word)
//}
//
//func (p *Page) contextFlashcardForWord(word string) (flashcards.Flashcard, bool) {
//	word = strings.TrimSpace(word)
//	if word != "" {
//		if card, ok := p.flashcardForWordExact(word); ok {
//			return card, true
//		}
//	}
//	if len(p.flashcards) > 0 {
//		return p.flashcards[0], true
//	}
//	return flashcards.Flashcard{}, false
//}
//
//func (p *Page) flashcardForWordExact(word string) (flashcards.Flashcard, bool) {
//	word = normalizeStructureMatchText(word)
//	if word == "" {
//		return flashcards.Flashcard{}, false
//	}
//	for _, card := range p.flashcards {
//		cardWords := []string{card.Text, card.Reading, card.PronunciationText}
//		for _, cardWord := range cardWords {
//			if normalizeStructureMatchText(cardWord) == word {
//				return card, true
//			}
//		}
//	}
//	return flashcards.Flashcard{}, false
//}
//
//func contextVocabPillText(hasCard bool) string {
//	if hasCard {
//		return "In Vocab"
//	}
//	return "Lookup"
//}
//
//func newTranslationLanguageOptions() []gui.DropdownOption {
//	labels := []string{
//		"English",
//		"Japanese",
//		"Spanish",
//		"French",
//		"German",
//		"Korean",
//		"Chinese",
//		"Italian",
//		"Portuguese",
//		"Russian",
//	}
//	options := make([]gui.DropdownOption, 0, len(labels))
//	for _, label := range labels {
//		options = append(options, gui.DropdownOption{
//			Label:     label,
//			Icon:      "mdi:translate",
//			Clickable: new(widget.Clickable),
//		})
//	}
//	return options
//}
//
//func tokenDetailText(token japanese.Token) string {
//	parts := []string{token.POSLabel()}
//	if base := strings.TrimSpace(token.BaseForm); base != "" && base != token.Surface {
//		parts = append(parts, "base: "+base)
//	}
//	if reading := strings.TrimSpace(token.Reading); reading != "" {
//		parts = append(parts, "reading: "+reading)
//	}
//	if inflection := token.InflectionLabel(); inflection != "" {
//		parts = append(parts, "inflection: "+inflection)
//	}
//	if token.POSMajor() == "助詞" {
//		parts = append(parts, "role: "+particleRole(token.Surface))
//	}
//	return strings.Join(parts, " | ")
//}
//
//func canCreateStructureFlashcard(token japanese.Token) bool {
//	switch token.POSMajor() {
//	case "名詞", "動詞", "形容詞":
//		return structureFlashcardWord(token) != ""
//	default:
//		return false
//	}
//}
//
//func structureFlashcardWord(token japanese.Token) string {
//	switch token.POSMajor() {
//	case "動詞", "形容詞":
//		return util.FirstNonEmpty(strings.TrimSpace(token.BaseForm), strings.TrimSpace(token.Surface))
//	default:
//		return strings.TrimSpace(token.Surface)
//	}
//}
//
//func structureTokenFlashcardCandidates(token japanese.Token) []string {
//	raw := []string{
//		token.Surface,
//		token.BaseForm,
//		token.Reading,
//		token.Pronunciation,
//		structureFlashcardWord(token),
//	}
//	seen := make(map[string]struct{}, len(raw))
//	out := make([]string, 0, len(raw))
//	for _, value := range raw {
//		value = normalizeStructureMatchText(value)
//		if value == "" {
//			continue
//		}
//		if _, ok := seen[value]; ok {
//			continue
//		}
//		seen[value] = struct{}{}
//		out = append(out, value)
//	}
//	return out
//}
//
//func focusTokens(tokens []japanese.Token, limit int) []japanese.Token {
//	if limit <= 0 {
//		return nil
//	}
//	out := make([]japanese.Token, 0, limit)
//	preferred := map[string]bool{
//		"名詞":  true,
//		"動詞":  true,
//		"形容詞": true,
//		"副詞":  true,
//	}
//	for _, token := range tokens {
//		if !preferred[token.POSMajor()] || structureFlashcardWord(token) == "" {
//			continue
//		}
//		out = append(out, token)
//		if len(out) == limit {
//			return out
//		}
//	}
//	for _, token := range tokens {
//		if structureFlashcardWord(token) == "" {
//			continue
//		}
//		out = append(out, token)
//		if len(out) == limit {
//			return out
//		}
//	}
//	return out
//}
//
//func focusedTokenReading(token japanese.Token) string {
//	if !util.ContainsKanji(token.Surface) {
//		return ""
//	}
//	reading := cleanInlineText(token.Reading)
//	if reading == "" || reading == cleanInlineText(token.Surface) {
//		return ""
//	}
//	return katakanaToHiragana(reading)
//}
//
//func focusedTokenDictionaryReady(token japanese.Token) bool {
//	return canCreateStructureFlashcard(token)
//}
//
//func isParticleToken(token japanese.Token) bool {
//	return token.POSMajor() == "助詞"
//}
//
//func normalizeFocusedFuriganaMode(mode string) string {
//	switch strings.ToLower(strings.TrimSpace(mode)) {
//	case focusedFuriganaAbove:
//		return focusedFuriganaAbove
//	case focusedFuriganaBelow:
//		return focusedFuriganaBelow
//	default:
//		return focusedFuriganaHidden
//	}
//}
//
//func focusedTokenColor(theme barethemes.Theme, token japanese.Token, selected, inFlashcards, dictionaryReady bool) color.NRGBA {
//	if selected {
//		return color.NRGBA{R: theme.Color.Primary.R, G: theme.Color.Primary.G, B: theme.Color.Primary.B, A: 88}
//	}
//	if inFlashcards {
//		return color.NRGBA{R: theme.Color.Primary.R, G: theme.Color.Primary.G, B: theme.Color.Primary.B, A: 54}
//	}
//	if dictionaryReady {
//		return color.NRGBA{R: theme.Color.SurfaceAlt.R, G: theme.Color.SurfaceAlt.G, B: theme.Color.SurfaceAlt.B, A: 210}
//	}
//	switch token.POSMajor() {
//	case "名詞":
//		return color.NRGBA{R: theme.Color.Secondary.R, G: theme.Color.Secondary.G, B: theme.Color.Secondary.B, A: 44}
//	case "動詞":
//		return color.NRGBA{R: theme.Color.Primary.R, G: theme.Color.Primary.G, B: theme.Color.Primary.B, A: 42}
//	case "形容詞", "副詞":
//		return color.NRGBA{R: theme.Color.Warning.R, G: theme.Color.Warning.G, B: theme.Color.Warning.B, A: 42}
//	case "助詞", "助動詞":
//		return color.NRGBA{R: theme.Color.Tertiary.R, G: theme.Color.Tertiary.G, B: theme.Color.Tertiary.B, A: 32}
//	default:
//		return color.NRGBA{R: theme.Color.SurfaceAlt.R, G: theme.Color.SurfaceAlt.G, B: theme.Color.SurfaceAlt.B, A: 180}
//	}
//}
//
//func katakanaToHiragana(text string) string {
//	var b strings.Builder
//	for _, r := range text {
//		if r >= 'ァ' && r <= 'ヶ' {
//			r -= 0x60
//		}
//		b.WriteRune(r)
//	}
//	return b.String()
//}
//
//func normalizeStructureMatchText(value string) string {
//	return strings.TrimSpace(value)
//}
//
//func structureTokenKey(token japanese.Token) string {
//	return strings.Join([]string{
//		strings.TrimSpace(token.Surface),
//		strings.TrimSpace(token.BaseForm),
//		token.POSLabel(),
//		token.InflectionLabel(),
//	}, "\x00")
//}
//
//func posMajorLabel(pos string) string {
//	switch pos {
//	case "名詞":
//		return "Noun"
//	case "動詞":
//		return "Verb"
//	case "形容詞":
//		return "Adjective"
//	case "副詞":
//		return "Adverb"
//	case "助詞":
//		return "Particle"
//	case "助動詞":
//		return "Auxiliary"
//	case "連体詞":
//		return "Prenoun"
//	case "接続詞":
//		return "Conjunction"
//	case "感動詞":
//		return "Interjection"
//	case "記号":
//		return "Symbol"
//	default:
//		if pos == "" {
//			return "Token"
//		}
//		return pos
//	}
//}
//
//func particleRole(surface string) string {
//	switch strings.TrimSpace(surface) {
//	case "は":
//		return "topic marker; sets what the sentence is about, often with contrast"
//	case "が":
//		return "subject marker; identifies the doer or thing being described"
//	case "を":
//		return "direct object marker; marks what the action affects"
//	case "に":
//		return "target, destination, time, indirect object, or location of existence"
//	case "へ":
//		return "direction marker; points toward a destination"
//	case "で":
//		return "place or means of an action; marks where/how something happens"
//	case "と":
//		return "and/with, quotation, or comparison partner"
//	case "も":
//		return "also/even; adds the marked item to the statement"
//	case "の":
//		return "possession, modification, or nominalizer"
//	case "から":
//		return "from/since; starting point or cause"
//	case "まで":
//		return "until/to; endpoint or limit"
//	case "より":
//		return "than/from; comparison baseline or source"
//	case "や":
//		return "non-exhaustive and; examples from a set"
//	case "か":
//		return "question marker or alternative"
//	case "ね":
//		return "seeks agreement or softens with shared feeling"
//	case "よ":
//		return "assertive emphasis; presents information to the listener"
//	case "な":
//		return "prohibition, emotion, or sentence-ending emphasis depending on form"
//	case "ぞ", "ぜ":
//		return "strong sentence-ending emphasis"
//	default:
//		return "particle function depends on the surrounding phrase"
//	}
//}
//
//func (p *Page) layoutComposerFocusTabs(gtx layout.Context) layout.Dimensions {
//	for p.composerFlashcardsTab.Clicked(gtx) {
//		p.composerFocus = composerFocusFlashcards
//	}
//	for p.composerSentenceTab.Clicked(gtx) {
//		p.composerFocus = composerFocusSentenceStructure
//	}
//	if p.composerFocus == "" {
//		p.composerFocus = composerFocusFlashcards
//	}
//
//	return bareutils.RoundedSurface(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.Inset{
//			Top:    unit.Dp(0),
//			Bottom: unit.Dp(0),
//			Left:   unit.Dp(0),
//			Right:  unit.Dp(0),
//		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutComposerFocusTab(gtx, &p.composerFlashcardsTab, composerFocusFlashcards, "Flashcards")
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutComposerFocusTab(gtx, &p.composerSentenceTab, composerFocusSentenceStructure, "Structure")
//				}),
//			)
//		})
//	})
//}
//
//func (p *Page) layoutComposerFocusTab(gtx layout.Context, click *widget.Clickable, id, label string) layout.Dimensions {
//	active := p.composerFocus == id
//	bg := p.theme.Color.SurfaceAlt
//	fg := p.theme.Color.TextMuted
//	if active {
//		bg = p.theme.Color.Primary
//		fg = bareutils.ReadableOn(bg)
//	} else if click.Hovered() {
//		bg = p.theme.Color.Surface
//		fg = p.theme.Color.Text
//	}
//
//	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//			return layout.Inset{
//				Top:    unit.Dp(10),
//				Bottom: unit.Dp(10),
//				Left:   unit.Dp(14),
//				Right:  unit.Dp(14),
//			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//				lbl := material.Body1(p.theme.Gio(), label)
//				lbl.Color = fg
//				return lbl.Layout(gtx)
//			})
//		})
//	})
//}
//
//func (p *Page) layoutFlashcardComposerHint(gtx layout.Context) layout.Dimensions {
//	expandButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-up", Icon: true, Prefix: "mdi:chevron-up", Variant: bareui.ButtonGhost}
//	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return p.layoutComposerHeader(gtx, &expandButton)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					switch p.composerFocus {
//					case composerFocusSentenceStructure:
//						return p.layoutSentenceStructurePanel(gtx, false)
//					default:
//						return p.layoutFlashcardComposerHintText(gtx)
//					}
//				}),
//			)
//		})
//	})
//}
//
//func (p *Page) layoutFlashcardComposerHintText(gtx layout.Context) layout.Dimensions {
//	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.H6(p.theme.Gio(), "New Flashcard")
//			lbl.Color = p.theme.Color.Text
//			return lbl.Layout(gtx)
//		}),
//		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//			lbl := material.Body1(p.theme.Gio(), "Highlight transcript text to open the flashcard editor, or click a vocab match to inspect it.")
//			lbl.Color = p.theme.Color.TextMuted
//			return lbl.Layout(gtx)
//		}),
//	)
//}
//
//func (p *Page) layoutFlashcardComposerMini(gtx layout.Context) layout.Dimensions {
//	expandButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-up", Icon: true, Prefix: "mdi:chevron-up", Variant: bareui.ButtonGhost}
//	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
//		return layout.Inset{
//			Top:    unit.Dp(10),
//			Bottom: unit.Dp(0),
//			Left:   unit.Dp(14),
//			Right:  unit.Dp(10),
//		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return p.layoutComposerHeader(gtx, &expandButton)
//		})
//	})
//}
//
//func (p *Page) layoutLookupResults(gtx layout.Context) layout.Dimensions {
//	maxHeight := gtx.Dp(unit.Dp(280))
//	if p.isCompactLayout(gtx) {
//		maxHeight = gtx.Dp(unit.Dp(240))
//	}
//	gtx.Constraints.Max.Y = maxHeight
//	if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y {
//		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
//	}
//	return material.List(p.theme.Gio(), &p.lookupResultsList).Layout(gtx, len(p.lookupResults), func(gtx layout.Context, index int) layout.Dimensions {
//		bottom := unit.Dp(0)
//		if index < len(p.lookupResults)-1 {
//			bottom = unit.Dp(10)
//		}
//		lookup := p.lookupResults[index]
//		return layout.Inset{Bottom: bottom}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return p.layoutLookupResultCard(gtx, lookup)
//		})
//	})
//}
//
//func (p *Page) layoutLookupResultCard(gtx layout.Context, lookup dictionary.Lookup) layout.Dimensions {
//	key := lookupResultKey(lookup)
//	alreadyAdded := p.lookupFlashcardExists(lookup)
//	addPending := p.lookupAddPending[key]
//	addButton := bareui.Button{Clickable: p.lookupResultAddClickable(key), Text: "mdi:plus-circle-outline", Icon: true, Variant: bareui.ButtonPrimary}
//	if alreadyAdded {
//		addButton = bareui.Button{Text: "mdi:check-circle-outline", Icon: true, Variant: bareui.ButtonSecondary}
//	} else if addPending {
//		addButton = bareui.Button{Text: "mdi:timer-sand", Icon: true, Variant: bareui.ButtonSecondary}
//	}
//	playButton := bareui.Button{Clickable: p.lookupResultPlayClickable(key), Text: "mdi:play-circle-outline", Icon: true, Variant: bareui.ButtonSecondary}
//	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.H6(p.theme.Gio(), util.FirstNonEmpty(lookup.Query, lookup.Headword))
//					lbl.Color = p.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					if strings.TrimSpace(lookup.Reading) == "" {
//						return layout.Dimensions{}
//					}
//					lbl := material.Body1(p.theme.Gio(), "Reading: "+lookup.Reading)
//					lbl.Color = p.theme.Color.TextMuted
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					lbl := material.Body1(p.theme.Gio(), lookup.Meaning)
//					lbl.Color = p.theme.Color.Text
//					return lbl.Layout(gtx)
//				}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
//				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
//						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//							if alreadyAdded || addPending {
//								return addButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//							}
//							return addButton.Layout(gtx, p.theme, p.iconify)
//						}),
//						layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
//						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//							if strings.TrimSpace(lookup.AudioPath) == "" && !p.hasTTSReference() {
//								return playButton.Layout(gtx.Disabled(), p.theme, p.iconify)
//							}
//							return playButton.Layout(gtx, p.theme, p.iconify)
//						}),
//					)
//				}),
//			)
//		})
//	})
//}
//
//func (p *Page) lookupCurrentWord() {
//	if selected := p.selectedTranscriptText(); selected != "" {
//		p.wordEditor.SetText(selected)
//	}
//	p.lookupResult = nil
//	p.lookupResults = nil
//	p.focusedLookupPendingKey = ""
//	p.meaningEditor.SetText("")
//
//	word := normalizeSelectionText(p.wordEditor.Text())
//	if word == "" {
//		p.showError("Dictionary Lookup Failed", "Flashcard word cannot be empty.")
//		return
//	}
//
//	lookups, err := dictionary.LookupWords(word)
//	if err != nil {
//		p.showError("Dictionary Lookup Failed", err.Error())
//		return
//	}
//	p.lookupResults = lookups
//	p.lookupResult = &lookups[0]
//	word = util.FirstNonEmpty(lookups[0].Query, lookups[0].Key, lookups[0].Headword)
//	p.wordEditor.SetText(word)
//	p.meaningEditor.SetText(lookups[0].Meaning)
//	p.hideReadingSet = false
//	p.lastAutoWord = ""
//	p.syncHideReadingDefault()
//}
//
//func (p *Page) playCurrentLookupAudio(ctx context.Context, w *app.Window) {
//	if p.lookupResult == nil || strings.TrimSpace(p.lookupResult.AudioPath) == "" {
//		text := ""
//		if p.lookupResult != nil {
//			text = util.FirstNonEmpty(p.lookupResult.Query, p.lookupResult.Key, p.lookupResult.Headword, p.lookupResult.Reading)
//		}
//		p.playTTSForText(ctx, w, text)
//		return
//	}
//	lookup := *p.lookupResult
//	p.startAudioPlayback(w, "Audio Playback Failed", func() error {
//		return dictionary.PlayLookupAudio(lookup)
//	})
//}
//
//func (p *Page) addLookupFlashcardByKey(key string, w *app.Window) {
//	key = strings.TrimSpace(key)
//	if key == "" {
//		return
//	}
//	if p.lookupAddPending[key] {
//		return
//	}
//	for _, lookup := range p.lookupResults {
//		if lookupResultKey(lookup) != key {
//			continue
//		}
//		card := p.flashcardFromLookup(lookup)
//		if p.lookupFlashcardExists(lookup) {
//			p.showNotification("Flashcard Exists", card.Text+" is already in your flashcards.", guitoast.NotificationTypeInfo)
//			return
//		}
//		p.startLookupFlashcardAdd(key, card, w)
//		return
//	}
//}
//
//func (p *Page) startLookupFlashcardAdd(key string, card flashcards.Flashcard, w *app.Window) {
//	if p.lookupAddPending == nil {
//		p.lookupAddPending = make(map[string]bool)
//	}
//	p.lookupAddPending[key] = true
//	go func() {
//		added, skipped, err := flashcards.AddFlashcards(card.GameName, []flashcards.Flashcard{card})
//		var cards []flashcards.Flashcard
//		if err == nil {
//			cards, err = flashcards.LoadFlashcards(card.GameName)
//		}
//		result := flashcardAddResult{
//			Key:     key,
//			Card:    card,
//			Cards:   cards,
//			Added:   added,
//			Skipped: skipped,
//			Err:     err,
//		}
//		select {
//		case p.flashcardAddResultCh <- result:
//		default:
//			slog.Warn("flashcard add result dropped", "key", key)
//		}
//		if w != nil {
//			w.Invalidate()
//		}
//	}()
//}
//
//func (p *Page) playLookupAudioByKey(ctx context.Context, w *app.Window, key string) {
//	key = strings.TrimSpace(key)
//	if key == "" {
//		return
//	}
//	for _, lookup := range p.lookupResults {
//		if lookupResultKey(lookup) != key {
//			continue
//		}
//		if strings.TrimSpace(lookup.AudioPath) == "" {
//			p.playTTSForText(ctx, w, util.FirstNonEmpty(lookup.Query, lookup.Key, lookup.Headword, lookup.Reading))
//			return
//		}
//		lookup := lookup
//		p.startAudioPlayback(w, "Audio Playback Failed", func() error {
//			return dictionary.PlayLookupAudio(lookup)
//		})
//		return
//	}
//}
//
//func (p *Page) addAllLookupFlashcards() {
//	if strings.TrimSpace(p.activeGameName) == "" {
//		p.showError("Create Flashcards Failed", "Select a game before creating flashcards.")
//		return
//	}
//	if len(p.lookupResults) == 0 {
//		p.showError("Create Flashcards Failed", "Run Dictionary Lookup first.")
//		return
//	}
//	cards := make([]flashcards.Flashcard, 0, len(p.lookupResults))
//	for _, lookup := range p.lookupResults {
//		cards = append(cards, p.flashcardFromLookup(lookup))
//	}
//	if _, _, err := flashcards.AddFlashcards(p.activeGameName, cards); err != nil {
//		p.showError("Create Flashcards Failed", err.Error())
//		return
//	}
//	_ = p.ReloadFlashcards()
//	p.resetFlashcardComposer()
//}
//
//func (p *Page) flashcardFromLookup(lookup dictionary.Lookup) flashcards.Flashcard {
//	word := util.FirstNonEmpty(lookup.Query, lookup.Headword, lookup.Key)
//	return flashcards.Flashcard{
//		GameName:           p.activeGameName,
//		Text:               word,
//		Meaning:            lookup.Meaning,
//		Reading:            lookup.Reading,
//		PronunciationText:  lookup.PronunciationText,
//		PronunciationPitch: lookup.Pitch,
//		AudioPath:          lookup.AudioPath,
//		SourcePath:         p.logPath,
//		SourceLine:         findFlashcardSourceLine(p.displayTranscript, word),
//	}
//}
//
//func (p *Page) lookupFlashcardExists(lookup dictionary.Lookup) bool {
//	card := p.flashcardFromLookup(lookup)
//	if _, ok := p.flashcardForWordExact(card.Text); ok {
//		return true
//	}
//	if _, ok := p.flashcardForWordExact(card.Reading); ok {
//		return true
//	}
//	if _, ok := p.flashcardForWordExact(card.PronunciationText); ok {
//		return true
//	}
//	if p.lookupMatchesSelectedFocusedToken(lookup) {
//		if _, ok := p.structureTokenFlashcardByKey(p.selectedFocusedTokenKey); ok {
//			return true
//		}
//	}
//	return false
//}
//
//func (p *Page) lookupMatchesSelectedFocusedToken(lookup dictionary.Lookup) bool {
//	key := strings.TrimSpace(p.selectedFocusedTokenKey)
//	if key == "" {
//		return false
//	}
//	candidates := p.focusedTokenCandidatesByKey(key)
//	if len(candidates) == 0 {
//		return false
//	}
//	lookupWords := []string{
//		lookup.Query,
//		lookup.Headword,
//		lookup.Key,
//		lookup.Reading,
//		lookup.PronunciationText,
//	}
//	for _, word := range lookupWords {
//		word = normalizeStructureMatchText(word)
//		if word == "" {
//			continue
//		}
//		if _, ok := candidates[word]; ok {
//			return true
//		}
//	}
//	return false
//}
//
//func (p *Page) focusedTokenCandidatesByKey(key string) map[string]struct{} {
//	key = strings.TrimSpace(key)
//	if key == "" {
//		return nil
//	}
//	analysis, _ := p.currentStructureAnalysis()
//	for _, token := range analysis.Tokens {
//		if structureTokenKey(token) != key {
//			continue
//		}
//		values := structureTokenFlashcardCandidates(token)
//		out := make(map[string]struct{}, len(values))
//		for _, value := range values {
//			if value != "" {
//				out[value] = struct{}{}
//			}
//		}
//		return out
//	}
//	return nil
//}
//
//func (p *Page) syncCurrentGameToAnki() error {
//	if strings.TrimSpace(p.activeGameName) == "" {
//		return fmt.Errorf("select a game before syncing Anki")
//	}
//	client := anki.New(p.ankiURL)
//	if _, err := client.SyncFlashcardsToAnki(p.activeGameName, p.ankiURL, p.pushSync); err != nil {
//		return err
//	}
//	return p.ReloadFlashcards()
//}
//
//func (p *Page) deleteCurrentLog() {
//	if p.currentConfig == nil {
//		p.showError("Delete Log Failed", "Select a game before deleting its transcript log.")
//		return
//	}
//	if p.OnDeleteLog != nil {
//		if err := p.OnDeleteLog(p.currentConfig); err != nil {
//			p.showError("Delete Log Failed", err.Error())
//			return
//		}
//	} else if err := p.currentConfig.DeleteLog(); err != nil {
//		p.showError("Delete Log Failed", err.Error())
//		return
//	}
//	p.ClearTranscript()
//	p.statusText = "Transcript log deleted; waiting for new dialogue."
//	p.showNotification("Transcript Log Deleted", "The saved transcript log was removed.", guitoast.NotificationTypeSuccess)
//}
//
//func (p *Page) logSizeText() string {
//	if p.currentConfig == nil {
//		return "Log: 0 B"
//	}
//	size, err := p.currentConfig.LogSize()
//	if err != nil {
//		return "Log: unavailable"
//	}
//	return "Log: " + vnutil.ByteCountSI(size)
//}
//
//func (p *Page) launchCurrentGameInBackground() {
//	if p.runnerStatus != nil {
//		p.statusText = p.transcriptRunningStatusText()
//		return
//	}
//	if p.currentConfig == nil || strings.TrimSpace(p.currentConfig.Name) == "" {
//		p.showError("Launch Failed", "The selected game configuration is not loaded yet.")
//		return
//	}
//	auto := runner.New()
//	status, err := auto.RunBackground(p.currentConfig)
//	if err != nil {
//		p.statusText = err.Error()
//		p.showError("Launch Failed", err.Error())
//		return
//	}
//	p.runnerStatus = status
//	p.statusText = fmt.Sprintf("Launching %s in the background.", p.currentConfig.Name)
//}
//
//func (p *Page) syncTranscriptEditor() {
//	if p.lastSyncedText == p.displayTranscript {
//		return
//	}
//	wasEmpty := p.lastSyncedText == ""
//	p.transcriptView.SetText(p.displayTranscript)
//	if wasEmpty {
//		runes := len([]rune(p.displayTranscript))
//		p.transcriptView.SetCaret(runes, runes)
//	}
//	p.lastSyncedText = p.displayTranscript
//}
//
//func (p *Page) syncFocusedSentenceView(text string) {
//	if p.lastFocusedText == text {
//		return
//	}
//	p.focusedSentenceView.SetText(text)
//	p.lastFocusedText = text
//}
//
//func (p *Page) paintTranscriptHighlights(gtx layout.Context) {
//	highlights := p.transcriptHighlights()
//	if len(highlights) == 0 || strings.TrimSpace(p.displayTranscript) == "" {
//		clear(p.transcriptHighlightBounds)
//		return
//	}
//	colorModeEnabled := p.colorizeHighlights && len(highlights) <= 160
//	var fill op.CallOp
//	var colorText op.CallOp
//	if colorModeEnabled {
//		colorMacro := op.Record(gtx.Ops)
//		p.layoutTranscriptLabel(gtx, p.theme.Color.Primary, nil)
//		colorText = colorMacro.Stop()
//	} else {
//		colorMacro := op.Record(gtx.Ops)
//		paint.ColorOp{Color: transcriptHighlightColor(p.theme.Color.Primary)}.Add(gtx.Ops)
//		fill = colorMacro.Stop()
//	}
//	regions := make([]widget.Region, 0, 8)
//	colorRects := make([]image.Rectangle, 0, len(highlights)*2)
//	validClicks := make(map[string]struct{}, len(highlights))
//	for _, match := range highlights {
//		validClicks[match.Key] = struct{}{}
//		if p.transcriptHighlightClicks[match.Key] == nil {
//			p.transcriptHighlightClicks[match.Key] = new(widget.Clickable)
//		}
//		regions = p.transcriptView.Regions(match.StartRune, match.EndRune, regions[:0])
//		var bounds image.Rectangle
//		for idx, region := range regions {
//			if idx == 0 {
//				bounds = region.Bounds
//			} else {
//				bounds = bounds.Union(region.Bounds)
//			}
//		}
//		if !bounds.Empty() {
//			p.transcriptHighlightBounds[match.Key] = bounds
//		}
//		for _, region := range regions {
//			if colorModeEnabled {
//				colorRects = append(colorRects, region.Bounds)
//			} else {
//				stack := clip.Rect(region.Bounds).Push(gtx.Ops)
//				fill.Add(gtx.Ops)
//				paint.PaintOp{}.Add(gtx.Ops)
//				stack.Pop()
//			}
//
//			offset := op.Offset(image.Pt(region.Bounds.Min.X, region.Bounds.Min.Y)).Push(gtx.Ops)
//			local := gtx
//			local.Constraints.Min = region.Bounds.Size()
//			local.Constraints.Max = region.Bounds.Size()
//			p.transcriptHighlightClicks[match.Key].Layout(local, func(gtx layout.Context) layout.Dimensions {
//				pointer.CursorPointer.Add(gtx.Ops)
//				return layout.Dimensions{Size: region.Bounds.Size()}
//			})
//			offset.Pop()
//		}
//	}
//	if colorModeEnabled && len(colorRects) > 0 {
//		for _, rect := range mergeHighlightRects(colorRects) {
//			stack := clip.Rect(rect).Push(gtx.Ops)
//			colorText.Add(gtx.Ops)
//			stack.Pop()
//		}
//	}
//	for key := range p.transcriptHighlightClicks {
//		if _, ok := validClicks[key]; !ok {
//			delete(p.transcriptHighlightClicks, key)
//			delete(p.transcriptHighlightBounds, key)
//		}
//	}
//	if p.popupFlashcard != nil {
//		found := false
//		for _, match := range highlights {
//			if p.popupMatchKey == match.Key {
//				if bounds, ok := p.transcriptHighlightBounds[match.Key]; ok {
//					p.popupAnchor = bounds
//				}
//				found = true
//				break
//			}
//		}
//		if !found {
//			p.DismissPopup()
//		}
//	}
//}
//
//func (p *Page) transcriptHighlights() []flashcards.Match {
//	cacheKey := p.highlightCacheKeyValue()
//	if cacheKey == p.highlightCacheKey {
//		return p.highlightCache
//	}
//	seen := make(map[string]flashcards.Flashcard, len(p.flashcards))
//	words := make([]string, 0, len(p.flashcards))
//	for _, card := range p.flashcards {
//		word := strings.TrimSpace(card.Text)
//		if word == "" {
//			continue
//		}
//		if _, ok := seen[word]; ok {
//			continue
//		}
//		seen[word] = card
//		words = append(words, word)
//	}
//	sort.SliceStable(words, func(i, j int) bool {
//		return len([]rune(words[i])) > len([]rune(words[j]))
//	})
//	matches := flashcards.FindMatches(p.displayTranscript, words)
//	for i := range matches {
//		matches[i].Card = seen[matches[i].Word]
//		matches[i].Key = fmt.Sprintf("%s-%d-%d", util.SanitizeName(matches[i].Card.ID), matches[i].StartRune, matches[i].EndRune)
//	}
//	p.highlightCacheKey = cacheKey
//	p.highlightCache = matches
//	return p.highlightCache
//}
//
//func (p *Page) openTranscriptHighlightPopup(key string) {
//	for _, match := range p.transcriptHighlights() {
//		if match.Key != key {
//			continue
//		}
//		if p.popupFlashcard != nil && p.popupMatchKey == match.Key {
//			p.DismissPopup()
//			return
//		}
//		cardCopy := match.Card
//		p.popupFlashcard = &cardCopy
//		p.popupAnchor = p.transcriptHighlightBounds[key]
//		p.popupMatchKey = match.Key
//		p.popupWord = match.Word
//		if p.autoPlayHighlightAudio {
//			_ = dictionary.PlayAudioForText(util.FirstNonEmpty(match.Card.Text, match.Card.Reading))
//		}
//		return
//	}
//}
//
//func (p *Page) statusColor() color.NRGBA {
//	status := strings.ToLower(p.statusText)
//	switch {
//	case strings.Contains(status, "failed"), strings.Contains(status, "error"):
//		return p.theme.Color.Error
//	case strings.Contains(status, "not found"):
//		return p.theme.Color.Warning
//	default:
//		return p.theme.Color.TextMuted
//	}
//}
//
//func (p *Page) isCompactLayout(gtx layout.Context) bool {
//	return gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(compactWidth))
//}
//
//func (p *Page) shouldStackTranscriptPage(gtx layout.Context) bool {
//	return gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(transcriptStackWidth))
//}
//
//func (p *Page) transcriptComposerWidth(gtx layout.Context) int {
//	if gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(transcriptMediumWidth)) {
//		return gtx.Dp(unit.Dp(360))
//	}
//	if gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(transcriptStackWidth)) {
//		return gtx.Dp(unit.Dp(380))
//	}
//	return gtx.Dp(unit.Dp(420))
//}
//
//func (p *Page) transcriptLaunchButtonLabel() string {
//	if p.runnerStatus != nil {
//		return "Game Running"
//	}
//	return "Launch Game"
//}
//
//func (p *Page) transcriptLaunchButtonIcon() string {
//	if p.runnerStatus != nil {
//		return "mdi:check-circle-outline"
//	}
//	return "mdi:play-box-outline"
//}
//
//func (p *Page) transcriptLaunchButtonVariant() bareui.ButtonVariant {
//	if p.runnerStatus != nil {
//		return bareui.ButtonSecondary
//	}
//	return bareui.ButtonPrimary
//}
//
//func (p *Page) transcriptRunningStatusText() string {
//	if p.runnerStatus != nil {
//		if p.runnerStatus.PID > 0 {
//			return fmt.Sprintf("Detected running game process (pid %d).", p.runnerStatus.PID)
//		}
//		return "Detected running game process."
//	}
//	return "Game process not detected."
//}
//
//func (p *Page) flashcardMetaText(card flashcards.Flashcard) string {
//	parts := make([]string, 0, 4)
//	if furigana := strings.TrimSpace(card.Furigana()); furigana != "" {
//		parts = append(parts, "Furigana: "+furigana)
//	}
//	if reading := strings.TrimSpace(card.Reading); reading != "" {
//		parts = append(parts, "Reading: "+reading)
//	}
//	if pronunciation := strings.TrimSpace(card.PronunciationText); pronunciation != "" {
//		if pitch := strings.TrimSpace(card.PronunciationPitch); pitch != "" {
//			pronunciation += " (" + pitch + ")"
//		}
//		parts = append(parts, "Pronunciation: "+pronunciation)
//	}
//	if strings.TrimSpace(card.AudioPath) != "" {
//		parts = append(parts, "Audio cached")
//	}
//	return strings.Join(parts, "\n")
//}
//
//func (p *Page) lookupResultAddClickable(key string) *widget.Clickable {
//	if p.lookupResultAddClicks[key] == nil {
//		p.lookupResultAddClicks[key] = new(widget.Clickable)
//	}
//	return p.lookupResultAddClicks[key]
//}
//
//func (p *Page) lookupResultPlayClickable(key string) *widget.Clickable {
//	if p.lookupResultPlayClicks[key] == nil {
//		p.lookupResultPlayClicks[key] = new(widget.Clickable)
//	}
//	return p.lookupResultPlayClicks[key]
//}
//
//func (p *Page) showError(title, body string) {
//	if p.OnError != nil {
//		p.OnError(title, body)
//	}
//}
//
//func (p *Page) showNotification(title, body string, kind guitoast.NotificationType) {
//	if p.OnNotify != nil {
//		p.OnNotify(title, body, kind)
//	}
//}
//
//func lookupResultKey(lookup dictionary.Lookup) string {
//	return util.FirstNonEmpty(lookup.Key, lookup.Query, lookup.Headword)
//}
//
//func (p *Page) selectedTranscriptText() string {
//	if selected := normalizeSelectionText(p.focusedSentenceView.SelectedText()); selected != "" {
//		return cleanInlineText(selected)
//	}
//	if selected := normalizeSelectionText(p.transcriptView.SelectedText()); selected != "" {
//		return cleanInlineText(selected)
//	}
//	return ""
//}
//
//func normalizeSelectionText(text string) string {
//	return cleanInlineText(text)
//}
//
//func cleanInlineText(text string) string {
//	text = strings.ReplaceAll(text, `\n`, " ")
//	text = strings.ReplaceAll(text, "\r\n", " ")
//	text = strings.ReplaceAll(text, "\r", " ")
//	text = strings.ReplaceAll(text, "\n", " ")
//	text = strings.ReplaceAll(text, "\t", " ")
//	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
//}
//
//func cleanTranscriptFocusText(text string) string {
//	parts := make([]string, 0, 1)
//	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
//		line = cleanInlineText(line)
//		if line == "" {
//			continue
//		}
//		_, body := splitTranscriptTimestamp(line)
//		body = cleanInlineText(body)
//		if body != "" {
//			parts = append(parts, body)
//		}
//	}
//	return cleanInlineText(strings.Join(parts, " "))
//}
//
//func findFlashcardSourceLine(transcriptText, word string) string {
//	word = strings.TrimSpace(word)
//	if word == "" {
//		return ""
//	}
//	for _, line := range strings.Split(transcriptText, "\n") {
//		trimmed := strings.TrimSpace(line)
//		if trimmed == "" {
//			continue
//		}
//		if strings.Contains(trimmed, word) {
//			return trimmed
//		}
//	}
//	return ""
//}
//
//func sanitizeTranscriptForDisplay(text string) string {
//	text = ansiRE.ReplaceAllString(text, "")
//	text = strings.ReplaceAll(text, `\r\n`, " ")
//	text = strings.ReplaceAll(text, `\n`, " ")
//	text = strings.ReplaceAll(text, `\r`, " ")
//	text = strings.ReplaceAll(text, "\r\n", "\n")
//	text = strings.ReplaceAll(text, "\r", "\n")
//	text = strings.ReplaceAll(text, "\t", "    ")
//	text = strings.Map(func(r rune) rune {
//		if r < 0x20 && r != '\n' && r != '\t' {
//			return -1
//		}
//		return r
//	}, text)
//	return text
//}
//
//func limitTranscriptLines(text string, recentLineLimit int) string {
//	if recentLineLimit <= 0 {
//		return text
//	}
//	lines := strings.Split(text, "\n")
//	if len(lines) <= recentLineLimit {
//		return text
//	}
//	return strings.Join(lines[len(lines)-recentLineLimit:], "\n")
//}
//
//func transcriptHighlightColor(base color.NRGBA) color.NRGBA {
//	return color.NRGBA{R: base.R, G: base.G, B: base.B, A: 72}
//}
//
//func transcriptPopupBorderColor(base color.NRGBA) color.NRGBA {
//	return color.NRGBA{R: base.R, G: base.G, B: base.B, A: 160}
//}
//
//func (p *Page) invalidateHighlights() {
//	p.highlightCacheKey = ""
//	p.highlightCache = nil
//}
//
//func (p *Page) highlightCacheKeyValue() string {
//	var b strings.Builder
//	b.Grow(len(p.displayTranscript) + len(p.flashcards)*24)
//	b.WriteString(p.displayTranscript)
//	b.WriteString("\x00")
//	for _, card := range p.flashcards {
//		b.WriteString(card.ID)
//		b.WriteString("\x1f")
//		b.WriteString(card.Text)
//		b.WriteString("\x1e")
//	}
//	return b.String()
//}
//
//func mergeHighlightRects(rects []image.Rectangle) []image.Rectangle {
//	if len(rects) <= 1 {
//		return rects
//	}
//	sort.Slice(rects, func(i, j int) bool {
//		if rects[i].Min.Y != rects[j].Min.Y {
//			return rects[i].Min.Y < rects[j].Min.Y
//		}
//		return rects[i].Min.X < rects[j].Min.X
//	})
//	merged := make([]image.Rectangle, 0, len(rects))
//	current := rects[0]
//	for _, rect := range rects[1:] {
//		if shouldMergeHighlightRect(current, rect) {
//			current = current.Union(rect)
//			continue
//		}
//		merged = append(merged, current)
//		current = rect
//	}
//	merged = append(merged, current)
//	return merged
//}
//
//func shouldMergeHighlightRect(a, b image.Rectangle) bool {
//	if a.Empty() || b.Empty() {
//		return false
//	}
//	if a.Min.Y > b.Max.Y || b.Min.Y > a.Max.Y {
//		return false
//	}
//	return b.Min.X <= a.Max.X+6
//}
//
//func (p *Page) playFlashcardAudio(ctx context.Context, card flashcards.Flashcard) error {
//	word := util.FirstNonEmpty(card.Text, card.Reading)
//	if strings.TrimSpace(word) == "" {
//		return fmt.Errorf("no audio is available for this flashcard")
//	}
//	if err := dictionary.PlayAudioForText(word); err == nil {
//		return nil
//	}
//	return p.playTTSForTextSync(ctx, word)
//}
//
//func (p *Page) playTTSForText(ctx context.Context, w *app.Window, text string) {
//	text = cleanInlineText(text)
//	if text == "" {
//		p.showError("TTS Playback Failed", "No text is available for TTS.")
//		return
//	}
//	if _, ok := p.selectedTTSReference(); !ok {
//		p.showError("TTS Playback Failed", "Select a TTS reference speaker from the transcript toolbar first.")
//		return
//	}
//	p.startAudioPlayback(w, "TTS Playback Failed", func() error {
//		return p.playTTSForTextSync(ctx, text)
//	})
//}
//
//func (p *Page) playTTSForTextSync(ctx context.Context, text string) error {
//	ref, ok := p.selectedTTSReference()
//	if !ok {
//		return fmt.Errorf("select a TTS reference speaker from the transcript toolbar first")
//	}
//	audioPath, err := p.cachedTranscriptVoicePath(ref.Voice)
//	if err != nil {
//		return err
//	}
//	path, err := wgltts.SpeakWithF5(ctx, p.activeGameName, text, wgltts.Reference{
//		Speaker: ref.Speaker,
//		Audio:   audioPath,
//		Text:    ref.Text,
//	})
//	if err != nil {
//		return err
//	}
//	player, err := audioplayer.NewPlayer(audioplayer.Config{Backend: audioplayer.BackendAuto})
//	if err != nil {
//		return err
//	}
//	return audioplayer.PlayAudioFile(player, path, true)
//}
//
//func engineFileCacheExt(fileInfo *engine.EngineFileInfo) string {
//	if fileInfo == nil {
//		return ""
//	}
//	if ext := strings.TrimSpace(fileInfo.Ext); ext != "" {
//		return ext
//	}
//	if ext := filepath.Ext(strings.TrimSpace(fileInfo.Name)); ext != "" {
//		return ext
//	}
//	switch strings.ToLower(strings.TrimSpace(fileInfo.MediaType)) {
//	case "audio/ogg", "audio/oga":
//		return ".ogg"
//	case "audio/wav", "audio/wave", "audio/x-wav":
//		return ".wav"
//	case "audio/mpeg", "audio/mp3":
//		return ".mp3"
//	case "audio/mp4", "audio/m4a":
//		return ".m4a"
//	case "audio/flac":
//		return ".flac"
//	case "audio/opus":
//		return ".opus"
//	default:
//		return ""
//	}
//}
