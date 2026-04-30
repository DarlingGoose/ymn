package transcript

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"os/exec"
	"regexp"
	"runtime"
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
	bareui "github.com/Seann-Moser/bare/pkg/ui"
	"github.com/Seann-Moser/bare/pkg/ui/icons"
	barethemes "github.com/Seann-Moser/bare/pkg/ui/themes"
	bareutils "github.com/Seann-Moser/bare/pkg/ui/utils"
	"github.com/Seann-Moser/wgl/pkg/anki"
	"github.com/Seann-Moser/wgl/pkg/dictionary"
	flashcards "github.com/Seann-Moser/wgl/pkg/flashcard"
	"github.com/Seann-Moser/wgl/pkg/game/gameconfig"
	"github.com/Seann-Moser/wgl/pkg/gui"
	"github.com/Seann-Moser/wgl/pkg/util"
)

const (
	compactWidth          = 1080
	transcriptStackWidth  = 1240
	transcriptMediumWidth = 1480
)

var _ gui.EvenHandler = &Page{}

var ansiRE = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|[@-Z\\-_])`)

type Page struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	transcriptView     widget.Selectable
	transcriptList     widget.List
	lookupResultsList  widget.List
	wordEditor         widget.Editor
	meaningEditor      widget.Editor
	searchWordButton   widget.Clickable
	playAudioButton    widget.Clickable
	addAllLookupButton widget.Clickable
	launchGameButton   widget.Clickable
	syncAnkiButton     widget.Clickable
	clearButton        widget.Clickable

	transcriptPopupAudioButton widget.Clickable
	transcriptPopupCloseButton widget.Clickable
	popupDismissClicks         [4]widget.Clickable
	composerToggleButton       widget.Clickable

	transcriptHighlightClicks map[string]*widget.Clickable
	transcriptHighlightBounds map[string]image.Rectangle
	lookupResultAddClicks     map[string]*widget.Clickable
	lookupResultPlayClicks    map[string]*widget.Clickable

	activeGameName         string
	logPath                string
	ankiURL                string
	pushSync               bool
	statusText             string
	gameRunning            bool
	gameRunningPID         int
	currentConfig          *gameconfig.GameConfig
	selectedTextSizeName   string
	selectedRecentLines    string
	transcriptTextSize     unit.Sp
	recentLineLimit        int
	autoPlayHighlightAudio bool
	colorizeHighlights     bool

	flashcards        []flashcards.Flashcard
	lookupResult      *dictionary.Lookup
	lookupResults     []dictionary.Lookup
	displayTranscript string
	lastSyncedText    string
	highlightCacheKey string
	highlightCache    []flashcards.Match
	popupFlashcard    *flashcards.Flashcard
	popupAnchor       image.Rectangle
	popupBounds       image.Rectangle
	popupMatchKey     string
	popupWord         string
	composerMinimized bool
	composerLastUsed  time.Time

	OnError func(title, body string)
}

func New(theme barethemes.Theme) *Page {
	p := &Page{
		theme:                     theme,
		pushSync:                  true,
		statusText:                "Start the game to show live transcript text here.",
		selectedTextSizeName:      "Medium",
		selectedRecentLines:       "All Lines",
		transcriptTextSize:        unit.Sp(16),
		composerMinimized:         true,
		composerLastUsed:          time.Now(),
		transcriptHighlightClicks: make(map[string]*widget.Clickable),
		transcriptHighlightBounds: make(map[string]image.Rectangle),
		lookupResultAddClicks:     make(map[string]*widget.Clickable),
		lookupResultPlayClicks:    make(map[string]*widget.Clickable),
	}
	p.wordEditor.SingleLine = true
	p.meaningEditor.SingleLine = false
	p.transcriptList.Axis = layout.Vertical
	p.transcriptList.ScrollToEnd = true
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

func (p *Page) SetContext(activeGameName, logPath, ankiURL string, cfg *gameconfig.GameConfig) *Page {
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
	p.gameRunning = running
	p.gameRunningPID = pid
	return p
}

func (p *Page) SetTranscriptOptions(textSize unit.Sp, textSizeName string, recentLineLimit int, recentLinesName string) *Page {
	p.transcriptTextSize = textSize
	p.selectedTextSizeName = strings.TrimSpace(textSizeName)
	p.recentLineLimit = recentLineLimit
	p.selectedRecentLines = strings.TrimSpace(recentLinesName)
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

func (p *Page) HandleEvents(gtx layout.Context, _ context.Context, _ *app.Window) {
	for p.launchGameButton.Clicked(gtx) {
		p.launchCurrentGameInBackground()
	}
	for p.syncAnkiButton.Clicked(gtx) {
		if err := p.syncCurrentGameToAnki(); err != nil {
			p.showError("Anki Sync Failed", err.Error())
		}
	}
	for p.clearButton.Clicked(gtx) {
		p.ClearTranscript()
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
	for key, click := range p.transcriptHighlightClicks {
		for click.Clicked(gtx) {
			p.openTranscriptHighlightPopup(key)
		}
	}
	for p.transcriptPopupAudioButton.Clicked(gtx) {
		if p.popupFlashcard == nil {
			p.showError("Audio Playback Failed", "No flashcard is selected.")
			continue
		}
		if err := playAudioFile(p.popupFlashcard.AudioPath); err != nil {
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
					Top:    unit.Dp(10),
					Left:   unit.Dp(20),
					Right:  unit.Dp(20),
					Bottom: unit.Dp(10),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					metaSpacing := unit.Dp(14)
					if !p.gameRunning {
						metaSpacing = 0
					}
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H5(p.theme.Gio(), util.FirstNonEmpty(p.activeGameName, "No game selected"))
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
						//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						//	lbl := material.Body1(p.theme.Gio(), util.FirstNonEmpty(p.logPath, "No transcript path resolved"))
						//	lbl.Color = p.theme.Color.TextMuted
						//	return lbl.Layout(gtx)
						//}),
						//layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(p.theme.Gio(), p.statusText)
							lbl.Color = p.statusColor()
							return lbl.Layout(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return p.layoutTranscriptActions(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(metaSpacing)),
						//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						//	if !p.gameRunning {
						//		return layout.Dimensions{}
						//	}
						//	return p.layoutTranscriptMeta(gtx)
						//}),
						//layout.Rigid(bareutils.SpacerH(metaSpacing)),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
								return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									if !p.gameRunning {
										return p.layoutTranscriptIdleState(gtx)
									}
									return p.layoutTranscriptEditor(gtx)
								})
							})
						}),
					)
				})
			}),
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return p.layoutFlashcardComposerOverlay(gtx)
			}),
		)
	})
}

func (p *Page) layoutFlashcardComposerOverlay(gtx layout.Context) layout.Dimensions {
	inset := layout.Inset{Right: unit.Dp(18), Bottom: unit.Dp(18)}
	width := p.transcriptComposerOverlayWidth(gtx)
	return layout.SE.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = width
			gtx.Constraints.Max.X = width
			return p.layoutFlashcardComposer(gtx)
		})
	})
}

func (p *Page) layoutTranscriptMeta(gtx layout.Context) layout.Dimensions {
	left := material.Body1(p.theme.Gio(), "Text Size: "+p.selectedTextSizeName)
	left.Color = p.theme.Color.TextMuted
	right := material.Body1(p.theme.Gio(), "Visible: "+p.selectedRecentLines)
	right.Color = p.theme.Color.TextMuted
	if p.isCompactLayout(gtx) {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(left.Layout),
			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
			layout.Rigid(right.Layout),
		)
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(left.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} }),
		layout.Rigid(right.Layout),
	)
}

func (p *Page) layoutTranscriptActions(gtx layout.Context) layout.Dimensions {
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
		Text:      "Clear View",
		Prefix:    "mdi:broom",
		Variant:   bareui.ButtonGhost,
	}
	if p.isCompactLayout(gtx) {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						if p.gameRunning {
							return launchButton.Layout(gtx.Disabled(), p.theme, p.iconify)
						}
						return launchButton.Layout(gtx, p.theme, p.iconify)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return syncButton.Layout(gtx, p.theme, p.iconify)
					}),
					layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return clearButton.Layout(gtx, p.theme, p.iconify)
					}),
				)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(p.theme.Gio(), p.transcriptRunningStatusText())
				lbl.Color = p.theme.Color.TextMuted
				return lbl.Layout(gtx)
			}),
		)
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if p.gameRunning {
				return launchButton.Layout(gtx.Disabled(), p.theme, p.iconify)
			}
			return launchButton.Layout(gtx, p.theme, p.iconify)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return syncButton.Layout(gtx, p.theme, p.iconify) }),
		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions { return clearButton.Layout(gtx, p.theme, p.iconify) }),
		layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(p.theme.Gio(), p.transcriptRunningStatusText())
			lbl.Color = p.theme.Color.TextMuted
			return lbl.Layout(gtx)
		}),
	)
}

func (p *Page) layoutTranscriptEditor(gtx layout.Context) layout.Dimensions {
	return material.List(p.theme.Gio(), &p.transcriptList).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				return p.layoutTranscriptLabel(gtx, p.theme.Color.Text, &p.transcriptView)
			}),
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				p.paintTranscriptHighlights(gtx)
				return layout.Dimensions{}
			}),
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				return p.layoutTranscriptPopup(gtx)
			}),
		)
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
	if normalizeSelectionText(p.transcriptView.SelectedText()) != "" {
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
	if normalizeSelectionText(p.transcriptView.SelectedText()) != "" {
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

	word := material.Editor(p.theme.Gio(), &p.wordEditor, "Word or phrase")
	word.Color = p.theme.Color.Text
	word.HintColor = p.theme.Color.TextMuted
	meaning := material.Editor(p.theme.Gio(), &p.meaningEditor, "Meaning")
	meaning.Color = p.theme.Color.Text
	meaning.HintColor = p.theme.Color.TextMuted

	searchButton := bareui.Button{Clickable: &p.searchWordButton, Text: "Lookup", Prefix: "mdi:book-search-outline", Variant: bareui.ButtonSecondary}
	playButton := bareui.Button{Clickable: &p.playAudioButton, Text: "mdi:play-circle-outline", Icon: true, Prefix: "mdi:play-circle-outline", Variant: bareui.ButtonSecondary}
	addAllButton := bareui.Button{Clickable: &p.addAllLookupButton, Text: "Add All Matches", Prefix: "mdi:playlist-plus", Variant: bareui.ButtonSecondary}
	minimizeButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-down", Icon: true, Prefix: "mdi:chevron-down", Variant: bareui.ButtonGhost}

	selected := normalizeSelectionText(p.transcriptView.SelectedText())
	if selected == "" {
		selected = "Select transcript text to prefill the flashcard word."
	}

	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(p.theme.Gio(), "New Flashcard")
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return minimizeButton.Layout(gtx, p.theme, p.iconify)
						}),
					)
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
		})
	})
}

func (p *Page) layoutFlashcardComposerHint(gtx layout.Context) layout.Dimensions {
	expandButton := bareui.Button{Clickable: &p.composerToggleButton, Text: "mdi:chevron-up", Icon: true, Prefix: "mdi:chevron-up", Variant: bareui.ButtonGhost}
	return bareutils.Panel(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(p.theme.Gio(), "New Flashcard")
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return expandButton.Layout(gtx, p.theme, p.iconify)
						}),
					)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), "Highlight transcript text to open the flashcard editor, or click a vocab match to inspect it.")
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
			)
		})
	})
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
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), "New Flashcard")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return expandButton.Layout(gtx, p.theme, p.iconify)
				}),
			)
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
	if selected := normalizeSelectionText(p.transcriptView.SelectedText()); selected != "" {
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
	p.wordEditor.SetText(util.FirstNonEmpty(lookups[0].Query, lookups[0].Key, lookups[0].Headword))
	p.meaningEditor.SetText(lookups[0].Meaning)
}

func (p *Page) playCurrentLookupAudio() {
	if p.lookupResult == nil || strings.TrimSpace(p.lookupResult.AudioPath) == "" {
		p.showError("Audio Playback Failed", "No audio is available for the current lookup.")
		return
	}
	if err := playAudioFile(p.lookupResult.AudioPath); err != nil {
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
		if err := playAudioFile(lookup.AudioPath); err != nil {
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
	if p.currentConfig == nil || strings.TrimSpace(p.currentConfig.Name) == "" {
		p.showError("Launch Failed", "The selected game configuration is not loaded yet.")
		return
	}
	if p.gameRunning {
		p.statusText = p.transcriptRunningStatusText()
		return
	}
	if err := p.currentConfig.LaunchInBackground(); err != nil {
		p.statusText = err.Error()
		p.showError("Launch Failed", err.Error())
		return
	}
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
			_ = playAudioFile(match.Card.AudioPath)
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

func (p *Page) transcriptComposerOverlayWidth(gtx layout.Context) int {
	if p.shouldCollapseFlashcardComposer() {
		if p.isCompactLayout(gtx) {
			return gtx.Dp(unit.Dp(280))
		}
		return gtx.Dp(unit.Dp(320))
	}
	if p.isCompactLayout(gtx) {
		return gtx.Dp(unit.Dp(320))
	}
	return min(gtx.Dp(unit.Dp(420)), p.transcriptComposerWidth(gtx))
}

func (p *Page) transcriptLaunchButtonLabel() string {
	if p.gameRunning {
		return "Game Running"
	}
	return "Launch Game"
}

func (p *Page) transcriptLaunchButtonIcon() string {
	if p.gameRunning {
		return "mdi:check-circle-outline"
	}
	return "mdi:play-box-outline"
}

func (p *Page) transcriptLaunchButtonVariant() bareui.ButtonVariant {
	if p.gameRunning {
		return bareui.ButtonSecondary
	}
	return bareui.ButtonPrimary
}

func (p *Page) transcriptRunningStatusText() string {
	if p.gameRunning {
		if p.gameRunningPID > 0 {
			return fmt.Sprintf("Detected running game process (pid %d).", p.gameRunningPID)
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

func lookupResultKey(lookup dictionary.Lookup) string {
	return util.FirstNonEmpty(lookup.Key, lookup.Query, lookup.Headword)
}

func normalizeSelectionText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
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

func playAudioFile(path string) error {
	path = strings.TrimSpace(path)
	if !util.IsExistingFile(path) {
		return fmt.Errorf("audio file not found: %s", path)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("afplay"); err == nil {
			cmd = exec.Command("afplay", path)
			break
		}
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		for _, candidate := range [][]string{
			{"xdg-open", path},
			{"mpv", path},
			{"ffplay", "-nodisp", "-autoexit", path},
			{"vlc", "--play-and-exit", path},
		} {
			if _, err := exec.LookPath(candidate[0]); err == nil {
				cmd = exec.Command(candidate[0], candidate[1:]...)
				break
			}
		}
	}
	if cmd == nil {
		return fmt.Errorf("no supported audio player found")
	}
	return cmd.Start()
}
