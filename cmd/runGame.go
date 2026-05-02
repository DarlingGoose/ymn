package cmd

import (
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
		//todo vntext wrapper + launch gui
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runGameCmd)
	runGameCmd.Flags().StringVarP(&runGameName, "game", "g", "", "name of the saved game to launch")
	runGameCmd.Flags().BoolVar(&runGameSaveLauncher, "save-launcher", false, "write a .desktop launcher so rofi and Linux app launchers can find the game")
	runGameCmd.Flags().StringVar(&runGameLauncherDir, "launcher-dir", "", "directory where the .desktop launcher is written (default: ~/.local/share/applications)")
}
