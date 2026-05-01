package launcher

import (
	"fmt"
	"strings"

	"github.com/DarlingGoose/wgl/pkg/game/gameconfig"
	"github.com/DarlingGoose/wgl/pkg/util"
	tea "github.com/charmbracelet/bubbletea"
)

type PickerModel struct {
	games    []gameconfig.GameConfig
	cursor   int
	selected *gameconfig.GameConfig
	quitting bool
	title    string
	action   string
}

func NewPicker(title, action string) (*PickerModel, error) {
	configs, err := gameconfig.ListConfigs()
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		return nil, fmt.Errorf("no saved games found in %s", util.ConfigBaseDir())
	}

	picker := &PickerModel{
		games:  configs,
		title:  title,
		action: action,
	}
	return picker, nil
}
func (m PickerModel) Init() tea.Cmd {
	return nil
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.games)-1 {
				m.cursor++
			}
		case "enter":
			selected := m.games[m.cursor]
			m.selected = &selected
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m PickerModel) View() string {
	if m.quitting && m.selected == nil {
		return "Selection cancelled.\n"
	}

	var builder strings.Builder
	builder.WriteString(m.title)
	builder.WriteString("\n\n")
	for i, game := range m.games {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		launchMode := "direct"
		if game.RequiresSteam || game.Runner == gameconfig.RunnerSteam {
			launchMode = "steam"
		}
		builder.WriteString(fmt.Sprintf("%s%s [%s/%s]\n", cursor, game.Name, game.Runner, launchMode))
	}
	builder.WriteString(fmt.Sprintf("\nUse ↑/↓ or j/k, press Enter to %s, q to cancel.\n", m.action))
	return builder.String()
}

func (m PickerModel) SelectGameConfig() (*gameconfig.GameConfig, error) {
	program := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	model, ok := finalModel.(PickerModel)
	if !ok {
		return nil, fmt.Errorf("unexpected picker state %T", finalModel)
	}
	if model.selected == nil {
		return nil, fmt.Errorf("game selection cancelled")
	}
	return model.selected, nil
}
