package transcript

import (
	"context"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/layouts/split"
	"github.com/DarlingGoose/wgl/pkg/yomuna/backend"
)

type TranscriptUI struct {
	th         *material.Theme
	theme      *theme.Client
	Overlay    overlay.Overlay
	bodySplit  split.SplitH
	gameByName map[string]*game.Game
	//backend
	//Perf    *components.PerformanceMonitor
	backend        backend.Backend
	transcriptList widget.List
	gameDropdown   *dropdowns.Dropdown

	transcriptFollower transcriptFollower
}

func NewTranscriptUI(th *material.Theme, tc *theme.Client, backend backend.Backend) *TranscriptUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	ui := &TranscriptUI{
		th:           th,
		theme:        tc,
		gameDropdown: dropdowns.NewDropdown([]dropdowns.DropdownItem{}),
		bodySplit: split.SplitH{
			Ratio:    0,
			Bar:      unit.Dp(4),
			MinRatio: -0.70,
			MaxRatio: 0.70,
		},
		//Perf: components.NewPerformanceMonitor(),

		backend:            backend,
		transcriptFollower: newTranscriptFollower(th),
	}

	ui.gameDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
		if !valid {
			return
		}

		ui.backend.SelectGameByName(item.Value)

		g := ui.gameByName[item.Value]
		if g == nil {
			return
		}

		//ui.StartFollowingGame(context.Background(), g)
	})
	go func() {
		for range 20 {
			ui.transcriptFollower.AddRows(transcriptRow{
				Time: time.Now().Format(time.RFC3339),
				Text: "hello there:" + time.Now().Format("04:05")})
			time.Sleep(time.Second)
		}
	}()
	ui.ReloadGames()
	return ui
}

func (ui *TranscriptUI) update(gtx layout.Context) {
	ui.transcriptFollower.HandeEvents(gtx)
}

func (ui *TranscriptUI) WithThemeClient(tc *theme.Client) *TranscriptUI {
	if ui == nil {
		return ui
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	ui.theme = tc
	if ui.gameDropdown != nil {
		ui.gameDropdown.WithThemeClient(tc)
	}
	return ui
}

func (ui *TranscriptUI) ReloadGames() {
	ui.gameByName = make(map[string]*game.Game)

	var items []dropdowns.DropdownItem
	for _, g := range ui.backend.GetGames() {
		if g == nil {
			continue
		}

		ui.gameByName[g.Name] = g

		items = append(items, dropdowns.DropdownItem{
			Label: g.Name,
			Value: g.Name,
		})
	}

	ui.gameDropdown.SetItems(items)
}
func (ui *TranscriptUI) Layout(gtx layout.Context, ctx context.Context) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ui.update(gtx)
	return panel.NewBackgroundPanel(ui.theme).
		WithFillMax(true).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(unit.Dp(15)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {

				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return ui.Overlay.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return ui.layoutHeader(gtx)
						})
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return ui.bodySplit.Layout(
							gtx,
							func(gtx layout.Context) layout.Dimensions {
								return ui.layoutTranscript(gtx)
							},
							func(gtx layout.Context) layout.Dimensions {
								return ui.layoutDetails(gtx)
							},
						)
					}),
				)

			})
		})
}

func (ui *TranscriptUI) layoutHeader(gtx layout.Context) layout.Dimensions {
	if ui.th == nil {
		ui.th = material.NewTheme()
	}

	if ui.theme.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}
	ui.update(gtx)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.gameDropdown == nil {
				return layout.Dimensions{}
			}

			return ui.gameDropdown.Layout(gtx, &ui.Overlay)
		}),
	)
}

func (ui *TranscriptUI) layoutTranscript(gtx layout.Context) layout.Dimensions {
	return panel.NewBackgroundPanel(ui.theme).
		WithFillMax(true).
		WithRadius(unit.Dp(10)).
		WithRole(panel.BackgroundRoleSurface).Layout(gtx, func(gtx layout.Context) layout.Dimensions {

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.transcriptFollower.Layout(gtx)
			}))
	})
}

func (ui *TranscriptUI) layoutDetails(gtx layout.Context) layout.Dimensions {
	return panel.NewBackgroundPanel(ui.theme).
		WithFillMax(true).
		WithRadius(unit.Dp(10)).
		WithRole(panel.BackgroundRoleSurfaceAlt).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{}
	})
}

//func (ui *TranscriptUI) layoutLiveTranscriptCard(gtx layout.Context) layout.Dimensions {
//	bg := panel.NewBackgroundPanel(ui.theme).
//		WithFillMax(true).
//		WithRadius(unit.Dp(10)).
//		WithRole(panel.BackgroundRoleBackground)
//	return bg.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
//			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return ui.layoutCardHeader(gtx, "Live Transcript", "Scanning mode: saved words are highlighted inline")
//			}),
//				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
//				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
//					if !ui.backend.IsGameRunning() { //should this be a bool var instead? sometimes it doesn't trigger on
//						return ui.layoutTranscriptIdleState(gtx)
//					}
//
//					return ui.layoutTranscriptIdleState(gtx) // todo ui.layoutTranscriptEditor(gtx)
//				}),
//			)
//		})
//	})
//
//}

func (ui *TranscriptUI) layoutCardHeader(gtx layout.Context, title, hint string) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				ui.th,
				ui.theme,
				theme.TextRoleH4,
				theme.ThemeColorTextPrimary,
				title,
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if strings.TrimSpace(hint) == "" {
				return layout.Dimensions{}
			}

			return theme.ThemedLabel(
				gtx,
				ui.th,
				ui.theme,
				theme.TextRoleCaption,
				theme.ThemeColorTextMuted,
				hint,
			)
		}),
	)
}

func (ui *TranscriptUI) layoutTranscriptIdleState(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Vertical,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(
					gtx,
					ui.th,
					ui.theme,
					theme.TextRoleH4,
					theme.ThemeColorTextPrimary,
					"Transcript Hidden",
				)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(
					gtx,
					ui.th,
					ui.theme,
					theme.TextRoleBody,
					theme.ThemeColorTextMuted,
					"Start the game to show live transcript text here.",
				)
			}),
			layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(
					gtx,
					ui.th,
					ui.theme,
					theme.TextRoleBodySmall,
					theme.ThemeColorTextMuted,
					"Captured lines will appear here as the backend receives hook text.",
				)
			}),
		)
	})
}

func (ui *TranscriptUI) WithMaxTranscriptRows(maxRows int) *TranscriptUI {
	if ui == nil {
		return ui
	}

	if maxRows <= 0 {
		maxRows = 200
	}

	ui.transcriptFollower.maxTranscriptRows = maxRows
	return ui
}
