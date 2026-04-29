package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var practiceGameName string

var practiceFlashcardsCmd = &cobra.Command{
	Use:   "practice [game-name]",
	Short: "Review saved flashcards for a game",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gameName, err := resolveSelectedGameName(practiceGameName, args, "Select a game deck to practice", "practice")
		if err != nil {
			return err
		}

		cards, err := loadFlashcards(gameName)
		if err != nil {
			return err
		}
		if len(cards) == 0 {
			return fmt.Errorf("no flashcards saved for %s", gameName)
		}

		_, err = tea.NewProgram(practiceModel{
			gameName: gameName,
			cards:    cards,
		}, tea.WithAltScreen()).Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(practiceFlashcardsCmd)
	practiceFlashcardsCmd.Flags().StringVarP(&practiceGameName, "game", "g", "", "name of the saved game deck to practice")
}

type practiceModel struct {
	gameName string
	cards    []Flashcard
	index    int
	showBack bool
	quitting bool
	width    int
	height   int
}

func (m practiceModel) Init() tea.Cmd {
	return nil
}

func (m practiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case " ", "enter":
			m.showBack = !m.showBack
		case "j", "down", "l", "right", "n":
			if m.index < len(m.cards)-1 {
				m.index++
			}
			m.showBack = false
		case "k", "up", "h", "left", "p":
			if m.index > 0 {
				m.index--
			}
			m.showBack = false
		}
	}
	return m, nil
}

func (m practiceModel) View() string {
	if m.quitting {
		return ""
	}
	card := m.cards[m.index]

	header := lipgloss.NewStyle().Bold(true).Padding(0, 1).
		Render(fmt.Sprintf("Practice %s (%d/%d)", m.gameName, m.index+1, len(m.cards)))
	faceTitle := "Front"
	faceBody := card.Text
	if m.showBack {
		faceTitle = "Meaning"
		faceBody = card.Meaning
	}

	width := m.width - 6
	if width < 30 {
		width = 30
	}
	height := m.height - 6
	if height < 8 {
		height = 8
	}

	body := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(width).
		Height(height).
		Render(strings.Join([]string{
			lipgloss.NewStyle().Bold(true).Render(faceTitle),
			"",
			faceBody,
			"",
			lipgloss.NewStyle().Faint(true).Render("Source"),
			truncateForWidth(card.SourceLine, width),
		}, "\n"))

	footer := lipgloss.NewStyle().Padding(0, 1).Faint(true).
		Render("Space/Enter reveal | n/p or arrows move | q quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}
