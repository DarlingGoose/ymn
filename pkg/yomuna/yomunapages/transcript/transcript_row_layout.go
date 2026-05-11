package transcript

import (
	"image/color"
	"strings"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

func (t *transcriptFollower) Layout(gtx layout.Context) layout.Dimensions {
	rows := t.GetRows()
	if len(rows) == 0 {
		return layout.Dimensions{}
		//return p.layoutTranscriptIdleState(gtx)
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
				)
			})
		})
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
					return t.layoutRowIcon(gtx, "mdi:information-outline", true)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					//todo update themed lable to allow for more dynamic sizes
					return theme.ThemedLabel(gtx, t.th, t.tc, theme.TextRoleBody, theme.ThemeColorTextMuted, row.Text)
				}),
			)
		})
	})
}

func (t *transcriptFollower) layoutRowIcon(gtx layout.Context, icon string, enabled bool) layout.Dimensions {
	clr := t.tc.GetCurrentColorToken().TextMutedNRGBA()
	if !enabled {
		clr = color.NRGBA{R: clr.R, G: clr.G, B: clr.B, A: 80}
	}
	return layout.Inset{Left: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return bareutils.RoundedSurface(gtx, color.NRGBA{}, unit.Dp(t.iconRadius), func(gtx layout.Context) layout.Dimensions {
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
