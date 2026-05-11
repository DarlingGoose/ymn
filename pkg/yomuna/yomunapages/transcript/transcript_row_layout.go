package transcript

import (
	"context"
	"image/color"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/wgl/pkg/translation"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

func (t *transcriptFollower) Layout(gtx layout.Context) layout.Dimensions {
	rows := t.GetRows()
	if len(rows) == 0 {
		return t.layoutTranscriptIdleState(gtx)
	}
	return material.List(t.th, &t.transcriptList).Layout(gtx, len(t.GetRows()), func(gtx layout.Context, index int) layout.Dimensions {
		if index < 0 || index >= len(t.transcriptRows) {
			return layout.Dimensions{}
		}
		return layout.Inset{
			Bottom: unit.Dp(8),
			Left:   unit.Dp(20),
			Right:  unit.Dp(20),
			Top:    unit.Dp(8),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return t.layoutTranscriptRow(gtx, rows[index])
		})
	})
}

func (t *transcriptFollower) layoutTranscriptIdleState(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.H6(t.th, "Transcript Hidden")
				lbl.Color = t.tc.GetCurrentColorToken().TextPrimaryNRGBA()
				return lbl.Layout(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(t.th, "Start the game to show live transcript text here.")
				lbl.Color = t.tc.GetCurrentColorToken().TextMutedNRGBA()
				return lbl.Layout(gtx)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(t.th, "The flashcard composer stays on this page next to the transcript.")
				lbl.Color = t.tc.GetCurrentColorToken().TextMutedNRGBA()
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (t *transcriptFollower) layoutTranscriptRow(gtx layout.Context, row transcriptRow) layout.Dimensions {
	if row.Info {
		return t.layoutTranscriptInfoRow(gtx, row)
	}
	ct := t.tc.GetCurrentColorToken()
	click := t.transcriptRowClickable(row.Key)
	selected := row.Key == t.currentTranscriptRowKey()
	bg := ct.BackgroundNRGBA()
	fg := ct.TextPrimaryNRGBA()
	timeColor := ct.TextMutedNRGBA()
	if selected {
		bg = theme.Mix(ct.PrimaryNRGBA(), ct.SurfaceAltNRGBA(), 0.22)
		timeColor = ct.PrimaryNRGBA()
	}
	if click.Hovered() && !selected {
		bg = ct.SurfaceAltNRGBA()
	}
	text := t.transcriptRowDisplayText(row)

	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		return utils.RoundedSurface(gtx, t.radius, bg, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(12),
				Bottom: unit.Dp(12),
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.X = gtx.Dp(t.transcriptTimestampWidth())
						lbl := material.Body2(t.th, t.transcriptTimestampText(row.Time))
						lbl.Color = timeColor
						return lbl.Layout(gtx)
					}),
					layout.Rigid(utils.SpacerW(unit.Dp(14))),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if strings.TrimSpace(row.Speaker) == "" {
							return layout.Dimensions{}
						}
						return t.layoutTranscriptSpeaker(gtx, row.Speaker, selected)
					}),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(t.th, text)
						lbl.Color = fg
						if t.isTranscriptRowTranslationShown(row) {
							lbl.Color = ct.PrimaryNRGBA()
						}
						lbl.TextSize = t.fontSize
						return lbl.Layout(gtx)
					}),
					layout.Rigid(utils.SpacerW(unit.Dp(8))),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return t.layoutTranscriptTranslateIcon(gtx, row)
					}),
				)
			})
		})
	})
}

func (t *transcriptFollower) layoutTranscriptTranslateIcon(gtx layout.Context, row transcriptRow) layout.Dimensions {
	enabled := strings.TrimSpace(row.Text) != "" && strings.TrimSpace(t.selectedTargetLanguage) != ""
	if !enabled {
		return t.layoutRowIcon(gtx, "mdi:translate", false, false)
	}
	click := t.transcriptRowTranslateClickable(row.Key)
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		pointer.CursorPointer.Add(gtx.Ops)
		icon := "mdi:translate"
		if t.isTranscriptRowTranslationShown(row) {
			icon = "mdi:translate-off"
		}
		return t.layoutRowIcon(gtx, icon, true, click.Hovered())
	})

}

func (t *transcriptFollower) layoutTranscriptSpeaker(gtx layout.Context, speaker string, selected bool) layout.Dimensions {
	speaker = strings.TrimSpace(speaker)
	if speaker == "" {
		return layout.Dimensions{}
	}
	fg := t.tc.GetCurrentColorToken().PrimaryNRGBA()
	bg := color.NRGBA{R: fg.R, G: fg.G, B: fg.B, A: 34}
	if selected {
		bg = color.NRGBA{R: fg.R, G: fg.G, B: fg.B, A: 54}
	}
	return layout.Inset{Right: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return utils.RoundedSurfaceWrap(gtx, bg, unit.Dp(t.iconRadius), func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top:    unit.Dp(4),
				Bottom: unit.Dp(4),
				Left:   unit.Dp(8),
				Right:  unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(t.th, speaker)
				lbl.Color = fg
				return lbl.Layout(gtx)
			})
		})
	})
}

func (t *transcriptFollower) layoutTranscriptInfoRow(gtx layout.Context, row transcriptRow) layout.Dimensions {
	return utils.RoundedSurface(gtx, t.radius, t.tc.GetCurrentColorToken().BackgroundNRGBA(), func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{
			Top:    unit.Dp(9),
			Bottom: unit.Dp(9),
			Left:   unit.Dp(14),
			Right:  unit.Dp(14),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					//todo update themed lable to allow for more dynamic sizes
					return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleBody, theme.ThemeColorTextMuted, t.transcriptTimestampText(row.Time))
				}),
				layout.Rigid(utils.SpacerW(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return t.layoutRowIcon(gtx, "mdi:information-outline", true, false)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					//todo update themed lable to allow for more dynamic sizes
					return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleBody, theme.ThemeColorTextMuted, row.Text)
				}),
			)
		})
	})
}

func (t *transcriptFollower) layoutRowIcon(gtx layout.Context, icon string, enabled bool, hovered bool) layout.Dimensions {
	clr := t.tc.GetCurrentColorToken().PrimaryNRGBA()
	bg := t.tc.GetCurrentColorToken().SurfaceNRGBA()
	if !enabled {
		clr = t.tc.GetCurrentColorToken().TextMutedNRGBA()
		clr = color.NRGBA{R: clr.R, G: clr.G, B: clr.B, A: 80}
		bg = t.tc.GetCurrentColorToken().DisabledNRGBA()
	}
	if hovered {
		bg = color.NRGBA{R: clr.R, G: clr.G, B: clr.B, A: 54}
	}

	return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return utils.RoundedSurface(gtx, unit.Dp(t.iconRadius), bg, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return iconify.DefaultIconify.Layout(gtx, icon, unit.Dp(16), clr)
			})
		})
	})
}

func (t *transcriptFollower) transcriptTimestampWidth() unit.Dp {
	if t.compactTimestamps {
		return unit.Dp(54)
	}
	return unit.Dp(78)
}

func (t *transcriptFollower) transcriptTimestampText(timestamp string) string {
	timestamp = strings.TrimSpace(timestamp)
	if timestamp == "" {
		return ""
	}
	if !t.compactTimestamps {
		return timestamp
	}
	fields := strings.Fields(timestamp)
	if len(fields) > 0 {
		return fields[len(fields)-1]
	}
	return timestamp
}

func (t *transcriptFollower) toggleTranscriptRowTranslation(ctx context.Context, rowKey string) {
	row, ok := t.transcriptRowByKey(rowKey)
	if !ok || row.Info {
		return
	}
	if t.rowTranslationShown[row.Key] {
		t.rowTranslationShown[row.Key] = false
		return
	}
	key := t.rowTranslationCacheKey(row)
	if key == "" {
		//t.showError("Translate Row Failed", "Select a target language before translating a row.")
		return
	}
	if _, ok := t.rowTranslations[key]; ok {
		t.rowTranslationShown[row.Key] = true
		return
	}
	//todo move to backend
	entry, ok, err := translation.Load(t.activeGameName, row.Text, t.selectedTargetLanguage)
	if err != nil {
		//p.showError("Translate Row Failed", err.Error())
		return
	}
	if ok {
		t.rowTranslations[key] = entry.Translation
		t.rowTranslationShown[row.Key] = true
		return
	}
	t.rowTranslationShown[row.Key] = true
	//p.generateTranscriptRowTranslation(ctx, w, row, key)
}
