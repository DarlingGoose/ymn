package transcript

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"regexp"
	"sort"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareui "github.com/DarlingGoose/bare/pkg/ui"
	"github.com/DarlingGoose/bare/pkg/ui/icons"
	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	vngame "github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/vntext/pkg/runner"
	"github.com/DarlingGoose/wgl/pkg/anki"
	"github.com/DarlingGoose/wgl/pkg/dictionary"
	flashcards "github.com/DarlingGoose/wgl/pkg/flashcard"
	"github.com/DarlingGoose/wgl/pkg/gui"
	guitoast "github.com/DarlingGoose/wgl/pkg/gui/toast"
	"github.com/DarlingGoose/wgl/pkg/japanese"
	"github.com/DarlingGoose/wgl/pkg/translation"
	"github.com/DarlingGoose/wgl/pkg/util"
)

const (
	compactWidth          = 1080
	transcriptStackWidth  = 1240
	transcriptMediumWidth = 1480

	composerFocusFlashcards        = "flashcards"
	composerFocusSentenceStructure = "sentence_structure"

	focusedFuriganaHidden = "hidden"
	focusedFuriganaAbove  = "above"
	focusedFuriganaBelow  = "below"
)

var _ gui.EvenHandler = &Page{}

var ansiRE = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|[@-Z\\-_])`)

type Page struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	transcriptView       widget.Selectable
	focusedSentenceView  widget.Selectable
	transcriptFocusSplit widget.Float
	transcriptList       widget.List
	structureList        widget.List
	lookupResultsList    widget.List
	wordEditor           widget.Editor
	meaningEditor        widget.Editor
	translationEditor    widget.Editor
	targetLanguageDrop   bareui.Dropdown
	hideReadingInAnki    widget.Bool
	searchWordButton     widget.Clickable
	playAudioButton      widget.Clickable
	addAllLookupButton   widget.Clickable
	launchGameButton     widget.Clickable
	syncAnkiButton       widget.Clickable
	clearButton          widget.Clickable

	playSentenceButton         widget.Clickable
	translateSentenceButton    widget.Clickable
	saveSentenceButton         widget.Clickable
	focusedLookupButton        widget.Clickable
	transcriptPopupAudioButton widget.Clickable
	transcriptPopupCloseButton widget.Clickable
	popupDismissClicks         [4]widget.Clickable
	composerToggleButton       widget.Clickable
	composerFlashcardsTab      widget.Clickable
	composerSentenceTab        widget.Clickable
	translationToggleButton    widget.Clickable
	saveTranslationButton      widget.Clickable
	generateTranslationButton  widget.Clickable
	autoTranslateMissing       widget.Bool
	furiganaHiddenButton       widget.Clickable
	furiganaAboveButton        widget.Clickable
	furiganaBelowButton        widget.Clickable
	focusedTokenAddButton      widget.Clickable
	focusedTokenAudioButton    widget.Clickable

	transcriptHighlightClicks map[string]*widget.Clickable
	transcriptHighlightBounds map[string]image.Rectangle
	lookupResultAddClicks     map[string]*widget.Clickable
	lookupResultPlayClicks    map[string]*widget.Clickable
	structureTokenAddClicks   map[string]*widget.Clickable
	structureTokenPlayClicks  map[string]*widget.Clickable
	transcriptRowClicks       map[string]*widget.Clickable
	focusedTokenClicks        map[string]*widget.Clickable
	targetLanguageOptions     []gui.DropdownOption

	activeGameName         string
	logPath                string
	ankiURL                string
	pushSync               bool
	statusText             string
	currentConfig          *vngame.Game
	runnerStatus           *runner.ProcessStatus
	selectedTextSizeName   string
	selectedRecentLines    string
	transcriptTextSize     unit.Sp
	focusedTextSize        unit.Sp
	translateDetailSize    unit.Sp
	recentLineLimit        int
	autoPlayHighlightAudio bool
	colorizeHighlights     bool

	flashcards                []flashcards.Flashcard
	lookupResult              *dictionary.Lookup
	lookupResults             []dictionary.Lookup
	displayTranscript         string
	lastSyncedText            string
	lastFocusedText           string
	structureCacheKey         string
	structureCache            japanese.Analysis
	structureCacheErr         string
	highlightCacheKey         string
	highlightCache            []flashcards.Match
	popupFlashcard            *flashcards.Flashcard
	popupAnchor               image.Rectangle
	popupBounds               image.Rectangle
	popupMatchKey             string
	popupWord                 string
	selectedLineKey           string
	selectedLineText          string
	selectedFocusedTokenKey   string
	selectedFocusedTokenWord  string
	selectedFocusedTokenNote  string
	translationCollapsed      bool
	focusedFuriganaMode       string
	focusedFuriganaDefault    string
	selectedTargetLanguage    string
	translationLoadedKey      string
	translationGeneratingKey  string
	autoTranslationAttemptKey string
	translationResultCh       chan translationResult
	translatorConfig          translation.Config
	composerFocus             string
	composerMinimized         bool
	composerLastUsed          time.Time
	lastAutoWord              string
	hideReadingSet            bool

	OnError  func(title, body string)
	OnNotify func(title, body string, kind guitoast.NotificationType)
}

type translationResult struct {
	Key   string
	Entry translation.Entry
	Err   error
}

type transcriptRow struct {
	Key        string
	Time       string
	Text       string
	VocabWords []string
}

func New(theme barethemes.Theme) *Page {
	p := &Page{
		theme:                     theme,
		pushSync:                  true,
		statusText:                "Start the game to show live transcript text here.",
		selectedTextSizeName:      "Medium",
		selectedRecentLines:       "All Lines",
		transcriptTextSize:        unit.Sp(16),
		focusedTextSize:           unit.Sp(26),
		translateDetailSize:       unit.Sp(15),
		focusedFuriganaMode:       focusedFuriganaHidden,
		focusedFuriganaDefault:    focusedFuriganaHidden,
		selectedTargetLanguage:    "English",
		translationResultCh:       make(chan translationResult, 1),
		translatorConfig:          translation.Config{Provider: translation.ProviderOllama},
		composerFocus:             composerFocusFlashcards,
		composerMinimized:         true,
		composerLastUsed:          time.Now(),
		transcriptHighlightClicks: make(map[string]*widget.Clickable),
		transcriptHighlightBounds: make(map[string]image.Rectangle),
		lookupResultAddClicks:     make(map[string]*widget.Clickable),
		lookupResultPlayClicks:    make(map[string]*widget.Clickable),
		structureTokenAddClicks:   make(map[string]*widget.Clickable),
		structureTokenPlayClicks:  make(map[string]*widget.Clickable),
		transcriptRowClicks:       make(map[string]*widget.Clickable),
		focusedTokenClicks:        make(map[string]*widget.Clickable),
	}
	p.transcriptFocusSplit.Value = 0.5
	p.wordEditor.SingleLine = true
	p.meaningEditor.SingleLine = false
	p.translationEditor.SingleLine = false
	gui.NewDropDownLayout(&p.targetLanguageDrop, "mdi:translate")
	p.targetLanguageDrop.Width = unit.Dp(190)
	p.targetLanguageDrop.OffsetY = unit.Dp(42)
	p.targetLanguageOptions = newTranslationLanguageOptions()
	p.transcriptList.Axis = layout.Vertical
	p.transcriptList.ScrollToEnd = true
	p.structureList.Axis = layout.Vertical
	p.lookupResultsList.Axis = layout.Vertical
	return p
}

func (p *Page) WithTheme(theme barethemes.Theme) *Page {
	p.theme = theme
	return p
}

func (p *Page) WithIcon(icon *icons.Iconify) *Page {
	p.iconify = icon
	return p
}

func (p *Page) SetContext(activeGameName, logPath, ankiURL string, cfg *vngame.Game) *Page {
	p.activeGameName = strings.TrimSpace(activeGameName)
	p.logPath = strings.TrimSpace(logPath)
	p.ankiURL = strings.TrimSpace(ankiURL)
	p.currentConfig = cfg
	return p
}

func (p *Page) SetPushSync(pushSync bool) *Page {
	p.pushSync = pushSync
	return p
}

func (p *Page) SetRunningState(running bool, pid int) *Page {
	if !running {
		p.runnerStatus = nil
		return p
	}
	p.runnerStatus = &runner.ProcessStatus{
		PID:    pid,
		Status: runner.StatusRunning,
	}
	return p
}

func (p *Page) SetTranscriptOptions(textSize unit.Sp, textSizeName string, recentLineLimit int, recentLinesName string) *Page {
	p.transcriptTextSize = textSize
	p.selectedTextSizeName = strings.TrimSpace(textSizeName)
	p.recentLineLimit = recentLineLimit
	p.selectedRecentLines = strings.TrimSpace(recentLinesName)
	return p
}

func (p *Page) SetTranslateTextOptions(focusedSize, detailSize unit.Sp) *Page {
	if focusedSize > 0 {
		p.focusedTextSize = focusedSize
	}
	if detailSize > 0 {
		p.translateDetailSize = detailSize
	}
	return p
}

func (p *Page) SetTranslatorConfig(cfg translation.Config) *Page {
	p.translatorConfig = cfg
	return p
}

func (p *Page) SetFocusedFuriganaDefault(mode string) *Page {
	mode = normalizeFocusedFuriganaMode(mode)
	oldDefault := p.focusedFuriganaDefault
	if oldDefault == "" || p.focusedFuriganaMode == "" || p.focusedFuriganaMode == oldDefault {
		p.focusedFuriganaMode = mode
	}
	p.focusedFuriganaDefault = mode
	return p
}

func (p *Page) SetAutoPlayHighlightAudio(enabled bool) *Page {
	p.autoPlayHighlightAudio = enabled
	return p
}

func (p *Page) SetColorizeHighlights(enabled bool) *Page {
	p.colorizeHighlights = enabled
	return p
}

func (p *Page) SetStatus(status string) *Page {
	p.statusText = strings.TrimSpace(status)
	return p
}

func (p *Page) SetRawTranscript(raw string) *Page {
	next := limitTranscriptLines(sanitizeTranscriptForDisplay(raw), p.recentLineLimit)
	if next != p.displayTranscript {
		p.displayTranscript = next
		p.invalidateHighlights()
	}
	return p
}

func (p *Page) ClearTranscript() {
	p.displayTranscript = ""
	p.lookupResult = nil
	p.lookupResults = nil
	p.invalidateHighlights()
	p.DismissPopup()
	p.selectedLineKey = ""
	p.selectedLineText = ""
	p.statusText = "Transcript view cleared; waiting for new dialogue."
	p.lastSyncedText = ""
}

func (p *Page) SetFlashcards(cards []flashcards.Flashcard) *Page {
	p.flashcards = append([]flashcards.Flashcard(nil), cards...)
	sort.Slice(p.flashcards, func(i, j int) bool {
		return p.flashcards[i].UpdatedAt.After(p.flashcards[j].UpdatedAt)
	})
	p.invalidateHighlights()
	return p
}

func (p *Page) Cards() []flashcards.Flashcard {
	return append([]flashcards.Flashcard(nil), p.flashcards...)
}

func (p *Page) ReloadFlashcards() error {
	if strings.TrimSpace(p.activeGameName) == "" {
		p.flashcards = nil
		return nil
	}
	cards, err := flashcards.LoadFlashcards(p.activeGameName)
	if err != nil {
		return err
	}
	p.SetFlashcards(cards)
	return nil
}

func (p *Page) PopupFlashcard() *flashcards.Flashcard {
	return p.popupFlashcard
}

func (p *Page) DismissPopup() {
	p.popupFlashcard = nil
	p.popupBounds = image.Rectangle{}
	p.popupMatchKey = ""
	p.popupWord = ""
}

func (p *Page) HandleEvents(gtx layout.Context, ctx context.Context, w *app.Window) {
	p.drainTranslationResults()
	p.maybeAutoGenerateTranslation(ctx, w)
	if p.hideReadingInAnki.Update(gtx) {
		p.hideReadingSet = true
	}
	p.targetLanguageDrop.Update(gtx)
	p.syncHideReadingDefault()
	for p.launchGameButton.Clicked(gtx) {
		p.launchCurrentGameInBackground()
	}

	for p.syncAnkiButton.Clicked(gtx) {
		if err := p.syncCurrentGameToAnki(); err != nil {
			p.showError("Anki Sync Failed", err.Error())
		} else {
			p.showNotification("Anki Sync Complete", "Transcript flashcards synced to Anki.", guitoast.NotificationTypeSuccess)
		}
	}
	for p.clearButton.Clicked(gtx) {
		p.ClearTranscript()
	}
	for p.playSentenceButton.Clicked(gtx) {
		p.playCurrentLookupAudio()
	}
	for p.translateSentenceButton.Clicked(gtx) {
		p.composerFocus = composerFocusSentenceStructure
		p.composerMinimized = false
		p.composerLastUsed = time.Now()
	}
	for p.saveSentenceButton.Clicked(gtx) {
		p.composerFocus = composerFocusFlashcards
		p.composerMinimized = false
		p.composerLastUsed = time.Now()
	}
	for p.focusedLookupButton.Clicked(gtx) {
		p.lookupCurrentWord()
	}
	for p.focusedTokenAddButton.Clicked(gtx) {
		p.addFocusedTokenFlashcard()
	}
	for p.focusedTokenAudioButton.Clicked(gtx) {
		p.playFocusedTokenAudio()
	}
	for p.translationToggleButton.Clicked(gtx) {
		p.translationCollapsed = !p.translationCollapsed
	}
	for p.saveTranslationButton.Clicked(gtx) {
		p.saveCurrentTranslation()
	}
	for p.generateTranslationButton.Clicked(gtx) {
		p.generateCurrentTranslation(ctx, w)
	}
	for p.furiganaHiddenButton.Clicked(gtx) {
		p.focusedFuriganaMode = focusedFuriganaHidden
	}
	for p.furiganaAboveButton.Clicked(gtx) {
		p.focusedFuriganaMode = focusedFuriganaAbove
	}
	for p.furiganaBelowButton.Clicked(gtx) {
		p.focusedFuriganaMode = focusedFuriganaBelow
	}
	for i := range p.targetLanguageOptions {
		opt := &p.targetLanguageOptions[i]
		for opt.Clickable.Clicked(gtx) {
			p.selectedTargetLanguage = opt.Label
			p.translationLoadedKey = ""
			p.autoTranslationAttemptKey = ""
			p.targetLanguageDrop.Close()
		}
	}
	for p.searchWordButton.Clicked(gtx) {
		p.lookupCurrentWord()
	}
	for p.playAudioButton.Clicked(gtx) {
		p.playCurrentLookupAudio()
	}
	for p.addAllLookupButton.Clicked(gtx) {
		p.addAllLookupFlashcards()
	}
	for key, click := range p.lookupResultAddClicks {
		for click.Clicked(gtx) {
			p.addLookupFlashcardByKey(key)
		}
	}
	for key, click := range p.lookupResultPlayClicks {
		for click.Clicked(gtx) {
			p.playLookupAudioByKey(key)
		}
	}
	for key, click := range p.structureTokenAddClicks {
		for click.Clicked(gtx) {
			p.addStructureTokenFlashcard(key)
		}
	}
	for key, click := range p.structureTokenPlayClicks {
		for click.Clicked(gtx) {
			p.playStructureTokenAudio(key)
		}
	}
	for key, click := range p.transcriptHighlightClicks {
		for click.Clicked(gtx) {
			p.openTranscriptHighlightPopup(key)
		}
	}
	for key, click := range p.transcriptRowClicks {
		for click.Clicked(gtx) {
			p.selectTranscriptRow(key)
		}
	}
	for key, click := range p.focusedTokenClicks {
		for click.Clicked(gtx) {
			p.selectFocusedToken(key)
		}
	}
	for p.transcriptPopupAudioButton.Clicked(gtx) {
		if p.popupFlashcard == nil {
			p.showError("Audio Playback Failed", "No flashcard is selected.")
			continue
		}
		if err := playFlashcardAudio(*p.popupFlashcard); err != nil {
			p.showError("Audio Playback Failed", err.Error())
		}
	}
	for p.transcriptPopupCloseButton.Clicked(gtx) {
		p.DismissPopup()
	}
	for p.composerToggleButton.Clicked(gtx) {
		p.composerMinimized = !p.composerMinimized
		p.composerLastUsed = time.Now()
	}
	for i := range p.popupDismissClicks {
		for p.popupDismissClicks[i].Clicked(gtx) {
			p.DismissPopup()
		}
	}
}

func (p *Page) LayoutPage(gtx layout.Context) layout.Dimensions {
	if p.iconify == nil {
		p.iconify = icons.NewIconify()
	}
	p.syncTranscriptEditor()
	return p.layoutTranscriptPanel(gtx)
}

func (p *Page) LayoutPopupContent(gtx layout.Context) layout.Dimensions {
	if p.popupFlashcard == nil {
		return layout.Dimensions{}
	}
	card := *p.popupFlashcard
	audioButton := bareui.Button{
		Clickable: &p.transcriptPopupAudioButton,
		Text:      "Play Audio",
		Prefix:    "mdi:play-circle-outline",
		Variant:   bareui.ButtonSecondary,
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(p.theme.Gio(), card.Text)
			lbl.Color = p.theme.Color.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(p.theme.Gio(), card.Meaning)
			lbl.Color = p.theme.Color.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			meta := p.flashcardMetaText(card)
			if meta == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(p.theme.Gio(), meta)
				lbl.Color = p.theme.Color.TextMuted
				return lbl.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !util.IsExistingFile(card.AudioPath) {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return audioButton.Layout(gtx, p.theme, p.iconify)
			})
		}),
	)
}

func (p *Page) layoutTranscriptPanel(gtx layout.Context) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(18),
					Left:   unit.Dp(20),
					Right:  unit.Dp(20),
					Bottom: unit.Dp(18),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return p.layoutTranscriptTopbar(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return p.layoutTranscriptWorkspace(gtx)
						}),
					)
				})
			}),
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{}
			}),
		)
	})
}

func (p *Page) layoutTranscriptTopbar(gtx layout.Context) layout.Dimensions {
	launchButton := bareui.Button{
		Clickable: &p.launchGameButton,
		Text:      p.transcriptLaunchButtonLabel(),
		Prefix:    p.transcriptLaunchButtonIcon(),
		Variant:   p.transcriptLaunchButtonVariant(),
	}
	syncButton := bareui.Button{
		Clickable: &p.syncAnkiButton,
		Text:      "Sync Anki",
		Prefix:    "mdi:cloud-sync-outline",
		Variant:   bareui.ButtonSecondary,
	}
	clearButton := bareui.Button{
		Clickable: &p.clearButton,
		Text:      "mdi:broom",
		Icon:      true,
		Variant:   bareui.ButtonGhost,
	}
	statusText := "IDLE"
	statusLive := false
	if p.runnerStatus != nil {
		statusText = "LIVE"
		statusLive = true
	}
	if p.isCompactLayout(gtx) {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.layoutStatusPill(gtx, statusText, statusLive)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(p.theme.Gio(), util.FirstNonEmpty(p.activeGameName, "No game selected"))
						lbl.Color = p.theme.Color.Text
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return clearButton.Layout(gtx, p.theme, p.iconify)
					}),
				)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						if p.runnerStatus != nil {
							return launchButton.Layout(gtx.Disabled(), p.theme, p.iconify)
						}
						return launchButton.Layout(gtx, p.theme, p.iconify)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return syncButton.Layout(gtx, p.theme, p.iconify)
					}),
				)
			}),
			//layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			//	lbl := material.Body1(p.theme.Gio(), p.statusText)
			//	lbl.Color = p.statusColor()
			//	return lbl.Layout(gtx)
			//}),
		)
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutStatusPill(gtx, statusText, statusLive)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(p.theme.Gio(), p.transcriptRunningStatusText())
			lbl.Color = p.theme.Color.TextMuted
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.runnerStatus != nil {
				return launchButton.Layout(gtx.Disabled(), p.theme, p.iconify)
			}
			return launchButton.Layout(gtx, p.theme, p.iconify)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return syncButton.Layout(gtx, p.theme, p.iconify)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(4))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return clearButton.Layout(gtx, p.theme, p.iconify)
		}),
	)
}

func (p *Page) layoutStatusPill(gtx layout.Context, text string, live bool) layout.Dimensions {
	bg := p.theme.Color.SurfaceAlt
	fg := p.theme.Color.TextMuted

	if live {
		fg = p.theme.Color.Primary
		bg = color.NRGBA{R: fg.R, G: fg.G, B: fg.B, A: 42}
	}

	return RoundedSurfaceWrap(
		gtx,
		bg,
		unit.Dp(p.theme.Radius.MD),
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(7),
				Bottom: unit.Dp(7),
				Left:   unit.Dp(10),
				Right:  unit.Dp(10),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(p.theme.Gio(), text)
				lbl.Color = fg
				return lbl.Layout(gtx)
			})
		},
	)
}

func RoundedSurfaceWrap(
	gtx layout.Context,
	bg color.NRGBA,
	radius unit.Dp,
	w layout.Widget,
) layout.Dimensions {
	macro := op.Record(gtx.Ops)

	dims := w(gtx)

	call := macro.Stop()

	rr := clip.RRect{
		Rect: image.Rectangle{
			Max: dims.Size,
		},
		NE: int(gtx.Dp(radius)),
		NW: int(gtx.Dp(radius)),
		SE: int(gtx.Dp(radius)),
		SW: int(gtx.Dp(radius)),
	}

	paint.FillShape(gtx.Ops, bg, rr.Op(gtx.Ops))
	call.Add(gtx.Ops)

	return dims
}

func (p *Page) layoutTranscriptWorkspace(gtx layout.Context) layout.Dimensions {
	if p.runnerStatus == nil || p.shouldStackTranscriptPage(gtx) {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = gtx.Constraints.Max
				return p.layoutTranscriptBodyPanel(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			//	if p.runnerStatus == nil {
			//		return layout.Dimensions{}
			//	}
			//	return p.layoutFlashcardComposer(gtx)
			//}),
		)
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return p.layoutTranscriptBodyPanel(gtx)
		}),
		//layout.Rigid(bareutils.SpacerW(unit.Dp(14))),
		//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		//	width := p.transcriptComposerWidth(gtx)
		//	gtx.Constraints.Min.X = width
		//	gtx.Constraints.Max.X = width
		//	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
		//	return p.layoutContextRail(gtx)
		//}),
	)
}

func (p *Page) layoutTranscriptBodyPanel(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min = gtx.Constraints.Max
	liveRatio := p.transcriptFocusRatio()
	focusedRatio := 1 - liveRatio

	if p.runnerStatus == nil {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Flexed(liveRatio, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = gtx.Constraints.Max
				return p.layoutLiveTranscriptCard(gtx)
			}),
		)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(liveRatio, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return p.layoutLiveTranscriptCard(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutTranscriptFocusResizeHandle(gtx)
		}),
		layout.Flexed(focusedRatio, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return p.layoutFocusedSentenceCard(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutBottomActions(gtx)
		}),
	)
}

func (p *Page) layoutTranscriptFocusResizeHandle(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(8),
		Bottom: unit.Dp(8),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return bareutils.RoundedSurface(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(6),
				Bottom: unit.Dp(6),
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body2(p.theme.Gio(), "Live")
						lbl.Color = p.theme.Color.TextMuted
						return lbl.Layout(gtx)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						pointer.CursorRowResize.Add(gtx.Ops)
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						slider := material.Slider(p.theme.Gio(), &p.transcriptFocusSplit)
						slider.Color = p.theme.Color.Primary
						return slider.Layout(gtx)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body2(p.theme.Gio(), "Focused")
						lbl.Color = p.theme.Color.TextMuted
						return lbl.Layout(gtx)
					}),
				)
			})
		})
	})
}

func (p *Page) transcriptFocusRatio() float32 {
	if p.transcriptFocusSplit.Value < 0.25 {
		p.transcriptFocusSplit.Value = 0.25
	}
	if p.transcriptFocusSplit.Value > 0.75 {
		p.transcriptFocusSplit.Value = 0.75
	}
	return p.transcriptFocusSplit.Value
}

func (p *Page) layoutLiveTranscriptCard(gtx layout.Context) layout.Dimensions {
	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutCardHeader(gtx, "Live Transcript", "Scanning mode: saved words are highlighted inline")
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = gtx.Constraints.Max
				if p.runnerStatus == nil {
					return p.layoutTranscriptIdleState(gtx)
				}
				return p.layoutTranscriptEditor(gtx)
			}),
		)
	})
}

func (p *Page) layoutFocusedSentenceCard(gtx layout.Context) layout.Dimensions {
	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutCardHeader(gtx, "Focused Sentence", "Select text, click a saved word, or use the latest transcript line")
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedSentenceText(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedFuriganaControls(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedTokenActions(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				analysis, errText := p.currentStructureAnalysis()
				if errText == "" && len(analysis.Tokens) == 0 {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return p.layoutFocusedSentenceChips(gtx, analysis, errText)
				})
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedLookupBar(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedTranslationSection(gtx)
			}),
			//layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
			//layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			//	gtx.Constraints.Min = gtx.Constraints.Max
			//	return p.layoutSentenceStructurePanel(gtx, true)
			//}),
		)
	})
}

func (p *Page) layoutFocusedSentenceText(gtx layout.Context) layout.Dimensions {
	text := p.structureSourceText()
	if text == "" {
		text = "Start the game to inspect the latest sentence."
	}
	text = cleanInlineText(text)
	p.syncFocusedSentenceView(text)
	if p.focusedFuriganaMode != focusedFuriganaHidden {
		return p.layoutFocusedSentenceWithFurigana(gtx, text)
	}

	lbl := material.H6(p.theme.Gio(), text)
	lbl.Color = p.theme.Color.Text
	lbl.TextSize = p.focusedSentenceTextSize(gtx)
	lbl.State = &p.focusedSentenceView
	return lbl.Layout(gtx)
}

func (p *Page) layoutFocusedSentenceWithFurigana(gtx layout.Context, sentence string) layout.Dimensions {
	analysis, err := japanese.AnalyzeSentence(sentence)
	if err != nil || len(analysis.Tokens) == 0 {
		lbl := material.H6(p.theme.Gio(), sentence)
		lbl.Color = p.theme.Color.Text
		lbl.TextSize = p.focusedSentenceTextSize(gtx)
		return lbl.Layout(gtx)
	}
	lines := p.focusedSentenceTokenLines(gtx, analysis.Tokens)
	children := make([]layout.FlexChild, 0, len(lines))
	for i, line := range lines {
		line := line
		if i > 0 {
			children = append(children, layout.Rigid(bareutils.SpacerH(unit.Dp(5))))
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lineChildren := make([]layout.FlexChild, 0, len(line))
			for _, token := range line {
				token := token
				lineChildren = append(lineChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutFocusedFuriganaToken(gtx, token)
				}))
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, lineChildren...)
		}))
	}
	p.pruneFocusedTokenClicks(analysis.Tokens)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (p *Page) focusedSentenceTokenLines(gtx layout.Context, tokens []japanese.Token) [][]japanese.Token {
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		return [][]japanese.Token{tokens}
	}
	lines := make([][]japanese.Token, 0, 2)
	line := make([]japanese.Token, 0, len(tokens))
	lineWidth := 0
	for _, token := range tokens {
		tokenWidth := p.focusedSentenceTokenWidth(gtx, token)
		if len(line) > 0 && lineWidth+tokenWidth > maxWidth {
			lines = append(lines, line)
			line = make([]japanese.Token, 0, len(tokens))
			lineWidth = 0
		}
		line = append(line, token)
		lineWidth += tokenWidth
	}
	if len(line) > 0 {
		lines = append(lines, line)
	}
	return lines
}

func (p *Page) focusedSentenceTokenWidth(gtx layout.Context, token japanese.Token) int {
	surfaceRunes := len([]rune(cleanInlineText(token.Surface)))
	readingRunes := len([]rune(focusedTokenReading(token)))
	runes := surfaceRunes
	if readingRunes > runes {
		runes = readingRunes
	}
	if runes <= 0 {
		runes = 1
	}
	size := float32(p.focusedSentenceTextSize(gtx))
	return gtx.Dp(unit.Dp(float32(runes)*size*0.72 + 16))
}

func (p *Page) layoutFocusedFuriganaToken(gtx layout.Context, token japanese.Token) layout.Dimensions {
	key := structureTokenKey(token)
	click := p.focusedTokenClickable(key)
	reading := focusedTokenReading(token)
	surface := cleanInlineText(token.Surface)
	if surface == "" {
		return layout.Dimensions{}
	}
	_, inFlashcards := p.structureTokenFlashcard(token)
	dictionaryReady := focusedTokenDictionaryReady(token)
	bg := focusedTokenColor(p.theme, token, p.selectedFocusedTokenKey == key, inFlashcards, dictionaryReady)
	children := make([]layout.FlexChild, 0, 4)
	if p.focusedFuriganaMode == focusedFuriganaAbove {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutFocusedTokenSlot(gtx, unit.Dp(18), func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedTokenReading(gtx, reading)
			})
		}), layout.Rigid(bareutils.SpacerH(unit.Dp(2))))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return p.layoutFocusedTokenSlot(gtx, p.focusedTokenSurfaceSlotHeight(gtx), func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(p.theme.Gio(), surface)
			lbl.Color = p.theme.Color.Text
			lbl.TextSize = p.focusedSentenceTextSize(gtx)
			return lbl.Layout(gtx)
		})
	}))
	if p.focusedFuriganaMode == focusedFuriganaBelow {
		children = append(children, layout.Rigid(bareutils.SpacerH(unit.Dp(2))), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutFocusedTokenSlot(gtx, unit.Dp(18), func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedTokenReading(gtx, reading)
			})
		}))
	}
	if p.focusedFuriganaMode == focusedFuriganaAbove {
		children = append(children, layout.Rigid(bareutils.SpacerH(unit.Dp(3))), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutFocusedTokenSlot(gtx, unit.Dp(14), func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusedTokenMarker(gtx, inFlashcards, dictionaryReady)
			})
		}))
	}
	return layout.Inset{Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			pointer.CursorPointer.Add(gtx.Ops)
			return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(5),
					Bottom: unit.Dp(5),
					Left:   unit.Dp(6),
					Right:  unit.Dp(6),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
				})
			})
		})
	})
}

func (p *Page) layoutFocusedTokenSlot(gtx layout.Context, height unit.Dp, w layout.Widget) layout.Dimensions {
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

func (p *Page) focusedTokenSurfaceSlotHeight(gtx layout.Context) unit.Dp {
	size := float32(p.focusedSentenceTextSize(gtx))
	return unit.Dp(size + 12)
}

func (p *Page) layoutFocusedTokenMarker(gtx layout.Context, inFlashcards, dictionaryReady bool) layout.Dimensions {
	text := " "
	fg := p.theme.Color.TextMuted
	if inFlashcards {
		text = "✓"
		fg = p.theme.Color.Success
	} else if dictionaryReady {
		text = "·"
		fg = p.theme.Color.Secondary
	}
	lbl := material.Body2(p.theme.Gio(), text)
	lbl.Color = fg
	lbl.TextSize = unit.Sp(12)
	return lbl.Layout(gtx)
}

func (p *Page) layoutFocusedTokenReading(gtx layout.Context, reading string) layout.Dimensions {
	if reading == "" {
		reading = " "
	}
	lbl := material.Body2(p.theme.Gio(), reading)
	lbl.Color = color.NRGBA{R: 255, G: 137, B: 103, A: 255}
	lbl.TextSize = p.translateDetailTextSize()
	return lbl.Layout(gtx)
}

func (p *Page) layoutFocusedFuriganaControls(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(p.theme.Gio(), "Furigana")
			lbl.Color = p.theme.Color.TextMuted
			lbl.TextSize = p.translateDetailTextSize()
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutFuriganaModeButton(gtx, &p.furiganaHiddenButton, focusedFuriganaHidden, "Hide")
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(6))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutFuriganaModeButton(gtx, &p.furiganaAboveButton, focusedFuriganaAbove, "Above")
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(6))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutFuriganaModeButton(gtx, &p.furiganaBelowButton, focusedFuriganaBelow, "Below")
		}),
	)
}

func (p *Page) layoutFocusedTokenActions(gtx layout.Context) layout.Dimensions {
	word := util.FirstNonEmpty(p.selectedFocusedTokenWord, p.selectedTranscriptText())
	meaning := "Click a word block above to inspect it."
	existingCard, hasExistingCard := p.focusedSelectedTokenFlashcard(word)
	if note := cleanInlineText(p.selectedFocusedTokenNote); note != "" {
		meaning = note
	} else if p.lookupResult != nil && cleanInlineText(p.lookupResult.Meaning) != "" {
		meaning = cleanInlineText(p.lookupResult.Meaning)
	} else if word != "" {
		if hasExistingCard && cleanInlineText(existingCard.Meaning) != "" {
			meaning = "Saved flashcard: " + cleanInlineText(existingCard.Meaning)
		}
	}
	addButton := bareui.Button{
		Clickable: &p.focusedTokenAddButton,
		Text:      "Add Flashcard",
		Prefix:    "mdi:plus-circle-outline",
		Variant:   bareui.ButtonSecondary,
	}
	audioButton := bareui.Button{
		Clickable: &p.focusedTokenAudioButton,
		Text:      "Play Audio",
		Prefix:    "mdi:volume-high",
		Variant:   bareui.ButtonSecondary,
	}
	return bareutils.RoundedSurface(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(9),
			Bottom: unit.Dp(9),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(p.theme.Gio(), meaning)
					lbl.Color = p.theme.Color.TextMuted
					lbl.TextSize = p.translateDetailTextSize()
					return lbl.Layout(gtx)
				}),
				//p.layoutStatusPill(gtx, contextVocabPillText(hasCard), hasCard),
				layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.lookupResult == nil || hasExistingCard {
						return addButton.Layout(gtx.Disabled(), p.theme, p.iconify)
					}
					return addButton.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					audioPath := ""
					if p.lookupResult != nil {
						audioPath = strings.TrimSpace(p.lookupResult.AudioPath)
					}
					if audioPath == "" && hasExistingCard {
						audioPath = strings.TrimSpace(existingCard.AudioPath)
					}
					if audioPath == "" {
						return audioButton.Layout(gtx.Disabled(), p.theme, p.iconify)
					}
					return audioButton.Layout(gtx, p.theme, p.iconify)
				}),
			)
		})
	})
}

func (p *Page) layoutFuriganaModeButton(gtx layout.Context, click *widget.Clickable, mode, label string) layout.Dimensions {
	active := p.focusedFuriganaMode == mode
	bg := p.theme.Color.Surface
	fg := p.theme.Color.TextMuted
	if active {
		bg = p.theme.Color.Primary
		fg = bareutils.ReadableOn(bg)
	} else if click.Hovered() {
		bg = p.theme.Color.SurfaceAlt
		fg = p.theme.Color.Text
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(6),
				Bottom: unit.Dp(6),
				Left:   unit.Dp(9),
				Right:  unit.Dp(9),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(p.theme.Gio(), label)
				lbl.Color = fg
				lbl.TextSize = p.translateDetailTextSize()
				return lbl.Layout(gtx)
			})
		})
	})
}

func (p *Page) layoutFocusedSentenceChips(gtx layout.Context, analysis japanese.Analysis, errText string) layout.Dimensions {
	if errText != "" {
		lbl := material.Body1(p.theme.Gio(), errText)
		lbl.Color = p.theme.Color.Warning
		lbl.TextSize = p.translateDetailTextSize()
		return lbl.Layout(gtx)
	}
	children := make([]layout.FlexChild, 0, min(4, len(analysis.Tokens)))
	for _, token := range focusTokens(analysis.Tokens, 4) {
		token := token
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Right: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutFocusChip(gtx, token)
			})
		}))
	}
	if len(children) == 0 {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
}

func (p *Page) layoutFocusChip(gtx layout.Context, token japanese.Token) layout.Dimensions {
	bg := barethemes.Mix(p.theme.Color.Primary, p.theme.Color.SurfaceAlt, 0.18)
	return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(9),
			Bottom: unit.Dp(9),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(p.theme.Gio(), posMajorLabel(token.POSMajor()))
					lbl.Color = p.theme.Color.TextMuted
					lbl.TextSize = p.translateDetailTextSize()
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), structureFlashcardWord(token))
					lbl.Color = p.theme.Color.Text
					lbl.TextSize = p.translateDetailTextSize()
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (p *Page) layoutFocusedLookupBar(gtx layout.Context) layout.Dimensions {
	selected := p.selectedTranscriptText()
	if selected == "" {
		selected = "Highlight a word in the focused sentence to look it up."
	}
	lookupButton := bareui.Button{
		Clickable: &p.focusedLookupButton,
		Text:      "Lookup Selection",
		Prefix:    "mdi:book-search-outline",
		Variant:   bareui.ButtonSecondary,
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(p.theme.Gio(), selected)
			lbl.Color = p.theme.Color.TextMuted
			lbl.TextSize = p.translateDetailTextSize()
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.selectedTranscriptText() == "" {
				return lookupButton.Layout(gtx.Disabled(), p.theme, p.iconify)
			}
			return lookupButton.Layout(gtx, p.theme, p.iconify)
		}),
	)
}

func (p *Page) layoutFocusedTranslationSection(gtx layout.Context) layout.Dimensions {
	p.syncTranslationEditor()
	return bareutils.RoundedSurface(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutTranslationHeader(gtx)
				}),
			}
			if !p.translationCollapsed {
				children = append(children,
					layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						editor := material.Editor(p.theme.Gio(), &p.translationEditor, "Type or edit the translation here")
						editor.Color = p.theme.Color.Text
						editor.HintColor = p.theme.Color.TextMuted
						editor.TextSize = p.translateDetailTextSize()
						maxHeight := min(gtx.Constraints.Max.Y, gtx.Dp(unit.Dp(96)))
						gtx.Constraints.Min.Y = min(maxHeight, gtx.Dp(unit.Dp(58)))
						gtx.Constraints.Max.Y = maxHeight
						return editor.Layout(gtx)
					}),
				)
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		})
	})
}

func (p *Page) layoutTranslationHeader(gtx layout.Context) layout.Dimensions {
	chevron := "mdi:chevron-down"
	if p.translationCollapsed {
		chevron = "mdi:chevron-right"
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.translationToggleButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				pointer.CursorPointer.Add(gtx.Ops)
				return layout.UniformInset(unit.Dp(4)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if p.iconify == nil {
						lbl := material.Body1(p.theme.Gio(), "+")
						lbl.Color = p.theme.Color.Text
						return lbl.Layout(gtx)
					}
					return p.iconify.Layout(gtx, chevron, unit.Dp(18), p.theme.Color.Text)
				})
			})
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(p.theme.Gio(), "Live Translation")
			lbl.Color = p.theme.Color.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(p.theme.Gio(), "to")
			lbl.Color = p.theme.Color.TextMuted
			lbl.TextSize = p.translateDetailTextSize()
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.targetLanguageDrop.Layout(gtx, p.theme, p.iconify, p.selectedTargetLanguage, func(gtx layout.Context) layout.Dimensions {
				return gui.LayoutOptionMenu(gtx, p.targetLanguageOptions, p.selectedTargetLanguage, p.theme, p.iconify)
			})
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			saveButton := bareui.Button{
				Clickable: &p.saveTranslationButton,
				Text:      "Save",
				Prefix:    "mdi:content-save-outline",
				Variant:   bareui.ButtonSecondary,
			}
			if strings.TrimSpace(p.translationEditor.Text()) == "" || strings.TrimSpace(p.structureSourceText()) == "" {
				return saveButton.Layout(gtx.Disabled(), p.theme, p.iconify)
			}
			return saveButton.Layout(gtx, p.theme, p.iconify)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(p.theme.Gio(), "Auto")
					lbl.Color = p.theme.Color.TextMuted
					lbl.TextSize = p.translateDetailTextSize()
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerW(unit.Dp(4))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					autoSwitch := material.Switch(p.theme.Gio(), &p.autoTranslateMissing, "Auto translate missing sentences")
					autoSwitch.Color.Enabled = p.theme.Color.Primary
					autoSwitch.Color.Disabled = p.theme.Color.Border
					return autoSwitch.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			text := "Generate"
			if p.translationGeneratingKey != "" {
				text = "Generating"
			}
			generateButton := bareui.Button{
				Clickable: &p.generateTranslationButton,
				Text:      text,
				Prefix:    "mdi:creation-outline",
				Variant:   bareui.ButtonPrimary,
			}
			if p.translationGeneratingKey != "" || strings.TrimSpace(p.structureSourceText()) == "" {
				return generateButton.Layout(gtx.Disabled(), p.theme, p.iconify)
			}
			return generateButton.Layout(gtx, p.theme, p.iconify)
		}),
	)
}

func (p *Page) layoutTranscriptCard(gtx layout.Context, bg color.NRGBA, child layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, bg, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, child)
	})
}

func (p *Page) layoutCardHeader(gtx layout.Context, title, hint string) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(p.theme.Gio(), title)
			lbl.Color = p.theme.Color.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.isCompactLayout(gtx) {
				return layout.Dimensions{}
			}
			lbl := material.Body2(p.theme.Gio(), hint)
			lbl.Color = p.theme.Color.TextMuted
			lbl.TextSize = p.translateDetailTextSize()
			return lbl.Layout(gtx)
		}),
	)
}

func (p *Page) focusedSentenceTextSize(gtx layout.Context) unit.Sp {
	if p.focusedTextSize > 0 {
		return p.focusedTextSize
	}
	if p.isCompactLayout(gtx) {
		return unit.Sp(20)
	}
	return unit.Sp(26)
}

func (p *Page) translateDetailTextSize() unit.Sp {
	if p.translateDetailSize > 0 {
		return p.translateDetailSize
	}
	return unit.Sp(15)
}

func (p *Page) layoutBottomActions(gtx layout.Context) layout.Dimensions {
	playButton := bareui.Button{
		Clickable: &p.playSentenceButton,
		Text:      "Play Audio",
		Prefix:    "mdi:volume-high",
		Variant:   bareui.ButtonSecondary,
	}
	structureButton := bareui.Button{
		Clickable: &p.translateSentenceButton,
		Text:      "Sentence Structure",
		Prefix:    "mdi:translate",
		Variant:   bareui.ButtonSecondary,
	}
	saveButton := bareui.Button{
		Clickable: &p.saveSentenceButton,
		Text:      "Save Sentence",
		Prefix:    "mdi:heart-outline",
		Variant:   bareui.ButtonSecondary,
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return playButton.Layout(gtx, p.theme, p.iconify) }),
		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return structureButton.Layout(gtx, p.theme, p.iconify) }),
		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return saveButton.Layout(gtx, p.theme, p.iconify) }),
	)
}

func (p *Page) layoutContextRail(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(5, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return p.layoutWordDetailsCard(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
		layout.Flexed(4, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = gtx.Constraints.Max
			return p.layoutChoicesCard(gtx)
		}),
	)
}

func (p *Page) layoutWordDetailsCard(gtx layout.Context) layout.Dimensions {
	card, hasCard := p.contextFlashcard()
	title := "Vocabulary"
	reading := "Select a saved word or run lookup"
	meaning := "Word details appear here while the transcript stays readable."
	meta := "No card selected"
	if hasCard {
		title = util.FirstNonEmpty(strings.TrimSpace(card.Text), strings.TrimSpace(p.popupWord), "Vocabulary")
		reading = util.FirstNonEmpty(strings.TrimSpace(card.Reading), strings.TrimSpace(card.PronunciationText), "No reading saved")
		meaning = util.FirstNonEmpty(strings.TrimSpace(card.Meaning), strings.TrimSpace(card.SourceLine), meaning)
		meta = util.FirstNonEmpty(p.flashcardMetaText(card), "Saved flashcard")
	} else if p.lookupResult != nil {
		title = util.FirstNonEmpty(p.lookupResult.Query, p.lookupResult.Headword, p.lookupResult.Key)
		reading = util.FirstNonEmpty(p.lookupResult.Reading, p.lookupResult.PronunciationText, "No reading found")
		meaning = util.FirstNonEmpty(p.lookupResult.Meaning, meaning)
		meta = "Dictionary lookup"
	}
	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
		playButton := bareui.Button{Clickable: &p.playAudioButton, Text: "mdi:volume-high", Icon: true, Variant: bareui.ButtonSecondary}
		//lookupButton := bareui.Button{Clickable: &p.searchWordButton, Text: "Lookup", Prefix: "mdi:book-search-outline", Variant: bareui.ButtonSecondary}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			//	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			//		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			//			lbl := material.H6(p.theme.Gio(), "Word Details")
			//			lbl.Color = p.theme.Color.Text
			//			return lbl.Layout(gtx)
			//		}),
			//		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			//			return p.layoutStatusPill(gtx, contextVocabPillText(hasCard), hasCard)
			//		}),
			//	)
			//}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(18))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.H5(p.theme.Gio(), title)
						lbl.Color = p.theme.Color.Text
						lbl.TextSize = unit.Sp(32)
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return playButton.Layout(gtx, p.theme, p.iconify)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(p.theme.Gio(), reading)
				lbl.Color = p.theme.Color.TextMuted
				lbl.TextSize = p.translateDetailTextSize()
				return lbl.Layout(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(p.theme.Gio(), meaning)
				lbl.Color = p.theme.Color.Text
				lbl.TextSize = p.translateDetailTextSize()
				return lbl.Layout(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(p.theme.Gio(), meta)
				lbl.Color = p.theme.Color.TextMuted
				lbl.TextSize = p.translateDetailTextSize()
				return lbl.Layout(gtx)
			}),
			//layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			//	return lookupButton.Layout(gtx, p.theme, p.iconify)
			//}),
		)
	})
}

func (p *Page) layoutChoicesCard(gtx layout.Context) layout.Dimensions {
	return p.layoutTranscriptCard(gtx, p.theme.Color.Background, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutCardHeader(gtx, "Dialog Choices", "Structure")
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutChoiceRow(gtx, "1", "Inspect sentence structure")
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutChoiceRow(gtx, "2", "Create a flashcard from selected text")
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutChoiceRow(gtx, "3", "Review saved transcript vocabulary")
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(p.theme.Gio(), "Highlight text in Focused Sentence, then use Lookup Selection or Word Details.")
				lbl.Color = p.theme.Color.TextMuted
				lbl.TextSize = p.translateDetailTextSize()
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (p *Page) layoutChoiceRow(gtx layout.Context, number, text string) layout.Dimensions {
	return bareutils.RoundedSurface(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(10),
			Bottom: unit.Dp(10),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutStatusPill(gtx, number, false)
				}),
				layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), text)
					lbl.Color = p.theme.Color.TextMuted
					lbl.TextSize = p.translateDetailTextSize()
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (p *Page) layoutTranscriptEditor(gtx layout.Context) layout.Dimensions {
	rows := p.transcriptRows()
	if len(rows) == 0 {
		return p.layoutTranscriptIdleState(gtx)
	}
	return material.List(p.theme.Gio(), &p.transcriptList).Layout(gtx, len(rows), func(gtx layout.Context, index int) layout.Dimensions {
		if index < 0 || index >= len(rows) {
			return layout.Dimensions{}
		}
		return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.layoutTranscriptRow(gtx, rows[index])
		})
	})
}

func (p *Page) layoutTranscriptRow(gtx layout.Context, row transcriptRow) layout.Dimensions {
	click := p.transcriptRowClickable(row.Key)
	selected := row.Key == p.currentTranscriptRowKey()
	bg := p.theme.Color.Surface
	fg := p.theme.Color.Text
	timeColor := p.theme.Color.TextMuted
	if selected {
		bg = barethemes.Mix(p.theme.Color.Primary, p.theme.Color.SurfaceAlt, 0.22)
		timeColor = p.theme.Color.Primary
	}
	if click.Hovered() && !selected {
		bg = p.theme.Color.SurfaceAlt
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(12),
				Bottom: unit.Dp(12),
				Left:   unit.Dp(14),
				Right:  unit.Dp(12),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(unit.Dp(78))
						lbl := material.Body2(p.theme.Gio(), row.Time)
						lbl.Color = timeColor
						return lbl.Layout(gtx)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(14))),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(p.theme.Gio(), row.Text)
						lbl.Color = fg
						lbl.TextSize = p.transcriptTextSize
						return lbl.Layout(gtx)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.layoutRowVocabIndicators(gtx, row.VocabWords)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.layoutRowIcon(gtx, "mdi:translate")
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.layoutRowIcon(gtx, "mdi:volume-high")
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return p.layoutRowIcon(gtx, "mdi:heart-outline")
					}),
				)
			})
		})
	})
}

func (p *Page) layoutRowVocabIndicators(gtx layout.Context, words []string) layout.Dimensions {
	if len(words) == 0 {
		return layout.Dimensions{}
	}
	visible := words
	if len(visible) > 2 {
		visible = visible[:2]
	}
	children := make([]layout.FlexChild, 0, len(visible)+1)
	for _, word := range visible {
		word := word
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutVocabChip(gtx, word)
			})
		}))
	}
	if extra := len(words) - len(visible); extra > 0 {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutVocabChip(gtx, fmt.Sprintf("+%d", extra))
			})
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (p *Page) layoutVocabChip(gtx layout.Context, text string) layout.Dimensions {
	bg := color.NRGBA{R: p.theme.Color.Primary.R, G: p.theme.Color.Primary.G, B: p.theme.Color.Primary.B, A: 34}
	return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(4),
			Bottom: unit.Dp(4),
			Left:   unit.Dp(7),
			Right:  unit.Dp(7),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(p.theme.Gio(), text)
			lbl.Color = p.theme.Color.Primary
			return lbl.Layout(gtx)
		})
	})
}

func (p *Page) layoutRowIcon(gtx layout.Context, icon string) layout.Dimensions {
	if p.iconify == nil {
		return layout.Dimensions{}
	}
	return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return bareutils.RoundedSurface(gtx, color.NRGBA{}, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.iconify.Layout(gtx, icon, unit.Dp(16), p.theme.Color.TextMuted)
			})
		})
	})
}

func (p *Page) layoutTranscriptLabel(gtx layout.Context, clr color.NRGBA, state *widget.Selectable) layout.Dimensions {
	label := material.Body1(p.theme.Gio(), p.displayTranscript)
	label.Color = clr
	label.TextSize = p.transcriptTextSize
	label.State = state
	return label.Layout(gtx)
}

func (p *Page) layoutTranscriptPopup(gtx layout.Context) layout.Dimensions {
	if p.popupFlashcard == nil {
		p.popupBounds = image.Rectangle{}
		return layout.Dimensions{}
	}

	card := *p.popupFlashcard
	popupWidth := gtx.Dp(unit.Dp(280))
	if popupWidth > gtx.Constraints.Max.X {
		popupWidth = gtx.Constraints.Max.X
	}
	if popupWidth <= 0 {
		return layout.Dimensions{}
	}
	popupHeightGuess := gtx.Dp(p.popupHeightGuess(card))

	x := p.popupAnchor.Min.X
	if x+popupWidth > gtx.Constraints.Max.X {
		x = gtx.Constraints.Max.X - popupWidth
	}
	if x < 0 {
		x = 0
	}

	y := p.popupAnchor.Min.Y - popupHeightGuess - gtx.Dp(unit.Dp(10))
	if y < 0 {
		y = p.popupAnchor.Max.Y + gtx.Dp(unit.Dp(10))
	}
	if y+popupHeightGuess > gtx.Constraints.Max.Y {
		y = max(0, gtx.Constraints.Max.Y-popupHeightGuess)
	}
	p.popupBounds = image.Rect(x, y, x+popupWidth, y+popupHeightGuess)

	p.layoutTranscriptPopupDismissRegions(gtx, p.popupBounds)

	offset := op.Offset(image.Pt(x, y)).Push(gtx.Ops)
	local := gtx
	local.Constraints.Min = image.Point{}
	local.Constraints.Max = image.Pt(popupWidth, popupHeightGuess)
	dims := p.layoutTranscriptPopupCard(local, card)
	offset.Pop()
	p.popupBounds = image.Rect(x, y, x+dims.Size.X, y+dims.Size.Y)
	return layout.Dimensions{}
}

func (p *Page) layoutTranscriptPopupCard(gtx layout.Context, card flashcards.Flashcard) layout.Dimensions {
	titleText := util.FirstNonEmpty(strings.TrimSpace(card.Text), strings.TrimSpace(p.popupWord), strings.TrimSpace(card.Reading), "Vocabulary")
	bodyText := util.FirstNonEmpty(strings.TrimSpace(card.Meaning), strings.TrimSpace(card.SourceLine), "No saved meaning for this word yet.")
	audioButton := bareui.Button{
		Clickable: &p.transcriptPopupAudioButton,
		Text:      "Play Audio",
		Prefix:    "mdi:play-circle-outline",
		Variant:   bareui.ButtonSecondary,
	}
	closeButton := bareui.Button{
		Clickable: &p.transcriptPopupCloseButton,
		Text:      "mdi:close",
		Icon:      true,
		Prefix:    "mdi:close",
		Variant:   bareui.ButtonGhost,
	}
	borderColor := transcriptPopupBorderColor(p.theme.Color.Primary)
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									lbl := material.Body1(p.theme.Gio(), "Vocabulary")
									lbl.Color = p.theme.Color.TextMuted
									return lbl.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return closeButton.Layout(gtx, p.theme, p.iconify)
								}),
							)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(p.theme.Gio(), titleText)
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(p.theme.Gio(), bodyText)
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							meta := p.flashcardMetaText(card)
							if meta == "" {
								return layout.Dimensions{}
							}
							return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Body1(p.theme.Gio(), meta)
								lbl.Color = p.theme.Color.TextMuted
								return lbl.Layout(gtx)
							})
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if !util.IsExistingFile(card.AudioPath) {
								return layout.Dimensions{}
							}
							return layout.Inset{Top: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return audioButton.Layout(gtx, p.theme, p.iconify)
							})
						}),
					)
				})
			})
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			border := clip.Stroke{
				Path:  clip.RRect{Rect: image.Rectangle{Max: gtx.Constraints.Max}, NW: gtx.Dp(unit.Dp(p.theme.Radius.MD)), NE: gtx.Dp(unit.Dp(p.theme.Radius.MD)), SW: gtx.Dp(unit.Dp(p.theme.Radius.MD)), SE: gtx.Dp(unit.Dp(p.theme.Radius.MD))}.Path(gtx.Ops),
				Width: float32(gtx.Dp(unit.Dp(1))),
			}.Op()
			paint.FillShape(gtx.Ops, borderColor, border)
			return layout.Dimensions{}
		}),
	)
}

func (p *Page) layoutTranscriptPopupDismissRegions(gtx layout.Context, popup image.Rectangle) {
	regions := [4]image.Rectangle{
		image.Rect(0, 0, gtx.Constraints.Max.X, popup.Min.Y),
		image.Rect(0, popup.Min.Y, popup.Min.X, popup.Max.Y),
		image.Rect(popup.Max.X, popup.Min.Y, gtx.Constraints.Max.X, popup.Max.Y),
		image.Rect(0, popup.Max.Y, gtx.Constraints.Max.X, gtx.Constraints.Max.Y),
	}
	for i, region := range regions {
		if region.Empty() {
			continue
		}
		offset := op.Offset(region.Min).Push(gtx.Ops)
		local := gtx
		local.Constraints.Min = region.Size()
		local.Constraints.Max = region.Size()
		p.popupDismissClicks[i].Layout(local, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: region.Size()}
		})
		offset.Pop()
	}
}

func (p *Page) popupHeightGuess(card flashcards.Flashcard) unit.Dp {
	bodyText := util.FirstNonEmpty(strings.TrimSpace(card.Meaning), strings.TrimSpace(card.SourceLine), "No saved meaning for this word yet.")
	height := 92
	height += min(3, 1+strings.Count(bodyText, "\n")) * 18
	if meta := p.flashcardMetaText(card); meta != "" {
		height += min(3, 1+strings.Count(meta, "\n")) * 14
	}
	if util.IsExistingFile(card.AudioPath) {
		height += 34
	}
	if height < 112 {
		height = 112
	}
	if height > 168 {
		height = 168
	}
	return unit.Dp(height)
}

func (p *Page) shouldCollapseFlashcardComposer() bool {
	if p.selectedTranscriptText() != "" {
		return false
	}
	if strings.TrimSpace(p.wordEditor.Text()) != "" {
		return false
	}
	if strings.TrimSpace(p.meaningEditor.Text()) != "" {
		return false
	}
	return len(p.lookupResults) == 0
}

func (p *Page) resetFlashcardComposer() {
	p.wordEditor.SetText("")
	p.meaningEditor.SetText("")
	p.hideReadingInAnki.Value = false
	p.lastAutoWord = ""
	p.hideReadingSet = false
	p.lookupResult = nil
	p.lookupResults = nil
	p.composerMinimized = true
	p.composerLastUsed = time.Now()
}

func (p *Page) syncComposerMinimized() {
	if p.composerHasActiveContent() {
		p.composerMinimized = false
		p.composerLastUsed = time.Now()
		return
	}
	if p.shouldCollapseFlashcardComposer() && time.Since(p.composerLastUsed) > 4*time.Second {
		p.composerMinimized = true
	}
}

func (p *Page) composerHasActiveContent() bool {
	if p.selectedTranscriptText() != "" {
		return true
	}
	if strings.TrimSpace(p.wordEditor.Text()) != "" {
		return true
	}
	if strings.TrimSpace(p.meaningEditor.Text()) != "" {
		return true
	}
	return len(p.lookupResults) > 0
}

func (p *Page) syncHideReadingDefault() {
	word := strings.TrimSpace(p.wordEditor.Text())
	if word == p.lastAutoWord {
		return
	}
	p.lastAutoWord = word
	if p.hideReadingSet {
		return
	}
	p.hideReadingInAnki.Value = util.ContainsKanji(word)
}

func (p *Page) layoutTranscriptIdleState(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H6(p.theme.Gio(), "Transcript Hidden")
				lbl.Color = p.theme.Color.Text
				return lbl.Layout(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(p.theme.Gio(), "Start the game to show live transcript text here.")
				lbl.Color = p.theme.Color.TextMuted
				return lbl.Layout(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(p.theme.Gio(), "The flashcard composer stays on this page next to the transcript.")
				lbl.Color = p.theme.Color.TextMuted
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (p *Page) layoutFlashcardComposer(gtx layout.Context) layout.Dimensions {
	p.syncComposerMinimized()
	if p.composerMinimized {
		return p.layoutFlashcardComposerMini(gtx)
	}
	if p.shouldCollapseFlashcardComposer() {
		return p.layoutFlashcardComposerHint(gtx)
	}

	minimizeButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-down", Icon: true, Prefix: "mdi:chevron-down", Variant: bareui.ButtonGhost}

	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutComposerHeader(gtx, &minimizeButton)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					switch p.composerFocus {
					case composerFocusSentenceStructure:
						return p.layoutSentenceStructurePanel(gtx, false)
					default:
						return p.layoutFlashcardComposerForm(gtx)
					}
				}),
			)
		})
	})
}

func (p *Page) layoutFlashcardComposerDocked(gtx layout.Context) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutComposerHeader(gtx, nil)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min = gtx.Constraints.Max
					switch p.composerFocus {
					case composerFocusSentenceStructure:
						return p.layoutSentenceStructurePanel(gtx, true)
					default:
						return p.layoutFlashcardComposerForm(gtx)
					}
				}),
			)
		})
	})
}

func (p *Page) layoutComposerHeader(gtx layout.Context, action *bareui.Button) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutComposerFocusTabs(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if action == nil {
				return layout.Dimensions{}
			}
			return layout.Inset{Left: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return action.Layout(gtx, p.theme, p.iconify)
			})
		}),
	)
}

func (p *Page) layoutFlashcardComposerForm(gtx layout.Context) layout.Dimensions {
	word := material.Editor(p.theme.Gio(), &p.wordEditor, "Word or phrase")
	word.Color = p.theme.Color.Text
	word.HintColor = p.theme.Color.TextMuted
	meaning := material.Editor(p.theme.Gio(), &p.meaningEditor, "Meaning")
	meaning.Color = p.theme.Color.Text
	meaning.HintColor = p.theme.Color.TextMuted
	//hideReadingCheck := material.CheckBox(p.theme.Gio(), &p.hideReadingInAnki, "Hide reading/furigana in Anki for this card")
	//hideReadingCheck.Color = p.theme.Color.Text

	searchButton := bareui.Button{Clickable: &p.searchWordButton, Text: "Lookup", Prefix: "mdi:book-search-outline", Variant: bareui.ButtonSecondary}
	playButton := bareui.Button{Clickable: &p.playAudioButton, Text: "mdi:play-circle-outline", Icon: true, Prefix: "mdi:play-circle-outline", Variant: bareui.ButtonSecondary}
	addAllButton := bareui.Button{Clickable: &p.addAllLookupButton, Text: "Add All Matches", Prefix: "mdi:playlist-plus", Variant: bareui.ButtonSecondary}

	selected := p.selectedTranscriptText()
	if selected == "" {
		selected = "Select focused sentence or transcript text to prefill the flashcard word."
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(p.theme.Gio(), "New Flashcard")
			lbl.Color = p.theme.Color.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(p.theme.Gio(), selected)
			lbl.Color = p.theme.Color.TextMuted
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
		layout.Rigid(word.Layout),
		layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			minHeight := unit.Dp(120)
			if p.isCompactLayout(gtx) {
				minHeight = unit.Dp(102)
			}
			gtx.Constraints.Min.Y = gtx.Dp(minHeight)
			return meaning.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return searchButton.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return playButton.Layout(gtx, p.theme, p.iconify)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(p.lookupResults) <= 1 {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return addAllButton.Layout(gtx, p.theme, p.iconify)
			})
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if len(p.lookupResults) == 0 {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(14)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutLookupResults(gtx)
			})
		}),
	)
}

func (p *Page) layoutSentenceStructurePanel(gtx layout.Context, fillHeight bool) layout.Dimensions {
	analysis, errText := p.currentStructureAnalysis()
	if fillHeight {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	} else if p.isCompactLayout(gtx) {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(unit.Dp(320)))
	} else {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(unit.Dp(380)))
	}
	if gtx.Constraints.Max.Y <= 0 {
		gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(260))
	}

	items := 1 + len(analysis.Tokens)
	if len(analysis.Particles) > 0 {
		items++
	}
	return material.List(p.theme.Gio(), &p.structureList).Layout(gtx, items, func(gtx layout.Context, index int) layout.Dimensions {
		switch {
		case index == 0:
			return p.layoutStructureSummary(gtx, analysis, errText)
		case len(analysis.Particles) > 0 && index == 1:
			return layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutParticleSummary(gtx, analysis.Particles)
			})
		default:
			tokenIndex := index - 1
			if len(analysis.Particles) > 0 {
				tokenIndex--
			}
			if tokenIndex < 0 || tokenIndex >= len(analysis.Tokens) {
				return layout.Dimensions{}
			}
			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.layoutStructureToken(gtx, analysis.Tokens[tokenIndex])
			})
		}
	})
}

func (p *Page) layoutStructureSummary(gtx layout.Context, analysis japanese.Analysis, errText string) layout.Dimensions {
	return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H6(p.theme.Gio(), "Sentence Structure")
				lbl.Color = p.theme.Color.Text
				return lbl.Layout(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := strings.TrimSpace(analysis.Sentence)
				if text == "" {
					text = "Select transcript text, or enter a flashcard word, to inspect sentence structure."
				}
				if errText != "" {
					text = errText
				}
				lbl := material.Body1(p.theme.Gio(), text)
				lbl.Color = p.theme.Color.TextMuted
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (p *Page) layoutParticleSummary(gtx layout.Context, particles []japanese.Token) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), "Particles")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			}
			for _, particle := range particles {
				particle := particle
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(p.theme.Gio(), particle.Surface+" - "+particleRole(particle.Surface))
						lbl.Color = p.theme.Color.TextMuted
						return lbl.Layout(gtx)
					})
				}))
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		})
	})
}

func (p *Page) layoutStructureToken(gtx layout.Context, token japanese.Token) layout.Dimensions {
	existingCard, hasExistingCard := p.structureTokenFlashcard(token)
	addButton := bareui.Button{
		Clickable: p.structureTokenAddClickable(structureTokenKey(token)),
		Text:      "mdi:plus-circle-outline",
		Icon:      true,
		Variant:   bareui.ButtonPrimary,
	}
	playButton := bareui.Button{
		Clickable: p.structureTokenPlayClickable(structureTokenKey(token)),
		Text:      "mdi:play-circle-outline",
		Icon:      true,
		Variant:   bareui.ButtonSecondary,
	}
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(p.theme.Gio(), token.Surface)
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(p.theme.Gio(), posMajorLabel(token.POSMajor()))
							lbl.Color = p.theme.Color.Primary
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if hasExistingCard {
								if strings.TrimSpace(existingCard.AudioPath) == "" {
									return layout.Dimensions{}
								}
								return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									return playButton.Layout(gtx, p.theme, p.iconify)
								})
							}
							if !canCreateStructureFlashcard(token) {
								return layout.Dimensions{}
							}
							return layout.Inset{Left: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return addButton.Layout(gtx, p.theme, p.iconify)
							})
						}),
					)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), tokenDetailText(token))
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !hasExistingCard || strings.TrimSpace(existingCard.Meaning) == "" {
						return layout.Dimensions{}
					}
					return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(p.theme.Gio(), existingCard.Meaning)
						lbl.Color = p.theme.Color.Text
						return lbl.Layout(gtx)
					})
				}),
			)
		})
	})
}

func (p *Page) currentStructureAnalysis() (japanese.Analysis, string) {
	text := p.structureSourceText()
	if text == "" {
		p.structureCacheKey = ""
		p.structureCache = japanese.Analysis{}
		p.structureCacheErr = ""
		return japanese.Analysis{}, ""
	}
	if text == p.structureCacheKey {
		return p.structureCache, p.structureCacheErr
	}
	analysis, err := japanese.AnalyzeSentence(text)
	p.structureCacheKey = text
	p.structureCache = analysis
	p.structureCacheErr = ""
	if err != nil {
		p.structureCache = japanese.Analysis{Sentence: text}
		p.structureCacheErr = err.Error()
	}
	return p.structureCache, p.structureCacheErr
}

func (p *Page) addStructureTokenFlashcard(key string) {
	if strings.TrimSpace(p.activeGameName) == "" {
		p.showError("Create Flashcard Failed", "Select a game before creating flashcards.")
		return
	}
	analysis, errText := p.currentStructureAnalysis()
	if errText != "" {
		p.showError("Create Flashcard Failed", errText)
		return
	}
	for _, token := range analysis.Tokens {
		if structureTokenKey(token) != key {
			continue
		}
		word := structureFlashcardWord(token)
		if word == "" {
			p.showError("Create Flashcard Failed", "This structure component cannot be turned into a flashcard.")
			return
		}
		lookups, err := dictionary.LookupWords(word)
		if err != nil {
			p.showError("Create Flashcard Failed", err.Error())
			return
		}
		if len(lookups) == 0 {
			p.showError("Create Flashcard Failed", "No dictionary matches were found for "+word+".")
			return
		}
		card := p.flashcardFromLookup(lookups[0])
		card.SourceLine = analysis.Sentence
		if err := flashcards.AddFlashcard(card); err != nil {
			p.showError("Create Flashcard Failed", err.Error())
			return
		}
		_ = p.ReloadFlashcards()
		p.showNotification("Flashcard Created", word+" was added from sentence structure.", guitoast.NotificationTypeSuccess)
		return
	}
}

func (p *Page) structureTokenAddClickable(key string) *widget.Clickable {
	if p.structureTokenAddClicks == nil {
		p.structureTokenAddClicks = make(map[string]*widget.Clickable)
	}
	if p.structureTokenAddClicks[key] == nil {
		p.structureTokenAddClicks[key] = new(widget.Clickable)
	}
	return p.structureTokenAddClicks[key]
}

func (p *Page) playStructureTokenAudio(key string) {
	tokenCard, ok := p.structureTokenFlashcardByKey(key)
	if !ok {
		return
	}
	if strings.TrimSpace(tokenCard.AudioPath) == "" {
		p.showError("Audio Playback Failed", "No audio is available for this flashcard.")
		return
	}
	if err := playFlashcardAudio(tokenCard); err != nil {
		p.showError("Audio Playback Failed", err.Error())
	}
}

func (p *Page) structureTokenPlayClickable(key string) *widget.Clickable {
	if p.structureTokenPlayClicks == nil {
		p.structureTokenPlayClicks = make(map[string]*widget.Clickable)
	}
	if p.structureTokenPlayClicks[key] == nil {
		p.structureTokenPlayClicks[key] = new(widget.Clickable)
	}
	return p.structureTokenPlayClicks[key]
}

func (p *Page) focusedTokenClickable(key string) *widget.Clickable {
	if p.focusedTokenClicks == nil {
		p.focusedTokenClicks = make(map[string]*widget.Clickable)
	}
	if p.focusedTokenClicks[key] == nil {
		p.focusedTokenClicks[key] = new(widget.Clickable)
	}
	return p.focusedTokenClicks[key]
}

func (p *Page) pruneFocusedTokenClicks(tokens []japanese.Token) {
	valid := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		valid[structureTokenKey(token)] = struct{}{}
	}
	for key := range p.focusedTokenClicks {
		if _, ok := valid[key]; !ok {
			delete(p.focusedTokenClicks, key)
		}
	}
	if p.selectedFocusedTokenKey != "" {
		if _, ok := valid[p.selectedFocusedTokenKey]; !ok {
			p.selectedFocusedTokenKey = ""
			p.selectedFocusedTokenWord = ""
			p.selectedFocusedTokenNote = ""
		}
	}
}

func (p *Page) selectFocusedToken(key string) {
	analysis, errText := p.currentStructureAnalysis()
	if errText != "" {
		p.showError("Dictionary Lookup Failed", errText)
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
		p.selectedFocusedTokenKey = key
		p.selectedFocusedTokenWord = word
		p.selectedFocusedTokenNote = ""
		p.wordEditor.SetText(word)
		p.meaningEditor.SetText("")
		p.lookupResult = nil
		p.lookupResults = nil
		if isParticleToken(token) {
			note := particleRole(token.Surface)
			p.selectedFocusedTokenNote = note
			p.meaningEditor.SetText(note)
			return
		}
		lookups, err := dictionary.LookupWords(word)
		if err != nil {
			p.showError("Dictionary Lookup Failed", err.Error())
			return
		}
		if len(lookups) == 0 {
			p.showError("Dictionary Lookup Failed", "No dictionary matches were found for "+word+".")
			return
		}
		p.lookupResults = lookups
		p.lookupResult = &lookups[0]
		p.meaningEditor.SetText(lookups[0].Meaning)
		return
	}
}

func (p *Page) addFocusedTokenFlashcard() {
	if p.lookupResult == nil {
		p.showError("Create Flashcard Failed", "Click a word block before adding a flashcard.")
		return
	}
	if _, ok := p.focusedSelectedTokenFlashcard(p.selectedFocusedTokenWord); ok {
		p.showNotification("Flashcard Exists", p.selectedFocusedTokenWord+" is already in your flashcards.", guitoast.NotificationTypeInfo)
		return
	}
	card := p.flashcardFromLookup(*p.lookupResult)
	if err := flashcards.AddFlashcard(card); err != nil {
		p.showError("Create Flashcard Failed", err.Error())
		return
	}
	_ = p.ReloadFlashcards()
	p.showNotification("Flashcard Created", card.Text+" was added.", guitoast.NotificationTypeSuccess)
}

func (p *Page) playFocusedTokenAudio() {
	if p.lookupResult != nil && strings.TrimSpace(p.lookupResult.AudioPath) != "" {
		if err := dictionary.PlayLookupAudio(*p.lookupResult); err != nil {
			p.showError("Audio Playback Failed", err.Error())
		}
		return
	}
	card, ok := p.focusedSelectedTokenFlashcard(p.selectedFocusedTokenWord)
	if !ok || strings.TrimSpace(card.AudioPath) == "" {
		p.showError("Audio Playback Failed", "No audio is available for the selected word.")
		return
	}
	if err := playFlashcardAudio(card); err != nil {
		p.showError("Audio Playback Failed", err.Error())
	}
}

func (p *Page) focusedSelectedTokenFlashcard(word string) (flashcards.Flashcard, bool) {
	if p.selectedFocusedTokenKey != "" {
		if card, ok := p.structureTokenFlashcardByKey(p.selectedFocusedTokenKey); ok {
			return card, true
		}
	}
	return p.flashcardForWordExact(word)
}

func (p *Page) structureTokenFlashcardByKey(key string) (flashcards.Flashcard, bool) {
	analysis, _ := p.currentStructureAnalysis()
	for _, token := range analysis.Tokens {
		if structureTokenKey(token) != key {
			continue
		}
		return p.structureTokenFlashcard(token)
	}
	return flashcards.Flashcard{}, false
}

func (p *Page) structureTokenFlashcard(token japanese.Token) (flashcards.Flashcard, bool) {
	candidates := structureTokenFlashcardCandidates(token)
	if len(candidates) == 0 {
		return flashcards.Flashcard{}, false
	}
	for _, card := range p.flashcards {
		cardWords := []string{card.Text, card.Reading, card.PronunciationText}
		for _, cardWord := range cardWords {
			cardWord = normalizeStructureMatchText(cardWord)
			if cardWord == "" {
				continue
			}
			for _, candidate := range candidates {
				if cardWord == candidate {
					return card, true
				}
			}
		}
	}
	return flashcards.Flashcard{}, false
}

func (p *Page) structureSourceText() string {
	selected := p.selectedTranscriptText()
	if selected != "" {
		if sentence := japanese.ExtractSenNtence(p.displayTranscript, selected); sentence != "" {
			return cleanTranscriptFocusText(sentence)
		}
		return cleanTranscriptFocusText(selected)
	}
	if p.selectedLineKey != "" {
		if selected := p.transcriptFocusTextForKey(p.selectedLineKey); selected != "" {
			return selected
		}
	}
	word := normalizeSelectionText(p.wordEditor.Text())
	if word != "" {
		if sentence := findFlashcardSourceLine(p.displayTranscript, word); sentence != "" {
			return cleanTranscriptFocusText(sentence)
		}
		return cleanInlineText(word)
	}
	if latest := p.transcriptFocusTextForKey(""); latest != "" {
		return latest
	}
	return ""
}

func (p *Page) translationCacheKey() string {
	source := p.structureSourceText()
	if strings.TrimSpace(source) == "" || strings.TrimSpace(p.selectedTargetLanguage) == "" {
		return ""
	}
	return strings.TrimSpace(p.activeGameName) + "\x00" + cleanInlineText(source) + "\x00" + strings.ToLower(strings.TrimSpace(p.selectedTargetLanguage))
}

func (p *Page) syncTranslationEditor() {
	key := p.translationCacheKey()
	if key == p.translationLoadedKey {
		return
	}
	p.translationLoadedKey = key
	if key == "" {
		p.translationEditor.SetText("")
		return
	}
	entry, ok, err := translation.Load(p.activeGameName, p.structureSourceText(), p.selectedTargetLanguage)
	if err != nil {
		p.showError("Translation Cache Failed", err.Error())
		return
	}
	if !ok {
		p.translationEditor.SetText("")
		return
	}
	p.translationEditor.SetText(entry.Translation)
}

func (p *Page) maybeAutoGenerateTranslation(ctx context.Context, w *app.Window) {
	if !p.autoTranslateMissing.Value || p.translationGeneratingKey != "" {
		return
	}
	key := p.translationCacheKey()
	if key == "" || key == p.autoTranslationAttemptKey {
		return
	}
	entry, ok, err := translation.Load(p.activeGameName, p.structureSourceText(), p.selectedTargetLanguage)
	if err != nil {
		p.autoTranslationAttemptKey = key
		p.showError("Translation Cache Failed", err.Error())
		return
	}
	if ok {
		p.translationLoadedKey = key
		p.translationEditor.SetText(entry.Translation)
		return
	}
	p.autoTranslationAttemptKey = key
	p.generateCurrentTranslation(ctx, w)
}

func (p *Page) saveCurrentTranslation() {
	source := p.structureSourceText()
	if strings.TrimSpace(source) == "" {
		p.showError("Save Translation Failed", "There is no focused sentence to save.")
		return
	}
	entry := translation.Entry{
		GameName:       p.activeGameName,
		SourceText:     source,
		TargetLanguage: p.selectedTargetLanguage,
		Translation:    p.translationEditor.Text(),
	}
	if err := translation.Save(entry); err != nil {
		p.showError("Save Translation Failed", err.Error())
		return
	}
	p.translationLoadedKey = p.translationCacheKey()
	p.autoTranslationAttemptKey = p.translationLoadedKey
	p.showNotification("Translation Saved", "Saved translation for "+p.selectedTargetLanguage+".", guitoast.NotificationTypeSuccess)
}

func (p *Page) generateCurrentTranslation(ctx context.Context, w *app.Window) {
	source := p.structureSourceText()
	if strings.TrimSpace(source) == "" {
		p.showError("Generate Translation Failed", "There is no focused sentence to translate.")
		return
	}
	key := p.translationCacheKey()
	if key == "" || p.translationGeneratingKey != "" {
		return
	}
	p.translationGeneratingKey = key
	p.translationEditor.SetText("Generating translation...")

	gameName := p.activeGameName
	targetLanguage := p.selectedTargetLanguage
	cfg := p.translatorConfig
	go func() {
		entry, err := translation.Generate(ctx, cfg, gameName, source, targetLanguage)
		result := translationResult{Key: key, Entry: entry, Err: err}
		select {
		case p.translationResultCh <- result:
		case <-ctx.Done():
		}
		if w != nil {
			w.Invalidate()
		}
	}()
}

func (p *Page) drainTranslationResults() {
	for {
		select {
		case result := <-p.translationResultCh:
			if result.Key != p.translationGeneratingKey {
				continue
			}
			p.translationGeneratingKey = ""
			if result.Key != p.translationCacheKey() {
				continue
			}
			if result.Err != nil {
				p.translationEditor.SetText("")
				p.showError("Generate Translation Failed", result.Err.Error())
				continue
			}
			p.translationLoadedKey = result.Key
			p.autoTranslationAttemptKey = result.Key
			p.translationEditor.SetText(result.Entry.Translation)
			p.showNotification("Translation Generated", "Generated and cached translation for "+result.Entry.TargetLanguage+".", guitoast.NotificationTypeSuccess)
		default:
			return
		}
	}
}

func (p *Page) transcriptFocusTextForKey(key string) string {
	rows := p.transcriptRows()
	if len(rows) == 0 {
		return ""
	}
	if strings.TrimSpace(key) != "" {
		for _, row := range rows {
			if row.Key == key {
				return cleanInlineText(row.Text)
			}
		}
		return ""
	}
	return cleanInlineText(rows[len(rows)-1].Text)
}

func (p *Page) transcriptRows() []transcriptRow {
	lines := strings.Split(strings.TrimSpace(p.displayTranscript), "\n")
	rows := make([]transcriptRow, 0, len(lines))
	var previousTimestamp string = unknownTimestamp
	for i, line := range lines {
		text := cleanInlineText(line)
		if text == "" {
			continue
		}
		timestamp, body := splitTranscriptTimestamp(text)
		if strings.HasPrefix(timestamp, "--") {
			timestamp = previousTimestamp
		} else {
			previousTimestamp = timestamp
		}
		body = cleanInlineText(body)
		key := fmt.Sprintf("%d:%s", i, text)
		rows = append(rows, transcriptRow{Key: key, Time: timestamp, Text: body, VocabWords: p.vocabWordsInText(body)})
	}
	p.pruneTranscriptRowClicks(rows)
	return rows
}

func (p *Page) vocabWordsInText(text string) []string {
	text = cleanInlineText(text)
	if text == "" || len(p.flashcards) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(p.flashcards))
	words := make([]string, 0, len(p.flashcards))
	for _, card := range p.flashcards {
		word := cleanInlineText(card.Text)
		if word == "" {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = struct{}{}
		words = append(words, word)
	}
	sort.SliceStable(words, func(i, j int) bool {
		return len([]rune(words[i])) > len([]rune(words[j]))
	})
	matches := flashcards.FindMatches(text, words)
	if len(matches) == 0 {
		return nil
	}
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if _, ok := seen[match.Word]; !ok {
			continue
		}
		out = append(out, match.Word)
		delete(seen, match.Word)
	}
	return out
}

func (p *Page) currentTranscriptRowKey() string {
	if p.selectedLineKey != "" {
		return p.selectedLineKey
	}
	rows := p.transcriptRows()
	if len(rows) == 0 {
		return ""
	}
	return rows[len(rows)-1].Key
}

func (p *Page) selectTranscriptRow(key string) {
	for _, row := range p.transcriptRows() {
		if row.Key != key {
			continue
		}
		p.selectedLineKey = row.Key
		p.selectedLineText = row.Text
		p.wordEditor.SetText("")
		p.meaningEditor.SetText("")
		p.lookupResult = nil
		p.lookupResults = nil
		p.DismissPopup()
		return
	}
}

func (p *Page) transcriptRowClickable(key string) *widget.Clickable {
	if p.transcriptRowClicks == nil {
		p.transcriptRowClicks = make(map[string]*widget.Clickable)
	}
	if p.transcriptRowClicks[key] == nil {
		p.transcriptRowClicks[key] = new(widget.Clickable)
	}
	return p.transcriptRowClicks[key]
}

func (p *Page) pruneTranscriptRowClicks(rows []transcriptRow) {
	valid := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		valid[row.Key] = struct{}{}
	}
	for key := range p.transcriptRowClicks {
		if _, ok := valid[key]; !ok {
			delete(p.transcriptRowClicks, key)
		}
	}
	if p.selectedLineKey != "" {
		if _, ok := valid[p.selectedLineKey]; !ok {
			p.selectedLineKey = ""
			p.selectedLineText = ""
		}
	}
}

const unknownTimestamp = "----:-- --:--:--"

func splitTranscriptTimestamp(line string) (string, string) {
	line = strings.TrimSpace(line)
	data, err := ParseLogLine(line)
	if err != nil {
		return unknownTimestamp, line
	}
	if data.RawTime == "" {
		return unknownTimestamp, line
	}
	return data.Time.Format("2006/01 15:04:05"), data.Text
}

func isClockTimestamp(value string) bool {
	if len(value) != 8 {
		return false
	}
	for i, r := range value {
		switch i {
		case 2, 5:
			if r != ':' {
				return false
			}
		default:
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func (p *Page) contextFlashcard() (flashcards.Flashcard, bool) {
	if p.popupFlashcard != nil {
		return *p.popupFlashcard, true
	}
	word := strings.TrimSpace(p.wordEditor.Text())
	if word == "" {
		word = strings.TrimSpace(p.popupWord)
	}
	return p.contextFlashcardForWord(word)
}

func (p *Page) contextFlashcardForWord(word string) (flashcards.Flashcard, bool) {
	word = strings.TrimSpace(word)
	if word != "" {
		if card, ok := p.flashcardForWordExact(word); ok {
			return card, true
		}
	}
	if len(p.flashcards) > 0 {
		return p.flashcards[0], true
	}
	return flashcards.Flashcard{}, false
}

func (p *Page) flashcardForWordExact(word string) (flashcards.Flashcard, bool) {
	word = normalizeStructureMatchText(word)
	if word == "" {
		return flashcards.Flashcard{}, false
	}
	for _, card := range p.flashcards {
		cardWords := []string{card.Text, card.Reading, card.PronunciationText}
		for _, cardWord := range cardWords {
			if normalizeStructureMatchText(cardWord) == word {
				return card, true
			}
		}
	}
	return flashcards.Flashcard{}, false
}

func contextVocabPillText(hasCard bool) string {
	if hasCard {
		return "In Vocab"
	}
	return "Lookup"
}

func newTranslationLanguageOptions() []gui.DropdownOption {
	labels := []string{
		"English",
		"Japanese",
		"Spanish",
		"French",
		"German",
		"Korean",
		"Chinese",
		"Italian",
		"Portuguese",
		"Russian",
	}
	options := make([]gui.DropdownOption, 0, len(labels))
	for _, label := range labels {
		options = append(options, gui.DropdownOption{
			Label:     label,
			Icon:      "mdi:translate",
			Clickable: new(widget.Clickable),
		})
	}
	return options
}

func tokenDetailText(token japanese.Token) string {
	parts := []string{token.POSLabel()}
	if base := strings.TrimSpace(token.BaseForm); base != "" && base != token.Surface {
		parts = append(parts, "base: "+base)
	}
	if reading := strings.TrimSpace(token.Reading); reading != "" {
		parts = append(parts, "reading: "+reading)
	}
	if inflection := token.InflectionLabel(); inflection != "" {
		parts = append(parts, "inflection: "+inflection)
	}
	if token.POSMajor() == "助詞" {
		parts = append(parts, "role: "+particleRole(token.Surface))
	}
	return strings.Join(parts, " | ")
}

func canCreateStructureFlashcard(token japanese.Token) bool {
	switch token.POSMajor() {
	case "名詞", "動詞", "形容詞":
		return structureFlashcardWord(token) != ""
	default:
		return false
	}
}

func structureFlashcardWord(token japanese.Token) string {
	switch token.POSMajor() {
	case "動詞", "形容詞":
		return util.FirstNonEmpty(strings.TrimSpace(token.BaseForm), strings.TrimSpace(token.Surface))
	default:
		return strings.TrimSpace(token.Surface)
	}
}

func structureTokenFlashcardCandidates(token japanese.Token) []string {
	raw := []string{
		token.Surface,
		token.BaseForm,
		token.Reading,
		token.Pronunciation,
		structureFlashcardWord(token),
	}
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, value := range raw {
		value = normalizeStructureMatchText(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func focusTokens(tokens []japanese.Token, limit int) []japanese.Token {
	if limit <= 0 {
		return nil
	}
	out := make([]japanese.Token, 0, limit)
	preferred := map[string]bool{
		"名詞":  true,
		"動詞":  true,
		"形容詞": true,
		"副詞":  true,
	}
	for _, token := range tokens {
		if !preferred[token.POSMajor()] || structureFlashcardWord(token) == "" {
			continue
		}
		out = append(out, token)
		if len(out) == limit {
			return out
		}
	}
	for _, token := range tokens {
		if structureFlashcardWord(token) == "" {
			continue
		}
		out = append(out, token)
		if len(out) == limit {
			return out
		}
	}
	return out
}

func focusedTokenReading(token japanese.Token) string {
	if !util.ContainsKanji(token.Surface) {
		return ""
	}
	reading := cleanInlineText(token.Reading)
	if reading == "" || reading == cleanInlineText(token.Surface) {
		return ""
	}
	return katakanaToHiragana(reading)
}

func focusedTokenDictionaryReady(token japanese.Token) bool {
	return canCreateStructureFlashcard(token)
}

func isParticleToken(token japanese.Token) bool {
	return token.POSMajor() == "助詞"
}

func normalizeFocusedFuriganaMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case focusedFuriganaAbove:
		return focusedFuriganaAbove
	case focusedFuriganaBelow:
		return focusedFuriganaBelow
	default:
		return focusedFuriganaHidden
	}
}

func focusedTokenColor(theme barethemes.Theme, token japanese.Token, selected, inFlashcards, dictionaryReady bool) color.NRGBA {
	if selected {
		return color.NRGBA{R: theme.Color.Primary.R, G: theme.Color.Primary.G, B: theme.Color.Primary.B, A: 88}
	}
	if inFlashcards {
		return color.NRGBA{R: theme.Color.Primary.R, G: theme.Color.Primary.G, B: theme.Color.Primary.B, A: 54}
	}
	if dictionaryReady {
		return color.NRGBA{R: theme.Color.SurfaceAlt.R, G: theme.Color.SurfaceAlt.G, B: theme.Color.SurfaceAlt.B, A: 210}
	}
	switch token.POSMajor() {
	case "名詞":
		return color.NRGBA{R: theme.Color.Secondary.R, G: theme.Color.Secondary.G, B: theme.Color.Secondary.B, A: 44}
	case "動詞":
		return color.NRGBA{R: theme.Color.Primary.R, G: theme.Color.Primary.G, B: theme.Color.Primary.B, A: 42}
	case "形容詞", "副詞":
		return color.NRGBA{R: theme.Color.Warning.R, G: theme.Color.Warning.G, B: theme.Color.Warning.B, A: 42}
	case "助詞", "助動詞":
		return color.NRGBA{R: theme.Color.Tertiary.R, G: theme.Color.Tertiary.G, B: theme.Color.Tertiary.B, A: 32}
	default:
		return color.NRGBA{R: theme.Color.SurfaceAlt.R, G: theme.Color.SurfaceAlt.G, B: theme.Color.SurfaceAlt.B, A: 180}
	}
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

func normalizeStructureMatchText(value string) string {
	return strings.TrimSpace(value)
}

func structureTokenKey(token japanese.Token) string {
	return strings.Join([]string{
		strings.TrimSpace(token.Surface),
		strings.TrimSpace(token.BaseForm),
		token.POSLabel(),
		token.InflectionLabel(),
	}, "\x00")
}

func posMajorLabel(pos string) string {
	switch pos {
	case "名詞":
		return "Noun"
	case "動詞":
		return "Verb"
	case "形容詞":
		return "Adjective"
	case "副詞":
		return "Adverb"
	case "助詞":
		return "Particle"
	case "助動詞":
		return "Auxiliary"
	case "連体詞":
		return "Prenoun"
	case "接続詞":
		return "Conjunction"
	case "感動詞":
		return "Interjection"
	case "記号":
		return "Symbol"
	default:
		if pos == "" {
			return "Token"
		}
		return pos
	}
}

func particleRole(surface string) string {
	switch strings.TrimSpace(surface) {
	case "は":
		return "topic marker; sets what the sentence is about, often with contrast"
	case "が":
		return "subject marker; identifies the doer or thing being described"
	case "を":
		return "direct object marker; marks what the action affects"
	case "に":
		return "target, destination, time, indirect object, or location of existence"
	case "へ":
		return "direction marker; points toward a destination"
	case "で":
		return "place or means of an action; marks where/how something happens"
	case "と":
		return "and/with, quotation, or comparison partner"
	case "も":
		return "also/even; adds the marked item to the statement"
	case "の":
		return "possession, modification, or nominalizer"
	case "から":
		return "from/since; starting point or cause"
	case "まで":
		return "until/to; endpoint or limit"
	case "より":
		return "than/from; comparison baseline or source"
	case "や":
		return "non-exhaustive and; examples from a set"
	case "か":
		return "question marker or alternative"
	case "ね":
		return "seeks agreement or softens with shared feeling"
	case "よ":
		return "assertive emphasis; presents information to the listener"
	case "な":
		return "prohibition, emotion, or sentence-ending emphasis depending on form"
	case "ぞ", "ぜ":
		return "strong sentence-ending emphasis"
	default:
		return "particle function depends on the surrounding phrase"
	}
}

func (p *Page) layoutComposerFocusTabs(gtx layout.Context) layout.Dimensions {
	for p.composerFlashcardsTab.Clicked(gtx) {
		p.composerFocus = composerFocusFlashcards
	}
	for p.composerSentenceTab.Clicked(gtx) {
		p.composerFocus = composerFocusSentenceStructure
	}
	if p.composerFocus == "" {
		p.composerFocus = composerFocusFlashcards
	}

	return bareutils.RoundedSurface(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(0),
			Bottom: unit.Dp(0),
			Left:   unit.Dp(0),
			Right:  unit.Dp(0),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutComposerFocusTab(gtx, &p.composerFlashcardsTab, composerFocusFlashcards, "Flashcards")
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutComposerFocusTab(gtx, &p.composerSentenceTab, composerFocusSentenceStructure, "Structure")
				}),
			)
		})
	})
}

func (p *Page) layoutComposerFocusTab(gtx layout.Context, click *widget.Clickable, id, label string) layout.Dimensions {
	active := p.composerFocus == id
	bg := p.theme.Color.SurfaceAlt
	fg := p.theme.Color.TextMuted
	if active {
		bg = p.theme.Color.Primary
		fg = bareutils.ReadableOn(bg)
	} else if click.Hovered() {
		bg = p.theme.Color.Surface
		fg = p.theme.Color.Text
	}

	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(10),
				Bottom: unit.Dp(10),
				Left:   unit.Dp(14),
				Right:  unit.Dp(14),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(p.theme.Gio(), label)
				lbl.Color = fg
				return lbl.Layout(gtx)
			})
		})
	})
}

func (p *Page) layoutFlashcardComposerHint(gtx layout.Context) layout.Dimensions {
	expandButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-up", Icon: true, Prefix: "mdi:chevron-up", Variant: bareui.ButtonGhost}
	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutComposerHeader(gtx, &expandButton)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					switch p.composerFocus {
					case composerFocusSentenceStructure:
						return p.layoutSentenceStructurePanel(gtx, false)
					default:
						return p.layoutFlashcardComposerHintText(gtx)
					}
				}),
			)
		})
	})
}

func (p *Page) layoutFlashcardComposerHintText(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(p.theme.Gio(), "New Flashcard")
			lbl.Color = p.theme.Color.Text
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(p.theme.Gio(), "Highlight transcript text to open the flashcard editor, or click a vocab match to inspect it.")
			lbl.Color = p.theme.Color.TextMuted
			return lbl.Layout(gtx)
		}),
	)
}

func (p *Page) layoutFlashcardComposerMini(gtx layout.Context) layout.Dimensions {
	expandButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-up", Icon: true, Prefix: "mdi:chevron-up", Variant: bareui.ButtonGhost}
	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(10),
			Bottom: unit.Dp(0),
			Left:   unit.Dp(14),
			Right:  unit.Dp(10),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.layoutComposerHeader(gtx, &expandButton)
		})
	})
}

func (p *Page) layoutLookupResults(gtx layout.Context) layout.Dimensions {
	maxHeight := gtx.Dp(unit.Dp(280))
	if p.isCompactLayout(gtx) {
		maxHeight = gtx.Dp(unit.Dp(240))
	}
	gtx.Constraints.Max.Y = maxHeight
	if gtx.Constraints.Min.Y > gtx.Constraints.Max.Y {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	return material.List(p.theme.Gio(), &p.lookupResultsList).Layout(gtx, len(p.lookupResults), func(gtx layout.Context, index int) layout.Dimensions {
		bottom := unit.Dp(0)
		if index < len(p.lookupResults)-1 {
			bottom = unit.Dp(10)
		}
		lookup := p.lookupResults[index]
		return layout.Inset{Bottom: bottom}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.layoutLookupResultCard(gtx, lookup)
		})
	})
}

func (p *Page) layoutLookupResultCard(gtx layout.Context, lookup dictionary.Lookup) layout.Dimensions {
	key := lookupResultKey(lookup)
	addButton := bareui.Button{Clickable: p.lookupResultAddClickable(key), Text: "mdi:plus-circle-outline", Icon: true, Variant: bareui.ButtonPrimary}
	playButton := bareui.Button{Clickable: p.lookupResultPlayClickable(key), Text: "mdi:play-circle-outline", Icon: true, Variant: bareui.ButtonSecondary}
	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), util.FirstNonEmpty(lookup.Query, lookup.Headword))
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if strings.TrimSpace(lookup.Reading) == "" {
						return layout.Dimensions{}
					}
					lbl := material.Body1(p.theme.Gio(), "Reading: "+lookup.Reading)
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), lookup.Meaning)
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return addButton.Layout(gtx, p.theme, p.iconify)
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							if strings.TrimSpace(lookup.AudioPath) == "" {
								return playButton.Layout(gtx.Disabled(), p.theme, p.iconify)
							}
							return playButton.Layout(gtx, p.theme, p.iconify)
						}),
					)
				}),
			)
		})
	})
}

func (p *Page) lookupCurrentWord() {
	if selected := p.selectedTranscriptText(); selected != "" {
		p.wordEditor.SetText(selected)
	}
	p.lookupResult = nil
	p.lookupResults = nil
	p.meaningEditor.SetText("")

	word := normalizeSelectionText(p.wordEditor.Text())
	if word == "" {
		p.showError("Dictionary Lookup Failed", "Flashcard word cannot be empty.")
		return
	}

	lookups, err := dictionary.LookupWords(word)
	if err != nil {
		p.showError("Dictionary Lookup Failed", err.Error())
		return
	}
	p.lookupResults = lookups
	p.lookupResult = &lookups[0]
	word = util.FirstNonEmpty(lookups[0].Query, lookups[0].Key, lookups[0].Headword)
	p.wordEditor.SetText(word)
	p.meaningEditor.SetText(lookups[0].Meaning)
	p.hideReadingSet = false
	p.lastAutoWord = ""
	p.syncHideReadingDefault()
}

func (p *Page) playCurrentLookupAudio() {
	if p.lookupResult == nil || strings.TrimSpace(p.lookupResult.AudioPath) == "" {
		p.showError("Audio Playback Failed", "No audio is available for the current lookup.")
		return
	}
	if err := dictionary.PlayLookupAudio(*p.lookupResult); err != nil {
		p.showError("Audio Playback Failed", err.Error())
	}
}

func (p *Page) addLookupFlashcardByKey(key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	for _, lookup := range p.lookupResults {
		if lookupResultKey(lookup) != key {
			continue
		}
		card := p.flashcardFromLookup(lookup)
		if err := flashcards.AddFlashcard(card); err != nil {
			p.showError("Create Flashcard Failed", err.Error())
			return
		}
		_ = p.ReloadFlashcards()
		p.resetFlashcardComposer()
		return
	}
}

func (p *Page) playLookupAudioByKey(key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	for _, lookup := range p.lookupResults {
		if lookupResultKey(lookup) != key {
			continue
		}
		if strings.TrimSpace(lookup.AudioPath) == "" {
			p.showError("Audio Playback Failed", "No audio is available for this lookup result.")
			return
		}
		if err := dictionary.PlayLookupAudio(lookup); err != nil {
			p.showError("Audio Playback Failed", err.Error())
		}
		return
	}
}

func (p *Page) addAllLookupFlashcards() {
	if strings.TrimSpace(p.activeGameName) == "" {
		p.showError("Create Flashcards Failed", "Select a game before creating flashcards.")
		return
	}
	if len(p.lookupResults) == 0 {
		p.showError("Create Flashcards Failed", "Run Dictionary Lookup first.")
		return
	}
	cards := make([]flashcards.Flashcard, 0, len(p.lookupResults))
	for _, lookup := range p.lookupResults {
		cards = append(cards, p.flashcardFromLookup(lookup))
	}
	if _, _, err := flashcards.AddFlashcards(p.activeGameName, cards); err != nil {
		p.showError("Create Flashcards Failed", err.Error())
		return
	}
	_ = p.ReloadFlashcards()
	p.resetFlashcardComposer()
}

func (p *Page) flashcardFromLookup(lookup dictionary.Lookup) flashcards.Flashcard {
	word := util.FirstNonEmpty(lookup.Query, lookup.Headword, lookup.Key)
	return flashcards.Flashcard{
		GameName:           p.activeGameName,
		Text:               word,
		Meaning:            lookup.Meaning,
		Reading:            lookup.Reading,
		PronunciationText:  lookup.PronunciationText,
		PronunciationPitch: lookup.Pitch,
		AudioPath:          lookup.AudioPath,
		SourcePath:         p.logPath,
		SourceLine:         findFlashcardSourceLine(p.displayTranscript, word),
	}
}

func (p *Page) syncCurrentGameToAnki() error {
	if strings.TrimSpace(p.activeGameName) == "" {
		return fmt.Errorf("select a game before syncing Anki")
	}
	client := anki.New(p.ankiURL)
	if _, err := client.SyncFlashcardsToAnki(p.activeGameName, p.ankiURL, p.pushSync); err != nil {
		return err
	}
	return p.ReloadFlashcards()
}

func (p *Page) launchCurrentGameInBackground() {
	if p.runnerStatus != nil {
		p.statusText = p.transcriptRunningStatusText()
		return
	}
	if p.currentConfig == nil || strings.TrimSpace(p.currentConfig.Name) == "" {
		p.showError("Launch Failed", "The selected game configuration is not loaded yet.")
		return
	}
	auto := runner.New()
	status, err := auto.RunBackground(p.currentConfig)
	if err != nil {
		p.statusText = err.Error()
		p.showError("Launch Failed", err.Error())
		return
	}
	p.runnerStatus = status
	p.statusText = fmt.Sprintf("Launching %s in the background.", p.currentConfig.Name)
}

func (p *Page) syncTranscriptEditor() {
	if p.lastSyncedText == p.displayTranscript {
		return
	}
	wasEmpty := p.lastSyncedText == ""
	p.transcriptView.SetText(p.displayTranscript)
	if wasEmpty {
		runes := len([]rune(p.displayTranscript))
		p.transcriptView.SetCaret(runes, runes)
	}
	p.lastSyncedText = p.displayTranscript
}

func (p *Page) syncFocusedSentenceView(text string) {
	if p.lastFocusedText == text {
		return
	}
	p.focusedSentenceView.SetText(text)
	p.lastFocusedText = text
}

func (p *Page) paintTranscriptHighlights(gtx layout.Context) {
	highlights := p.transcriptHighlights()
	if len(highlights) == 0 || strings.TrimSpace(p.displayTranscript) == "" {
		clear(p.transcriptHighlightBounds)
		return
	}
	colorModeEnabled := p.colorizeHighlights && len(highlights) <= 160
	var fill op.CallOp
	var colorText op.CallOp
	if colorModeEnabled {
		colorMacro := op.Record(gtx.Ops)
		p.layoutTranscriptLabel(gtx, p.theme.Color.Primary, nil)
		colorText = colorMacro.Stop()
	} else {
		colorMacro := op.Record(gtx.Ops)
		paint.ColorOp{Color: transcriptHighlightColor(p.theme.Color.Primary)}.Add(gtx.Ops)
		fill = colorMacro.Stop()
	}
	regions := make([]widget.Region, 0, 8)
	colorRects := make([]image.Rectangle, 0, len(highlights)*2)
	validClicks := make(map[string]struct{}, len(highlights))
	for _, match := range highlights {
		validClicks[match.Key] = struct{}{}
		if p.transcriptHighlightClicks[match.Key] == nil {
			p.transcriptHighlightClicks[match.Key] = new(widget.Clickable)
		}
		regions = p.transcriptView.Regions(match.StartRune, match.EndRune, regions[:0])
		var bounds image.Rectangle
		for idx, region := range regions {
			if idx == 0 {
				bounds = region.Bounds
			} else {
				bounds = bounds.Union(region.Bounds)
			}
		}
		if !bounds.Empty() {
			p.transcriptHighlightBounds[match.Key] = bounds
		}
		for _, region := range regions {
			if colorModeEnabled {
				colorRects = append(colorRects, region.Bounds)
			} else {
				stack := clip.Rect(region.Bounds).Push(gtx.Ops)
				fill.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				stack.Pop()
			}

			offset := op.Offset(image.Pt(region.Bounds.Min.X, region.Bounds.Min.Y)).Push(gtx.Ops)
			local := gtx
			local.Constraints.Min = region.Bounds.Size()
			local.Constraints.Max = region.Bounds.Size()
			p.transcriptHighlightClicks[match.Key].Layout(local, func(gtx layout.Context) layout.Dimensions {
				pointer.CursorPointer.Add(gtx.Ops)
				return layout.Dimensions{Size: region.Bounds.Size()}
			})
			offset.Pop()
		}
	}
	if colorModeEnabled && len(colorRects) > 0 {
		for _, rect := range mergeHighlightRects(colorRects) {
			stack := clip.Rect(rect).Push(gtx.Ops)
			colorText.Add(gtx.Ops)
			stack.Pop()
		}
	}
	for key := range p.transcriptHighlightClicks {
		if _, ok := validClicks[key]; !ok {
			delete(p.transcriptHighlightClicks, key)
			delete(p.transcriptHighlightBounds, key)
		}
	}
	if p.popupFlashcard != nil {
		found := false
		for _, match := range highlights {
			if p.popupMatchKey == match.Key {
				if bounds, ok := p.transcriptHighlightBounds[match.Key]; ok {
					p.popupAnchor = bounds
				}
				found = true
				break
			}
		}
		if !found {
			p.DismissPopup()
		}
	}
}

func (p *Page) transcriptHighlights() []flashcards.Match {
	cacheKey := p.highlightCacheKeyValue()
	if cacheKey == p.highlightCacheKey {
		return p.highlightCache
	}
	seen := make(map[string]flashcards.Flashcard, len(p.flashcards))
	words := make([]string, 0, len(p.flashcards))
	for _, card := range p.flashcards {
		word := strings.TrimSpace(card.Text)
		if word == "" {
			continue
		}
		if _, ok := seen[word]; ok {
			continue
		}
		seen[word] = card
		words = append(words, word)
	}
	sort.SliceStable(words, func(i, j int) bool {
		return len([]rune(words[i])) > len([]rune(words[j]))
	})
	matches := flashcards.FindMatches(p.displayTranscript, words)
	for i := range matches {
		matches[i].Card = seen[matches[i].Word]
		matches[i].Key = fmt.Sprintf("%s-%d-%d", util.SanitizeName(matches[i].Card.ID), matches[i].StartRune, matches[i].EndRune)
	}
	p.highlightCacheKey = cacheKey
	p.highlightCache = matches
	return p.highlightCache
}

func (p *Page) openTranscriptHighlightPopup(key string) {
	for _, match := range p.transcriptHighlights() {
		if match.Key != key {
			continue
		}
		if p.popupFlashcard != nil && p.popupMatchKey == match.Key {
			p.DismissPopup()
			return
		}
		cardCopy := match.Card
		p.popupFlashcard = &cardCopy
		p.popupAnchor = p.transcriptHighlightBounds[key]
		p.popupMatchKey = match.Key
		p.popupWord = match.Word
		if p.autoPlayHighlightAudio {
			_ = playFlashcardAudio(match.Card)
		}
		return
	}
}

func (p *Page) statusColor() color.NRGBA {
	status := strings.ToLower(p.statusText)
	switch {
	case strings.Contains(status, "failed"), strings.Contains(status, "error"):
		return p.theme.Color.Error
	case strings.Contains(status, "not found"):
		return p.theme.Color.Warning
	default:
		return p.theme.Color.TextMuted
	}
}

func (p *Page) isCompactLayout(gtx layout.Context) bool {
	return gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(compactWidth))
}

func (p *Page) shouldStackTranscriptPage(gtx layout.Context) bool {
	return gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(transcriptStackWidth))
}

func (p *Page) transcriptComposerWidth(gtx layout.Context) int {
	if gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(transcriptMediumWidth)) {
		return gtx.Dp(unit.Dp(360))
	}
	if gtx.Constraints.Max.X <= gtx.Dp(unit.Dp(transcriptStackWidth)) {
		return gtx.Dp(unit.Dp(380))
	}
	return gtx.Dp(unit.Dp(420))
}

func (p *Page) transcriptLaunchButtonLabel() string {
	if p.runnerStatus != nil {
		return "Game Running"
	}
	return "Launch Game"
}

func (p *Page) transcriptLaunchButtonIcon() string {
	if p.runnerStatus != nil {
		return "mdi:check-circle-outline"
	}
	return "mdi:play-box-outline"
}

func (p *Page) transcriptLaunchButtonVariant() bareui.ButtonVariant {
	if p.runnerStatus != nil {
		return bareui.ButtonSecondary
	}
	return bareui.ButtonPrimary
}

func (p *Page) transcriptRunningStatusText() string {
	if p.runnerStatus != nil {
		if p.runnerStatus.PID > 0 {
			return fmt.Sprintf("Detected running game process (pid %d).", p.runnerStatus.PID)
		}
		return "Detected running game process."
	}
	return "Game process not detected."
}

func (p *Page) flashcardMetaText(card flashcards.Flashcard) string {
	parts := make([]string, 0, 4)
	if furigana := strings.TrimSpace(card.Furigana()); furigana != "" {
		parts = append(parts, "Furigana: "+furigana)
	}
	if reading := strings.TrimSpace(card.Reading); reading != "" {
		parts = append(parts, "Reading: "+reading)
	}
	if pronunciation := strings.TrimSpace(card.PronunciationText); pronunciation != "" {
		if pitch := strings.TrimSpace(card.PronunciationPitch); pitch != "" {
			pronunciation += " (" + pitch + ")"
		}
		parts = append(parts, "Pronunciation: "+pronunciation)
	}
	if strings.TrimSpace(card.AudioPath) != "" {
		parts = append(parts, "Audio cached")
	}
	return strings.Join(parts, "\n")
}

func (p *Page) lookupResultAddClickable(key string) *widget.Clickable {
	if p.lookupResultAddClicks[key] == nil {
		p.lookupResultAddClicks[key] = new(widget.Clickable)
	}
	return p.lookupResultAddClicks[key]
}

func (p *Page) lookupResultPlayClickable(key string) *widget.Clickable {
	if p.lookupResultPlayClicks[key] == nil {
		p.lookupResultPlayClicks[key] = new(widget.Clickable)
	}
	return p.lookupResultPlayClicks[key]
}

func (p *Page) showError(title, body string) {
	if p.OnError != nil {
		p.OnError(title, body)
	}
}

func (p *Page) showNotification(title, body string, kind guitoast.NotificationType) {
	if p.OnNotify != nil {
		p.OnNotify(title, body, kind)
	}
}

func lookupResultKey(lookup dictionary.Lookup) string {
	return util.FirstNonEmpty(lookup.Key, lookup.Query, lookup.Headword)
}

func (p *Page) selectedTranscriptText() string {
	if selected := normalizeSelectionText(p.focusedSentenceView.SelectedText()); selected != "" {
		return cleanInlineText(selected)
	}
	if selected := normalizeSelectionText(p.transcriptView.SelectedText()); selected != "" {
		return cleanInlineText(selected)
	}
	return ""
}

func normalizeSelectionText(text string) string {
	return cleanInlineText(text)
}

func cleanInlineText(text string) string {
	text = strings.ReplaceAll(text, `\n`, " ")
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func cleanTranscriptFocusText(text string) string {
	parts := make([]string, 0, 1)
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		line = cleanInlineText(line)
		if line == "" {
			continue
		}
		_, body := splitTranscriptTimestamp(line)
		body = cleanInlineText(body)
		if body != "" {
			parts = append(parts, body)
		}
	}
	return cleanInlineText(strings.Join(parts, " "))
}

func findFlashcardSourceLine(transcriptText, word string) string {
	word = strings.TrimSpace(word)
	if word == "" {
		return ""
	}
	for _, line := range strings.Split(transcriptText, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, word) {
			return trimmed
		}
	}
	return ""
}

func sanitizeTranscriptForDisplay(text string) string {
	text = ansiRE.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, `\n`, "\n")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\t", "    ")
	text = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, text)
	return text
}

func limitTranscriptLines(text string, recentLineLimit int) string {
	if recentLineLimit <= 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	if len(lines) <= recentLineLimit {
		return text
	}
	return strings.Join(lines[len(lines)-recentLineLimit:], "\n")
}

func transcriptHighlightColor(base color.NRGBA) color.NRGBA {
	return color.NRGBA{R: base.R, G: base.G, B: base.B, A: 72}
}

func transcriptPopupBorderColor(base color.NRGBA) color.NRGBA {
	return color.NRGBA{R: base.R, G: base.G, B: base.B, A: 160}
}

func (p *Page) invalidateHighlights() {
	p.highlightCacheKey = ""
	p.highlightCache = nil
}

func (p *Page) highlightCacheKeyValue() string {
	var b strings.Builder
	b.Grow(len(p.displayTranscript) + len(p.flashcards)*24)
	b.WriteString(p.displayTranscript)
	b.WriteString("\x00")
	for _, card := range p.flashcards {
		b.WriteString(card.ID)
		b.WriteString("\x1f")
		b.WriteString(card.Text)
		b.WriteString("\x1e")
	}
	return b.String()
}

func mergeHighlightRects(rects []image.Rectangle) []image.Rectangle {
	if len(rects) <= 1 {
		return rects
	}
	sort.Slice(rects, func(i, j int) bool {
		if rects[i].Min.Y != rects[j].Min.Y {
			return rects[i].Min.Y < rects[j].Min.Y
		}
		return rects[i].Min.X < rects[j].Min.X
	})
	merged := make([]image.Rectangle, 0, len(rects))
	current := rects[0]
	for _, rect := range rects[1:] {
		if shouldMergeHighlightRect(current, rect) {
			current = current.Union(rect)
			continue
		}
		merged = append(merged, current)
		current = rect
	}
	merged = append(merged, current)
	return merged
}

func shouldMergeHighlightRect(a, b image.Rectangle) bool {
	if a.Empty() || b.Empty() {
		return false
	}
	if a.Min.Y > b.Max.Y || b.Min.Y > a.Max.Y {
		return false
	}
	return b.Min.X <= a.Max.X+6
}

func playFlashcardAudio(card flashcards.Flashcard) error {
	word := util.FirstNonEmpty(card.Text, card.Reading)
	if strings.TrimSpace(word) == "" {
		return fmt.Errorf("no audio is available for this flashcard")
	}
	return dictionary.PlayAudioForText(word)
}
