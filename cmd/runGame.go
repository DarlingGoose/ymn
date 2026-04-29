package cmd

import (
	"fmt"
	"strings"

	"github.com/Seann-Moser/wgl/pkg/game/gameconfig"
	"github.com/Seann-Moser/wgl/pkg/game/launcher"
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

		var cfg *gameconfig.GameConfig
		var err error
		if selectedName == "" {
			picker, err := launcher.NewPicker("Select a game to launch", "launch")
			if err != nil {
				return err
			}
			cfg, err = picker.SelectGameConfig()
			if err != nil {
				return err
			}
		} else {
			cfg, err = gameconfig.FindConfig(selectedName)
			if err != nil {
				return err
			}
		}

		if runGameSaveLauncher {
			desktopPath, err := launcher.SaveDesktopEntry(cfg, runGameLauncherDir)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "saved launcher: %s\n", desktopPath)
			return nil
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "launching %s with %s\n", cfg.Name, cfg.Runner)
		return cfg.Launch()
	},
}

func init() {
	rootCmd.AddCommand(runGameCmd)
	runGameCmd.Flags().StringVarP(&runGameName, "game", "g", "", "name of the saved game to launch")
	runGameCmd.Flags().BoolVar(&runGameSaveLauncher, "save-launcher", false, "write a .desktop launcher so rofi and Linux app launchers can find the game")
	runGameCmd.Flags().StringVar(&runGameLauncherDir, "launcher-dir", "", "directory where the .desktop launcher is written (default: ~/.local/share/applications)")
}
