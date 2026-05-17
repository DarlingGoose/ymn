package transcript

import (
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/jpndict"
	"github.com/DarlingGoose/ymn/pkg/japanese"
	"github.com/DarlingGoose/ymn/pkg/util"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

func (t *SentenceAnalysis) Layout(gtx layout.Context) layout.Dimensions {
	t.ensureFlashcardsCurrent()
	return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return t.layoutHeader(gtx)
			}),
			//layout.Rigid(spacerH(unit.Dp(14))),
			//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			//	return t.layoutFocusedSentenceText(gtx)
			//}),
			layout.Rigid(spacerH(unit.Dp(14))),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return t.layoutSentenceStructure(gtx)
			}),
		)
	})
}

func (t *SentenceAnalysis) layoutHeader(gtx layout.Context) layout.Dimensions {
	meta := "No sentence selected"
	if t.line != nil {
		parts := make([]string, 0, 2)
		if speaker := strings.TrimSpace(t.line.Speaker); speaker != "" {
			parts = append(parts, speaker)
		}
		if ts := strings.TrimSpace(t.line.Time); ts != "" {
			parts = append(parts, ts)
		}
		if len(parts) > 0 {
			meta = strings.Join(parts, " - ")
		} else {
			meta = "Selected transcript row"
		}
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleH4, theme.ThemeColorTextPrimary, "Sentence Analysis")
		}),
		layout.Rigid(spacerW(unit.Dp(12))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return t.layoutHeaderTokenSummary(gtx)
		}),
		layout.Rigid(spacerW(unit.Dp(12))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleCaption, theme.ThemeColorTextMuted, meta)
		}),
	)
}

func (t *SentenceAnalysis) layoutHeaderTokenSummary(gtx layout.Context) layout.Dimensions {
	if strings.TrimSpace(t.structureSourceText()) == "" {
		return layout.Dimensions{}
	}
	analysis, errText := t.currentAnalysis()
	if errText != "" || len(analysis.Tokens) == 0 {
		return layout.Dimensions{}
	}
	return t.layoutTokenSummary(gtx, analysis)
}

func (t *SentenceAnalysis) layoutFocusedSentenceText(gtx layout.Context) layout.Dimensions {
	text := t.structureSourceText()
	if text == "" {
		text = "Start the game to inspect the latest sentence."
	}
	text = utils.CleanInlineText(text)
	//	p.syncFocusedSentenceView(text)
	if t.focusedFuriganaMode != focusedFuriganaHidden {
		return layout.Dimensions{}
		//return t.layoutFocusedSentenceWithFurigana(gtx, text)
	}

	lbl := material.H6(t.th, text)
	theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleH1)
	lbl.Color = t.tc.GetCurrentColorToken().TextPrimaryNRGBA()
	//lbl.Color = p.theme.Color.Text
	//lbl.TextSize = p.focusedSentenceTextSize(gtx)
	//lbl.State = &p.focusedSentenceView
	return utils.RoundedSurface(gtx, unit.Dp(10), t.tc.GetCurrentColorToken().SurfaceNRGBA(), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, lbl.Layout)
	})
}

func (t *SentenceAnalysis) layoutSentenceStructure(gtx layout.Context) layout.Dimensions {
	text := t.structureSourceText()
	if strings.TrimSpace(text) == "" {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleBody, theme.ThemeColorTextMuted, "Select a transcript row to inspect its tokens.")
		})
	}

	analysis, errText := t.currentAnalysis()
	if errText != "" {
		return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleBody, theme.ThemeColorWarning, errText)
	}
	if len(analysis.Tokens) == 0 {
		return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleBody, theme.ThemeColorTextMuted, "No tokens found.")
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutTokenSectionHeader(gtx)
		}),
		layout.Rigid(spacerH(unit.Dp(8))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return t.layoutTokenLines(gtx, analysis, analysis.Tokens)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutSelectedTokenLookup(gtx)
		}),
	)
}

func (t *SentenceAnalysis) layoutTokenSummary(gtx layout.Context, analysis japanese.Analysis) layout.Dimensions {
	items := []struct {
		label string
		value int
	}{
		{"Tokens", len(analysis.Tokens)},
		{"Particles", len(analysis.Particles)},
		{"Verbs", len(analysis.Verbs)},
		{"Aux", len(analysis.Auxiliaries)},
	}

	children := make([]layout.FlexChild, 0, len(items)*2)
	for i, item := range items {
		item := item
		if i > 0 {
			children = append(children, layout.Rigid(spacerW(unit.Dp(8))))
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutSummaryPill(gtx, item.label, item.value)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (t *SentenceAnalysis) layoutTokenSectionHeader(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleLabel, theme.ThemeColorTextSecondary, "Tokens")
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutFuriganaControls(gtx)
		}),
	)
}

func (t *SentenceAnalysis) layoutFuriganaControls(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutFuriganaModeButton(gtx, &t.furiganaHiddenClick, "Hidden", t.focusedFuriganaMode == focusedFuriganaHidden)
		}),
		layout.Rigid(spacerW(unit.Dp(6))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutFuriganaModeButton(gtx, &t.furiganaAboveClick, "Above", t.focusedFuriganaMode == focusedFuriganaAbove)
		}),
	)
}

func (t *SentenceAnalysis) layoutFuriganaModeButton(gtx layout.Context, click *widget.Clickable, text string, selected bool) layout.Dimensions {
	ct := t.tc.GetCurrentColorToken()
	bg := color.NRGBA{A: 0}
	fg := ct.TextMutedNRGBA()
	if selected {
		bg = theme.Mix(ct.PrimaryNRGBA(), ct.SurfaceNRGBA(), 0.20)
		fg = ct.PrimaryNRGBA()
	} else if click.Hovered() {
		bg = ct.SurfaceNRGBA()
	}

	return utils.ClickableSurface(gtx, click, bg, unit.Dp(7), func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Inset{Top: unit.Dp(5), Bottom: unit.Dp(5), Left: unit.Dp(9), Right: unit.Dp(9)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(t.th, text)
			theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleLabelSmall)
			lbl.Color = fg
			return lbl.Layout(gtx)
		})
	})
}

func (t *SentenceAnalysis) layoutSummaryPill(gtx layout.Context, label string, value int) layout.Dimensions {
	ct := t.tc.GetCurrentColorToken()
	bg := theme.Mix(ct.PrimaryNRGBA(), ct.SurfaceNRGBA(), 0.14)
	return utils.RoundedSurface(gtx, unit.Dp(8), bg, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(7), Bottom: unit.Dp(7), Left: unit.Dp(10), Right: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleLabelSmall, theme.ThemeColorPrimary, label)
				}),
				layout.Rigid(spacerW(unit.Dp(6))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleLabelSmall, theme.ThemeColorTextPrimary, strconv.Itoa(value))
				}),
			)
		})
	})
}

func (t *SentenceAnalysis) layoutTokenLines(gtx layout.Context, analysis japanese.Analysis, tokens []japanese.Token) layout.Dimensions {
	lines := t.focusedSentenceTokenLines(gtx, tokens)
	t.pruneFocusedTokenClicks(analysis.Tokens)

	top := unit.Dp(0)
	if t.focusedFuriganaMode == focusedFuriganaAbove {
		top = unit.Dp(4)
	}
	return layout.Inset{Top: top}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return material.List(t.th, &t.structureList).Layout(gtx, len(lines), func(gtx layout.Context, index int) layout.Dimensions {
			if index < 0 || index >= len(lines) {
				return layout.Dimensions{}
			}
			line := lines[index]
			bottom := unit.Dp(8)
			if index == len(lines)-1 {
				bottom = unit.Dp(0)
			}
			return layout.Inset{Bottom: bottom}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lineChildren := make([]layout.FlexChild, 0, len(line))
				for _, token := range line {
					token := token
					lineChildren = append(lineChildren, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return t.layoutFocusedFuriganaToken(gtx, token)
					}))
				}
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, lineChildren...)
			})
		})
	})
}

func structureTokenKey(token japanese.Token) string {
	return strings.Join([]string{
		strings.TrimSpace(token.Surface),
		strings.TrimSpace(token.BaseForm),
		token.POSLabel(),
		token.InflectionLabel(),
	}, "\x00")
}

func (t *SentenceAnalysis) layoutFocusedFuriganaToken(gtx layout.Context, token japanese.Token) layout.Dimensions {
	key := structureTokenKey(token)
	click := t.focusedTokenClickable(key)
	reading := focusedTokenReading(token)
	surface := utils.CleanInlineText(token.Surface)
	if surface == "" {
		return layout.Dimensions{}
	}
	_, inFlashcards := t.structureTokenFlashcard(token)
	dictionaryReady := false
	if t.selectedFocusedTokenKey == key {
		_, _, errText, pending, results := t.lookupSnapshot()
		dictionaryReady = errText == "" && !pending && len(results) > 0
	}
	bg := focusedTokenColor(t.tc.GetCurrentColorToken(), token, t.selectedFocusedTokenKey == key, inFlashcards, dictionaryReady)
	children := make([]layout.FlexChild, 0, 4)
	if t.focusedFuriganaMode == focusedFuriganaAbove {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutFocusedTokenSlot(gtx, t.focusedTokenReadingSlotHeight(), func(gtx layout.Context) layout.Dimensions {
				return t.layoutFocusedTokenReading(gtx, reading)
			})
		}), layout.Rigid(spacerH(unit.Dp(1))))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return t.layoutFocusedTokenSlot(gtx, t.focusedTokenSurfaceSlotHeight(gtx), func(gtx layout.Context) layout.Dimensions {
			lbl := material.H6(t.th, surface)
			theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleH2)
			lbl.TextSize = t.sentenceFontSize
			lbl.Color = t.tc.GetCurrentColorToken().TextPrimaryNRGBA()
			return lbl.Layout(gtx)
		})
	}))
	if t.focusedFuriganaMode == focusedFuriganaBelow {
		children = append(children, layout.Rigid(spacerH(unit.Dp(2))), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutFocusedTokenSlot(gtx, t.focusedTokenReadingSlotHeight(), func(gtx layout.Context) layout.Dimensions {
				return t.layoutFocusedTokenReading(gtx, reading)
			})
		}))
	}
	if t.focusedFuriganaMode == focusedFuriganaAbove {
		children = append(children, layout.Rigid(spacerH(unit.Dp(3))), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutFocusedTokenSlot(gtx, unit.Dp(14), func(gtx layout.Context) layout.Dimensions {
				return t.layoutFocusedTokenMarker(gtx, inFlashcards, dictionaryReady)
			})
		}))
	}
	return layout.Inset{Right: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			pointer.CursorPointer.Add(gtx.Ops)
			return utils.RoundedSurface(gtx, unit.Dp(7), bg, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top:    unit.Dp(2),
					Bottom: unit.Dp(2),
					Left:   unit.Dp(4),
					Right:  unit.Dp(4),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
				})
			})
		})
	})
}

func (t *SentenceAnalysis) layoutFocusedTokenReading(gtx layout.Context, reading string) layout.Dimensions {
	if reading == "" {
		reading = " "
	}
	lbl := material.Body2(t.th, reading)
	theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleBodyLarge)
	lbl.TextSize = t.focusedTokenReadingFontSize()
	lbl.Color = t.tc.GetCurrentColorToken().SecondaryNRGBA()
	return lbl.Layout(gtx)
}

func (t *SentenceAnalysis) layoutFocusedTokenMarker(gtx layout.Context, inFlashcards, dictionaryReady bool) layout.Dimensions {
	fg := t.tc.GetCurrentColorToken().TextMutedNRGBA()
	if inFlashcards {
		fg = t.tc.GetCurrentColorToken().PrimaryNRGBA()
		return iconify.DefaultIconify.Layout(gtx, "lucide:check", unit.Dp(12), fg)
	}
	if dictionaryReady {
		fg := t.tc.GetCurrentColorToken().SecondaryNRGBA()
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			size := gtx.Dp(unit.Dp(4))
			if size < 2 {
				size = 2
			}
			return utils.RoundedSurface(gtx, unit.Dp(2), fg, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(size, size)}
			})
		})
	}
	return layout.Dimensions{Size: image.Pt(gtx.Dp(unit.Dp(12)), gtx.Dp(unit.Dp(12)))}
}

func (t *SentenceAnalysis) layoutSelectedTokenLookup(gtx layout.Context) layout.Dimensions {
	query, _, errText, pending, results := t.lookupSnapshot()
	visible := strings.TrimSpace(query) != ""
	if visible {
		t.lastLookupQuery = query
		t.lastLookupErr = errText
		t.lastLookupPending = pending
		t.lastLookupResults = append([]*jpndict.Response(nil), results...)
		if t.lookupBarFlip != nil && !t.lookupBarFlip.Expanded() {
			t.lookupBarFlip.Expand()
		}
	} else if t.lookupBarFlip != nil && t.lookupBarFlip.Expanded() {
		t.lookupBarFlip.Collapse()
	}

	progress := 1.0
	running := false
	if t.lookupBarFlip != nil {
		progress, running = t.lookupBarFlip.Value(time.Now())
	}
	if running {
		gtx.Execute(op.InvalidateCmd{})
	}
	if progress <= 0 {
		return layout.Dimensions{}
	}
	if !visible {
		query = t.lastLookupQuery
		errText = t.lastLookupErr
		pending = t.lastLookupPending
		results = append([]*jpndict.Response(nil), t.lastLookupResults...)
	}
	if strings.TrimSpace(query) == "" {
		return layout.Dimensions{}
	}

	macro := op.Record(gtx.Ops)
	dims := t.layoutSelectedTokenLookupContent(gtx, query, errText, pending, results)
	call := macro.Stop()

	height := tween.MapInt(progress, 0, dims.Size.Y)
	if height <= 0 {
		return layout.Dimensions{}
	}
	clipStack := clip.Rect{Max: image.Pt(dims.Size.X, height)}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return layout.Dimensions{Size: image.Pt(dims.Size.X, height)}
}

func (t *SentenceAnalysis) layoutSelectedTokenLookupContent(gtx layout.Context, query, errText string, pending bool, results []*jpndict.Response) layout.Dimensions {
	ct := t.tc.GetCurrentColorToken()
	return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return utils.Surface(gtx, ct.BackgroundNRGBA(), unit.Dp(8), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(10), Bottom: unit.Dp(10), Left: unit.Dp(12), Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := []layout.FlexChild{
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return t.layoutLookupHeader(gtx, query, pending, len(results))
					}),
				}

				if errText != "" {
					children = append(children,
						layout.Rigid(spacerH(unit.Dp(6))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleBodySmall, theme.ThemeColorWarning, errText)
						}),
					)
				} else if len(results) > 0 {
					children = append(children,
						layout.Rigid(spacerH(unit.Dp(8))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return t.layoutLookupResults(gtx, results)
						}),
					)
				}
				if status := strings.TrimSpace(t.flashcardStatus); status != "" {
					children = append(children,
						layout.Rigid(spacerH(unit.Dp(6))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleCaption, theme.ThemeColorTextMuted, status)
						}),
					)
				}
				if errText := strings.TrimSpace(t.flashcardLoadErr); errText != "" {
					children = append(children,
						layout.Rigid(spacerH(unit.Dp(6))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleCaption, theme.ThemeColorWarning, errText)
						}),
					)
				}

				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			})
		})
	})
}

func (t *SentenceAnalysis) layoutLookupHeader(gtx layout.Context, query string, pending bool, count int) layout.Dimensions {
	status := "Dictionary"
	if pending {
		status = "Looking up..."
	} else if count > 0 {
		status = strconv.Itoa(count) + " matches"
	}

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(t.th, query)
			theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleLabel)
			t.applyLookupTextStyle(&lbl, 2)
			lbl.Color = t.tc.GetCurrentColorToken().TextPrimaryNRGBA()
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutLookupFontControls(gtx)
		}),
		layout.Rigid(spacerW(unit.Dp(10))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutAddFlashcardButton(gtx, pending)
		}),
		layout.Rigid(spacerW(unit.Dp(10))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleCaption, theme.ThemeColorTextMuted, status)
		}),
	)
}

func (t *SentenceAnalysis) layoutAddFlashcardButton(gtx layout.Context, pending bool) layout.Dimensions {
	ct := t.tc.GetCurrentColorToken()
	fg := ct.PrimaryNRGBA()
	bg := color.NRGBA{A: 0}
	if pending {
		fg = ct.TextMutedNRGBA()
	} else if t.addFlashcardClick.Hovered() {
		bg = ct.SurfaceNRGBA()
	}

	return utils.ClickableSurface(gtx, &t.addFlashcardClick, bg, unit.Dp(7), func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Inset{Top: unit.Dp(5), Bottom: unit.Dp(5), Left: unit.Dp(7), Right: unit.Dp(7)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iconify.DefaultIconify.Layout(gtx, "lucide:plus", unit.Dp(15), fg)
				}),
				layout.Rigid(spacerW(unit.Dp(5))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := "Add"
					if pending {
						label = "Wait"
					}
					lbl := material.Body2(t.th, label)
					theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleCaption)
					lbl.Color = fg
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (t *SentenceAnalysis) layoutLookupFontControls(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutLookupFontButton(gtx, &t.lookupFontDownClick, "A-")
		}),
		layout.Rigid(spacerW(unit.Dp(4))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutLookupFontButton(gtx, &t.lookupFontUpClick, "A+")
		}),
	)
}

func (t *SentenceAnalysis) layoutLookupFontButton(gtx layout.Context, click *widget.Clickable, text string) layout.Dimensions {
	ct := t.tc.GetCurrentColorToken()
	bg := color.NRGBA{A: 0}
	if click.Hovered() {
		bg = ct.SurfaceNRGBA()
	}
	return utils.ClickableSurface(gtx, click, bg, unit.Dp(6), func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Inset{Top: unit.Dp(4), Bottom: unit.Dp(4), Left: unit.Dp(7), Right: unit.Dp(7)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(t.th, text)
			theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleCaption)
			lbl.Color = ct.TextMutedNRGBA()
			return lbl.Layout(gtx)
		})
	})
}

func (t *SentenceAnalysis) layoutLookupResults(gtx layout.Context, results []*jpndict.Response) layout.Dimensions {
	limit := len(results)
	if limit > 3 {
		limit = 3
	}
	children := make([]layout.FlexChild, 0, limit*2)
	for i := 0; i < limit; i++ {
		resp := results[i]
		if i > 0 {
			children = append(children, layout.Rigid(spacerH(unit.Dp(6))))
		}
		index := i
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutLookupResult(gtx, resp, index)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (t *SentenceAnalysis) layoutLookupResult(gtx layout.Context, resp *jpndict.Response, index int) layout.Dimensions {
	if resp == nil {
		return layout.Dimensions{}
	}
	headword, reading, meaning := lookupResponseText(resp)
	if strings.TrimSpace(meaning) == "" {
		meaning = strings.TrimSpace(resp.Text)
	}
	audioKey := lookupResultKey(resp, index)
	audioQuery, expectedReading := t.lookupAudioContext(resp, headword)
	t.registerLookupAudio(audioKey, audioQuery, expectedReading, resp)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body2(t.th, headword)
							theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleLabel)
							t.applyLookupTextStyle(&lbl, 1)
							lbl.Color = t.tc.GetCurrentColorToken().TextPrimaryNRGBA()
							return lbl.Layout(gtx)
						}),
						layout.Rigid(spacerW(unit.Dp(8))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if strings.TrimSpace(reading) == "" || reading == headword {
								return layout.Dimensions{}
							}
							lbl := material.Body2(t.th, reading)
							theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleCaption)
							t.applyLookupTextStyle(&lbl, -1)
							lbl.Color = t.tc.GetCurrentColorToken().SecondaryNRGBA()
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutLookupAudioButton(gtx, audioKey)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return t.layoutLookupAudioStatus(gtx, audioKey)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if strings.TrimSpace(meaning) == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(t.th, meaning)
				theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleBodySmall)
				t.applyLookupTextStyle(&lbl, 0)
				lbl.Color = t.tc.GetCurrentColorToken().TextMutedNRGBA()
				return lbl.Layout(gtx)
			})
		}),
	)
}

func (t *SentenceAnalysis) layoutLookupAudioButton(gtx layout.Context, key string) layout.Dimensions {
	pending, cached, _ := t.lookupAudioSnapshot(key)
	ct := t.tc.GetCurrentColorToken()
	fg := ct.TextMutedNRGBA()
	bg := color.NRGBA{A: 0}
	if cached {
		fg = ct.PrimaryNRGBA()
	}
	click := t.lookupAudioClickable(key)
	if click.Hovered() {
		bg = ct.SurfaceNRGBA()
	}
	if pending {
		fg = ct.TextMutedNRGBA()
	}

	return utils.ClickableSurface(gtx, click, bg, unit.Dp(7), func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		return layout.Inset{Top: unit.Dp(5), Bottom: unit.Dp(5), Left: unit.Dp(7), Right: unit.Dp(7)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iconify.DefaultIconify.Layout(gtx, "lucide:volume-2", unit.Dp(15), fg)
				}),
				layout.Rigid(spacerW(unit.Dp(5))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := "Audio"
					if pending {
						label = "Loading"
					} else if cached {
						label = "Cached"
					}
					lbl := material.Body2(t.th, label)
					theme.ApplyTypography(&lbl, t.tc.GetCurrentTypography(), theme.TextRoleCaption)
					lbl.Color = fg
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (t *SentenceAnalysis) layoutLookupAudioStatus(gtx layout.Context, key string) layout.Dimensions {
	_, _, errText := t.lookupAudioSnapshot(key)
	if strings.TrimSpace(errText) == "" {
		return layout.Dimensions{}
	}
	return layout.Inset{Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleCaption, theme.ThemeColorWarning, errText)
	})
}

func (t *SentenceAnalysis) applyLookupTextStyle(lbl *material.LabelStyle, offset unit.Sp) {
	if lbl == nil {
		return
	}
	size := t.lookupFontSize + offset
	if size < unit.Sp(10) {
		size = unit.Sp(10)
	}
	lbl.TextSize = size
	lbl.LineHeight = size + unit.Sp(7)
	lbl.Font.Typeface = font.Typeface("Noto Sans CJK JP")
}

func lookupResultKey(resp *jpndict.Response, index int) string {
	if resp == nil {
		return strconv.Itoa(index)
	}
	parts := []string{
		strconv.Itoa(index),
		strings.TrimSpace(resp.Query),
		strings.TrimSpace(resp.Key),
	}
	if resp.Entry != nil {
		parts = append(parts, strings.TrimSpace(resp.Entry.Headword), strings.TrimSpace(resp.Entry.Reading))
	}
	return strings.Join(parts, "\x00")
}

func (t *SentenceAnalysis) lookupAudioContext(resp *jpndict.Response, headword string) (query, expectedReading string) {
	if token, ok := t.selectedToken(); ok {
		expectedReading = normalizeAudioReading(util.FirstNonEmpty(token.Pronunciation, token.Reading))
		candidates := []string{
			token.Surface,
			structureFlashcardWord(token),
			token.BaseForm,
			headword,
		}
		if resp != nil {
			if resp.Entry != nil {
				candidates = append(candidates, resp.Entry.Headword, resp.Entry.Reading)
			}
			candidates = append(candidates, resp.Query, resp.Key)
		}
		for _, candidate := range candidates {
			if text := strings.TrimSpace(candidate); text != "" {
				return text, expectedReading
			}
		}
		return "", expectedReading
	}

	candidates := []string{headword}
	if resp != nil {
		if resp.Entry != nil {
			candidates = append(candidates, resp.Entry.Headword, resp.Entry.Reading)
		}
		candidates = append(candidates, resp.Query, resp.Key)
	}
	for _, candidate := range candidates {
		if text := strings.TrimSpace(candidate); text != "" {
			return text, ""
		}
	}
	return "", ""
}

func lookupResponseText(resp *jpndict.Response) (headword, reading, meaning string) {
	if resp == nil {
		return "", "", ""
	}
	headword = strings.TrimSpace(resp.Query)
	if resp.Entry != nil {
		headword = strings.TrimSpace(resp.Entry.Headword)
		reading = strings.TrimSpace(resp.Entry.Reading)
		meaning = summarizeLookupEntry(resp.Entry)
	}
	if headword == "" {
		headword = strings.TrimSpace(resp.Key)
	}
	if headword == "" {
		headword = strings.TrimSpace(resp.Query)
	}
	if meaning == "" {
		meaning = strings.TrimSpace(resp.Text)
	}
	return headword, reading, meaning
}

func summarizeLookupEntry(entry *jpndict.Entry) string {
	if entry == nil {
		return ""
	}
	lines := make([]string, 0, len(entry.Senses))
	for i, sense := range entry.Senses {
		gloss := strings.TrimSpace(strings.Join(sense.Glosses, "; "))
		if gloss == "" {
			continue
		}
		if len(sense.PartsOfSpeech) > 0 {
			gloss = "[" + strings.Join(sense.PartsOfSpeech, ", ") + "] " + gloss
		}
		if len(entry.Senses) > 1 {
			gloss = strconv.Itoa(i+1) + ". " + gloss
		}
		lines = append(lines, gloss)
		if len(lines) == 2 {
			break
		}
	}
	return strings.Join(lines, "\n")
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
