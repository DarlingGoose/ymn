package transcript

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/wgl/pkg/japanese"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

func (t *SentenceAnalysis) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{} //card header??
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutFocusedSentenceText(gtx)
		}),
	)
}

func (t *SentenceAnalysis) layoutFocusedSentenceText(gtx layout.Context) layout.Dimensions {
	text := t.structureSourceText()
	if text == "" {
		text = "Start the game to inspect the latest sentence."
	}
	text = utils.CleanInlineText(text)
	//	p.syncFocusedSentenceView(text)
	if t.focusedFuriganaMode != focusedFuriganaHidden {
		return t.layoutFocusedSentenceWithFurigana(gtx, text)
	}

	lbl := material.H6(t.th, text)
	theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleH1)
	lbl.Color = t.tc.GetCurrentColorToken().TextPrimaryNRGBA()
	//lbl.Color = p.theme.Color.Text
	//lbl.TextSize = p.focusedSentenceTextSize(gtx)
	//lbl.State = &p.focusedSentenceView
	return lbl.Layout(gtx)
}

func (t *SentenceAnalysis) layoutFocusedSentenceWithFurigana(gtx layout.Context, sentence string) layout.Dimensions {
	analysis, err := japanese.AnalyzeSentence(sentence) //todo move to backend
	if err != nil || len(analysis.Tokens) == 0 {
		lbl := material.H6(t.th, sentence)
		theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleH2)
		lbl.Color = t.tc.GetCurrentColorToken().TextPrimaryNRGBA()
		return lbl.Layout(gtx)
	}
	lines := t.focusedSentenceTokenLines(gtx, analysis.Tokens)
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
					token.StructureLabel()
					return layout.Dimensions{}
					//	return p.layoutFocusedFuriganaToken(gtx, token)
				}))
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, lineChildren...)
		}))
	}
	//p.pruneFocusedTokenClicks(analysis.Tokens)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (t *SentenceAnalysis) focusedSentenceTokenLines(gtx layout.Context, tokens []japanese.Token) [][]japanese.Token {
	maxWidth := gtx.Constraints.Max.X
	if maxWidth <= 0 {
		return [][]japanese.Token{tokens}
	}
	lines := make([][]japanese.Token, 0, 2)
	line := make([]japanese.Token, 0, len(tokens))
	lineWidth := 0
	for _, token := range tokens {
		tokenWidth := t.focusedSentenceTokenWidth(gtx, token)
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
