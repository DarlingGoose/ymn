package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var runGameName string
var runGameSaveLauncher bool
var runGameLauncherDir string

var runGameCmd = &cobra.Command{
	Use:     "run-game [game-name]",
	Aliases: []string{"runGame"},
	Short:   "Launch a previously added game",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		selectedName := strings.TrimSpace(runGameName)
		if selectedName == "" && len(args) > 0 {
			selectedName = args[0]
		}

		var cfg GameConfig
		var err error
		if selectedName == "" {
			cfg, err = selectGameConfigWithTUI("Select a game to launch", "launch")
			if err != nil {
				return err
			}
		} else {
			cfg, err = findGameConfig(selectedName)
			if err != nil {
				return err
			}
		}

		if runGameSaveLauncher {
			desktopPath, err := saveDesktopEntry(cfg, runGameLauncherDir)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved launcher: %s\n", desktopPath)
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "launching %s with %s\n", cfg.Name, cfg.Runner)
		return launchGame(cfg)
	},
}

func init() {
	rootCmd.AddCommand(runGameCmd)
	runGameCmd.Flags().StringVarP(&runGameName, "game", "g", "", "name of the saved game to launch")
	runGameCmd.Flags().BoolVar(&runGameSaveLauncher, "save-launcher", false, "write a .desktop launcher so rofi and Linux app launchers can find the game")
	runGameCmd.Flags().StringVar(&runGameLauncherDir, "launcher-dir", "", "directory where the .desktop launcher is written (default: ~/.local/share/applications)")
}

type gamePickerModel struct {
	games    []GameConfig
	cursor   int
	selected *GameConfig
	quitting bool
	title    string
	action   string
}

func selectGameConfigWithTUI(title, action string) (GameConfig, error) {
	configs, err := listGameConfigs()
	if err != nil {
		return GameConfig{}, err
	}
	if len(configs) == 0 {
		return GameConfig{}, fmt.Errorf("no saved games found in %s", configBaseDir())
	}

	program := tea.NewProgram(gamePickerModel{
		games:  configs,
		title:  title,
		action: action,
	}, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return GameConfig{}, err
	}

	model, ok := finalModel.(gamePickerModel)
	if !ok {
		return GameConfig{}, fmt.Errorf("unexpected picker state %T", finalModel)
	}
	if model.selected == nil {
		return GameConfig{}, fmt.Errorf("game selection cancelled")
	}
	return *model.selected, nil
}

func (m gamePickerModel) Init() tea.Cmd {
	return nil
}

func (m gamePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m gamePickerModel) View() string {
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
		if game.RequiresSteam || game.Runner == RunnerSteam {
			launchMode = "steam"
		}
		builder.WriteString(fmt.Sprintf("%s%s [%s/%s]\n", cursor, game.Name, game.Runner, launchMode))
	}
	builder.WriteString(fmt.Sprintf("\nUse ↑/↓ or j/k, press Enter to %s, q to cancel.\n", m.action))
	return builder.String()
}
