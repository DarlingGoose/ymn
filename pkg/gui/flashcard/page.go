package flashcard

import (
	"context"
	"fmt"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareui "github.com/Seann-Moser/bare/pkg/ui"
	"github.com/Seann-Moser/bare/pkg/ui/icons"
	barethemes "github.com/Seann-Moser/bare/pkg/ui/themes"
	bareutils "github.com/Seann-Moser/bare/pkg/ui/utils"
	"github.com/Seann-Moser/wgl/pkg/anki"
	flashcards "github.com/Seann-Moser/wgl/pkg/flashcard"
	"github.com/Seann-Moser/wgl/pkg/gui"
	"github.com/Seann-Moser/wgl/pkg/util"
)

var _ gui.EvenHandler = &Page{}

const (
	flashcardTabDeck   = "deck"
	flashcardTabInfo   = "info"
	flashcardTabEditor = "editor"
)

type Page struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	flashcardList widget.List
	pageTabs      *bareui.Tabs

	searchEditor  widget.Editor
	wordEditor    widget.Editor
	meaningEditor widget.Editor

	saveButton   widget.Clickable
	deleteButton widget.Clickable
	newButton    widget.Clickable
	reloadButton widget.Clickable
	syncButton   widget.Clickable

	selectedFlashcardID string
	cards               []flashcards.Flashcard
	selectClicks        map[string]*widget.Clickable
	deleteClicks        map[string]*widget.Clickable

	activeGameName string
	ankiURL        string
	pushSync       bool
	statusText     string

	OnError func(title, body string)
}

func New(theme barethemes.Theme) *Page {
	p := &Page{
		theme:      theme,
		pushSync:   true,
		statusText: "Create a new card or pick one from the list to edit.",
		pageTabs: bareui.NewTabs([]bareui.TabItem{
			{ID: flashcardTabDeck, Label: "Deck", Icon: "mdi:view-list-outline"},
			{ID: flashcardTabInfo, Label: "Info", Icon: "mdi:information-outline"},
			{ID: flashcardTabEditor, Label: "Editor", Icon: "mdi:pencil-box-outline"},
		}, flashcardTabDeck),
		selectClicks: make(map[string]*widget.Clickable),
		deleteClicks: make(map[string]*widget.Clickable),
	}
	p.flashcardList.Axis = layout.Vertical
	p.searchEditor.SingleLine = true
	p.wordEditor.SingleLine = true
	p.meaningEditor.SingleLine = false
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

func (p *Page) SetContext(activeGameName, ankiURL string) *Page {
	p.activeGameName = strings.TrimSpace(activeGameName)
	p.ankiURL = strings.TrimSpace(ankiURL)
	return p
}

func (p *Page) SetPushSync(pushSync bool) *Page {
	p.pushSync = pushSync
	return p
}

func (p *Page) SetCards(cards []flashcards.Flashcard) *Page {
	p.cards = append([]flashcards.Flashcard(nil), cards...)
	p.syncRowState()
	if p.selectedFlashcardID == "" {
		return p
	}
	for _, card := range p.cards {
		if card.ID == p.selectedFlashcardID {
			return p
		}
	}
	p.prepareNewFlashcard()
	return p
}

func (p *Page) Cards() []flashcards.Flashcard {
	return append([]flashcards.Flashcard(nil), p.cards...)
}

func (p *Page) HandleEvents(gtx layout.Context, _ context.Context, _ *app.Window) {
	for p.saveButton.Clicked(gtx) {
		p.saveCurrentCard()
	}
	for p.deleteButton.Clicked(gtx) {
		p.deleteSelectedCard()
	}
	for p.newButton.Clicked(gtx) {
		p.prepareNewFlashcard()
	}
	for p.reloadButton.Clicked(gtx) {
		if err := p.Reload(); err != nil {
			p.showError("Reload Flashcards Failed", err.Error())
		}
	}
	for p.syncButton.Clicked(gtx) {
		if err := p.SyncToAnki(); err != nil {
			p.showError("Anki Sync Failed", err.Error())
		}
	}
	for cardID, click := range p.selectClicks {
		for click.Clicked(gtx) {
			p.selectCard(cardID)
		}
	}
	for cardID, click := range p.deleteClicks {
		for click.Clicked(gtx) {
			p.deleteCardByID(cardID)
		}
	}
}

func (p *Page) LayoutPage(gtx layout.Context) layout.Dimensions {
	if p.iconify == nil {
		p.iconify = icons.NewIconify()
	}
	p.syncRowState()

	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(p.theme.Gio(), "Flashcards")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), "Review saved cards for the active game, update meanings, and sync the current deck to Anki.")
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					p.pageTabs.Axis = layout.Horizontal
					return p.pageTabs.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					switch p.pageTabs.Selected() {
					case flashcardTabInfo:
						return p.layoutInfoPanel(gtx)
					case flashcardTabEditor:
						return p.layoutEditorPanel(gtx)
					default:
						return p.layoutFlashcardList(gtx)
					}
				}),
			)
		})
	})
}

func (p *Page) Reload() error {
	if strings.TrimSpace(p.activeGameName) == "" {
		p.cards = nil
		p.prepareNewFlashcard()
		return nil
	}
	cards, err := flashcards.LoadFlashcards(p.activeGameName)
	if err != nil {
		return err
	}
	p.SetCards(cards)
	return nil
}

func (p *Page) SyncToAnki() error {
	if strings.TrimSpace(p.activeGameName) == "" {
		return fmt.Errorf("select a game before syncing Anki")
	}
	client := anki.New(p.ankiURL)
	if _, err := client.SyncFlashcardsToAnki(p.activeGameName, p.ankiURL, p.pushSync); err != nil {
		return err
	}
	return p.Reload()
}

func (p *Page) layoutFlashcardList(gtx layout.Context) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			filtered := p.filteredCards()
			search := material.Editor(p.theme.Gio(), &p.searchEditor, "Search flashcards")
			search.Color = p.theme.Color.Text
			search.HintColor = p.theme.Color.TextMuted

			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutInfoRow(gtx, "Library", fmt.Sprintf("%d cards", len(filtered)), search.Layout)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if len(filtered) == 0 {
						return p.layoutEmptyState(gtx, "No matching flashcards for this game.")
					}
					return material.List(p.theme.Gio(), &p.flashcardList).Layout(gtx, len(filtered), func(gtx layout.Context, index int) layout.Dimensions {
						card := filtered[index]
						return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return p.layoutFlashcardRow(gtx, card)
						})
					})
				}),
			)
		})
	})
}

func (p *Page) layoutFlashcardRow(gtx layout.Context, card flashcards.Flashcard) layout.Dimensions {
	selectButton := bareui.Button{
		Clickable: p.selectClickable(card.ID),
		Text:      "Edit",
		Prefix:    "mdi:pencil-outline",
		Variant:   p.rowButtonVariant(card.ID),
	}
	deleteButton := bareui.Button{
		Clickable: p.deleteClickable(card.ID),
		Text:      "Delete",
		Prefix:    "mdi:delete-outline",
		Variant:   bareui.ButtonGhost,
	}

	surfaceColor := p.theme.Color.Surface
	if card.ID == p.selectedFlashcardID {
		surfaceColor = p.theme.Color.SurfaceAlt
	}

	return bareutils.Panel(gtx, surfaceColor, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), card.Text)
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), card.Meaning)
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					meta := p.cardMetaText(card)
					if meta == "" {
						return layout.Dimensions{}
					}
					return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body1(p.theme.Gio(), meta)
						lbl.Color = p.theme.Color.TextMuted
						return lbl.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					source := strings.TrimSpace(card.SourceLine)
					if source == "" {
						return layout.Dimensions{}
					}
					return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body2(p.theme.Gio(), source)
						lbl.Color = p.theme.Color.TextMuted
						return lbl.Layout(gtx)
					})
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return selectButton.Layout(gtx, p.theme, p.iconify)
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return deleteButton.Layout(gtx, p.theme, p.iconify)
						}),
					)
				}),
			)
		})
	})
}

func (p *Page) layoutEditorPanel(gtx layout.Context) layout.Dimensions {
	word := material.Editor(p.theme.Gio(), &p.wordEditor, "Word or phrase")
	word.Color = p.theme.Color.Text
	word.HintColor = p.theme.Color.TextMuted

	meaning := material.Editor(p.theme.Gio(), &p.meaningEditor, "Meaning")
	meaning.Color = p.theme.Color.Text
	meaning.HintColor = p.theme.Color.TextMuted

	newButton := bareui.Button{
		Clickable: &p.newButton,
		Text:      "New Flashcard",
		Prefix:    "mdi:plus-box-outline",
		Variant:   bareui.ButtonSecondary,
	}
	saveButton := bareui.Button{
		Clickable: &p.saveButton,
		Text:      p.saveButtonLabel(),
		Prefix:    "mdi:content-save-outline",
		Variant:   bareui.ButtonPrimary,
	}
	deleteButton := bareui.Button{
		Clickable: &p.deleteButton,
		Text:      "Delete",
		Prefix:    "mdi:delete-outline",
		Variant:   bareui.ButtonGhost,
	}
	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), "Editor")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), p.editorStatus())
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(word.Layout),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(220))
					return meaning.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return newButton.Layout(gtx, p.theme, p.iconify)
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return deleteButton.Layout(gtx, p.theme, p.iconify)
						}),
					)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return saveButton.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), "Use the Info tab for deck status, reload, and Anki sync.")
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (p *Page) layoutInfoPanel(gtx layout.Context) layout.Dimensions {
	reloadButton := bareui.Button{
		Clickable: &p.reloadButton,
		Text:      "Reload Cards",
		Prefix:    "mdi:refresh",
		Variant:   bareui.ButtonSecondary,
	}
	syncButton := bareui.Button{
		Clickable: &p.syncButton,
		Text:      "Sync Anki",
		Prefix:    "mdi:cloud-upload-outline",
		Variant:   bareui.ButtonPrimary,
	}
	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), "Deck Info")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), p.editorStatus())
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutInfoRow(gtx, "Deck", util.FirstNonEmpty(util.AnkiDeckName(p.activeGameName), "No deck selected"), nil)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutInfoRow(gtx, "Cards", fmt.Sprintf("%d saved", len(p.cards)), nil)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.layoutInfoRow(gtx, "AnkiConnect URL", util.FirstNonEmpty(p.ankiURL, "Not configured"), nil)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return reloadButton.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(10))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return syncButton.Layout(gtx, p.theme, p.iconify)
				}),
			)
		})
	})
}

func (p *Page) layoutInfoRow(gtx layout.Context, label, current string, control layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(p.theme.Gio(), label)
							lbl.Color = p.theme.Color.TextMuted
							return lbl.Layout(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(p.theme.Gio(), current)
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
					)
				}),
			}
			if control != nil {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return control(gtx)
				}))
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
		})
	})
}

func (p *Page) layoutEmptyState(gtx layout.Context, message string) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(p.theme.Gio(), message)
			lbl.Color = p.theme.Color.TextMuted
			return lbl.Layout(gtx)
		})
	})
}

func (p *Page) syncRowState() {
	valid := make(map[string]struct{}, len(p.cards))
	for _, card := range p.cards {
		valid[card.ID] = struct{}{}
		if p.selectClicks[card.ID] == nil {
			p.selectClicks[card.ID] = new(widget.Clickable)
		}
		if p.deleteClicks[card.ID] == nil {
			p.deleteClicks[card.ID] = new(widget.Clickable)
		}
	}
	for id := range p.selectClicks {
		if _, ok := valid[id]; !ok {
			delete(p.selectClicks, id)
		}
	}
	for id := range p.deleteClicks {
		if _, ok := valid[id]; !ok {
			delete(p.deleteClicks, id)
		}
	}
}

func (p *Page) filteredCards() []flashcards.Flashcard {
	query := strings.TrimSpace(strings.ToLower(p.searchEditor.Text()))
	if query == "" {
		return p.cards
	}
	filtered := make([]flashcards.Flashcard, 0, len(p.cards))
	for _, card := range p.cards {
		haystack := strings.ToLower(strings.Join([]string{
			card.Text,
			card.Meaning,
			card.Reading,
			card.SourceLine,
		}, "\n"))
		if strings.Contains(haystack, query) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

func (p *Page) selectClickable(cardID string) *widget.Clickable {
	if p.selectClicks[cardID] == nil {
		p.selectClicks[cardID] = new(widget.Clickable)
	}
	return p.selectClicks[cardID]
}

func (p *Page) deleteClickable(cardID string) *widget.Clickable {
	if p.deleteClicks[cardID] == nil {
		p.deleteClicks[cardID] = new(widget.Clickable)
	}
	return p.deleteClicks[cardID]
}

func (p *Page) rowButtonVariant(cardID string) bareui.ButtonVariant {
	if cardID == p.selectedFlashcardID {
		return bareui.ButtonPrimary
	}
	return bareui.ButtonSecondary
}

func (p *Page) saveButtonLabel() string {
	if strings.TrimSpace(p.selectedFlashcardID) != "" {
		return "Save Changes"
	}
	return "Create Flashcard"
}

func (p *Page) editorStatus() string {
	if strings.TrimSpace(p.statusText) != "" {
		return p.statusText
	}
	return "Create a new card or pick one from the list to edit."
}

func (p *Page) prepareNewFlashcard() {
	p.selectedFlashcardID = ""
	p.wordEditor.SetText("")
	p.meaningEditor.SetText("")
	p.statusText = "Create a new card or pick one from the list to edit."
	p.pageTabs.Active = flashcardTabEditor
}

func (p *Page) selectCard(cardID string) {
	for _, card := range p.cards {
		if card.ID == cardID {
			p.selectedFlashcardID = card.ID
			p.wordEditor.SetText(card.Text)
			p.meaningEditor.SetText(card.Meaning)
			p.statusText = "Editing selected flashcard."
			p.pageTabs.Active = flashcardTabEditor
			return
		}
	}
}

func (p *Page) saveCurrentCard() {
	if strings.TrimSpace(p.activeGameName) == "" {
		p.showError("Save Flashcard Failed", "Select a game before editing flashcards.")
		return
	}

	word := strings.TrimSpace(p.wordEditor.Text())
	meaning := strings.TrimSpace(p.meaningEditor.Text())
	if word == "" {
		p.showError("Save Flashcard Failed", "Flashcard word cannot be empty.")
		return
	}
	if meaning == "" {
		p.showError("Save Flashcard Failed", "Flashcard meaning cannot be empty.")
		return
	}

	if strings.TrimSpace(p.selectedFlashcardID) == "" {
		card := flashcards.Flashcard{
			GameName:  p.activeGameName,
			Text:      word,
			Meaning:   meaning,
			AnkiDeck:  util.AnkiDeckName(p.activeGameName),
			AnkiModel: flashcards.DefaultAnkiModel,
		}
		if err := flashcards.AddFlashcard(card); err != nil {
			p.showError("Save Flashcard Failed", err.Error())
			return
		}
		p.prepareNewFlashcard()
		if err := p.Reload(); err != nil {
			p.showError("Reload Flashcards Failed", err.Error())
		}
		return
	}

	for _, card := range p.cards {
		if card.ID != p.selectedFlashcardID {
			continue
		}
		card.Text = word
		card.Meaning = meaning
		if err := flashcards.UpdateFlashcard(card); err != nil {
			p.showError("Save Flashcard Failed", err.Error())
			return
		}
		p.statusText = "Editing selected flashcard."
		if err := p.Reload(); err != nil {
			p.showError("Reload Flashcards Failed", err.Error())
		}
		return
	}

	p.showError("Save Flashcard Failed", "Selected flashcard could not be found.")
}

func (p *Page) deleteSelectedCard() {
	if strings.TrimSpace(p.selectedFlashcardID) == "" {
		p.showError("Delete Flashcard Failed", "Select a flashcard before deleting it.")
		return
	}
	p.deleteCardByID(p.selectedFlashcardID)
}

func (p *Page) deleteCardByID(cardID string) {
	client := anki.New(p.ankiURL)
	for _, card := range p.cards {
		if card.ID != cardID {
			continue
		}
		if err := client.DeleteFlashcardFromAnki(card, p.ankiURL, p.pushSync); err != nil {
			p.showError("Delete Flashcard Failed", err.Error())
			return
		}
		if err := flashcards.DeleteFlashcard(card.GameName, card.ID); err != nil {
			p.showError("Delete Flashcard Failed", err.Error())
			return
		}
		if p.selectedFlashcardID == cardID {
			p.prepareNewFlashcard()
		}
		if err := p.Reload(); err != nil {
			p.showError("Reload Flashcards Failed", err.Error())
		}
		return
	}
	p.showError("Delete Flashcard Failed", "Selected flashcard could not be found.")
}

func (p *Page) cardMetaText(card flashcards.Flashcard) string {
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

func (p *Page) showError(title, body string) {
	if p.OnError != nil {
		p.OnError(title, body)
	}
}
