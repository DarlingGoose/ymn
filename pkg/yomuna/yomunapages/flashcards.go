package yomunapages

import (
	"context"
	"fmt"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/anki"
	flashcards "github.com/DarlingGoose/ymn/pkg/flashcard"
	"github.com/DarlingGoose/ymn/pkg/util"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/modal"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/layouts/grid"
	"github.com/DarlingGoose/ymn/pkg/yomuna/backend"
)

type FlashcardsUI struct {
	th      *material.Theme
	theme   *theme.Client
	backend backend.Backend

	grid   *grid.ScrollGrid
	cards  []flashcards.Flashcard
	status string

	newButton  *components.IconButton
	syncButton *components.IconButton
	saveButton *components.IconButton
	deleteBtn  *components.IconButton
	cancelBtn  *components.IconButton

	editButtons   map[string]*components.IconButton
	deleteButtons map[string]*components.IconButton

	syncing    bool
	syncGame   string
	syncResult chan flashcardSyncResult

	editorModal *modal.Modal
	editingID   string
	wordEditor  widget.Editor
	meaningEdit widget.Editor
	readingEdit widget.Editor
	sourceEdit  widget.Editor
}

type flashcardSyncResult struct {
	gameName string
	result   anki.SyncResult
	err      error
}

func NewFlashcardsUI(th *material.Theme, tc *theme.Client, backend backend.Backend) *FlashcardsUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	g := grid.NewScrollGrid()
	g.Grid.MinCellWidth = unit.Dp(260)
	g.Grid.Gap = unit.Dp(12)
	g.Grid.MinColumns = 1

	ui := &FlashcardsUI{
		th:            th,
		theme:         tc,
		backend:       backend,
		grid:          g,
		status:        "Select a game to review saved flashcards.",
		editButtons:   make(map[string]*components.IconButton),
		deleteButtons: make(map[string]*components.IconButton),
		syncResult:    make(chan flashcardSyncResult, 1),
	}
	ui.wordEditor.SingleLine = true
	ui.readingEdit.SingleLine = true
	ui.meaningEdit.SingleLine = false
	ui.sourceEdit.SingleLine = false

	ui.newButton = components.NewIconButton("New", nil, mustFlashcardIcon("lucide:plus")).WithThemeClient(tc)
	ui.syncButton = components.NewIconButton("Sync Anki", nil, mustFlashcardIcon("lucide:cloud-upload")).WithThemeClient(tc)
	ui.saveButton = components.NewIconButton("Save", nil, mustFlashcardIcon("lucide:save")).WithThemeClient(tc)
	ui.deleteBtn = components.NewIconButton("Delete", nil, mustFlashcardIcon("lucide:trash-2")).WithThemeClient(tc)
	ui.cancelBtn = components.NewIconButton("Cancel", nil, mustFlashcardIcon("lucide:x")).WithThemeClient(tc)
	for _, btn := range []*components.IconButton{ui.newButton, ui.syncButton, ui.saveButton, ui.deleteBtn, ui.cancelBtn} {
		btn.FillWidth = false
		btn.MinWidth = unit.Dp(96)
		btn.Height = unit.Dp(38)
		btn.Radius = unit.Dp(10)
	}
	ui.deleteBtn.MinWidth = unit.Dp(104)
	loaderIcon := mustFlashcardIcon("lucide:loader-circle")
	ui.syncButton.LoadingIcon = loaderIcon

	ui.editorModal = modal.New("flashcard-editor", "Flashcard", ui.layoutEditorModal).
		WithMaterialTheme(th).
		WithThemeClient(tc).
		WithSize(unit.Dp(560), unit.Dp(620)).
		WithFooter(ui.layoutEditorFooter)

	return ui
}

func (ui *FlashcardsUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}

	gameName := ui.activeGameName()
	ui.drainSyncResults()
	ui.reload(gameName)
	ui.handleEvents(gtx, gameName)
	if layer != nil && ui.editorModal != nil && ui.editorModal.Visible {
		layer.Add(gtx, ui.editorModal)
	}

	return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.layoutHeader(gtx, gameName)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutCards(gtx, ui.cards)
			}),
		)
	})
}

func (ui *FlashcardsUI) handleEvents(gtx layout.Context, gameName string) {
	if ui.newButton.Clicked(gtx) {
		ui.openNewModal()
	}
	if ui.syncButton.Clicked(gtx) {
		ui.startSyncAnki(gameName)
	}
	if ui.saveButton.Clicked(gtx) {
		ui.saveModal(gameName)
	}
	if ui.deleteBtn.Clicked(gtx) {
		ui.deleteCard(gameName, ui.editingID)
		if ui.editorModal != nil {
			ui.editorModal.Dismiss()
		}
	}
	if ui.cancelBtn.Clicked(gtx) && ui.editorModal != nil {
		ui.editorModal.Dismiss()
	}
	for id, btn := range ui.editButtons {
		if btn.Clicked(gtx) {
			ui.openEditModal(id)
		}
	}
	for id, btn := range ui.deleteButtons {
		if btn.Clicked(gtx) {
			ui.deleteCard(gameName, id)
		}
	}
}

func (ui *FlashcardsUI) layoutHeader(gtx layout.Context, gameName string) layout.Dimensions {
	title := "Flashcards"
	if strings.TrimSpace(gameName) != "" {
		title = "Flashcards: " + gameName
	}
	ui.syncButton.SetLoading(ui.syncing)
	ui.syncButton.Disabled = strings.TrimSpace(gameName) == ""

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutCJKLabel(gtx, title, theme.TextRoleH2, theme.ThemeColorTextPrimary)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutCJKLabel(gtx, ui.status, theme.TextRoleBodySmall, theme.ThemeColorTextMuted)
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.newButton.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.syncButton.Layout(gtx)
		}),
	)
}

func (ui *FlashcardsUI) layoutCards(gtx layout.Context, cards []flashcards.Flashcard) layout.Dimensions {
	if len(cards) == 0 {
		return panel.NewBackgroundPanel(ui.theme).
			WithRole(panel.BackgroundRoleSurface).
			WithRadius(unit.Dp(8)).
			WithInset(layout.UniformInset(unit.Dp(16))).
			WithFillMax(false).
			Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return ui.layoutCJKLabel(gtx, "No flashcards saved for this game yet.", theme.TextRoleBody, theme.ThemeColorTextMuted)
			})
	}

	return grid.LayoutScrollSlice(gtx, ui.grid, cards, func(gtx layout.Context, card flashcards.Flashcard, index int) layout.Dimensions {
		return ui.layoutCard(gtx, card)
	})
}

func (ui *FlashcardsUI) layoutCard(gtx layout.Context, card flashcards.Flashcard) layout.Dimensions {
	return panel.NewBackgroundPanel(ui.theme).
		WithRole(panel.BackgroundRoleSurface).
		WithRadius(unit.Dp(8)).
		WithInset(layout.UniformInset(unit.Dp(14))).
		WithFillMax(false).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutCJKLabel(gtx, card.Text, theme.TextRoleH4, theme.ThemeColorTextPrimary)
				}),
			}
			if reading := strings.TrimSpace(card.Reading); reading != "" && reading != strings.TrimSpace(card.Text) {
				children = append(children,
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return ui.layoutCJKLabel(gtx, reading, theme.TextRoleCaption, theme.ThemeColorPrimary)
					}),
				)
			}
			if meaning := strings.TrimSpace(card.Meaning); meaning != "" {
				children = append(children,
					layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return ui.layoutCJKLabel(gtx, meaning, theme.TextRoleBodySmall, theme.ThemeColorTextSecondary)
					}),
				)
			}
			if source := strings.TrimSpace(card.SourceLine); source != "" {
				children = append(children,
					layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return ui.layoutCJKLabel(gtx, source, theme.TextRoleCaption, theme.ThemeColorTextMuted)
					}),
				)
			}
			children = append(children,
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutCardActions(gtx, card)
				}),
			)
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		})
}

func (ui *FlashcardsUI) layoutCardActions(gtx layout.Context, card flashcards.Flashcard) layout.Dimensions {
	edit := ui.editButton(card.ID)
	del := ui.deleteButton(card.ID)

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(edit.Layout),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(del.Layout),
	)
}

func (ui *FlashcardsUI) layoutEditorModal(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutEditorField(gtx, "Text", "Word or phrase", &ui.wordEditor, unit.Dp(42))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutEditorField(gtx, "Reading", "Optional reading", &ui.readingEdit, unit.Dp(42))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutEditorField(gtx, "Meaning", "Meaning", &ui.meaningEdit, unit.Dp(120))
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutEditorField(gtx, "Source", "Optional source sentence", &ui.sourceEdit, unit.Dp(96))
		}),
	)
}

func (ui *FlashcardsUI) layoutEditorField(gtx layout.Context, label, hint string, editor *widget.Editor, height unit.Dp) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.layoutCJKLabel(gtx, label, theme.TextRoleLabelSmall, theme.ThemeColorTextMuted)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(5)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return panel.NewBackgroundPanel(ui.theme).
				WithRole(panel.BackgroundRoleSurfaceAlt).
				WithRadius(unit.Dp(8)).
				WithInset(layout.UniformInset(unit.Dp(10))).
				WithFillMax(false).
				Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					heightPx := gtx.Dp(height)
					gtx.Constraints.Min.Y = heightPx
					gtx.Constraints.Max.Y = heightPx
					ed := material.Editor(ui.th, editor, hint)
					ed.Color = ui.theme.GetCurrentColorToken().TextPrimaryNRGBA()
					ed.HintColor = ui.theme.GetCurrentColorToken().TextMutedNRGBA()
					ed.Font.Typeface = font.Typeface("Noto Sans CJK JP")
					return ed.Layout(gtx)
				})
		}),
	)
}

func (ui *FlashcardsUI) layoutEditorFooter(gtx layout.Context) layout.Dimensions {
	children := []layout.FlexChild{
		layout.Rigid(ui.cancelBtn.Layout),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
	}
	if strings.TrimSpace(ui.editingID) != "" {
		children = append(children,
			layout.Rigid(ui.deleteBtn.Layout),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		)
	}
	children = append(children, layout.Rigid(ui.saveButton.Layout))

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
}

func (ui *FlashcardsUI) openNewModal() {
	ui.editingID = ""
	ui.wordEditor.SetText("")
	ui.readingEdit.SetText("")
	ui.meaningEdit.SetText("")
	ui.sourceEdit.SetText("")
	if ui.editorModal != nil {
		ui.editorModal.Title = "New Flashcard"
		ui.editorModal.Open()
	}
}

func (ui *FlashcardsUI) openEditModal(id string) {
	for _, card := range ui.cards {
		if card.ID != id {
			continue
		}
		ui.editingID = card.ID
		ui.wordEditor.SetText(card.Text)
		ui.readingEdit.SetText(card.Reading)
		ui.meaningEdit.SetText(card.Meaning)
		ui.sourceEdit.SetText(card.SourceLine)
		if ui.editorModal != nil {
			ui.editorModal.Title = "Edit Flashcard"
			ui.editorModal.Open()
		}
		return
	}
}

func (ui *FlashcardsUI) saveModal(gameName string) {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		ui.status = "Select a game before saving flashcards."
		return
	}

	card := flashcards.Flashcard{
		ID:         strings.TrimSpace(ui.editingID),
		GameName:   gameName,
		Text:       strings.TrimSpace(ui.wordEditor.Text()),
		Reading:    strings.TrimSpace(ui.readingEdit.Text()),
		Meaning:    strings.TrimSpace(ui.meaningEdit.Text()),
		SourceLine: strings.TrimSpace(ui.sourceEdit.Text()),
		AnkiDeck:   util.AnkiDeckName(gameName),
		AnkiModel:  flashcards.DefaultAnkiModel,
	}

	if strings.TrimSpace(ui.editingID) == "" {
		if err := flashcards.AddFlashcard(card); err != nil {
			ui.status = err.Error()
			return
		}
		ui.status = card.Text + " added."
	} else {
		for _, existing := range ui.cards {
			if existing.ID == ui.editingID {
				card.AnkiNoteID = existing.AnkiNoteID
				card.AnkiLastSyncAt = existing.AnkiLastSyncAt
				break
			}
		}
		if err := flashcards.UpdateFlashcard(card); err != nil {
			ui.status = err.Error()
			return
		}
		ui.status = card.Text + " saved."
	}
	ui.reload(gameName)
	if ui.editorModal != nil {
		ui.editorModal.Dismiss()
	}
}

func (ui *FlashcardsUI) deleteCard(gameName, id string) {
	gameName = strings.TrimSpace(gameName)
	id = strings.TrimSpace(id)
	if gameName == "" || id == "" {
		return
	}
	var target flashcards.Flashcard
	for _, card := range ui.cards {
		if card.ID == id {
			target = card
			break
		}
	}
	if target.ID == "" {
		ui.status = "Flashcard not found."
		return
	}
	client := anki.New(anki.DefaultAnkiConnectURL)
	if err := client.DeleteFlashcardFromAnki(target, anki.DefaultAnkiConnectURL, false); err != nil {
		ui.status = err.Error()
		return
	}
	if err := flashcards.DeleteFlashcard(gameName, id); err != nil {
		ui.status = err.Error()
		return
	}
	ui.status = target.Text + " deleted."
	ui.reload(gameName)
}

func (ui *FlashcardsUI) startSyncAnki(gameName string) {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		ui.status = "Select a game before syncing Anki."
		return
	}
	if ui.syncing {
		return
	}
	ui.syncing = true
	ui.syncGame = gameName
	ui.status = "Syncing Anki..."

	go func() {
		result, err := anki.New(anki.DefaultAnkiConnectURL).SyncFlashcardsToAnki(gameName, anki.DefaultAnkiConnectURL, true)
		ui.syncResult <- flashcardSyncResult{gameName: gameName, result: result, err: err}
	}()
}

func (ui *FlashcardsUI) drainSyncResults() {
	for {
		select {
		case result := <-ui.syncResult:
			if result.gameName != ui.syncGame {
				continue
			}
			ui.syncing = false
			ui.syncGame = ""
			if result.err != nil {
				ui.status = result.err.Error()
			} else {
				ui.status = fmt.Sprintf("Synced %s: %d created, %d updated.", result.result.DeckName, result.result.Created, result.result.Updated)
			}
			ui.reload(result.gameName)
		default:
			return
		}
	}
}

func (ui *FlashcardsUI) reload(gameName string) {
	gameName = strings.TrimSpace(gameName)
	cards, err := ui.loadCards(gameName)
	if err != nil {
		ui.status = err.Error()
		return
	}
	ui.cards = cards
	ui.pruneClicks()
	if gameName == "" {
		ui.status = "Select a game to review saved flashcards."
		return
	}
	if ui.status == "" || strings.Contains(ui.status, "flashcards for ") || strings.HasPrefix(ui.status, "Select a game") {
		ui.status = fmt.Sprintf("%d flashcards for %s", len(cards), gameName)
	}
}

func (ui *FlashcardsUI) pruneClicks() {
	valid := make(map[string]struct{}, len(ui.cards))
	for _, card := range ui.cards {
		valid[card.ID] = struct{}{}
	}
	for id := range ui.editButtons {
		if _, ok := valid[id]; !ok {
			delete(ui.editButtons, id)
		}
	}
	for id := range ui.deleteButtons {
		if _, ok := valid[id]; !ok {
			delete(ui.deleteButtons, id)
		}
	}
}

func (ui *FlashcardsUI) editButton(id string) *components.IconButton {
	if ui.editButtons[id] == nil {
		btn := components.NewIconButton("Edit", &widget.Clickable{}, mustFlashcardIcon("lucide:pencil")).WithThemeClient(ui.theme)
		btn.FillWidth = false
		btn.MinWidth = unit.Dp(86)
		btn.Height = unit.Dp(34)
		btn.Radius = unit.Dp(8)
		ui.editButtons[id] = btn
	}
	return ui.editButtons[id]
}

func (ui *FlashcardsUI) deleteButton(id string) *components.IconButton {
	if ui.deleteButtons[id] == nil {
		btn := components.NewIconButton("Delete", &widget.Clickable{}, mustFlashcardIcon("lucide:trash-2")).WithThemeClient(ui.theme)
		btn.FillWidth = false
		btn.MinWidth = unit.Dp(104)
		btn.Height = unit.Dp(34)
		btn.Radius = unit.Dp(8)
		ui.deleteButtons[id] = btn
	}
	return ui.deleteButtons[id]
}

func (ui *FlashcardsUI) layoutCJKLabel(gtx layout.Context, value string, role theme.TextRole, colorRole theme.TextColorRole) layout.Dimensions {
	lbl := material.Body1(ui.th, value)
	theme.ApplyTypography(&lbl, ui.theme.GetCurrentTypography(), role)
	lbl.Color = theme.SelectTextColor(ui.theme.GetCurrentColorToken(), colorRole)
	lbl.Alignment = text.Start
	lbl.Font.Typeface = font.Typeface("Noto Sans CJK JP")
	return lbl.Layout(gtx)
}

func (ui *FlashcardsUI) activeGameName() string {
	if ui == nil || ui.backend == nil {
		return ""
	}
	g := ui.backend.CurrentGame()
	if g == nil {
		return ""
	}
	return strings.TrimSpace(g.Name)
}

func (ui *FlashcardsUI) loadCards(gameName string) ([]flashcards.Flashcard, error) {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		return nil, nil
	}
	return flashcards.LoadFlashcards(gameName)
}

func mustFlashcardIcon(name string) *iconify.SVGIcon {
	ic, err := iconify.DefaultIconify.Icon(context.Background(), name)
	if err != nil {
		return nil
	}
	return ic
}
