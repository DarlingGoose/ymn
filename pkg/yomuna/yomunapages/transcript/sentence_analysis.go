package transcript

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
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
	autoTranslate          bool

	focusedFuriganaMode    string
	focusedFuriganaDefault string

	sentenceFontSize  unit.Sp
	furigiganFontSize unit.Sp

	line *transcriptRow
}

func NewSentenceAnalysis(th *material.Theme, backend backend.Backend) *SentenceAnalysis {
	return &SentenceAnalysis{
		tc:                     theme.DefaultThemeClient,
		backend:                backend,
		th:                     th,
		selectedTargetLanguage: "english",
		focusedFuriganaMode:    focusedFuriganaHidden,
		focusedFuriganaDefault: focusedFuriganaHidden,
	}
}

func (t *SentenceAnalysis) WithTranslatorConfig(cfg translation.Config) *SentenceAnalysis {
	t.translatorConfig = cfg
	return t
}

func (t *SentenceAnalysis) WithAutoTranslate(at bool) *SentenceAnalysis {
	t.autoTranslate = at
	return t
}

func (t *SentenceAnalysis) SetSentence(line *transcriptRow) {

}
func (t *SentenceAnalysis) Reset() {

}
func (t *SentenceAnalysis) HandeEvents(gtx layout.Context) {

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
	if readingRunes > runes {
		runes = readingRunes
	}
	if runes <= 0 {
		runes = 1
	}
	size := float32(t.sentenceFontSize)
	return gtx.Dp(unit.Dp(float32(runes)*size*0.72 + 16))
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
