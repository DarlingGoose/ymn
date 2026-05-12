package transcript

import (
	"context"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
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

	backend backend.Backend

	transcriptList widget.List
	gameDropdown   *dropdowns.Dropdown

	autoTranslateToggle *toggles.Toggle
	runGameButton       *components.IconButton
	stopGameButton      *components.IconButton

	running  bool
	starting bool
	stopping bool

	selectedGameName string
	gameStatus       string
	following        bool

	followCtx    context.Context
	followCancel context.CancelFunc

	transcriptFollower transcriptFollower
	sentenceAnalysis   *SentenceAnalysis
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
		backend:             backend,
		gameStatus:          "No game selected",
		transcriptFollower:  newTranscriptFollower(th, backend),
		sentenceAnalysis:    NewSentenceAnalysis(th, backend),
		autoTranslateToggle: toggles.NewToggle("Auto Translate", false),
	}
	playIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:play")
	stopIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:square")

	ui.runGameButton = components.NewIconButton(
		"Run",
		&widget.Clickable{},
		playIcon,
	).WithThemeClient(tc)

	ui.runGameButton.FillWidth = false
	ui.runGameButton.TextCollapseMode = components.TextCollapseNever
	ui.runGameButton.MinWidth = unit.Dp(92)
	ui.runGameButton.Height = unit.Dp(38)
	ui.runGameButton.Radius = unit.Dp(10)

	ui.stopGameButton = components.NewIconButton(
		"Stop",
		&widget.Clickable{},
		stopIcon,
	).WithThemeClient(tc)

	ui.stopGameButton.FillWidth = false
	ui.stopGameButton.TextCollapseMode = components.TextCollapseNever
	ui.stopGameButton.MinWidth = unit.Dp(92)
	ui.stopGameButton.Height = unit.Dp(38)
	ui.stopGameButton.Radius = unit.Dp(10)

	ui.gameDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
		if !valid {
			return
		}

		ui.selectedGameName = item.Value
		ui.backend.SelectGameByName(item.Value)

		g := ui.gameByName[item.Value]
		if g == nil {
			ui.gameStatus = "Game not found"
			ui.stopFollowing()
			ui.transcriptFollower.Reset(item.Value)
			return
		}

		ui.gameStatus = "Selected"
		ui.transcriptFollower.SetGame(g.Name)

		ui.StartFollowingGame(context.Background(), g)
	})

	ui.transcriptFollower.WithSelectedRow(func(row transcriptRow) {
		ui.sentenceAnalysis.SetSentence(&row) //todo doesn't work or show it??
	})
	ui.ReloadGames()
	return ui
}

func (ui *TranscriptUI) update(gtx layout.Context) {
	ui.transcriptFollower.HandeEvents(gtx)
	ui.transcriptFollower.WithAutoTranslate(ui.autoTranslateToggle.Changed(gtx))
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
	if ui.runGameButton != nil {
		ui.runGameButton.WithThemeClient(tc)
	}
	if ui.stopGameButton != nil {
		ui.stopGameButton.WithThemeClient(tc)
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

	if ui.theme.ColorTweenRunning() || ui.following {
		gtx.Execute(op.InvalidateCmd{})
	}

	ui.update(gtx)

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if ui.gameDropdown == nil {
				return layout.Dimensions{}
			}

			return ui.gameDropdown.Layout(gtx, &ui.Overlay)
		}),

		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			status := strings.TrimSpace(ui.gameStatus)
			if status == "" {
				status = "Idle"
			}

			return theme.ThemedLabel(
				gtx,
				ui.th,
				ui.theme,
				theme.TextRoleBodySmall,
				theme.ThemeColorTextMuted,
				status,
			)
		}),

		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutGameActionButtons(gtx)
		}),
		layout.Rigid(bareutils.SpacerW(unit.Dp(10))),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.autoTranslateToggle.Layout(gtx)
		}),
	)
}
func (ui *TranscriptUI) layoutGameActionButtons(gtx layout.Context) layout.Dimensions {
	if ui.runGameButton == nil || ui.stopGameButton == nil {
		return layout.Dimensions{}
	}

	isRunning := ui.running
	if ui.backend != nil && ui.backend.IsGameRunning() {
		isRunning = true
	}

	ui.runGameButton.Disabled = ui.starting || ui.stopping || isRunning || ui.selectedGame() == nil
	ui.runGameButton.SetLoading(ui.starting)

	ui.stopGameButton.Disabled = ui.starting || ui.stopping || !isRunning
	ui.stopGameButton.SetLoading(ui.stopping)

	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.runGameButton.Clicked(gtx) {
				ui.RunSelectedGame(context.Background())
			}
			return ui.runGameButton.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.stopGameButton.Clicked(gtx) {
				ui.StopSelectedGame()
			}
			return ui.stopGameButton.Layout(gtx)
		}),
	)
}

func (ui *TranscriptUI) StopSelectedGame() {
	if ui == nil || ui.backend == nil {
		return
	}

	g := ui.selectedGame()

	ui.stopping = true
	ui.gameStatus = "Stopping..."

	if g != nil {
		ui.transcriptFollower.AddRows(transcriptRow{
			Info: true,
			Text: "Stopping " + g.Name + "...",
		})
	} else {
		ui.transcriptFollower.AddRows(transcriptRow{
			Info: true,
			Text: "Stopping current game...",
		})
	}

	ui.stopFollowing()

	go func() {
		ui.backend.StopCurrentGame()

		ui.running = false
		ui.stopping = false
		ui.gameStatus = "Stopped"

		if g != nil {
			ui.transcriptFollower.AddRows(transcriptRow{
				Info: true,
				Text: "Stopped " + g.Name,
			})
			return
		}

		ui.transcriptFollower.AddRows(transcriptRow{
			Info: true,
			Text: "Stopped current game.",
		})
	}()
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
		return ui.sentenceAnalysis.Layout(gtx)
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

func (ui *TranscriptUI) selectedGame() *game.Game {
	if ui == nil || ui.selectedGameName == "" {
		return nil
	}

	return ui.gameByName[ui.selectedGameName]
}

func (ui *TranscriptUI) stopFollowing() {
	if ui.followCancel != nil {
		ui.followCancel()
		ui.followCancel = nil
	}

	ui.followCtx = nil
	ui.following = false
}

func (ui *TranscriptUI) StartFollowingGame(ctx context.Context, g *game.Game) {
	if ui == nil || ui.backend == nil || g == nil {
		return
	}

	ui.stopFollowing()
	if !ui.backend.IsGameRunning() {
		return
	}

	followCtx, cancel := context.WithCancel(ctx)
	ui.followCtx = followCtx
	ui.followCancel = cancel
	ui.following = true
	ui.gameStatus = "Following logs"

	ch, err := ui.backend.FollowGameText(followCtx, g)
	if err != nil {
		ui.gameStatus = "Follow failed"
		ui.transcriptFollower.AddRows(transcriptRow{
			Info: true,
			Text: "Failed to follow game logs: " + err.Error(),
		})
		ui.following = false
		return
	}

	ui.transcriptFollower.AddRows(transcriptRow{
		Info: true,
		Text: "Following logs for " + g.Name,
	})

	go func() {
		for {
			select {
			case <-followCtx.Done():
				return

			case line, ok := <-ch:
				if !ok {
					ui.transcriptFollower.AddRows(transcriptRow{
						Info: true,
						Text: "Log stream closed for " + g.Name,
					})
					return
				}

				ui.transcriptFollower.AddRows(transcriptRowFromEngineLine(line))
			}
		}
	}()
}

func (ui *TranscriptUI) RunSelectedGame(ctx context.Context) {
	if ui == nil || ui.backend == nil {
		return
	}

	g := ui.selectedGame()
	if g == nil {
		ui.gameStatus = "Select a game first"
		return
	}

	if ui.starting || ui.running {
		return
	}

	ui.starting = true
	ui.running = false
	ui.gameStatus = "Starting..."

	ui.transcriptFollower.AddRows(transcriptRow{
		Info: true,
		Text: "Starting " + g.Name + "...",
	})

	go func() {
		proc, err := ui.backend.RunGame(ctx, g)
		if err != nil {
			ui.starting = false
			ui.running = false
			ui.gameStatus = "Run failed"
			ui.transcriptFollower.AddRows(transcriptRow{
				Info: true,
				Text: "Failed to run " + g.Name + ": " + err.Error(),
			})
			return
		}

		ui.starting = false
		ui.running = true
		ui.gameStatus = "Running"

		if proc != nil {
			ui.transcriptFollower.AddRows(transcriptRow{
				Info: true,
				Text: "Started " + g.Name,
			})
		}

		ui.StartFollowingGame(ctx, g)
	}()
}
