package transcript

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/tr/pkg/textractor"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/layouts/split"
	"github.com/DarlingGoose/ymn/pkg/yomuna/backend"
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
	hookDropdown   *dropdowns.Dropdown

	autoTranslateToggle *toggles.Toggle
	runGameButton       *components.IconButton
	stopGameButton      *components.IconButton

	running  bool
	starting bool
	stopping bool

	selectedGameName string
	gameStatus       string
	hookStatus       string
	hookDropdownOpen bool
	selectedHook     string
	hookRefreshToken int
	hookRefreshBusy  bool
	following        bool

	followCtx    context.Context
	followCancel context.CancelFunc

	transcriptFollower transcriptFollower
	sentenceAnalysis   *SentenceAnalysis
	invalidate         func()
	preferences        transcriptPreferences
}

func NewTranscriptUI(th *material.Theme, tc *theme.Client, backend backend.Backend) *TranscriptUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	prefs := loadTranscriptPreferences()

	ui := &TranscriptUI{
		th:           th,
		theme:        tc,
		gameDropdown: dropdowns.NewDropdown([]dropdowns.DropdownItem{}),
		hookDropdown: newHookDropdown(),
		bodySplit: split.SplitH{
			Ratio:    0,
			Bar:      unit.Dp(8),
			MinRatio: -0.70,
			MaxRatio: 0.70,
		},
		backend:             backend,
		gameStatus:          "No game selected",
		transcriptFollower:  newTranscriptFollower(th, backend),
		sentenceAnalysis:    NewSentenceAnalysis(th, backend),
		autoTranslateToggle: toggles.NewToggle("Auto Translate", false),
		preferences:         prefs,
	}
	ui.sentenceAnalysis.SetLookupFontSize(unit.Sp(prefs.LookupFontSizeSp))
	ui.sentenceAnalysis.sentenceFontSize = unit.Sp(prefs.SentenceFontSizeSp)
	ui.transcriptFollower.fontSize = unit.Sp(prefs.TranscriptFontSizeSp)
	ui.transcriptFollower.SetMaxTranscriptRows(prefs.MaxTranscriptRows)

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
		ui.selectGameByName(item.Value, true)
	})
	ui.hookDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
		if !valid {
			return
		}
		ui.setSelectedHookFilter(context.Background(), item.Value)
	})

	ui.transcriptFollower.WithSelectedRow(func(row transcriptRow) {
		ui.sentenceAnalysis.SetSentence(&row)
		ui.invalidateUI()
	})
	ui.transcriptFollower.WithThemeClient(tc)
	ui.sentenceAnalysis.WithThemeClient(tc)
	ui.ReloadGames()
	return ui
}

func (ui *TranscriptUI) WithInvalidate(invalidate func()) *TranscriptUI {
	if ui == nil {
		return ui
	}
	ui.invalidate = invalidate
	ui.transcriptFollower.WithInvalidate(invalidate)
	ui.sentenceAnalysis.WithInvalidate(invalidate)
	return ui
}

func (ui *TranscriptUI) invalidateUI() {
	if ui != nil && ui.invalidate != nil {
		ui.invalidate()
	}
}

func (ui *TranscriptUI) update(gtx layout.Context) {
	ui.transcriptFollower.HandeEvents(gtx)
	ui.transcriptFollower.SetShowHookLabels(ui.hookDropdownOpen && strings.TrimSpace(ui.selectedHook) == "")
	ui.autoTranslateToggle.Update(gtx)
	ui.transcriptFollower.WithAutoTranslate(ui.autoTranslateToggle.Checked)
	ui.sentenceAnalysis.HandeEvents(gtx)
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
	if ui.hookDropdown != nil {
		ui.hookDropdown.WithThemeClient(tc)
	}
	if ui.runGameButton != nil {
		ui.runGameButton.WithThemeClient(tc)
	}
	if ui.stopGameButton != nil {
		ui.stopGameButton.WithThemeClient(tc)
	}
	ui.transcriptFollower.WithThemeClient(tc)
	if ui.sentenceAnalysis != nil {
		ui.sentenceAnalysis.WithThemeClient(tc)
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
	if ui.selectedGameName == "" && ui.preferences.SelectedGameName != "" {
		ui.selectGameByName(ui.preferences.SelectedGameName, false)
	}
}

func (ui *TranscriptUI) selectGameByName(name string, followIfRunning bool) {
	if ui == nil {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if ui.gameDropdown != nil {
		ui.gameDropdown.SelectItem(name)
	}

	ui.selectedGameName = name
	if ui.backend != nil {
		ui.backend.SelectGameByName(name)
	}

	g := ui.gameByName[name]
	if g == nil {
		ui.gameStatus = "Game not found"
		ui.stopFollowing()
		ui.transcriptFollower.Reset(name)
		ui.sentenceAnalysis.Reset()
		return
	}

	ui.gameStatus = "Selected"
	ui.selectedHook = firstTextHookFilter(g.TextHookFilter)
	ui.setHookDropdownItems(nil, false)
	ui.transcriptFollower.SetGame(g.Name)
	ui.sentenceAnalysis.Reset()
	ui.refreshHookDropdown(context.Background(), g)

	if followIfRunning {
		ui.StartFollowingGame(context.Background(), g)
	}
}

func (ui *TranscriptUI) SavePreferences() error {
	if ui == nil {
		return nil
	}
	prefs := transcriptPreferences{
		SelectedGameName:     strings.TrimSpace(ui.selectedGameName),
		LookupFontSizeSp:     spToFloat(ui.sentenceAnalysis.LookupFontSize()),
		SentenceFontSizeSp:   spToFloat(ui.sentenceAnalysis.sentenceFontSize),
		TranscriptFontSizeSp: spToFloat(ui.transcriptFollower.fontSize),
		MaxTranscriptRows:    ui.transcriptFollower.MaxTranscriptRows(),
	}
	if err := saveTranscriptPreferences(prefs); err != nil {
		return err
	}
	ui.preferences = prefs
	return nil
}

func (ui *TranscriptUI) SelectedGameName() string {
	if ui == nil {
		return ""
	}
	return strings.TrimSpace(ui.selectedGameName)
}

func (ui *TranscriptUI) TranscriptFontSize() unit.Sp {
	if ui == nil {
		return unit.Sp(22)
	}
	return ui.transcriptFollower.fontSize
}

func (ui *TranscriptUI) SetTranscriptFontSize(size unit.Sp) {
	if ui == nil {
		return
	}
	ui.transcriptFollower.fontSize = clampFontSize(size, unit.Sp(14), unit.Sp(34), unit.Sp(22))
	ui.invalidateUI()
}

func (ui *TranscriptUI) SentenceFontSize() unit.Sp {
	if ui == nil || ui.sentenceAnalysis == nil {
		return unit.Sp(24)
	}
	return ui.sentenceAnalysis.sentenceFontSize
}

func (ui *TranscriptUI) SetSentenceFontSize(size unit.Sp) {
	if ui == nil || ui.sentenceAnalysis == nil {
		return
	}
	ui.sentenceAnalysis.sentenceFontSize = clampFontSize(size, unit.Sp(16), unit.Sp(40), unit.Sp(24))
	ui.invalidateUI()
}

func (ui *TranscriptUI) LookupFontSize() unit.Sp {
	if ui == nil || ui.sentenceAnalysis == nil {
		return unit.Sp(14)
	}
	return ui.sentenceAnalysis.LookupFontSize()
}

func (ui *TranscriptUI) SetLookupFontSize(size unit.Sp) {
	if ui == nil || ui.sentenceAnalysis == nil {
		return
	}
	ui.sentenceAnalysis.SetLookupFontSize(size)
	ui.invalidateUI()
}

func (ui *TranscriptUI) MaxTranscriptRows() int {
	if ui == nil {
		return 200
	}
	return ui.transcriptFollower.MaxTranscriptRows()
}

func (ui *TranscriptUI) SetMaxTranscriptRows(maxRows int) {
	if ui == nil {
		return
	}
	ui.transcriptFollower.SetMaxTranscriptRows(maxRows)
	ui.invalidateUI()
}

func clampFontSize(size, min, max, fallback unit.Sp) unit.Sp {
	if size <= 0 {
		return fallback
	}
	if size < min {
		return min
	}
	if size > max {
		return max
	}
	return size
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
	ui.layoutHeaderResponsiveControls(gtx)

	gap := unit.Dp(10)
	if gtx.Constraints.Max.X > 0 && gtx.Constraints.Max.X < gtx.Dp(unit.Dp(760)) {
		gap = unit.Dp(6)
	}

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

		layout.Rigid(bareutils.SpacerW(gap)),

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

		layout.Rigid(bareutils.SpacerW(gap)),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutHookDropdown(gtx)
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.hookDropdownOpen {
				return bareutils.SpacerW(gap)(gtx)
			}
			return layout.Dimensions{}
		}),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutGameActionButtons(gtx)
		}),
		layout.Rigid(bareutils.SpacerW(gap)),

		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.autoTranslateToggle.Layout(gtx)
		}),
	)
}

func (ui *TranscriptUI) layoutHeaderResponsiveControls(gtx layout.Context) {
	if ui == nil || ui.autoTranslateToggle == nil {
		return
	}
	width := gtx.Constraints.Max.X
	switch {
	case width > 0 && width < gtx.Dp(unit.Dp(650)):
		ui.autoTranslateToggle.WithLabel("")
	case width > 0 && width < gtx.Dp(unit.Dp(820)):
		ui.autoTranslateToggle.WithLabel("Auto")
	default:
		ui.autoTranslateToggle.WithLabel("Auto Translate")
	}
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

func (ui *TranscriptUI) layoutHookDropdown(gtx layout.Context) layout.Dimensions {
	if ui == nil || ui.hookDropdown == nil || !ui.hookDropdownOpen {
		return layout.Dimensions{}
	}
	return ui.hookDropdown.Layout(gtx, &ui.Overlay)
}

func newHookDropdown() *dropdowns.Dropdown {
	d := dropdowns.NewDropdown([]dropdowns.DropdownItem{
		{Label: "All Hooks", Value: ""},
	})
	d.Width = unit.Dp(240)
	d.Height = unit.Dp(38)
	d.ItemHeight = unit.Dp(34)
	d.MaxMenuHeight = unit.Dp(340)
	d.Radius = unit.Dp(10)
	d.Inset = unit.Dp(10)
	d.WithRole(theme.TextRoleBodySmall)
	return d
}

func (ui *TranscriptUI) refreshHookDropdown(ctx context.Context, g *game.Game) {
	if ui == nil || ui.backend == nil || g == nil {
		return
	}

	ui.hookRefreshBusy = true
	ui.hookRefreshToken++
	token := ui.hookRefreshToken
	gameName := g.Name
	ui.invalidateUI()

	go func() {
		hooks, hasTextractor, err := ui.backend.GetGameTextHooks(ctx, g)
		if ctx != nil && ctx.Err() != nil {
			return
		}
		if token != ui.hookRefreshToken || strings.TrimSpace(ui.selectedGameName) != strings.TrimSpace(gameName) {
			return
		}

		ui.hookRefreshBusy = false
		if err != nil {
			ui.hookStatus = err.Error()
			ui.setHookDropdownItems(nil, false)
			ui.invalidateUI()
			return
		}

		ui.hookStatus = ""
		ui.setHookDropdownItems(hooks, hasTextractor)
		ui.invalidateUI()
	}()
}

func (ui *TranscriptUI) setHookDropdownItems(hooks []string, hasTextractor bool) {
	if ui == nil || ui.hookDropdown == nil {
		return
	}

	ui.hookDropdownOpen = hasTextractor
	if !hasTextractor {
		ui.hookDropdown.SetItems([]dropdowns.DropdownItem{{Label: "All Hooks", Value: ""}})
		ui.hookDropdown.SelectItem("")
		return
	}

	selected := strings.TrimSpace(ui.selectedHook)
	selectedPresent := selected == ""
	seen := map[string]struct{}{}
	normalizedHooks := make([]string, 0, len(hooks))
	for _, hook := range hooks {
		hook = normalizeHookGroup(hook)
		if hook == "" {
			continue
		}
		if _, ok := seen[hook]; ok {
			continue
		}
		seen[hook] = struct{}{}
		if hook == selected {
			selectedPresent = true
		}
		normalizedHooks = append(normalizedHooks, hook)
	}
	sort.Strings(normalizedHooks)

	allHooksLabel := "All Hooks"
	if len(normalizedHooks) > 0 {
		allHooksLabel = "All Hooks (" + strconv.Itoa(len(normalizedHooks)) + ")"
	}
	items := []dropdowns.DropdownItem{{Label: allHooksLabel, Value: ""}}
	for _, hook := range normalizedHooks {
		items = append(items, dropdowns.DropdownItem{
			Label: hookDropdownLabel(hook),
			Value: hook,
		})
	}
	if !selectedPresent {
		items = append(items, dropdowns.DropdownItem{
			Label: hookDropdownLabel(selected),
			Value: selected,
		})
	}

	ui.hookDropdown.SetItems(items)
	ui.hookDropdown.SelectItem(selected)
}

func (ui *TranscriptUI) ensureHookDropdownOption(hook string) {
	if ui == nil || ui.hookDropdown == nil || !ui.hookDropdownOpen {
		return
	}
	hook = normalizeHookGroup(hook)
	if hook == "" {
		return
	}
	for _, item := range ui.hookDropdown.Items {
		if normalizeHookGroup(item.Value) == hook {
			return
		}
	}

	hooks := currentHookDropdownValues(ui.hookDropdown)
	hooks = append(hooks, hook)
	ui.setHookDropdownItems(hooks, true)
	ui.invalidateUI()
}

func (ui *TranscriptUI) setSelectedHookFilter(ctx context.Context, hook string) {
	if ui == nil {
		return
	}
	hook = normalizeHookGroup(hook)
	if hook == strings.TrimSpace(ui.selectedHook) {
		return
	}

	ui.selectedHook = hook
	ui.transcriptFollower.SetShowHookLabels(ui.hookDropdownOpen && hook == "")
	g := ui.selectedGame()
	if g == nil || ui.backend == nil {
		ui.invalidateUI()
		return
	}

	filters := []string(nil)
	if hook != "" {
		filters = []string{hook}
	}
	if err := ui.backend.SetGameTextHookFilter(g, filters); err != nil {
		ui.hookStatus = err.Error()
		ui.invalidateUI()
		return
	}

	g.TextHookFilter = append([]string(nil), filters...)
	ui.hookStatus = ""
	ui.setHookDropdownItems(currentHookDropdownValues(ui.hookDropdown), ui.hookDropdownOpen)
	ui.invalidateUI()

	if ui.following && ui.backend.IsGameRunning() {
		ui.transcriptFollower.Reset(g.Name)
		ui.StartFollowingGame(ctx, g)
	}
}

func (ui *TranscriptUI) transcriptLinePassesHookFilter(hook string) bool {
	if ui == nil {
		return true
	}
	selected := normalizeHookGroup(ui.selectedHook)
	if selected == "" {
		return true
	}
	return hookMatchesFilter(selected, hook)
}

func currentHookDropdownValues(d *dropdowns.Dropdown) []string {
	if d == nil {
		return nil
	}
	out := make([]string, 0, len(d.Items))
	for _, item := range d.Items {
		if value := normalizeHookGroup(item.Value); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func firstTextHookFilter(filters []string) string {
	for _, filter := range filters {
		if hook := normalizeHookGroup(filter); hook != "" {
			return hook
		}
	}
	return ""
}

func normalizeHookGroup(hook string) string {
	return strings.TrimSpace(textractor.HookGroup(hook))
}

func hookDropdownLabel(hook string) string {
	hook = normalizeHookGroup(hook)
	if hook == "" {
		return "All Hooks"
	}
	return compactHookLabel(hook)
}

func hookMatchesFilter(filter string, hook string) bool {
	filter = normalizeHookGroup(filter)
	hook = normalizeHookGroup(hook)
	if filter == "" {
		return true
	}
	if hook == "" {
		return false
	}
	if strings.EqualFold(filter, hook) {
		return true
	}
	return false
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
		ui.invalidateUI()

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
		gtx.Constraints.Min = gtx.Constraints.Max
		return ui.transcriptFollower.Layout(gtx)
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
	ui.SetMaxTranscriptRows(maxRows)
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
	ui.invalidateUI()
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
	ui.invalidateUI()

	followGame := *g
	followGame.TextHookFilter = nil

	ch, err := ui.backend.FollowGameText(followCtx, &followGame)
	if err != nil {
		ui.gameStatus = "Follow failed"
		ui.transcriptFollower.AddRows(transcriptRow{
			Info: true,
			Text: "Failed to follow game logs: " + err.Error(),
		})
		ui.following = false
		ui.invalidateUI()
		return
	}

	ui.transcriptFollower.AddRows(transcriptRow{
		Info: true,
		Text: "Following logs for " + g.Name,
	})
	ui.refreshHookDropdown(followCtx, g)

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
					ui.following = false
					ui.invalidateUI()
					return
				}

				ui.ensureHookDropdownOption(line.Hook)
				if !ui.transcriptLinePassesHookFilter(line.Hook) {
					continue
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
			ui.invalidateUI()
			ui.transcriptFollower.AddRows(transcriptRow{
				Info: true,
				Text: "Failed to run " + g.Name + ": " + err.Error(),
			})
			return
		}

		ui.starting = false
		ui.running = true
		ui.gameStatus = "Running"
		ui.invalidateUI()

		if proc != nil {
			ui.transcriptFollower.AddRows(transcriptRow{
				Info: true,
				Text: "Started " + g.Name,
			})
		}

		ui.StartFollowingGame(ctx, g)
	}()
}
